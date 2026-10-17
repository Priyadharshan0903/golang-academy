---
title: "Monzo: a bank built as Go microservices on Kubernetes"
linkTitle: "Monzo"
weight: 9
description: "Monzo built a bank backend from day one as Go microservices: about 150 services by 2016, Docker on Kubernetes, HTTP RPC through a service mesh proxy, Kafka for events, and infrastructure costs cut to a quarter by moving to Kubernetes."
---

| | |
|---|---|
| **Company** | Monzo, a UK bank |
| **Architecture** | Go microservices on Kubernetes from the start |
| **Source** | Oliver Beattie, [Building a Modern Bank Backend](https://monzo.com/blog/2016/09/19/building-a-modern-bank-backend), Monzo blog, 19 September 2016 |
| **Connects to** | Modules 7, 12, 13, 15 |

## The story

Most startups begin with one application and split it later. Monzo, building a bank, started with **distributed microservices from day one**, a choice its head of engineering, Oliver Beattie, describes as unusual for an early-stage company. Three backend engineers started it; by the public beta there were about **100 services**, and about **150** when the post was written in 2016.

The platform, as described in the post:

- **Go** as the main language, chosen because it makes it easy to build low-latency, high-concurrency servers that suit microservices; some services in Java, Python and Scala, since no single ecosystem has everything a bank needs.
- **Docker containers on Kubernetes**, running on CoreOS machines in AWS, after an earlier setup with Mesos and Marathon. Moving to Kubernetes cut production infrastructure costs to about **a quarter** of what they had been.
- **HTTP-based RPC** between services through **linkerd**, a local proxy that handles load balancing, routing and retries of idempotent requests: an early service mesh.
- **Kafka** for asynchronous messaging, chosen for availability, horizontal scaling, persistence, replay and at-least-once delivery.
- Shared infrastructure exposed **as services** too, such as a distributed lock service in front of etcd.

The targets stated: payments processed end to end in under a second, card authorizations answered within tens of milliseconds, and an architecture designed for hundreds of millions of customers, available all the time.

## What changed since

The building blocks Monzo assembled by hand in 2016 have become standard: Kubernetes is the default platform, service meshes are common, and Go remains the language of that ecosystem. The 2016 post deliberately left out data storage for a later article, so this case study does not cover it.

## What you can apply

- **Retries must be safe**: the mesh only retries **idempotent** requests. Design operations so retrying them is harmless (lessons 13.8 and 15.4).
- **At-least-once delivery** from Kafka means consumers must handle duplicates, just like the job queue in lesson 13.8.
- **Low-latency, high-concurrency servers** with timeouts, limits and graceful shutdown are what Module 12 builds.
- **Kubernetes' resource model** (requests, limits, bin packing) is how Monzo cut costs; Module 15 shows how Go services behave inside it.

## Sources

- Oliver Beattie: [Building a Modern Bank Backend](https://monzo.com/blog/2016/09/19/building-a-modern-bank-backend), Monzo blog, 19 September 2016
