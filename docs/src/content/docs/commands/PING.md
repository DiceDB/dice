---
title: PING
description: The `PING` command in DiceDB is used to test the connection between the client and the DiceDB server. It's a crucial command for ensuring that the DiceDB server is reachable and functional. It is often utilized to check the health of the server and ensure that it is responsive.
---

The `PING` command in DiceDB is used to test the connection between the client and the DiceDB server. It's a crucial command for ensuring that the DiceDB server is reachable and functional. It is often utilized to check the health of the server and ensure that it is responsive.

## Syntax

The basic syntax of the `PING` command is as follows:

```bash
PING [message]
```

- If `[message]` is provided, it will be echoed back in the response.
- If `[message]` is not provided, default response is `PONG`.

## Parameters

| Parameter       | Description                                      | Type    | Required |
|-----------------|--------------------------------------------------|---------|----------|
| `message`       | `message` echoed in the response                 | String  | No       |

## Return values

| Condition                                      | Return Value                                      |
|------------------------------------------------|---------------------------------------------------|
| No `message` provided                          | `PONG`                                            |
| `message` provided                             | `message`                                         |

## Behaviour

When the `PING` command is fired:

- If no `message` is given, it sends back a `PONG`.
- If a `message` is provided, it sends back the same message in the response.
- If `message` provided is non-string value, the value is internally coerced to string and echoed as response. 

This helps clients determine if the server is up and responsive. Essentially, it acts as a keep-alive or heartbeat mechanism for clients to validate their connection with the DiceDB server.

## Errors

1. `Syntax Error`: If the syntax is incorrect, such as including unexpected additional parameters, an error will be raised:

   - `Error message`: `(error) ERR wrong number of arguments for 'ping' command`
   - `Scenario`: If more than one argument is provided.

   ```bash
   127.0.0.1:7379> PING "Message 1" "Message 2"
   (error) ERR wrong number of arguments for 'ping' command
   ```

## Example Usage

### Basic Usage 

`PING` DiceDB server without a message and the server echoes with `PONG`.

```bash
127.0.0.1:7379> PING
PONG
```

### Pinging the server with a message

DiceDB server is pinged with `Hello, DiceDB!` and the server echoes with `Hello, DiceDB!`.

```bash
127.0.0.1:7379>  PING "Hello, DiceDB!"
"Hello, DiceDB!"
```

### Pinging the server with int message

DiceDB server is pinged with `1234` and the server echoes with `"1234"` coerced to string internally.

```bash
127.0.0.1:7379>  PING 1234
"1234"
```

### Pinging the server with list message

DiceDB server is pinged with `[1234]` and the server echoes with `"[1234]"` coerced to string internally.

```bash
127.0.0.1:7379>  PING [1234]
"[1234]"
```
