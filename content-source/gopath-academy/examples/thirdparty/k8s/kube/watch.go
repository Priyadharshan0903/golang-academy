package kube

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

// [start informer]

// PodEvent is something that happened to a pod.
type PodEvent struct {
	Type  string // "added", "updated" or "deleted"
	Name  string
	Phase corev1.PodPhase
}

// WatchPods reports pod events in namespace until ctx is done. It uses
// a shared informer: one LIST, then a WATCH stream, with a local cache
// that is kept in sync and automatically re-listed if the watch breaks.
func WatchPods(ctx context.Context, cs kubernetes.Interface, namespace string, events chan<- PodEvent) error {
	factory := informers.NewSharedInformerFactoryWithOptions(cs, 10*time.Minute,
		informers.WithNamespace(namespace))
	pods := factory.Core().V1().Pods().Informer()

	send := func(typ string, obj any) {
		if p, ok := obj.(*corev1.Pod); ok {
			select {
			case events <- PodEvent{typ, p.Name, p.Status.Phase}:
			case <-ctx.Done():
			}
		}
	}
	_, err := pods.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj any) { send("added", obj) },
		UpdateFunc: func(_, obj any) { send("updated", obj) },
		DeleteFunc: func(obj any) {
			if d, ok := obj.(cache.DeletedFinalStateUnknown); ok {
				obj = d.Obj // we missed the delete event; this is the last known state
			}
			send("deleted", obj)
		},
	})
	if err != nil {
		return err
	}

	factory.Start(ctx.Done())
	defer factory.Shutdown() // waits for the informer goroutines to stop
	if !cache.WaitForCacheSync(ctx.Done(), pods.HasSynced) {
		return ctx.Err() // cancelled before the first LIST completed
	}
	<-ctx.Done()
	return nil
}

// [end informer]
