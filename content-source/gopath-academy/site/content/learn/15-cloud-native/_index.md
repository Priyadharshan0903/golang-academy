---
title: "15. Go for cloud native"
weight: 15
description: "Containers and Kubernetes from the Go side: small secure images, services that behave well under rolling updates, client-go, and a real operator built with controller-runtime, all run on a Kubernetes cluster."
---

Kubernetes, Docker, containerd, Prometheus, etcd, Helm, Argo CD: the cloud native world is largely written in Go. This module covers both directions of that relationship: **running Go in Kubernetes** well, and **writing Go that drives Kubernetes**.

Everything here was run on a real cluster (k3s, Kubernetes 1.37) while the lessons were written: the images were built with the Dockerfiles shown, the manifests were applied, rolling updates were measured under load, and the operator created and deleted real short links. The output you see is from those runs.

**Time:** about 5 to 6 hours, plus the exercise and the project. You need Docker and a local cluster (kind, k3d, k3s or Docker Desktop) to follow along.

{{< cards >}}
  {{< card link="01-containers" title="15.1 Containerizing Go" subtitle="Multi-stage builds, distroless, 24 MB images." >}}
  {{< card link="02-go-in-kubernetes" title="15.2 Go in Kubernetes" subtitle="Probes, SIGTERM, zero-downtime rollouts, limits." >}}
  {{< card link="03-client-go" title="15.3 client-go basics" subtitle="Clientsets, informers, conflicts, fakes." >}}
  {{< card link="04-controller" title="15.4 Writing a controller" subtitle="A ShortLink operator with controller-runtime." >}}
  {{< card link="05-operators-in-practice" title="15.5 Operators in practice" subtitle="Idempotency, finalizers, status, leader election." >}}
  {{< card link="06-exercise" title="15.6 Exercise: drain gracefully" subtitle="The shutdown sequence, by hand." >}}
  {{< card link="07-project-shortener-k8s" title="15.7 Project: shortener on Kubernetes" subtitle="Milestone 4: the full stack on a cluster." >}}
{{< /cards >}}
