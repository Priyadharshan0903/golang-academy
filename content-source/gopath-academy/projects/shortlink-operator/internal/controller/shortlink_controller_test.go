package controller

import (
	"context"
	"errors"
	"sync"
	"testing"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/Priyadharshan0903/gopath-academy/projects/shortlink-operator/api/v1alpha1"
	"github.com/Priyadharshan0903/gopath-academy/projects/shortlink-operator/internal/shortener"
)

// fakeAPI is an in-memory shortener that records the calls it gets.
type fakeAPI struct {
	mu    sync.Mutex
	links map[string]string // code -> url
	calls []string
	down  bool
}

func newFakeAPI() *fakeAPI { return &fakeAPI{links: map[string]string{}} }

func (f *fakeAPI) record(call string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, call)
	if f.down {
		return errors.New("connection refused")
	}
	return nil
}

func (f *fakeAPI) Get(_ context.Context, code string) (shortener.Link, error) {
	if err := f.record("get " + code); err != nil {
		return shortener.Link{}, err
	}
	u, ok := f.links[code]
	if !ok {
		return shortener.Link{}, shortener.ErrNotFound
	}
	return shortener.Link{Code: code, URL: u, ShortURL: "https://sho.rt/" + code}, nil
}

func (f *fakeAPI) Create(_ context.Context, code, url string) (shortener.Link, error) {
	if err := f.record("create " + code); err != nil {
		return shortener.Link{}, err
	}
	if _, ok := f.links[code]; ok {
		return shortener.Link{}, shortener.ErrConflict
	}
	f.links[code] = url
	return shortener.Link{Code: code, URL: url, ShortURL: "https://sho.rt/" + code}, nil
}

func (f *fakeAPI) Delete(_ context.Context, code string) error {
	if err := f.record("delete " + code); err != nil {
		return err
	}
	if _, ok := f.links[code]; !ok {
		return shortener.ErrNotFound
	}
	delete(f.links, code)
	return nil
}

func (f *fakeAPI) takeCalls() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	c := f.calls
	f.calls = nil
	return c
}

// [start setup]

func setup(t *testing.T, objs ...client.Object) (*ShortLinkReconciler, client.Client, *fakeAPI) {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	// The fake client keeps objects in memory and behaves like the API
	// server for gets, updates, finalizers and the status subresource.
	c := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(objs...).
		WithStatusSubresource(&v1alpha1.ShortLink{}).
		Build()
	api := newFakeAPI()
	return &ShortLinkReconciler{Client: c, API: api}, c, api
}

func reconcile(t *testing.T, r *ShortLinkReconciler, name string) (ctrl.Result, error) {
	t.Helper()
	return r.Reconcile(context.Background(), ctrl.Request{NamespacedName: types.NamespacedName{Namespace: "default", Name: name}})
}

// [end setup]

func get(t *testing.T, c client.Client, name string) v1alpha1.ShortLink {
	t.Helper()
	var sl v1alpha1.ShortLink
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: "default", Name: name}, &sl); err != nil {
		t.Fatal(err)
	}
	return sl
}

func shortLink(name, url, code string) *v1alpha1.ShortLink {
	return &v1alpha1.ShortLink{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "default", Generation: 1},
		Spec:       v1alpha1.ShortLinkSpec{URL: url, Code: code},
	}
}

// [start test]

func TestCreatesLinkAndIsIdempotent(t *testing.T) {
	r, c, api := setup(t, shortLink("go-docs", "https://go.dev/doc", ""))

	if _, err := reconcile(t, r, "go-docs"); err != nil {
		t.Fatal(err)
	}
	sl := get(t, c, "go-docs")
	ready := meta.FindStatusCondition(sl.Status.Conditions, "Ready")
	if sl.Status.Code != "go-docs" || sl.Status.ShortURL != "https://sho.rt/go-docs" || ready == nil || ready.Status != metav1.ConditionTrue {
		t.Fatalf("status = %+v", sl.Status)
	}
	if len(sl.Finalizers) != 1 {
		t.Fatalf("finalizers = %v", sl.Finalizers)
	}
	t.Logf("first reconcile: %v", api.takeCalls())

	// Reconciling again must not create anything.
	if _, err := reconcile(t, r, "go-docs"); err != nil {
		t.Fatal(err)
	}
	if calls := api.takeCalls(); len(calls) != 1 || calls[0] != "get go-docs" {
		t.Fatalf("second reconcile made calls %v, want only a get", calls)
	}
}

