---
title: "15.2 Go in Kubernetes: probes, signals and graceful shutdown"
linkTitle: "15.2 Go in Kubernetes"
weight: 2
description: "Run a Go service well on Kubernetes: what happens when a pod stops, why a shutdown delay prevents failed requests (59 failures versus 0, measured), liveness, readiness and startup probes, resources with GOMEMLIMIT and GOMAXPROCS, security contexts, and Kustomize."
goVersion: "1.27"
---

## What you will learn

- What happens, step by step, when Kubernetes stops a pod
- Why a Go service should keep serving for a few seconds after SIGTERM, with measurements
- Liveness, readiness and startup probes, and what each should check
- Requests, limits, `GOMEMLIMIT`, and how Go 1.25+ sets `GOMAXPROCS` in a pod
- A locked-down security context
- Managing manifests with Kustomize

## The Deployment

This lesson walks through the shortener's Deployment from milestone 4 (lesson 15.7). Here it is in full; the sections below explain each part.

{{< example file="projects/shortener/k8s/base/deployment.yaml" >}}

## How a pod stops

During a rolling update, a scale-down or a node drain, Kubernetes terminates pods. Two things then happen **at the same time**, driven by different components:

1. The kubelet sends the container **SIGTERM**, and starts the `terminationGracePeriodSeconds` countdown (30 s here). If the process is still running at the end, it gets **SIGKILL**.
2. The endpoints controller removes the pod from the Service's endpoints, and every kube-proxy, ingress controller and load balancer then has to notice and update its own routing.

Step 2 takes time, typically a second or more. If your program exits as soon as it receives SIGTERM (which is what a Go program does by default), new requests are still being sent to it for that time, and they fail.

The shortener handles SIGTERM like this (lesson 12.8): mark itself not ready, **keep serving for `SHUTDOWN_DELAY`**, then call `http.Server.Shutdown`, which stops accepting connections and waits up to `SHUTDOWN_TIMEOUT` for in-flight requests.

### Measured

To see the difference, `examples/m15/rollcheck` sends a steady stream of requests through the ingress (4 clients, a new connection per request) while the Deployment is restarted:

{{< example file="examples/m15/rollcheck/main.go" region="run" >}}

```bash
rollcheck -url http://127.0.0.1/healthz -host shortener.localtest.me -duration 50s &
kubectl -n shortener rollout restart deployment/shortener
```

Results on a k3s cluster with Traefik as the ingress controller, same image, only `SHUTDOWN_DELAY` changed:

```text
SHUTDOWN_DELAY=5s:   15364 requests, 0 failed
                       HTTP 200: 15364

SHUTDOWN_DELAY=0s:   13889 requests, 59 failed
                       HTTP 200: 13830
                       HTTP 502: 55
                       error "...context deadline exceeded (Client.Timeout exceeded while awaiting headers)": 4
```

Without the delay, the ingress kept routing to pods that had already stopped listening: 55 requests got `502 Bad Gateway` and 4 hung until the client gave up. With a 5 second delay, nothing failed. The same test on this website's own Deployment (`deploy/k8s`, with `-shutdown-delay 5s`) gave 12,454 requests and 0 failures.

Rules that follow:

- Handle **SIGTERM** (`signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)`); `os.Interrupt` alone only covers Ctrl+C.
- Keep serving for a few seconds after SIGTERM before calling `Shutdown`. (Some teams use a `preStop` hook with `sleep` instead; that needs a `sleep` binary in the image, which distroless does not have. Kubernetes 1.30+ also offers a built-in `sleep` action for `preStop`.)
- Set `terminationGracePeriodSeconds` longer than delay plus shutdown timeout (here 30 > 5 + 20), or the kubelet kills the process mid-shutdown.
- Do not exit with an error code for a normal SIGTERM: the shortener's `run` returns `nil`.

## Probes

| Probe | Question | On failure | The shortener |
|---|---|---|---|
| **startupProbe** | Has it finished starting? | Keep waiting (up to `failureThreshold × periodSeconds`), then restart | `/healthz`, up to 60 s |
| **readinessProbe** | Should it get traffic now? | Removed from Service endpoints; **not** restarted | `/readyz` |
| **livenessProbe** | Is it alive, or stuck? | **Restarted** | `/healthz` |

The startup probe holds the other two back until the program is up, so a slow start (migrations, cache warm-up) is not mistaken for a hang. Liveness checks must be **trivial**: if `/healthz` checked the database, a database outage would restart every replica at once, which helps nobody (lesson 12.5). Readiness is where "not now" belongs: during start-up, during shutdown, and optionally when a dependency the pod cannot work without is unavailable.

