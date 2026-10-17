// Package controller reconciles ShortLink resources with the shortener
// service.
package controller

import (
	"context"
	"errors"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/Priyadharshan0903/gopath-academy/projects/shortlink-operator/api/v1alpha1"
	"github.com/Priyadharshan0903/gopath-academy/projects/shortlink-operator/internal/shortener"
)

// finalizer makes Kubernetes wait for us before a ShortLink disappears,
// so we can delete the link in the shortener first.
const finalizer = "shortener.example.com/delete-link"

// API is the part of the shortener client the reconciler uses; tests
// substitute a fake.
type API interface {
	Get(ctx context.Context, code string) (shortener.Link, error)
	Create(ctx context.Context, code, url string) (shortener.Link, error)
	Delete(ctx context.Context, code string) error
}

// ShortLinkReconciler makes the shortener match the ShortLink resources.
type ShortLinkReconciler struct {
	client.Client
	API API
}

//go:generate go tool controller-gen rbac:roleName=shortlink-operator paths=. output:rbac:dir=../../config/rbac

// The RBAC markers below generate config/rbac/role.yaml.
//
// +kubebuilder:rbac:groups=shortener.example.com,resources=shortlinks,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=shortener.example.com,resources=shortlinks/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=shortener.example.com,resources=shortlinks/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// [start reconcile]

// Reconcile is called whenever a ShortLink changes (and periodically).
// It must be idempotent: it looks at the current state of the world and
// moves it one step closer to the desired state, however many times it
// runs, and whatever happened before.
func (r *ShortLinkReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var sl v1alpha1.ShortLink
	if err := r.Get(ctx, req.NamespacedName, &sl); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err) // deleted: nothing to do
	}

	// 1. Being deleted: clean up in the shortener, then let it go.
	if !sl.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(&sl, finalizer) {
			return ctrl.Result{}, nil
		}
		if sl.Status.Code != "" {
			err := r.API.Delete(ctx, sl.Status.Code)
			if err != nil && !errors.Is(err, shortener.ErrNotFound) {
				return ctrl.Result{}, fmt.Errorf("delete link %q: %w", sl.Status.Code, err) // retried with backoff
			}
			logger.Info("deleted link", "code", sl.Status.Code)
		}
		controllerutil.RemoveFinalizer(&sl, finalizer)
		return ctrl.Result{}, r.Update(ctx, &sl)
	}

	// 2. Make sure we get a chance to clean up later.
	if controllerutil.AddFinalizer(&sl, finalizer) {
		if err := r.Update(ctx, &sl); err != nil {
			return ctrl.Result{}, err
		}
	}

	// 3. Make the link exist, with the right target.
	code := sl.Spec.Code
	if code == "" {
		code = sl.Name // deterministic: retries never create a second link
	}
	link, err := r.ensureLink(ctx, &sl, code)
	if err != nil {
		reason := "APIError"
		if errors.Is(err, shortener.ErrConflict) {
			reason = "CodeTaken"
		}
		r.setReady(&sl, metav1.ConditionFalse, reason, err.Error())
		if uerr := r.Status().Update(ctx, &sl); uerr != nil {
			return ctrl.Result{}, uerr
		}
		if reason == "CodeTaken" {
			// Retrying soon will not help; check again later.
			return ctrl.Result{RequeueAfter: 5 * time.Minute}, nil
		}
		return ctrl.Result{}, err
	}

	// 4. Report what we observed.
	sl.Status.Code = link.Code
	sl.Status.ShortURL = link.ShortURL
	r.setReady(&sl, metav1.ConditionTrue, "LinkReady", "the short link exists")
	return ctrl.Result{}, r.Status().Update(ctx, &sl)
}

// ensureLink creates the link for code, or confirms it already points
// at spec.url. A link this ShortLink created earlier (status.code ==
// code) is recreated when the URL changes; anyone else's link is left
// alone.
func (r *ShortLinkReconciler) ensureLink(ctx context.Context, sl *v1alpha1.ShortLink, code string) (shortener.Link, error) {
	// spec.code changed: remove the link we created under the old code.
	if old := sl.Status.Code; old != "" && old != code {
		if err := r.API.Delete(ctx, old); err != nil && !errors.Is(err, shortener.ErrNotFound) {
			return shortener.Link{}, err
		}
		sl.Status.Code = ""
	}
	existing, err := r.API.Get(ctx, code)
	switch {
	case err == nil && existing.URL == sl.Spec.URL:
		return existing, nil // already correct: nothing to do
	case err == nil && sl.Status.Code == code:
		if err := r.API.Delete(ctx, code); err != nil && !errors.Is(err, shortener.ErrNotFound) {
			return shortener.Link{}, err
		}
	case err == nil:
		return shortener.Link{}, fmt.Errorf("%w: %q already points to %s", shortener.ErrConflict, code, existing.URL)
	case !errors.Is(err, shortener.ErrNotFound):
		return shortener.Link{}, err
	}
	return r.API.Create(ctx, code, sl.Spec.URL)
}

// [end reconcile]

func (r *ShortLinkReconciler) setReady(sl *v1alpha1.ShortLink, status metav1.ConditionStatus, reason, msg string) {
	sl.Status.ObservedGeneration = sl.Generation
	meta.SetStatusCondition(&sl.Status.Conditions, metav1.Condition{
		Type:               "Ready",
		Status:             status,
		Reason:             reason,
		Message:            msg,
		ObservedGeneration: sl.Generation,
	})
}

// [start setup]

// SetupWithManager registers the reconciler: watch ShortLinks, and call
// Reconcile with the name of each one that changes.
func (r *ShortLinkReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.ShortLink{}).
		Named("shortlink").
		Complete(r)
}

// [end setup]
