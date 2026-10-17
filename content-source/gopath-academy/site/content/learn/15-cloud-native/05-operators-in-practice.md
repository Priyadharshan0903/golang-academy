---
title: "15.5 Operators in practice"
linkTitle: "15.5 Operators in practice"
weight: 5
description: "The patterns that make operators reliable in production: level-based reconciliation, owner references and garbage collection, status and observedGeneration, high availability with leader election (failover measured), RBAC, API versioning, and when not to write an operator."
goVersion: "1.27"
---

## What you will learn

- The principles behind every reliable operator
- Owner references and garbage collection, the other half of cleanup
- Status, conditions and `observedGeneration`
- Running several replicas with leader election, and how fast failover is (measured)
- Least-privilege RBAC, API versions, and when not to write an operator at all

## Level-based, not edge-based

The single most important idea: a controller reacts to **state**, not to **events**. Reconcile receives a name and works out what to do by comparing desired state (the spec) with actual state (the world). Events can be missed, merged or repeated (after a restart, a re-list, a network blip), and a level-based controller does not care: it converges anyway. Every rule below follows from this.

- **Idempotent actions.** Running Reconcile twice must be harmless (lesson 15.4 tests it).
- **No state in memory.** Everything the controller needs to know must be readable from the cluster or the external system, because the process can restart at any moment. Put what you learn in **status**, not in variables.
- **Deterministic names** for anything you create, so you can find it again after a crash.
- **One step at a time.** Do a bit of work, update status, return; let the next reconcile continue. Long blocking calls hold a worker and delay other objects.

## Owner references and garbage collection

Finalizers (lesson 15.4) clean up resources **outside** Kubernetes. For resources **inside** Kubernetes, use **owner references** instead: when an operator creates, say, a Deployment and a Service for a custom resource, it sets the custom resource as their owner:

```go
if err := controllerutil.SetControllerReference(&app, deployment, r.Scheme); err != nil {
	return ctrl.Result{}, err
}
```

When the owner is deleted, Kubernetes' garbage collector deletes the owned objects automatically: no finalizer, no code. And with `Owns(&appsv1.Deployment{})` in `SetupWithManager`, a change to an owned Deployment (someone scales it by hand, say) triggers a reconcile of its owner, which puts it back. This is how a Deployment owns ReplicaSets, which own Pods.

## Status, conditions and observedGeneration

`metadata.generation` increases whenever the **spec** changes. A controller writes `status.observedGeneration` to say "my status describes this generation". A client that changed the spec can then wait until `observedGeneration` catches up before trusting the status. `kubectl rollout status` works exactly like that for Deployments.

Conditions (`Ready`, `Progressing`, `Degraded`, ...) follow the shape of `metav1.Condition`: a type, a status (`True`, `False`, `Unknown`), a machine-readable `Reason` in CamelCase and a human-readable `Message`. The ShortLink operator uses `LinkReady`, `CodeTaken` and `APIError` as reasons, so `kubectl describe` explains every failure, and automation can wait on readiness:

```bash
kubectl -n shortener wait --for=condition=Ready shortlink/go-docs --timeout=60s
```

## High availability: leader election

Run one replica and a node failure stops reconciliation until the pod is rescheduled. Run two without coordination, and both reconcile the same objects at once. **Leader election** solves this: replicas compete for a **Lease** object, only the holder runs controllers, and the others wait to take over.

The operator's Deployment runs 2 replicas with `--leader-elect`. On the cluster:

```text
$ kubectl -n shortlink-system get lease
NAME                                       HOLDER
shortlink-operator.shortener.example.com   shortlink-operator-64c9df47c9-rb7rw_cfd2...
```

Deleting the leader pod and measuring how long until another replica holds the lease:

```text
default settings:                       new leader after ~15 s
with LeaderElectionReleaseOnCancel:     new leader after 1.3 s
```

By default, a stopping leader just stops renewing, and the others must wait for the lease to expire (15 seconds). With `LeaderElectionReleaseOnCancel: true`, the manager gives up the lease when it shuts down, so a standby takes over almost at once. It is safe only if the process exits right after the manager stops, as `main` does here. During those seconds nothing is lost: changes queue up as state in the cluster and are reconciled by the new leader. (Leader election needs RBAC for `leases`, which `config/rbac/leader_election.yaml` grants in the operator's own namespace only.)

## Least-privilege RBAC

An operator's service account is a powerful identity. The `+kubebuilder:rbac` markers generate exactly the permissions the code declares; review them like code. Grant cluster-wide permissions only for cluster-wide work; prefer namespaced `Role`s, and restrict a controller's cache to the namespaces it manages (`cache.Options{DefaultNamespaces: ...}`) when it does not need the whole cluster. Never request `*` verbs or resources.

## Evolving the API

`v1alpha1` signals that the API may change. Kubernetes API conventions describe the path: `v1alpha1` → `v1beta1` → `v1`, with a **conversion webhook** when the stored representation changes between versions, so old objects keep working. Within a version, only add **optional** fields; never remove or rename fields others depend on. Treat your CRD like any public API, because once other teams write YAML against it, it is one.

## When not to write an operator

Operators shine when there is ongoing, domain-specific operational knowledge to automate: databases that need failover and backups, certificates that need renewal, external services that must stay in sync with declarations, like the ShortLink. They are overkill when:

- a Deployment, a Helm chart or Kustomize already express what you need;
- the work is a one-off: a Job or a script is simpler;
- nobody will maintain it: an operator is software running with cluster privileges, with all that implies.

## Tools

- **Kubebuilder** scaffolds projects like `shortlink-operator` (`kubebuilder init`, `kubebuilder create api`) with Makefile targets for controller-gen, envtest and deployment. The structure in this lesson follows its conventions, so moving to it is easy.
- **Operator SDK** builds on Kubebuilder and adds integration with the Operator Lifecycle Manager.
- **controller-runtime** alone, as in this lesson, when you want to see every moving part.

## Key takeaways

- Reconcile on state, not events: idempotent, no memory state, deterministic names, small steps.
- Owner references garbage-collect Kubernetes resources; finalizers clean up external ones.
- Report `observedGeneration` and standard conditions so humans and tools can tell where things stand.
- Run several replicas with leader election; releasing the lease on shutdown cut failover from about 15 s to 1.3 s here.
- Generate least-privilege RBAC from markers, version your API carefully, and write an operator only when there is real operational logic to automate.

## Further reading

- [Kubernetes API conventions](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md)
- [Kubebuilder Book: Good practices](https://book.kubebuilder.io/reference/good-practices)
- [Kubernetes: Garbage collection and owner references](https://kubernetes.io/docs/concepts/architecture/garbage-collection/)
- [Kubernetes: Leases](https://kubernetes.io/docs/concepts/architecture/leases/)
