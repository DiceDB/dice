---
title: QWATCH
description: The `QWATCH` command is a novel feature designed to provide real-time updates to clients based on changes in underlying data.
---

The `QWATCH` command is a novel feature designed to provide real-time updates to clients based on changes in underlying data. It operates similarly to the `SUBSCRIBE` command but focuses on SQL-like queries over data structures. Whenever data modifications affect the query's results, the updated result set is pushed to the subscribed client. This eliminates the need for clients to constantly poll for changes.

This command is what makes DiceDB different from Redis and uniquely positions it as the easiest and most intuitive way to build real-time reactive applications like leaderboards.

## Protocol Support

| Protocol | Supported |
| -------- | --------- |
| TCP-RESP | ✅        |
| HTTP     | ✅        |
| WebSocket| ❌        |

## Syntax

```
QWATCH dsql-query
```

## Parameters

- dsql-query: A SQL-like query specifying the data to be monitored and operation to be performed. The query syntax is as follows:

  - `SELECT`: Specifies the fields to be returned, `$key`, `$value`, and `$value.<attr>`.
  - `WHERE`: Optional clause for filtering results based on conditions.
  - `LIKE`: Optional clause within WHERE to specify the key pattern and supports the `%` operator from SQL.
  - `ORDER BY`: Optional clause for sorting results.
  - `LIMIT`: Optional clause to limit the number of results.

## Query Syntax

```sql
SELECT $key, $value
WHERE condition
ORDER BY field [ASC | DESC] LIMIT n
```

Special column names:

- `$key`: Refers to the key of the key-value pair
- `$value`: Refers to the value of the key-value pair

### WHERE Clause

The WHERE clause allows you to filter results based on conditions applied to the key or value. It supports various comparison operators and can include complex logical expressions.
It also supports the LIKE clause as one of the conditions.

Supported features:

- Comparison operators: `=`, `<`, `>`, `<=`, `>=`, `!=`
- Logical operators: `AND`, `OR`
- Parentheses for grouping conditions
- Comparison between key and value fields
- LIKE clause: `like 'pattern'`

## Return Value

- A subscription confirmation message similar to `SUBSCRIBE`.

## Behavior

1. `Query Parsing`: The provided query is parsed to extract the `SELECT`, `WHERE`, `LIKE`, `ORDER BY`, and `LIMIT` clauses.
2. `Subscription Establishment`: The client establishes a subscription to the specified query.
3. `Initial Result`: The initial result set based on the current data is sent to the client.
4. `Data Change Monitoring`: DiceDB continuously monitors the data sources specified in the LIKE clause.
5. `Query Reevaluation`: Whenever data changes that might affect the query result, the query is reevaluated.
6. `Result Push`: If the reevaluated result differs from the previous result, the new result set is pushed to the subscribed client.

## Example Usage with Query Flow

Let's explore a practical example of using the `QWATCH` command to create a real-time leaderboard for a game match, including filtering with a `WHERE` clause.

### Query

```bash
127.0.0.1:7379> QWATCH "SELECT $key, $value WHERE $key like 'match:100:*' AND $value > 10 ORDER BY $value DESC LIMIT 3"
```

This query does the following:

- Monitors all keys matching the pattern `match:100:*`
- Filters results to include only scores greater than 10
- Orders the results by their values in descending order
- Limits the results to the top 3 entries

### Scenario

Imagine we're tracking player scores in a game match with ID 100. Each player's score is stored in a key formatted as `match:100:user:<userID>`.

Let's walk through a series of updates and see how the `QWATCH` command responds. Please note
that the response will be RESP encoded and parsing will be handled by the SDK that you are using.

1. Initial state (empty leaderboard):
   QWATCH response: `[] (empty array)`

2. Player 0 scores 5 points:

   ```bash
   127.0.0.1:7379> SET match:100:user:0 5
   ```

   QWATCH response: `[] (no change, score <= 10)`

3. Player 1 scores 15 points:

   ```bash
   127.0.0.1:7379> SET match:100:user:1 15
   ```

   QWATCH response: `[["match:100:user:1", "15"]]`

4. Player 2 scores 20 points:

   ```bash
   127.0.0.1:7379> SET match:100:user:2 20
   ```

   QWATCH response: `[["match:100:user:2", "20"], ["match:100:user:1", "15"]]`

5. Player 3 scores 12 points:

   ```bash
   127.0.0.1:7379> SET match:100:user:3 12
   ```

   QWATCH response: `[["match:100:user:2", "20"], ["match:100:user:1", "15"], ["match:100:user:3", "12"]]`

6. Player 4 scores 25 points:

   ```bash
   127.0.0.1:7379> SET match:100:user:4 25
   ```

   QWATCH response: `[["match:100:user:4", "25"], ["match:100:user:2", "20"], ["match:100:user:1", "15"]]`

7. Player 0 improves their score to 30:
   ```bash
   127.0.0.1:7379> SET match:100:user:0 30
   ```
   QWATCH response: `[["match:100:user:0", "30"], ["match:100:user:4", "25"], ["match:100:user:2", "20"]]`

This example demonstrates how `QWATCH` provides real-time updates as the leaderboard changes, always keeping clients informed of the top 3 scores above 10, without the need for constant polling.

## Error Handling

- `ERR invalid query`: If the provided query is malformed or unsupported.
- `ERR syntax error`: If the query syntax is incorrect.
- `ERR unknown command`: If the `QWATCH` command is not implemented.
- `ERR max number of subscriptions reached`: If the maximum number of allowed subscriptions is exceeded.

## Best Practices

1. Use specific key patterns in the LIKE clause wherever possible to limit the scope of the query.
2. Keep `WHERE` conditions as simple as possible for better performance.
3. Ensure type consistency in comparisons to avoid runtime errors, this is best done in your application layer.
