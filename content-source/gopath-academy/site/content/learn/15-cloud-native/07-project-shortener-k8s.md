---
title: "15.7 Project milestone: the shortener on Kubernetes"
linkTitle: "15.7 Project: shortener on Kubernetes"
weight: 7
description: "Milestone 4 of the URL shortener: a container image, Kubernetes manifests with Kustomize, PostgreSQL and Redis in the cluster, an ingress, zero-downtime rollouts, and the ShortLink operator managing links declaratively."
goVersion: "1.27"
---

## The milestone

| Milestone | After module | Adds | Status |
|---|---|---|---|
| 1. Core | 5. Errors | Domain types, errors, validation, `Service`, in-memory store | done |
| 2. HTTP API | 12. HTTP services | JSON API, redirects, middleware, metrics, shutdown | done |
| 3. Postgres | 13. Working with data | PostgreSQL store, migrations, Redis cache | done |
| **4. Kubernetes** (this lesson) | 15. Go for cloud native | Image, manifests, operator | done |

Over four milestones the shortener went from a package with an in-memory map to a service running on Kubernetes with a database, a cache, metrics and its own operator. Milestone 4 adds:

```text
projects/shortener/
  Dockerfile, .dockerignore         the image (lesson 15.1)
  k8s/base/                         Deployment, Service, Ingress, PodDisruptionBudget (lesson 15.2)
  k8s/overlays/local/               namespace, PostgreSQL, Redis, secrets
  internal/... Delete               DELETE /api/links/{code}, needed by the operator
projects/shortlink-operator/        the ShortLink operator (lesson 15.4)
```

## Run the whole thing

You need Docker and a local cluster with an ingress controller (k3s or k3d include Traefik; for kind, install ingress-nginx).

```bash
# 1. Images
cd projects/shortener && go mod tidy && docker build -t shortener:dev . && cd -
cd projects/shortlink-operator && go mod tidy && docker build -t shortlink-operator:dev . && cd -

# 2. Load them into the cluster (pick yours)
kind load docker-image shortener:dev shortlink-operator:dev
# k3s:  docker save shortener:dev shortlink-operator:dev | sudo k3s ctr images import -

# 3. The shortener, PostgreSQL and Redis
kubectl apply -k projects/shortener/k8s/overlays/local
kubectl -n shortener rollout status deployment/shortener

# 4. The operator, then some links
kubectl apply -k projects/shortlink-operator/config/default
kubectl -n shortener apply -f projects/shortlink-operator/config/samples/shortlink.yaml
kubectl -n shortener get shortlinks

# 5. Use it (*.localtest.me resolves to 127.0.0.1)
curl -i http://shortener.localtest.me/k8s
```

On the test cluster, the first start showed one more real-world behavior: the shortener pods started faster than PostgreSQL, failed to connect, and restarted (`CrashLoopBackOff`) a few times until the database was ready, then ran normally:

```text
$ kubectl -n shortener get pods
NAME                         READY   STATUS    RESTARTS      AGE
postgres-0                   1/1     Running   0             48s
redis-84f65c4d6f-hwpz7       1/1     Running   0             48s
shortener-6dd697b989-bktgh   1/1     Running   3 (33s ago)   48s
shortener-6dd697b989-ngkgr   1/1     Running   3 (33s ago)   48s
```

The logs of a restarted pod show why (reproduced later by stopping PostgreSQL during a rollout):

```text
shortener: database: failed to connect to `user=postgres database=postgres`: 10.43.115.125:5432 (postgres): dial error: dial tcp 10.43.115.125:5432: connect: connection refused
```

During that reproduction, the old pods kept serving, because the rollout uses `maxUnavailable: 0`: a new pod must become ready before an old one is removed. Kubernetes' restarts make this self-healing, but restarts back off exponentially, so a service that waits for its dependencies starts faster. See "Your turn".

## What runs where

```text
                          ┌─────────────────────── namespace shortener ───────────────────────┐
 curl ──► Ingress ──► Service shortener ──► 2 × shortener pod ──► Service postgres ──► postgres-0
 (Traefik)                (port 80)            :8080 API          │                   (StatefulSet, PVC)
                                               :9090 metrics      └─► Service redis ──► redis
                          └───────────────────────────────────────────────────────────────────┘
 kubectl apply ShortLink ──► API server ──► shortlink-operator (2 replicas, 1 leader) ──► shortener API
                                            namespace shortlink-system
```

- **PostgreSQL** runs as a **StatefulSet** with a PersistentVolumeClaim, so its data survives pod restarts. On a laptop that is fine; in production, most teams use a managed database or a dedicated operator such as CloudNativePG.
- **Redis** is a plain Deployment with no persistence, because it is a cache: losing it costs only speed.
- The **database URL** reaches the shortener as a **file** from a Secret (`DATABASE_URL_FILE`, lesson 12.4), never as a plain environment variable in the manifest.
- The **operator** talks to the shortener through its cluster-internal Service: `http://shortener.shortener.svc.cluster.local`.

## Verified behavior

Everything below was measured or observed on the test cluster (k3s, Kubernetes 1.37):

| Check | Result |
|---|---|
| Rolling update under load, `SHUTDOWN_DELAY=5s` | 15,364 requests, 0 failed |
| Same, `SHUTDOWN_DELAY=0s` | 13,889 requests, 59 failed (55 × 502, 4 timeouts) |
| `GOMEMLIMIT` from the manifest | `go_gc_gomemlimit_bytes 1.2058624e+08` (115 MiB) |
| Migrations on first start | `migrations applied=["0001_create_links.sql"]` |
| `kubectl apply` of a ShortLink | Ready within seconds; redirect works through the ingress |
| Changing `spec.url` | link recreated; redirect follows the new URL |
| Deleting a ShortLink | link deleted in the shortener (404); the delete waited until cleanup succeeded |
| Deleting the operator's leader pod | new leader in 1.3 s (with lease release on shutdown) |

## Your turn

1. **Wait for the database.** Instead of crashing when PostgreSQL is not up yet, retry `pool.Ping` with backoff for up to 30 seconds in `openStore`. Measure the start-up time before and after.
2. **A `prod` overlay.** Image from a registry with an immutable tag (`ghcr.io/you/shortener:v1.4.0`), 3 replicas, a HorizontalPodAutoscaler on CPU, TLS on the Ingress with cert-manager, and the database URL from your secret manager.
3. **Scrape metrics.** Install the Prometheus Operator (kube-prometheus-stack) and add a `ServiceMonitor` for the `admin` port. Graph request rate by route.
4. **Network policies.** Allow traffic to PostgreSQL and Redis only from the shortener pods.
5. **Extend the operator**: a `ShortLink` that sets `spec.code` to an existing link of someone else should show `CodeTaken`. Add a printer column for the condition's reason, and an event (`r.Recorder.Event`) when a link is created.

## Course recap

This is the end of the main curriculum. Across 15 modules you have gone from `go mod init` to an operator running on Kubernetes: the language and its idioms, concurrency, the standard library, testing and profiling, the runtime underneath, and the services, data, tools and deployments of production Go. The [case studies](/case-studies/) show the same ideas at work in real companies, from Docker and Kubernetes to Uber, Twitch and Discord.
