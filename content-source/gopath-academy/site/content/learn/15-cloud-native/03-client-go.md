---
title: "15.3 client-go basics"
linkTitle: "15.3 client-go"
weight: 3
description: "Talk to Kubernetes from Go with client-go: connecting in and out of the cluster, typed clientsets, label selectors, optimistic concurrency and RetryOnConflict, informers and their local cache, and testing with a fake clientset."
goVersion: "1.27"
---

## What you will learn

- How a Go program connects to Kubernetes, inside a pod or from your laptop
- Listing and reading objects with the typed clientset
- Updates, `resourceVersion`, and retrying conflicts
- Informers: LIST plus WATCH plus a local cache, the foundation of every controller
- Testing without a cluster, using the fake clientset

## The idea

Everything in Kubernetes goes through one HTTP API. `kubectl`, controllers, the scheduler and your own programs all read and write objects through it. **client-go** is the official Go library for that API; `kubectl` itself is built on it.

The examples live in `examples/thirdparty/k8s`. The module needs `go mod tidy` once; the tests need no cluster.

## Connecting

{{< example file="examples/thirdparty/k8s/kube/kube.go" region="connect" >}}

Inside a pod, `rest.InClusterConfig` uses the service account token and CA certificate that Kubernetes mounts into every pod, and the API server address from environment variables. Outside, the same loading rules as `kubectl` read `$KUBECONFIG` or `~/.kube/config`, so the program uses whatever cluster and context `kubectl` uses. Setting a **user agent** makes your program identifiable in API server audit logs.

The program can then only do what its identity is allowed to do: in a pod, what **RBAC** grants the pod's service account (lesson 15.4 generates those rules).

## Reading objects

{{< example file="examples/thirdparty/k8s/kube/kube.go" region="list" >}}

`cs.CoreV1().Pods(namespace).List(...)` mirrors the API path `/api/v1/namespaces/{ns}/pods`. Every built-in kind has the same shape: `AppsV1().Deployments(ns)`, `NetworkingV1().Ingresses(ns)`, and so on, with `Get`, `List`, `Create`, `Update`, `Patch`, `Delete` and `Watch`. A **label selector** filters on the server, which matters in big clusters.

Run against the cluster from lesson 15.2, the `podlist` command prints:

```text
$ go run ./k8s/podlist -n shortener
NAME                        READY  STATUS   RESTARTS  AGE
postgres-0                  1/1    Running  0         13m58s
redis-84f65c4d6f-hwpz7      1/1    Running  0         13m58s
shortener-6b7df6c565-d7vfr  1/1    Running  0         7m45s
shortener-6b7df6c565-d9g7b  1/1    Running  0         7m48s

$ go run ./k8s/podlist -n shortener -l app.kubernetes.io/name=shortener
NAME                        READY  STATUS   RESTARTS  AGE
shortener-6b7df6c565-d7vfr  1/1    Running  0         7m46s
shortener-6b7df6c565-d9g7b  1/1    Running  0         7m49s
```

## Updates and conflicts

Every object has a **`resourceVersion`**, which changes on every write. An `Update` sends back the object you read, including that version; if someone else wrote the object in the meantime, the API server rejects your update with **409 Conflict** instead of silently overwriting their change. This is optimistic concurrency, the same idea as the job queue's attempts check in lesson 13.8.

The standard response is to read again, reapply your change, and retry:

{{< example file="examples/thirdparty/k8s/kube/kube.go" region="update" >}}

`retry.RetryOnConflict` retries only on conflicts, with a short backoff. The function must **re-read** the object each time; retrying the same stale object would conflict forever. For changes that touch only a few fields, a **patch** (`Patch` with a JSON merge patch, or server-side apply) avoids the read entirely, and with it most conflicts.

## Informers

A program that reacts to changes could poll with `List` every few seconds. With many programs and many objects that would overload the API server, and it would still react late. Instead, client-go programs use **informers**:

{{< example file="examples/thirdparty/k8s/kube/watch.go" region="informer" >}}

An informer does one **LIST** to fill a local cache, then opens a **WATCH**: a long-lived stream of changes. It keeps the cache up to date, calls your handlers for each change, and if the watch breaks (it will, eventually) it re-lists and resumes. Reads from the cache cost nothing on the API server. A **shared** informer factory lets every part of a program share one cache per kind.

Two details in the code are easy to miss. `WaitForCacheSync` blocks until the initial LIST has been processed: acting before that means acting on an incomplete picture. And a delete handler can receive a `DeletedFinalStateUnknown` instead of the object, when the watch missed the delete and the informer only found out on re-list.

Watching the shortener's namespace during `kubectl rollout restart`:

```text
$ go run ./k8s/podlist -n shortener -watch
added    shortener-6d454586f7-jln2b    Running     <- initial LIST
added    shortener-6d454586f7-hvfwg    Running
added    shortener-656cf6c78-g56kq     Pending     <- the rollout creates a new pod
updated  shortener-656cf6c78-g56kq     Pending
updated  shortener-656cf6c78-g56kq     Running
updated  shortener-6d454586f7-jln2b    Running     <- an old pod is told to terminate
added    shortener-656cf6c78-v2kvc     Pending
...
```

The old pods' `deleted` events follow a few seconds later, after their `SHUTDOWN_DELAY` and graceful shutdown (lesson 15.2).

Informers, a work queue, and a reconcile function together make a **controller**. Writing that machinery by hand is how the core Kubernetes controllers are built; for your own, controller-runtime (lesson 15.4) provides it.

## Testing without a cluster

`k8s.io/client-go/kubernetes/fake` provides a clientset backed by an in-memory object tracker. It implements the same `kubernetes.Interface`, so code written against the interface works with either:

{{< example file="examples/thirdparty/k8s/kube/kube_test.go" region="fake" >}}

**Reactors** inject behavior, such as a conflict on the first update, to test the retry path:

{{< example file="examples/thirdparty/k8s/kube/kube_test.go" region="conflict" run="cd examples/thirdparty && go test -v ./k8s/kube" >}}

The fake does not implement everything the real API server does (validation, defaulting, admission, server-side apply), so add tests against a real API server for those: `envtest` from controller-runtime runs a local `etcd` and `kube-apiserver`, and a throwaway kind or k3s cluster covers the rest.

## Common mistakes

### Polling with List

Use informers: one LIST, then a WATCH, and cheap cached reads.

### Updating a stale object in a retry loop

Re-read inside the retry function, or use a patch.

### Acting before the cache has synced

Your program then believes objects do not exist. Wait for `HasSynced`.

### Accepting kubernetes.Clientset instead of kubernetes.Interface

Then the fake cannot be substituted. Accept the interface (lesson 4.8).

## Key takeaways

- `rest.InClusterConfig` inside pods, kubeconfig loading rules outside; RBAC decides what the program may do.
- The typed clientset mirrors the API: `CoreV1().Pods(ns).List(ctx, opts)`.
- Updates carry `resourceVersion`; on 409 Conflict, re-read and retry with `retry.RetryOnConflict`, or patch.
- Informers LIST once, then WATCH, keeping a local cache; wait for it to sync.
- Test with the fake clientset and reactors; use envtest or a real cluster for API server behavior.

## Further reading

- [client-go](https://github.com/kubernetes/client-go) and its [examples](https://github.com/kubernetes/client-go/tree/master/examples)
- [Kubernetes API concepts: resource versions, watches](https://kubernetes.io/docs/reference/using-api/api-concepts/)
- [Sample controller](https://github.com/kubernetes/sample-controller), a controller written with client-go only
