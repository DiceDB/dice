---
title: Q.WATCH
description: The `Q.WATCH` command is a novel feature designed to provide real-time updates to clients based on changes in underlying data.
---

The `Q.WATCH` command is a novel feature designed to provide real-time updates to clients based on changes in underlying
data. It operates similarly to the `SUBSCRIBE` command but focuses on SQL-like queries over data structures. Whenever
data modifications affect the query's results, the updated result set is pushed to the subscribed client. This
eliminates the need for clients to constantly poll for changes.

This command uniquely positions DiceDB as the easiest and most intuitive way to build real-time reactive applications like leaderboards.

## Protocol Support

| Protocol  | Supported |
| --------- | --------- |
| TCP-RESP  | ✅        |
| HTTP      | ✅        |
| WebSocket | ✅        |

## Syntax

```bash
Q.WATCH <dsql-query>
```

## Parameters

| Parameter    | Description                                                                         | Type   | Required |
| ------------ | ----------------------------------------------------------------------------------- | ------ | -------- |
| `dsql-query` | A SQL-like query specifying the data to be monitored and operation to be performed. | String | Yes      |

## DSQL Query

A SQL-like query specifying the data to be monitored and operation to be performed.

### Syntax

```sql
SELECT $key, $value
WHERE condition
ORDER BY field [ASC | DESC] LIMIT n
```

### SELECT Clause

Specifies the fields to be returned, `$key`, `$value`, and `$value.<attr>`.

### Special Column Names

- `$key`: Refers to the key of the key-value pair
- `$value`: Refers to the value of the key-value pair

### WHERE Clause

**Optional clause** for filtering results based on conditions applied to the key or value.

Supported conditions:

- Comparison operators: `=`, `<`, `>`, `<=`, `>=`, `!=`
- Logical operators: `AND`, `OR`
- Parentheses for grouping conditions
- Comparison between key and value fields
- LIKE clause: `LIKE 'pattern'`

### LIKE Clause

**Optional clause** within WHERE to specify the key pattern and supports the `%` operator from SQL.

### ORDER BY Clause

**Optional clause** for sorting results.

### LIMIT Clause

**Optional clause** to limit the number of results.

## Return Value

| Condition             | Return Value                                               |
| --------------------- | ---------------------------------------------------------- |
| Command is successful | A subscription confirmation message similar to `SUBSCRIBE` |
| DSQL Query is invalid | error                                                      |

## Behavior

- The provided query is parsed to extract the `SELECT`, `WHERE`, `LIKE`, `ORDER BY`, and `LIMIT` clauses.
- The client establishes a subscription to the specified query.
- The initial result set based on the current data is sent to the client.
- DiceDB continuously monitors the data sources specified in the LIKE clause.
- Whenever data changes that might affect the query result, the query is reevaluated.

## Errors

1. `Missing query`

   - Error Message: `(error) ERROR wrong number of arguments for 'q.watch' command`
   - Occurs if no DSQL Query is provided.

2. `Invalid query`:

   - Error Message: `(error) ERROR error parsing SQL statement: syntax error at position <n>`
   - Occurs if the provided query is malformed or has unsupported clauses.

3. `Max number of subscriptions reached`:

   - Error Message: `(error) ERROR could not perform this operation on a key that doesn't exist`
   - Occurs if the maximum number of allowed subscriptions is exceeded.

## Example Usage

### Basic Usage

Let's explore a practical example of using the `Q.WATCH` command to create a real-time leaderboard for a game match,
including filtering with a `WHERE` clause.

```bash
127.0.0.1:7379> Q.WATCH "SELECT $key, $value WHERE $key like 'match:100:*' AND $value > 10 ORDER BY $value DESC LIMIT 3"
q.watch    from SELECT $key, $value WHERE $key like 'match:100:*' AND $value > 10 ORDER BY $value asc: []
```

This query does the following:

- Monitors all keys matching the pattern `match:100:*`
- Filters results to include only scores greater than 10
- Orders the results by their values in descending order
- Limits the results to the top 3 entries

### Scenario

Imagine we're tracking player scores in a game match with ID 100. Each player's score is stored in a key formatted as
`match:100:user:<userID>`.

Let's walk through a series of updates and see how the `Q.WATCH` command responds. Please note
that the response will be RESP encoded and parsing will be handled by the SDK that you are using.

1. Initial state (empty leaderboard): `[]`

2. Player 0 scores 5 points:

   ```bash
   127.0.0.1:7379> SET match:100:user:0 5
   ```

   Subscription does not return response as `value < 10`.

3. Player 1 scores 15 points:

   ```bash
   127.0.0.1:7379> SET match:100:user:1 15
   ```

   Q.WATCH Response:

   ```bash
   q.watch    from SELECT $key, $value WHERE $key like 'match:100:*' and $value > 100 ORDER BY $value asc: `[["match:100:user:1", "15"]]`
   ```

4. Player 2 scores 20 points:

   ```bash
   127.0.0.1:7379> SET match:100:user:2 20
   ```

   Q.WATCH Response:

   ```bash
   q.watch    from SELECT $key, $value WHERE $key like 'match:100:*' and $value > 100 ORDER BY $value asc: `[["match:100:user:2", "20"], ["match:100:user:1", "15"]]`
   ```

5. Player 3 scores 12 points:

   ```bash
   127.0.0.1:7379> SET match:100:user:3 12
   ```

   Q.WATCH Response:

   ```bash
   q.watch    from SELECT $key, $value WHERE $key like 'match:100:*' and $value > 100 ORDER BY $value asc: `[["match:100:user:2", "20"], ["match:100:user:1", "15"], ["match:100:user:3", "12"]]`
   ```

6. Player 4 scores 25 points:

   ```bash
   127.0.0.1:7379> SET match:100:user:4 25
   ```

   Q.WATCH Response:

   ```bash
   q.watch    from SELECT $key, $value WHERE $key like 'match:100:*' and $value > 100 ORDER BY $value asc: `[["match:100:user:4", "25"], ["match:100:user:2", "20"], ["match:100:user:1", "15"]]`
   ```

7. Player 0 improves their score to 30:

   ```bash
   127.0.0.1:7379> SET match:100:user:0 30
   ```

   Q.WATCH Response:

   ```bash
   q.watch    from SELECT $key, $value WHERE $key like 'match:100:*' and $value > 100 ORDER BY $value asc: `[["match:100:user:0", "30"], ["match:100:user:4", "25"], ["match:100:user:2", "20"]]`
   ```

This example demonstrates how `Q.WATCH` provides real-time updates as the leaderboard changes, always keeping clients
informed of the top 3 scores above 10, without the need for constant polling.

## Best Practices

1. Use specific key patterns in the LIKE clause wherever possible to limit the scope of the query.
2. Keep `WHERE` conditions as simple as possible for better performance.
3. Ensure type consistency in comparisons to avoid runtime errors, this is best done in your application layer.
