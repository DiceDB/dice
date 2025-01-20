# Reactivity

In traditional databases, we query data when it's needed. This is a classic process but it can become highly inefficient when large number of expensive and redundant queries are frequently fired on the database.

- expensive: query takes a relatively longer time to execute
- redundant: large number of same query is fired on the database

## Leaderboard

Imagine a traditional database that maintains a leaderboard for a competitive game. Here, an asynchronous flow updates the leaderboard data continuously in the database, reflecting new scores or rankings as they come in.

![naive-leaderboard](https://github.com/user-attachments/assets/afeb35c5-c005-4435-ac9b-d681d9993444)

Now consider the users, thousands of clients query this leaderboard every 10 seconds to get the latest rankings. Every query requires the database to compute the current leaderboard standings based on the updated data and then respond to the client with the latest results. Upon receiving the result, each client proceeds to render the leaderboard in its application, ensuring users see the current standings. There are several issues in this setup,

1. Redundant Computation: Since thousands of clients are querying the database every few seconds, the database repeatedly computes the leaderboard, even if there has been little or no change in scores. This results in substantial redundant processing, significantly taxing the database resources.

2. High Load and Latency: The high frequency and volume of queries introduce a considerable load on the database, leading to slower response times.

3. Increased Infrastructure Costs: To handle the high query volume, the system may require powerful hardware and extensive scaling strategies, leading to increased operational costs.

This traditional query model is inefficient for high-read, rapidly updating data structures like leaderboards, where many clients want real-time information. This setup incurs significant resource overhead and diminishes the overall system performance, especially under heavy load.

This is where a reactive database chimes in and shines.

## What is a reactive database?

A reactive database pushes updated query results to clients as soon as the underlying data changes. DiceDB, upon detecting the change in data, automatically re-evaluates the relevant queries and immediately streams the updated result sets to all clients who have subscribed to it.

![reactive leaderboard](https://github.com/user-attachments/assets/08ab40ab-3a34-4d74-9b2f-7d21bbeef73c)

This real-time reactivity ensures that clients always receive the latest data without the need to send repeated queries, dramatically reducing network load, database load, and latency while providing an efficient, seamless experience for data-driven applications.

## DiceDB, a reactive database

DiceDB exemplifies the reactive database model and also focuses on real-time reactivity and efficiency. In DiceDB, clients can set up query subscriptions for specific keys and queries and when a value tied to a subscribed query changes, the updated result set is pushed directly to the subscribed clients. This push model eliminates the need for clients to continuously poll for updates.

### Creating a query subscription

DiceDB makes creating query subscriptions straightforward. For all read-only commands, such as `GET` and `ZRANGE`, DiceDB introduces a `.WATCH` variant, enabling reactive capabilities:

- [GET.WATCH](/commands/getwatch)
- [ZRANGE.WATCH](/commands/zrangewatch)

The `.WATCH` variant signifies a query subscription. For example:

- Executing `GET k1` simply retrieves the value associated with the key, say `v1`.
- Executing `GET.WATCH k1` establishes a query subscription. Whenever the data changes (e.g., `v1` updates to `v2`), DiceDB automatically evaluates the query and streams the updated result (`v2`) over the same database connection.

This makes DiceDB super powerful for building real-time reactive applications. Support for `.WATCH` will soon extend to additional read-only commands, unlocking more reactive use cases.

### But, isn't it same as CDC

No, DiceDB is not the same as CDC. While CDC captures and streams data changes as deltas, DiceDB streams the evaluated output of a query whenever the underlying data changes. This goes beyond mere change notifications, making DiceDB a truly real-time reactive database.

## Efficiency Gains and Cost Reduction

DiceDB’s reactivity model delivers tangible benefits:

- **Reduced Query Load**: Shifting from client polling to server push eliminates redundant queries, particularly in high-read scenarios with multiple clients requesting the same data.

- **Lower Latency**: Updates are pushed instantly upon data changes, ensuring clients receive real-time information with minimal delay, enhancing user experience.

- **Resource Savings**: By processing queries once and streaming results to all subscribers, DiceDB optimizes CPU, memory, and network usage, reducing infrastructure requirements and lowering operational costs.

## In conclusion

Traditional polling-based systems struggle to efficiently deliver real-time data at scale, often incurring high costs, latency, and redundant computations. DiceDB's reactive approach revolutionizes this by pushing updated query results directly to clients as data changes, eliminating inefficiencies while enhancing performance. For applications demanding real-time responsiveness, like leaderboards, DiceDB offers a seamless, cost-effective solution tailored for the modern era of reactive computing.
