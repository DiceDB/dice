---
title: SLEEP
description: The `SLEEP` command in DiceDB is used to pause the execution of the current client connection for a specified number of seconds. This command is primarily useful for testing purposes, such as simulating network latency or delaying operations to observe the behavior of other commands in a controlled environment.
---

The `SLEEP` command in DiceDB is used to pause the execution of the current client connection for a specified number of seconds. This command is primarily useful for testing purposes, such as simulating network latency or delaying operations to observe the behavior of other commands in a controlled environment.

## Parameters

### Syntax

```
SLEEP seconds
```

### Parameters

- `seconds`: (Required) A floating-point number representing the number of seconds to sleep. This value can be an integer or a decimal, allowing for sub-second precision.

## Return Value

The `SLEEP` command returns a simple string reply of "OK" after the specified sleep duration has elapsed.

## Behaviour

When the `SLEEP` command is executed, the following behavior is observed:

1. The client connection that issued the `SLEEP` command will be paused for the specified duration.
2. During this sleep period, the client will not be able to send or receive any other commands.
3. Other clients connected to the DiceDB server will not be affected and can continue to execute commands normally.
4. After the sleep duration has elapsed, the client will receive an "OK" response and can resume normal operations.

## Error Handling

The `SLEEP` command can raise errors under the following conditions:

1. `Invalid Number of Arguments`:

   - `Error Message`: `ERR wrong number of arguments for 'sleep' command`
   - `Condition`: This error occurs if the `SLEEP` command is called without the required `seconds` parameter or with more than one parameter.

2. `Invalid Parameter Type`:

   - `Error Message`: `ERR value is not a valid float`
   - `Condition`: This error occurs if the `seconds` parameter is not a valid floating-point number.

## Example Usage

### Example 1: Basic Usage

Pause the client for 5 seconds.

```shell
127.0.0.1:7379> SLEEP 5
OK
```

### Example 2: Sub-second Precision

Pause the client for 1.5 seconds.

```shell
127.0.0.1:7379> SLEEP 1.5
OK
```

### Example 3: Error Handling - Missing Parameter

Attempt to call `SLEEP` without specifying the `seconds` parameter.

```shell
127.0.0.1:7379> SLEEP
(error) ERR wrong number of arguments for 'sleep' command
```

### Example 4: Error Handling - Invalid Parameter Type

Attempt to call `SLEEP` with a non-numeric parameter.

```shell
127.0.0.1:7379> SLEEP abc
(error) ERR value is not a valid float
```

## Notes

- The `SLEEP` command is a blocking operation for the client that issues it. It does not affect the server's ability to handle other clients.
- This command is useful for testing and debugging purposes but should be used with caution in production environments to avoid unintended delays.
