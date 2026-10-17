# Kubernetes track

Status (2026-10-16): steps 1 to 3 work on a local cluster. Manifests are in `deploy/k8s` (the site) and `projects/shortener/k8s` (the shortener, Module 15). The ShortLink operator in `projects/shortlink-operator` runs on the same cluster.

Goal: run this site on Kubernetes, learning one concept per step. Each step has a "done when" check so we know it actually works.

## Step 1: Container image (Phase 0)

- [x] `deploy/server`: a small Go static file server with `/healthz`, graceful shutdown on SIGTERM
- [x] `deploy/Dockerfile`: multi-stage (Hugo build, Go build, distroless nonroot runtime)
- [x] Build locally: `docker build -f deploy/Dockerfile -t gopath-academy:dev .` (verified in the authoring sandbox with stand-in base images; the Hugo stage still needs a real build)
- [x] Run locally: `docker run --rm -p 8080:8080 gopath-academy:dev`

Done when: http://localhost:8080 shows the site and `curl localhost:8080/healthz` returns `ok`.

## Step 2: Local cluster with kind (Phase 1)

- [~] Install `kind` and `kubectl` (the sandbox used k3s v1.37 instead of kind; the manifests are the same)
- [~] `kind create cluster --name academy`
- [x] Load the image into the cluster (`kind load docker-image`, or `docker save | k3s ctr images import -`)
- [x] Write `deploy/k8s/base/deployment.yaml` (2 replicas)
- [x] Write `deploy/k8s/base/service.yaml` (ClusterIP)
- [x] `kubectl port-forward svc/academy 8080:80` (or through the ingress at academy.localtest.me)

Done when: the site loads through port-forward and deleting one pod does not cause downtime.

Concepts: Pod, ReplicaSet, Deployment, Service, labels and selectors.

## Step 3: Ingress and probes (Phase 1)

- [x] An ingress controller (Traefik, bundled with k3s; ingress-nginx in kind)
- [x] Add `ingress.yaml`
- [x] Add readiness and liveness probes on `/healthz`
- [x] Observe a rolling update with `kubectl rollout status`: 12,454 requests during a rollout, 0 failed (with `-shutdown-delay 5s`, measured with examples/m15/rollcheck)

Concepts: Ingress, probes, rolling updates.

## Step 4: CI to registry (Phase 2)

- [ ] GitHub Actions builds the image on push to `main`
- [ ] Push to `ghcr.io/priyadharshan0903/gopath-academy:<git-sha>`
- [~] Kustomize overlays: `local` done, `prod` pending

Concepts: image tags, immutability, Kustomize.

## Step 5: GitOps (Phase 3)

- [ ] Install Argo CD in the cluster
- [ ] Argo CD Application pointing at `deploy/k8s/overlays/prod`
- [x] Resource requests and limits
- [ ] HorizontalPodAutoscaler
- [x] PodDisruptionBudget

Concepts: GitOps, reconciliation, autoscaling, disruption budgets.

## Step 6: Real cluster (Phase 4)

- [ ] k3s on a free-tier VM (or another low-cost option)
- [ ] cert-manager with Let's Encrypt
- [ ] Domain and DNS

## Step 7: Observability (Phase 5)

- [ ] Expose Prometheus metrics from `deploy/server`
- [ ] kube-prometheus-stack, a basic dashboard
- [x] Structured JSON logs with `log/slog`
