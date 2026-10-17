// Package kube holds the client-go examples for lesson 15.3.
package kube

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/retry"
)

// [start connect]

// NewClientset connects to the cluster the program runs in, or, outside
// a cluster, to the one selected by the kubeconfig ($KUBECONFIG or
// ~/.kube/config), exactly like kubectl.
func NewClientset() (kubernetes.Interface, error) {
	cfg, err := rest.InClusterConfig() // the pod's service account
	if err != nil {
		rules := clientcmd.NewDefaultClientConfigLoadingRules()
		cfg, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, nil).ClientConfig()
		if err != nil {
			return nil, fmt.Errorf("no in-cluster config and no kubeconfig: %w", err)
		}
	}
	cfg.UserAgent = "gopath-academy-example"
	return kubernetes.NewForConfig(cfg)
}

// [end connect]

// [start list]

// PodSummary is one line of output, like `kubectl get pods`.
type PodSummary struct {
	Name     string
	Phase    corev1.PodPhase
	Ready    string // "1/1"
	Restarts int32
	Age      time.Duration
}

// Pods lists the pods in namespace that match a label selector such as
// "app.kubernetes.io/name=shortener" (empty means all).
func Pods(ctx context.Context, cs kubernetes.Interface, namespace, selector string, now time.Time) ([]PodSummary, error) {
	list, err := cs.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return nil, err
	}
	var out []PodSummary
	for _, p := range list.Items {
		ready, restarts := 0, int32(0)
		for _, s := range p.Status.ContainerStatuses {
			if s.Ready {
				ready++
			}
			restarts += s.RestartCount
		}
		out = append(out, PodSummary{
			Name:     p.Name,
			Phase:    p.Status.Phase,
			Ready:    fmt.Sprintf("%d/%d", ready, len(p.Spec.Containers)),
			Restarts: restarts,
			Age:      now.Sub(p.CreationTimestamp.Time).Round(time.Second),
		})
	}
	slices.SortFunc(out, func(a, b PodSummary) int { return strings.Compare(a.Name, b.Name) })
	return out, nil
}

// [end list]

// [start update]

// SetConfigValue sets one key of a ConfigMap, creating the ConfigMap if
// needed. Updates use optimistic concurrency: if someone else changed
// the object since we read it, the API server answers 409 Conflict, and
// RetryOnConflict reads it again and reapplies the change.
func SetConfigValue(ctx context.Context, cs kubernetes.Interface, namespace, name, key, value string) error {
	cms := cs.CoreV1().ConfigMaps(namespace)
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		cm, err := cms.Get(ctx, name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			cm = &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
				Data:       map[string]string{key: value},
			}
			_, err = cms.Create(ctx, cm, metav1.CreateOptions{})
			return err
		}
		if err != nil {
			return err
		}
		if cm.Data == nil {
			cm.Data = map[string]string{}
		}
		cm.Data[key] = value
		_, err = cms.Update(ctx, cm, metav1.UpdateOptions{}) // carries the resourceVersion we read
		return err
	})
}

// [end update]
