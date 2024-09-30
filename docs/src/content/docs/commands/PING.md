---
title: PING
description: The `PING` command in DiceDB is used to test the connection between the client and the DiceDB server. It's a crucial command for ensuring that the DiceDB server is reachable and functional. It is often utilized to check the health of the server and ensure that it is responsive.
---

The `PING` command in DiceDB is used to test the connection between the client and the DiceDB server. It's a crucial command for ensuring that the DiceDB server is reachable and functional. It is often utilized to check the health of the server and ensure that it is responsive.

## Parameters

The `PING` command can optionally take a single argument which is a string.

- `message`: Optional. A string that you want the DiceDB server to return as a response. If a message is provided, the DiceDB server will respond with the same message.

### Syntax

The basic syntax of the `PING` command is as follows:

```
PING [message]
```

- If `[message]` is provided, it will be echoed back in the response.
- If `[message]` is not provided, default response is `PONG`.

## Return Value

The return value varies based on the presence of the optional `[message]` parameter.

- `No message provided`: If no message is provided, `PING` will return a simple string reply `PONG`.
- `Message provided`: If a message is provided, `PING` will return a simple string that echoes the given message.

## Example Usage

### Example 1: Pinging the server without a message

```bash
127.0.0.1:7379> PING
PONG
```

### Example 2: Pinging the server with a message

```bash
127.0.0.1:7379>  PING "Hello, DiceDB!"
"Hello, DiceDB!"
```

## Behaviour

When the `PING` command is fired:

1. If no message is given, it sends back a `PONG`.
2. If a message is provided, it sends back the same message in the response.

This helps clients determine if the server is up and responsive. Essentially, it acts as a keep-alive or heartbeat mechanism for clients to validate their connection with the DiceDB server.

## Error Handling

### Potential Errors

1. `Syntax Error`: If the syntax is incorrect, such as including unexpected additional parameters, an error will be raised:

   - `Error message`: `(error) ERR wrong number of arguments for 'ping' command`
   - `Scenario`: If more than one argument is provided.

   ```bash
   > PING "Message 1" "Message 2"
   (error) ERR wrong number of arguments for 'ping' command
   ```

2. `Data Type Error`: If the argument provided is not a string (for example, if a list or other data type is provided), a type error will be raised.

   - `Error message`: This specific type of error handling is internal, and improper types are generally implicitly converted or rejected by the protocol, raising basic syntax errors or allowing them based on DiceDB internal type coercion rules.

## Additional Notes

- The `PING` command works similarly in both standalone and clustered DiceDB environments.
- It is typically sent periodically by clients to ensure the connection is still active.
- The `PING` command does not modify the data within the DiceDB server or affect any ongoing transactions.
