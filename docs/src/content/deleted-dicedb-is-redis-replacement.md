---
title: "Why DiceDB chose to be a Redis replacement?"
description: "Redis has become synonymous with in-memory caching, creating an ecosystem of tools, best practices, and mental models that many engineers already know well. Redis's simplicity and reliability have helped it scale across countless production systems, so building a completely new caching paradigm risked reinventing the wheel."
published_at: 2024-11-04
author: arpit
---

Redis has become synonymous with in-memory caching, creating an ecosystem of tools, best practices, and mental models that many engineers already know well. Redis's simplicity and reliability have helped it scale across countless production systems, so building a completely new caching paradigm risked reinventing the wheel.

Instead, we focused on delivering a solution that would align closely with Redis's familiarity while introducing major improvements for modern applications. Here’s why we chose to position DiceDB as a Redis-compatible replacement.

# Redis Compatibility for Easy Adoption

Adoption hurdles for new databases are often high. Engineers and developers must rework existing codebases, learn new APIs, and update dependencies, all of which introduce friction and slow down adoption. By making DiceDB Redis-compatible, we enable engineers and developers to start with minimal modifications and existing code often works as-is. This seamless integration not only reduces migration costs but also empowers teams to leverage DiceDB without substantial rewrites or operational overhead.

# Targeting Redis’s Architectural Gaps

Redis was designed over a decade ago, before multi-core systems became as prevalent and powerful as they are today. [Redis's single-threaded](https://redis.io/learn/operate/redis-at-scale/talking-to-redis/redis-server-overview) event loop approach, while effective in its time, does not fully harness the compute potential of modern, multi-core hardware. We identified this architectural gap as [a bottleneck](/benchmarks), limiting Redis's throughput on today's processors.

![DiceDB vs Redis - Throughput vs Number of Cores](https://github.com/user-attachments/assets/367f25c8-bc59-4787-bb37-63fd5d4314e4)

Following in the footsteps of architectures like [DragonflyDB](https://www.youtube.com/watch?v=XbV1LoVsbME), DiceDB employs a multi-threaded architecture that can more efficiently utilize all CPU cores, resulting in higher throughput and significantly better resource utilization. This design choice allows DiceDB to serve high-request-rate applications with lower latency and greater efficiency, without the need to sacrifice the familiar simplicity of Redis commands.

# Building Toward Horizontal Scalability

Unlike solutions that push single-node performance limits, DiceDB's goal is to serve as a unified, horizontally scalable cache capable of holding and querying large volumes of cached data across distributed systems and microservices. DiceDB's architecture is designed to scale out and up both, enabling applications to get max out of single node while being able to add nodes dynamically as demand increases. This approach provides a more practical path to scaling for large applications.

# End

By addressing these architectural limitations, DiceDB aims to be more than just a faster Redis clone. It's designed to be the foundational infrastructure for the next generation of applications and micro-services offering Redis familiarity combined with high performance and scalability across distributed environments.
