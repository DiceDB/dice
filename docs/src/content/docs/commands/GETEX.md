---
title: GETEX
description: Documentation for the DiceDB command GETEX
---

The `GETEX` command in DiceDB is used to retrieve the value of a specified key and simultaneously set its expiration time. This command is particularly useful when you want to access the value of a key and ensure that it expires after a certain period, all in a single atomic operation.

## Syntax

```bash
GETEX key [EX seconds|PX milliseconds|EXAT timestamp|PXAT milliseconds-timestamp|PERSIST]
```

## Parameters

| Parameter | Description                                                               | Type    | Required |
|-----------|---------------------------------------------------------------------------|---------|----------|
| `key`     | The key whose value you want to retrieve and set an expiration for.       | String  | Yes      |
| `EX`      | Set the expiration time in seconds.                                       | Integer | No       |
| `EXAT`    | Set the expiration time as a Unix timestamp in seconds.                   | Integer | No       |
| `PX`      | Set the expiration time in milliseconds.                                  | Integer | No       |
| `PXAT`    | Set the expiration time as a Unix timestamp in milliseconds.              | Integer | No       |
| `PERSIST` | Remove the existing expiration time, making the key persistent.           | None    | No       |

## Return Values

| Condition                                         | Return Value                                      |
|---------------------------------------------------|---------------------------------------------------|
| The specified key exists and holds a value        | The value stored at the key                       |
| The specified key does not exist                  | `nil`                                             |
| Syntax or specified constraints are invalid       | error                                             |

## Behaviour

When the `GETEX` command is executed, it performs the following actions:

1. Retrieves the value of the specified key.
2. Sets the expiration time for the key based on the provided option (`EX`, `PX`, `EXAT`, `PXAT`, or `PERSIST`).
3. Returns the value of the key.

If the key does not exist, the command will return `nil` and no expiration time will be set.

## Errors

The `GETEX` command can raise errors in the following scenarios:

1. `Wrong number of arguments`: If the command is not provided with the correct number of arguments, it will return an error.
   - Error message: `(error) ERR wrong number of arguments for 'getex' command`
2. `Invalid expiration option`: If an invalid expiration option is provided, it will return an error.
   - Error message: `(error) ERR syntax error`
3. `Invalid expiration time`: If the expiration time is not a valid integer or timestamp, it will return an error.
   - Error message: `(error) ERR value is not an integer or out of range`

## Example Usage

### Example 1: Retrieve value and set expiration in seconds


```bash
127.0.0.1:7379> SET mykey "Hello"
OK
127.0.0.1:7379> GETEX mykey EX 10
127.0.0.1:7379> "Hello"
```

- This command will return `"Hello"` and set the expiration time of `mykey` to 10 seconds.

### Example 2: Retrieve value and set expiration in milliseconds

```bash
127.0.0.1:7379> SET mykey "Hello"
OK
127.0.0.1:7379> GETEX mykey PX 10000
127.0.0.1:7379> "Hello"
```

- This command will return `"Hello"` and set the expiration time of `mykey` to 10,000 milliseconds (10 seconds).

### Example 3: Retrieve value and set expiration as Unix timestamp in seconds

```bash
127.0.0.1:7379> SET mykey "Hello"
OK
127.0.0.1:7379> GETEX mykey EXAT 1672531199
127.0.0.1:7379> "Hello"
```

- This command will return `"Hello"` and set the expiration time of `mykey` to the Unix timestamp `1672531199`.

### Example 4: Retrieve value and set expiration as Unix timestamp in milliseconds

```bash
127.0.0.1:7379> SET mykey "Hello"
OK
127.0.0.1:7379> GETEX mykey PXAT 1672531199000
127.0.0.1:7379> "Hello"
```

- This command will return `"Hello"` and set the expiration time of `mykey` to the Unix timestamp `1672531199000` milliseconds.

### Example 5: Retrieve value and remove expiration

```bash
127.0.0.1:7379> SET mykey "Hello"
OK
127.0.0.1:7379> EXPIRE mykey 10
OK
127.0.0.1:7379> GETEX mykey PERSIST
127.0.0.1:7379> "Hello"
```

- This command will return `"Hello"` and remove the expiration time of `mykey`, making it persistent.

## Notes

- The `GETEX` command is atomic, meaning it ensures that the retrieval of the value and the setting of the expiration time happen as a single, indivisible operation.
- If the key does not exist, the command will return `nil` and no expiration time will be set.
- The `PERSIST` option is useful when you want to make a key persistent again after it has been set to expire.

By understanding and utilizing the `GETEX` command, you can efficiently manage key-value pairs in DiceDB with precise control over their expiration times.