// [end test]

func TestURLChangeRecreatesOwnLink(t *testing.T) {
	r, c, api := setup(t, shortLink("docs", "https://go.dev/doc", "docs"))
	reconcile(t, r, "docs")
	api.takeCalls()

	sl := get(t, c, "docs")
	sl.Spec.URL = "https://pkg.go.dev"
	if err := c.Update(context.Background(), &sl); err != nil {
		t.Fatal(err)
	}
	if _, err := reconcile(t, r, "docs"); err != nil {
		t.Fatal(err)
	}
	if got := api.links["docs"]; got != "https://pkg.go.dev" {
		t.Fatalf("link points to %q", got)
	}
	t.Logf("after URL change: %v", api.takeCalls())
}

func TestCodeChangeDeletesOldLink(t *testing.T) {
	r, c, api := setup(t, shortLink("docs", "https://go.dev/doc", "old-code"))
	reconcile(t, r, "docs")
	sl := get(t, c, "docs")
	sl.Spec.Code = "new-code"
	c.Update(context.Background(), &sl)
	if _, err := reconcile(t, r, "docs"); err != nil {
		t.Fatal(err)
	}
	if _, ok := api.links["old-code"]; ok || api.links["new-code"] != "https://go.dev/doc" {
		t.Fatalf("links = %v", api.links)
	}
}

func TestCodeTakenBySomeoneElse(t *testing.T) {
	r, c, api := setup(t, shortLink("docs", "https://go.dev/doc", "popular"))
	api.links["popular"] = "https://example.com" // created by a user, not by us

	res, err := reconcile(t, r, "docs")
	if err != nil || res.RequeueAfter == 0 {
		t.Fatalf("res=%v err=%v: want a delayed retry, not an error", res, err)
	}
	ready := meta.FindStatusCondition(get(t, c, "docs").Status.Conditions, "Ready")
	if ready == nil || ready.Status != metav1.ConditionFalse || ready.Reason != "CodeTaken" {
		t.Fatalf("Ready = %+v", ready)
	}
	if api.links["popular"] != "https://example.com" {
		t.Fatal("someone else's link was changed")
	}
}

func TestAPIDownReturnsError(t *testing.T) {
	r, c, api := setup(t, shortLink("docs", "https://go.dev/doc", ""))
	api.down = true
	if _, err := reconcile(t, r, "docs"); err == nil {
		t.Fatal("want an error, so the controller retries with backoff")
	}
	ready := meta.FindStatusCondition(get(t, c, "docs").Status.Conditions, "Ready")
	if ready == nil || ready.Reason != "APIError" {
		t.Fatalf("Ready = %+v", ready)
	}
}

// [start deletetest]

func TestDeletionCleansUp(t *testing.T) {
	r, c, api := setup(t, shortLink("docs", "https://go.dev/doc", ""))
	reconcile(t, r, "docs")

	sl := get(t, c, "docs")
	if err := c.Delete(context.Background(), &sl); err != nil {
		t.Fatal(err)
	}
	// The finalizer keeps the object around, marked for deletion.
	if sl := get(t, c, "docs"); sl.DeletionTimestamp.IsZero() {
		t.Fatal("expected a deletion timestamp")
	}
	if _, err := reconcile(t, r, "docs"); err != nil {
		t.Fatal(err)
	}
	if _, ok := api.links["docs"]; ok {
		t.Fatal("link still exists in the shortener")
	}
	var gone v1alpha1.ShortLink
	err := c.Get(context.Background(), types.NamespacedName{Namespace: "default", Name: "docs"}, &gone)
	if err == nil {
		t.Fatal("object still exists after its finalizer was removed")
	}
}

// [end deletetest]
