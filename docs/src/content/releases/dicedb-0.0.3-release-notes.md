---
title: "DiceDB v0.0.3 Release Notes"
description: "The release of DiceDB version 0.0.3 is a major step forward in our mission to deliver an in-memory real-time database with SQL-based reactivity. In this version, DiceDB betters its SQL-based reactivity, advanced concurrency controls, and lock-free architecture."
published_at: 2024-09-11
author: arpit
---

The release of DiceDB version 0.0.3 is a major step forward in our mission to deliver an in-memory real-time database with SQL-based reactivity. In this version, DiceDB betters its SQL-based reactivity, advanced concurrency controls, and lock-free architecture. Let's break down each significant change, enhancement, and addition from a technical standpoint.

## Query and DSQL Improvements

Two significant improvements in query execution and performance handling were introduced in this release, both of which contribute directly to the efficiency of real-time data operations.

### Fingerprint-Based Caching in `QWATCH`

We have added **common cache support per fingerprint** in `QWATCH`, our query watcher subsystem. This caching mechanism allows for more efficient query execution, particularly for frequently accessed data patterns. By caching queries based on their fingerprint, we minimize repeated computation for identical queries, reducing query execution times and alleviating the burden on system resources.

### Exponential Backoff for `QWATCH` Writes

For scenarios involving high-load or network instability, we have implemented **retry with exponential backoff** for `QWATCH` writes. This improvement enhances the system's resilience by gradually increasing retry intervals during transient failures, which avoids overwhelming the system during network issues or spikes in traffic. This retry mechanism ensures that `QWATCH` remains responsive under duress without risking excessive load.

### Support for LIKE and NOT LIKE Operators

The addition of the `LIKE` and `NOT LIKE` operators allows for more flexible string matching in SQL queries. These operators are particularly useful in scenarios requiring pattern-based searches, enabling developers to implement more sophisticated querying logic. The inclusion of these operators enhances DiceDB's SQL reactivity layer, making it a more powerful tool for real-time applications requiring flexible querying.

### Refined DSQL Behavior

One of the key changes in DiceDB's Structured Query Language (DSQL) is the **deprecation of the FROM clause**. In this release, key filtering is now handled directly in the `WHERE` clause, aligning DiceDB's query structure more closely with standard SQL practices. This shift simplifies query syntax while maintaining the unique optimizations that DiceDB brings to real-time data processing.

## SQL Parser Optimization: `ORDER BY` Handling

Improvements to the SQL parser include more robust handling of the `ORDER BY` clause. This enhancement provides more efficient sorting capabilities, particularly for large datasets. Optimizing the parsing and execution of `ORDER BY` operations leads to lower query latencies and more predictable performance, especially in scenarios involving complex sorting criteria or large data volumes.

## Enhanced Compatibility with Redis

DiceDB continues to solidify its role as a drop-in replacement for Redis. We have focused on ensuring compatibility with Redis's most critical features while optimizing performance for modern hardware.

### Bitmap Command Compatibility

One of the notable changes in this release is the resolution of deviations in bitmap commands between Redis and DiceDB. This ensures that users migrating from Redis to DiceDB can maintain existing application logic without encountering unexpected behavior. Full bitmap compatibility aligns DiceDB more closely with Redis, reducing the friction involved in migration.

### Expanded Hash Operations: `HGET` Command

A crucial addition to this release is the **HGET command**, extending DiceDB's support for hash-based operations. The ability to efficiently retrieve specific fields from a hash enhances DiceDB's flexibility in handling structured data and is especially useful for applications with dynamic or nested data structures.

### Improved Expiry Command Handling

The `EXPIRE` command's behavior has been refined to handle various flag combinations more consistently. This ensures that edge cases, particularly those that could lead to unexpected TTL (Time-to-Live) behavior, are resolved. The command now behaves predictably across all combinations of flags, contributing to more reliable session management and key expiration logic in real-time applications.

## JSON Manipulation Capabilities

As part of our ongoing effort to support complex data structures, we have introduced the `JSON.ARRLEN` command. This command allows for efficient retrieval of the length of JSON arrays stored within DiceDB, further expanding our database’s capability to handle JSON-based datasets. The addition of this command is crucial for applications relying heavily on dynamic and nested data structures, such as document-based systems or microservices with REST-like responses.

## Architectural Improvements: Lock-Free Store and Asynchronous Notifications

This release introduces two significant architectural improvements: the removal of locking structures for the store and the implementation of asynchronous notifications for key changes.

### Lock-Free Store Design

The removal of locking structures in the store is one of the most impactful optimizations introduced in this release. By transitioning to a fully **lock-free design**, DiceDB reduces contention and overhead, especially in highly concurrent environments. This change is vital for scaling DiceDB in multicore systems, as it enables more efficient parallelism and improves the overall throughput of data operations.

### Asynchronous Notifications for Key Changes

We've also decoupled the write path from query processing by introducing **asynchronous notifications** to the Query Manager for key changes. This decoupling improves write performance by offloading the responsibility of notifying queries to a separate asynchronous process. As a result, write operations are no longer bottlenecked by synchronous query updates, leading to significant latency reductions and improved responsiveness in high-throughput systems.

## Code Refactoring and Structural Improvements

The foundation of this release focused on substantial code refactoring and architectural enhancements. By improving the modularity and maintainability of DiceDB, we have created a more efficient development environment and laid the groundwork for future scalability.

### Separation of the Store Package

One of the key architectural changes is the separation of the **store package**. This separation introduces a cleaner modular design, making it easier to maintain, extend, and test the system. A modular codebase simplifies debugging, reduces the chances of unintended coupling between unrelated parts of the system, and isolates specific functionality for targeted optimizations.

### Internalization of Packages

Several packages have been moved to **internal** directories, enforcing encapsulation and eliminating access to certain components from external modules. This improves code hygiene, reduces the risk of misuse by external applications, and ensures that internal components are only modified in controlled environments. This change aligns with best practices for building scalable software systems, where clear boundaries between components are critical.

## Error Handling Optimization

Error handling has been another critical focus area in this release. We've refined the **type error return** mechanisms, leading to improved system reliability and easier error tracing during runtime. By enhancing the granularity of error types and return paths, DiceDB is now better equipped to handle unexpected conditions without cascading failures.

### Refined Type Error Returns

The refactoring of the error return types enhances our ability to diagnose and fix issues. By providing more specific error messages and ensuring consistent error handling across the codebase, we have reduced the surface area for potential bugs. This improvement also simplifies unit testing and improves the overall robustness of the system, particularly under heavy workloads where error tolerance is critical.

## Summary

DiceDB v0.0.3 introduces substantial improvements in performance, compatibility, and functionality. The refactored codebase, optimized query execution, and enhanced SQL capabilities make DiceDB a more powerful tool for building truly real-time applications. The lock-free store design, coupled with improved error handling and retry mechanisms, significantly improves the system's reliability and scalability, making DiceDB a compelling choice for modern, high-performance workloads. This release continues to position DiceDB as a leading solution in the real-time database space, combining the best of Redis compatibility with the flexibility and power of SQL-based reactivity.
