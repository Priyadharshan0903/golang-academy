package kube

import (
	"context"
	"errors"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func pod(name string, ready bool, restarts int32, created time.Time) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "demo", CreationTimestamp: metav1.NewTime(created),
			Labels: map[string]string{"app": "web"}},
		Spec:   corev1.PodSpec{Containers: []corev1.Container{{Name: "c"}}},
		Status: corev1.PodStatus{Phase: corev1.PodRunning, ContainerStatuses: []corev1.ContainerStatus{{Ready: ready, RestartCount: restarts}}},
	}
}

// [start fake]

func TestPods(t *testing.T) {
	now := time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)
	// A fake clientset is an in-memory API server: no cluster needed.
	cs := fake.NewClientset(
		pod("web-b", true, 0, now.Add(-time.Hour)),
		pod("web-a", false, 3, now.Add(-90*time.Second)),
	)
	got, err := Pods(context.Background(), cs, "demo", "app=web", now)
	if err != nil {
		t.Fatal(err)
	}
	want := []PodSummary{
		{"web-a", corev1.PodRunning, "0/1", 3, 90 * time.Second},
		{"web-b", corev1.PodRunning, "1/1", 0, time.Hour},
	}
	if len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("Pods = %+v", got)
	}
}

// [end fake]

// [start conflict]

func TestSetConfigValueRetriesConflicts(t *testing.T) {
	cs := fake.NewClientset(&corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "settings", Namespace: "demo"}})

	// Make the first update fail with 409 Conflict, as if another client
	// had changed the ConfigMap between our read and our write.
	conflicts := 0
	cs.PrependReactor("update", "configmaps", func(k8stesting.Action) (bool, runtime.Object, error) {
		if conflicts == 0 {
			conflicts++
			return true, nil, apierrors.NewConflict(schema.GroupResource{Resource: "configmaps"}, "settings", nil)
		}
		return false, nil, nil // fall through to the normal fake behavior
	})

	if err := SetConfigValue(context.Background(), cs, "demo", "settings", "mode", "fast"); err != nil {
		t.Fatal(err)
	}
	cm, _ := cs.CoreV1().ConfigMaps("demo").Get(context.Background(), "settings", metav1.GetOptions{})
	if conflicts != 1 || cm.Data["mode"] != "fast" {
		t.Fatalf("conflicts=%d data=%v", conflicts, cm.Data)
	}
}

// [end conflict]

func TestSetConfigValueCreates(t *testing.T) {
	cs := fake.NewClientset()
	if err := SetConfigValue(context.Background(), cs, "demo", "new", "k", "v"); err != nil {
		t.Fatal(err)
	}
	cm, err := cs.CoreV1().ConfigMaps("demo").Get(context.Background(), "new", metav1.GetOptions{})
	if err != nil || cm.Data["k"] != "v" {
		t.Fatalf("%v %v", cm, err)
	}
}

func TestWatchPods(t *testing.T) {
	cs := fake.NewClientset(pod("existing", true, 0, time.Now()))
	ctx, cancel := context.WithCancel(context.Background())
	events := make(chan PodEvent, 10)
	done := make(chan error)
	go func() { done <- WatchPods(ctx, cs, "demo", events) }()

	expect := func(typ, name string) {
		t.Helper()
		select {
		case e := <-events:
			if e.Type != typ || e.Name != name {
				t.Fatalf("got %+v, want %s %s", e, typ, name)
			}
		case <-time.After(5 * time.Second):
			t.Fatalf("no %s event for %s", typ, name)
		}
	}
	expect("added", "existing") // from the initial LIST
	cs.CoreV1().Pods("demo").Create(ctx, pod("new", false, 0, time.Now()), metav1.CreateOptions{})
	expect("added", "new") // from the WATCH
	cs.CoreV1().Pods("demo").Delete(ctx, "new", metav1.DeleteOptions{})
	expect("deleted", "new")

	cancel()
	// nil, or context.Canceled if we cancelled before the cache reported synced.
	if err := <-done; err != nil && !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}