## Resources, GOMEMLIMIT and GOMAXPROCS

```yaml
resources:
  requests: {cpu: 100m, memory: 64Mi}   # what the scheduler reserves
  limits:   {cpu: "1",  memory: 128Mi}  # what the kernel enforces
env:
  - {name: GOMEMLIMIT, value: 115MiB}   # about 90% of the memory limit
```

- **Memory limit.** Exceeding it gets the container killed (`OOMKilled`). The Go GC does not know about the limit unless you tell it: `GOMEMLIMIT` at about 90% of the limit makes the GC work harder as memory approaches it, instead of the kernel killing the pod (lesson 11.3). The admin port confirms it: `go_gc_gomemlimit_bytes 1.2058624e+08`, which is 115 MiB.
- **CPU limit and GOMAXPROCS.** Since Go 1.25, when the module's `go` line is 1.25 or later, the runtime reads the cgroup CPU limit and sets `GOMAXPROCS` to match (lesson 11.1), rounding up, with a **minimum of 2**. On the 2-CPU test node, the shortener with a limit of 1 CPU reported `go_sched_gomaxprocs_threads 2`, exactly as that rule says. On a large node, the same limit prevents a pod from running 64 threads against a 1 CPU quota and being throttled. Older Go versions, or modules whose `go.mod` says `go 1.24` (check with `go version -m`, lesson 14.4), keep the old default; use `go.uber.org/automaxprocs` there.
- Whether to set CPU limits at all is debated: limits cause throttling, while requests alone share spare CPU. Either way, set **requests** for every container, so the scheduler can place pods sensibly.

## Security context

```yaml
securityContext:              # pod level
  runAsNonRoot: true
  runAsUser: 65532
  seccompProfile: {type: RuntimeDefault}
containers:
  - securityContext:          # container level
      allowPrivilegeEscalation: false
      readOnlyRootFilesystem: true
      capabilities: {drop: ["ALL"]}
```

A Go service on distroless needs none of the things these settings take away: no root, no writable filesystem (the shortener writes nothing to disk), no Linux capabilities. These settings satisfy the **restricted** Pod Security Standard, which a namespace can enforce with a label (the read-only root filesystem goes further than it requires). As lesson 15.1 showed, `runAsNonRoot` needs a numeric user in the image or in `runAsUser`.

## Kustomize

The manifests are split into a **base** that holds what every environment shares, and **overlays** that change it per environment:

```text
projects/shortener/k8s/
  base/                 Deployment, Service, Ingress, PodDisruptionBudget
  overlays/local/       namespace, PostgreSQL and Redis in the cluster, secrets, image settings
```

{{< example file="projects/shortener/k8s/overlays/local/kustomization.yaml" >}}

`kubectl apply -k projects/shortener/k8s/overlays/local` renders and applies it; `kubectl kustomize ...` only renders, which is useful in reviews. `secretGenerator` adds a hash of the content to the Secret's name and rewrites every reference, so changing a secret automatically rolls the Deployment. Kustomize is built into `kubectl`; Helm is the main alternative, with templates and packaging for distributing software to others.

## Common mistakes

### Exiting immediately on SIGTERM

Requests in the routing window fail, as measured above. Keep serving for a few seconds.

### Liveness probes that check dependencies

One database hiccup restarts every replica. Keep liveness trivial.

### No GOMEMLIMIT in a memory-limited pod

The GC lets the heap grow past the limit and the kernel kills the pod.

### A grace period shorter than your shutdown

The kubelet sends SIGKILL in the middle of draining.

## Key takeaways

- SIGTERM and endpoint removal happen concurrently; keep serving, unready, for a few seconds after SIGTERM. Measured: 59 failed requests without a delay, 0 with one.
- Startup probes protect slow starts, readiness controls traffic, liveness restarts: keep liveness trivial.
- Set `GOMEMLIMIT` to about 90% of the memory limit; Go 1.25+ derives `GOMAXPROCS` from the CPU limit (minimum 2).
- Run as non-root with a read-only root filesystem and no capabilities.
- Kustomize keeps a shared base and per-environment overlays, built into `kubectl`.

## Further reading

- [Kubernetes: Pod lifecycle and termination](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#pod-termination)
- [Kubernetes: Configure liveness, readiness and startup probes](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/)
- [Pod Security Standards](https://kubernetes.io/docs/concepts/security/pod-security-standards/)
- [Kustomize](https://kubectl.docs.kubernetes.io/references/kustomize/)
