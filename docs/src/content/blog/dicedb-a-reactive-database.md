---
title: "DiceDB, a reactive database"
description: ""
published_at: 2024-11-12
author: arpit
---

In traditional databases, we query data when it's needed; a process that is simple but can become highly inefficient when large number of expensive and redundant queries are frequently fired on the database. Let's understand this by building a leaderboard.

Imagine a traditional database that maintains a leaderboard for a competitive game or application. Here, an asynchronous flow updates the leaderboard data continuously in the database, reflecting new scores or rankings as they come in.

Now consider the users: thousands of clients query this leaderboard every 10 seconds to get the latest rankings. Every query requires the database to compute the current leaderboard standings based on the updated data and then respond to the client with the latest results. Upon receiving the result, each client proceeds to render the leaderboard in its application, ensuring users see the current standings.

TODO: image.

In this setup, several issues arise:

1. Redundant Computation: Since thousands of clients are querying the database every few seconds, the database repeatedly computes the leaderboard, even if there has been little or no change in scores. This results in substantial redundant processing, significantly taxing the database resources.

2. High Load and Latency: The high frequency and volume of queries introduce a considerable load on the database, leading to latency. The database must allocate resources to process each query individually, slowing down response times as the number of clients scales.

3. Increased Infrastructure Costs: To handle the high query volume, the system may require powerful hardware and extensive scaling strategies, leading to increased operational costs.

This traditional query model is inefficient for high-read, rapidly updating data structures like leaderboards, where many clients want real-time information. This setup incurs significant resource overhead and diminishes the overall system performance, especially under heavy load.

This is where DiceDB, a reactive database chimes in ...

## What is a reactive database?

The concept of a reactive database is simple, a reactive database pushes updated query results to clients as soon as the underlying data changes. DiceDB, upon detecting the change in data, automatically re-evaluates the relevant queries and immediately streams the updated result sets to all clients who have subscribed to it.

TODO Image.

This real-time reactivity ensures that clients always receive the latest data without the need to send repeated queries, dramatically reducing network load and latency while providing an efficient, seamless experience for data-driven applications.

### Reactive Databases are not CDC

One common comparison that is draw to reactive data is of CDC. Dice DB is not same as CDC. CDC is about capturing the delta of changes in data, but in DiceDB the output of the Query is streamed and it is not just a notification about the change. This makes DiceDB a truly real-time reactive database.

## DiceDB: A Reactive Database

**DiceDB** exemplifies the reactive database model. Designed as a Redis-compatible, in-memory, multi-threaded database, it focuses on real-time reactivity and efficiency, enabling it to handle modern application demands for both speed and scale. In DiceDB, clients can set up **query subscriptions** for specific data keys or patterns. When a value tied to a subscribed query changes, the updated result set is **pushed directly to the subscribed clients**. This push model eliminates the need for clients to continuously poll for updates.

### How Query Subscriptions Work in DiceDB

DiceDB's query subscription model is designed around the concept of **data-driven reactivity**:

1. **Subscription Creation**: Clients initiate a subscription by registering a query or specific data key on DiceDB. For example, if an application wants to monitor changes to a leaderboard score, it subscribes to the relevant leaderboard key.
  
2. **Data Monitoring**: DiceDB continuously monitors data for any changes affecting the subscribed queries. Unlike traditional databases, where monitoring is external and client-driven, DiceDB handles this process internally and in real time.

3. **Automatic Query Execution on Data Changes**: When DiceDB detects a change in a subscribed key, it evaluates the relevant query automatically. This step is only triggered by a change, meaning no computational resources are used to check for updates unless necessary.

4. **Result Streaming to Subscribed Clients**: After query evaluation, DiceDB streams the updated result set to each client that subscribed to the query. This eliminates the need for each client to independently query the database, thus reducing network traffic, CPU usage, and latency.

### Real-World Application: Leaderboards

To illustrate the power of DiceDB’s reactive architecture, consider the use case of a leaderboard in a gaming application:

- In a traditional setup, if there are \( n \) users, each user might query the database every 10 seconds to see if the leaderboard has changed. This would result in \( n \) redundant queries every 10 seconds, all asking for the same information.

- With DiceDB, each client can **subscribe** to the leaderboard key. As players’ scores change, DiceDB will automatically evaluate the leaderboard query only when a score update occurs. The updated leaderboard is then **pushed to all subscribed clients**. This approach drastically reduces the number of database operations, as DiceDB processes the query once per update instead of once per client, per interval.

For instance, if the leaderboard updates once every 10 seconds, DiceDB would process the query once, and stream the result to all clients simultaneously. This efficiency means DiceDB can scale to support significantly larger user bases without experiencing the query-based bottlenecks seen in traditional databases.

### Efficiency Gains and Cost Reduction

The reactivity model provided by DiceDB offers several key benefits:

- **Reduced Query Load**: DiceDB eliminates the need for repeated, redundant queries by moving from a client-polling approach to a server-push model. This is especially beneficial in high-read applications where many clients request the same information.
  
- **Lower Latency**: With data being pushed as soon as it changes, clients receive updates faster, leading to lower latency and a more seamless user experience.

- **Resource Savings**: By processing a query once and distributing the result to all subscribers, DiceDB significantly reduces CPU, memory, and network usage. This approach can lead to cost savings by reducing the need for scaling infrastructure to handle repetitive loads.

### Conclusion: DiceDB and the Future of Reactive Databases

DiceDB exemplifies a shift towards a data-reactive paradigm, where the database is not merely a repository for queries but an active participant in maintaining data synchronization with minimal client intervention. Through query subscriptions and real-time data streaming, DiceDB provides a model that can efficiently handle the demands of modern applications, enabling real-time, highly reactive user experiences at a reduced cost.

This transition to reactivity represents a substantial evolution in database technology. As applications increasingly require real-time interactivity and scalable performance, systems like DiceDB are poised to play a critical role in powering the next generation of real-time applications.
