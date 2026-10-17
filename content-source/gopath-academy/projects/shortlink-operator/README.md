# shortlink-operator

A Kubernetes operator for the course's URL shortener (lessons 15.4 and 15.5). A `ShortLink` custom resource declares a short link; the operator creates it through the shortener's HTTP API, keeps it pointing at `spec.url`, reports the short URL in its status, and deletes it when the resource is deleted.

```yaml
apiVersion: shortener.example.com/v1alpha1
kind: ShortLink
metadata:
  name: k8s-docs
spec:
  url: https://kubernetes.io/docs
  code: k8s        # optional; defaults to the resource name
```

```text
$ kubectl get shortlinks
NAME       CODE      SHORT URL                               READY   AGE
k8s-docs   k8s       http://shortener.localtest.me/k8s       True    28s
```

This directory is its own Go module (controller-runtime, client-go). `controller-gen` is pinned as a `go tool`.

## Layout

```text
api/v1alpha1/                 ShortLink types and markers; generated deep-copy code
internal/controller/          the reconciler and its tests (fake client)
internal/shortener/           a small client for the shortener API
cmd/manager/                  main: the controller-runtime manager
config/crd/                   generated CRD
config/rbac/                  generated ClusterRole, service account, bindings, leader election
config/manager/               Deployment (2 replicas, leader election)
config/default/               kustomization tying it together
config/samples/               example ShortLinks
```

## Develop

```bash
go mod tidy
go generate ./...            # regenerate deep-copy code, CRD and RBAC after changing markers
go test ./...
go run ./cmd/manager -shortener-url=http://localhost:8080   # against the current kubectl context
```

## Deploy

The shortener must already run in the cluster (see `projects/shortener`, lesson 15.7).

```bash
docker build -t shortlink-operator:dev .     # then load it into your cluster
kubectl apply -k config/default
kubectl -n shortener apply -f config/samples/shortlink.yaml
```
