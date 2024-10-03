---
title: "DiceDB v0.0.2 Release Notes"
description: "We are proud to announce the release of DiceDB version 0.0.2, a significant update that enhances our high-performance, in-memory real-time database. This release ships new SQL capabilities, advanced data type handling, performance optimizations, and expanded support for Redis-compatible commands, further solidifying DiceDB as a superior choice for building and scaling real-time applications on modern hardware."
published_at: 2024-09-06
author: arpit
---

We are proud to announce the release of DiceDB version 0.0.2, a significant update that enhances our high-performance, in-memory real-time database. This release ships new SQL capabilities, advanced data type handling, performance optimizations, and expanded support for Redis-compatible commands, further solidifying DiceDB as a superior choice for building and scaling real-time applications on modern hardware.

DiceDB is engineered for real-time applications that demand low-latency, reactive data handling. It is designed as a drop-in replacement for Redis, with SQL-based reactivity that makes it uniquely suited for modern, scalable systems. In version 0.0.2, we introduce several key features that elevate the performance, flexibility, and usability of the database. Let’s dive into the technical details.

## `QWATCH` Command and Real-Time SQL Querying

### 2.1 Real-Time Subscription with QWATCH

One of the cornerstone features of this release is the introduction of the `QWATCH` command, which allows clients to subscribe to SQL queries and receive real-time updates as the underlying data changes. This feature is essential for applications where the responsiveness to data changes is critical, such as live dashboards, collaborative editing tools, or financial trading platforms.

The `QWATCH` command leverages the power of DiceDB SQL (DSQL) to enable real-time reactive queries. When a client issues a `QWATCH` command, it subscribes to a specific query, and whenever the data related to that query is modified, the client is notified instantly. This approach provides an efficient mechanism for propagating data changes to clients, minimizing latency while ensuring consistent query results.

### SQL Parsing for QWATCH

To support `QWATCH`, DiceDB v0.0.2 includes initial support for SQL parsing within the DSQL engine. This means that complex SQL queries can now be parsed, executed, and subscribed to, allowing for powerful query capabilities with real-time updates.

### Support for `ORDER BY $value` and `LIMIT`

With the addition of `ORDER BY $value`, DiceDB now allows the sorting of query results based on dynamic value fields, adding a layer of flexibility in result set manipulation. Additionally, the inclusion of `LIMIT` clauses allows developers to fine-tune query results, controlling the number of rows returned by a query. These additions make DiceDB more versatile in handling complex queries that require result customization.

## Data Type Support and Operations

### Byte Arrays and JSON Support

In this release, we’ve expanded the range of data types that DiceDB supports by introducing byte arrays and JSON. Byte arrays provide an efficient way to store and manipulate binary data, which is crucial for applications dealing with non-textual content such as multimedia or encrypted data.

JSON support, on the other hand, opens up new possibilities for structuring and querying hierarchical or semi-structured data. JSON has become a standard format for data interchange in modern applications, and DiceDB’s ability to natively handle JSON makes it easier to integrate with various systems.

### JSONPath Querying

To complement JSON support, DiceDB now includes JSONPath querying, which enables complex querying and manipulation of JSON documents. JSONPath is analogous to XPath for XML, allowing developers to extract specific elements from JSON structures. This functionality is critical for applications that need to efficiently parse and manipulate large JSON documents without resorting to external libraries or custom code.

## Performance Optimizations

### Functional Locking Strategy

Concurrency is always a challenge in real-time databases, especially when dealing with high transaction volumes. In DiceDB v0.0.2, we’ve introduced a functional locking strategy that significantly improves concurrency handling. By adopting a more granular locking mechanism, we reduce contention between threads, ensuring that multiple operations can be processed concurrently without sacrificing data integrity.

### Thread Safety with `sync.Map`

Another performance optimization comes from the adoption of `sync.Map` for managing query watchers. This change enhances the thread safety of the system, ensuring that real-time data watching can scale horizontally with minimal synchronization overhead. `sync.Map` is optimized for concurrent read-heavy workloads, making it an ideal choice for the real-time subscription model of `QWATCH`.

## Expanded Redis-Compatible Command Set

In DiceDB v0.0.2, we’ve significantly expanded the set of Redis-compatible commands, making it easier for developers to transition from Redis without requiring substantial changes to their codebases. The newly supported commands include:

- `HELLO`: Allows clients to initiate a connection and exchange protocol information.
- `KEYS`: Returns a list of all keys matching a given pattern.
- `MSET/MGET`: Supports setting and getting multiple key-value pairs in a single command, optimizing batch operations.
- `RENAME/COPY`: Provides functionality to rename or copy keys within the database.
- `TOUCH`: Updates the last access time of keys, useful for managing TTL-based expiration.
- `OBJECT IDLETIME`: Returns the idle time of a key, which is valuable for cache management and performance monitoring.
- `EXPIRETIME`: Retrieves the expiration time for keys, ensuring that applications can programmatically manage key lifecycles.
- `DBSIZE`: Returns the number of keys in the current database, useful for monitoring and capacity planning.

Additionally, we’ve introduced several JSON-specific commands, including:

- `JSON.TYPE`: Returns the type of JSON element (e.g., object, array, string).
- `JSON.CLEAR`: Clears the contents of a JSON object or array.
- `JSON.DEL`: Deletes a specific JSON element from a key.

## Improved Error Handling and Messaging

DiceDB v0.0.2 brings significant improvements to error handling and messaging. We’ve standardized error messages to be more informative, making it easier for developers to debug issues and understand the system's behavior. Whether it's a syntax error in a SQL query or an operational issue with a Redis command, the error messages are now consistent and meaningful, improving the developer experience.

## CI, Testing, and Code Quality Enhancements

### Comprehensive Benchmarks and Test Coverage

In this release, we’ve added comprehensive benchmarks to assess DiceDB’s performance across various workloads. These benchmarks provide critical insights into the database's scalability and latency characteristics under different scenarios. In addition, we’ve improved our test coverage, ensuring that every major feature and edge case is thoroughly validated in both the parser and execution layers.

### Continuous Integration (CI) Testing

As part of our commitment to delivering a robust and stable database, we’ve integrated CI testing into our development workflow. With each push-to-master or pull request, a suite of automated tests is executed, ensuring that no regressions are introduced.

### Linting for Code Quality

To maintain code quality and consistency, we’ve introduced linting across the codebase. Linting ensures adherence to coding standards, improving readability and reducing potential errors in the code.

## Updated Licensing: BSL 1.1

We’ve updated DiceDB’s licensing to the Business Source License (BSL) 1.1, which strikes a balance between open-source principles and sustainable development. This licensing model allows developers to use DiceDB freely in development and small-scale production environments while providing us with the flexibility to offer commercial licenses for larger deployments.

## Summary

DiceDB v0.0.2 goes a step closer to becoming a truly real-time and in-memory database. With powerful new SQL features, enhanced data type support, performance optimizations, and expanded Redis-compatible commands, DiceDB continues to push the boundaries of what’s possible in real-time applications. Our focus on scalability, performance, and developer experience makes DiceDB a compelling choice for modern applications looking to scale without compromising on speed or functionality.
