---
title: GETEX
description: The `GETEX` command in DiceDB is used to retrieve the value of a specified key and simultaneously set its expiration time. This command is particularly useful when you want to access the value of a key and ensure that it expires after a certain period, all in a single atomic operation.
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

- If the specified key does not exist, `GETEX` will return nil, and no expiration time will be set.
- If the key exists, `GETEX` retrieves and returns its current value.
- When used with the `EX`, `PX`, `EXAT`, or `PXAT` options, `GETEX` will adjust the key's expiration time according to the specified option.
- Using the `PERSIST` option removes any existing expiration, making the key persist indefinitely.

## Errors

1. `Wrong number of arguments`:

   - Error Message: `(error) ERR wrong number of arguments for 'getex' command`
   - Occurs when the command is executed with an incorrect number of arguments.

2. `Invalid expiration option`:

   - Error Message: `(error) ERR syntax error`
   - Occurs when an unrecognized or incorrect expiration option is specified.

3. `Invalid expiration time`:

   - Error Message: `(error) ERR value is not an integer or out of range`
   - Occurs when the expiration time provided is not a valid integer or timestamp.

## Example Usage

### Retrieve value and set expiration in seconds

This command will return `"Hello"` and set the expiration time of `mykey` to 10 seconds.

```bash
127.0.0.1:7379> SET mykey "Hello"
OK
127.0.0.1:7379> GETEX mykey EX 10
127.0.0.1:7379> "Hello"
```

### Retrieve value and set expiration in milliseconds

This command will return `"Hello"` and set the expiration time of `mykey` to 10,000 milliseconds (10 seconds).

```bash
127.0.0.1:7379> SET mykey "Hello"
OK
127.0.0.1:7379> GETEX mykey PX 10000
127.0.0.1:7379> "Hello"
```

### Retrieve value and set expiration as Unix timestamp in seconds

This command will return `"Hello"` and set the expiration time of `mykey` to the Unix timestamp `1672531199`.

```bash
127.0.0.1:7379> SET mykey "Hello"
OK
127.0.0.1:7379> GETEX mykey EXAT 1672531199
127.0.0.1:7379> "Hello"
```

### Retrieve value and set expiration as Unix timestamp in milliseconds

This command will return `"Hello"` and set the expiration time of `mykey` to the Unix timestamp `1672531199000` milliseconds.

```bash
127.0.0.1:7379> SET mykey "Hello"
OK
127.0.0.1:7379> GETEX mykey PXAT 1672531199000
127.0.0.1:7379> "Hello"
```

### Retrieve value and remove expiration

This command will return `"Hello"` and remove the expiration time of `mykey`, making it persistent.

```bash
127.0.0.1:7379> SET mykey "Hello"
OK
127.0.0.1:7379> EXPIRE mykey 10
OK
127.0.0.1:7379> GETEX mykey PERSIST
127.0.0.1:7379> "Hello"
```

## Notes

- The `GETEX` command is atomic, meaning it ensures that the retrieval of the value and the setting of the expiration time happen as a single, indivisible operation.
- If the key does not exist, the command will return `nil` and no expiration time will be set.
- The `PERSIST` option is useful when you want to make a key persistent again after it has been set to expire.

By understanding and utilizing the `GETEX` command, you can efficiently manage key-value pairs in DiceDB with precise control over their expiration times.

