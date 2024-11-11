---
title: SET
description: The `SET` command in DiceDB is used to set the value of a key. If the key already holds a value, it is overwritten, regardless of its type. This is one of the most fundamental operations in DiceDB as it allows for both creating and updating key-value pairs.
---

<!-- description in 2 to 3 sentences, following is an example -->

The `SET` command in DiceDB is used to set the value of a key. If the key already holds a value, it is overwritten, regardless of its type. This is one of the most fundamental operations in DiceDB as it allows for both creating and updating key-value pairs.

## Syntax

```bash
SET key value [EX seconds | PX milliseconds | EXAT unix-time-seconds | PXAT unix-time-milliseconds | KEEPTTL] [NX | XX]
```

<!-- If the command have subcommands please mention but do not consider them as arguments -->
<!-- please mention them in subcommands section and create their individual documents -->

## Parameters

<!-- please add all parameters, small description, type and required, see example for SET command-->

| Parameter | Description                                                               | Type    | Required |
| --------- | ------------------------------------------------------------------------- | ------- | -------- |
| `key`     | The name of the key to be set.                                            | String  | Yes      |
| `value`   | The value to be set for the key.                                          | String  | Yes      |
| `EX`      | Set the specified expire time, in seconds.                                | Integer | No       |
| `EXAT`    | Set the specified Unix time at which the key will expire, in seconds      | Integer | No       |
| `PX`      | Set the specified expire time, in milliseconds.                           | Integer | No       |
| `PXAT`    | Set the specified Unix time at which the key will expire, in milliseconds | Integer | No       |
| `NX`      | Only set the key if it does not already exist.                            | None    | No       |
| `XX`      | Only set the key if it already exists.                                    | None    | No       |
| `KEEPTTL` | Retain the time-to-live associated with the key.                          | None    | No       |

## Return values

<!-- add all scenarios, see below example for SET -->

| Condition                              | Return Value |
| -------------------------------------- | ------------ |
| if key is set successfully             | `OK`         |
| If `NX` is used and key already exists | `nil`        |
| If `XX` is used and key doesn't exist  | `nil`        |

## Behaviour

<!-- How does the command execute goes here, kind of explaining the underlying algorithm -->
<!-- see below example for SET command -->
<!-- Please modify for the command by going through the code -->

- If the specified key already exists, the `SET` command will overwrite the existing key-value pair with the new value unless the `NX` option is provided.
- If the `NX` option is present, the command will set the key only if it does not already exist. If the key exists, no operation is performed and `nil` is returned.
- If the `XX` option is present, the command will set the key only if it already exists. If the key does not exist, no operation is performed and `nil` is returned.
- Using the `EX`, `EXAT`, `PX` or `PXAT` options together with `KEEPTTL` is not allowed and will result in an error.
- When provided, `EX` sets the expiry time in seconds and `PX` sets the expiry time in milliseconds.
- The `KEEPTTL` option ensures that the key's existing TTL is retained.

## Errors

<!-- sample errors, please update for commands-->
<!-- please add all the errors here -->
<!-- incase of a dynamic error message, feel free to use variable names -->

1. `Wrong type of value or key`:

   - Error Message: `(error) WRONGTYPE Operation against a key holding the wrong kind of value`
   - Occurs when attempting to use the command on a key that contains a non-string value.

2. `Invalid syntax or conflicting options`:

   - Error Message: `(error) ERR syntax error`
   - Occurs if the command's syntax is incorrect, such as incompatible options like `EX` and `KEEPTTL` used together, or a missing required parameter.

3. `Non-integer value for `EX`or`PX\`\`:

   - Error Message: `(error) ERR value is not an integer or out of range`
   - Occurs when the expiration time provided is not a valid integer.

## Example Usage

### Basic Usage

<!-- examples here are for set, please update them for the command -->

Setting a key `foo` with the value `bar`

```bash
127.0.0.1:7379> SET foo bar
OK
```

<!-- Please use detailed scenarios and edges cases if possible -->

### Using expiration time (in seconds)

Setting a key `foo` with the value `bar` to expire in 10 seconds

```bash
127.0.0.1:7379> SET foo bar EX 10
OK
```

### Using expiration time (in milliseconds)

Setting a key `foo` with the value `bar` to expire in 10000 milliseconds (10 seconds)

```bash
127.0.0.1:7379> SET foo bar PX 10000
OK
```

### Setting only if key does not exist

Setting a key `foo` only if it does not already exist

```bash
127.0.0.1:7379> SET foo bar NX
```

`Response`:

- If the key does not exist: `OK`
- If the key exists: `nil`

### Setting only if key exists

Setting a key `foo` only if it exists

```bash
127.0.0.1:7379> SET foo bar XX
```

`Response`:

- If the key exists: `OK`
- If the key does not exist: `nil`

### Retaining existing TTL

Setting a key `foo` with a value `bar` and retaining existing TTL

```bash
127.0.0.1:7379> SET foo bar KEEPTTL
OK
```

### Invalid usage

Trying to set key `foo` with both `EX` and `KEEPTTL` will result in an error

```bash
127.0.0.1:7379> SET foo bar EX 10 KEEPTTL
(error) ERR syntax error
```

<!-- Optional: Used when additional information is to conveyed to users -->
<!-- For example warnings about usage ex: Keys * -->
<!-- OR alternatives of the commands -->
<!-- Or perhaps deprecation warning -->
<!-- anything related to the command which cannot be shared in other sections -->

<!-- Optional -->

## Best Practices

<!-- below example from Keys command -->

- `Avoid in Production`: Due to its potential to slow down the server, avoid using the `KEYS` command in a production environment. Instead, consider using the `SCAN` command, which is more efficient for large keyspaces.
- `Use Specific Patterns`: When using the `KEYS` command, try to use the most specific pattern possible to minimize the number of keys returned and reduce the load on the server.

<!-- Optional -->

## Alternatives

<!-- below example from keys command -->

- `SCAN`: The `SCAN` command is a cursor-based iterator that allows you to incrementally iterate over the keyspace without blocking the server. It is a more efficient alternative to `KEYS` for large datasets.

<!-- Optional -->

## Notes

<!-- below example from json.get command -->

- JSONPath expressions allow you to navigate and retrieve specific parts of a JSON document. Ensure that your JSONPath expressions are correctly formatted to avoid errors.

By understanding the `JSON.GET` command, you can efficiently retrieve JSON data stored in your DiceDB database, enabling you to build powerful and flexible applications that leverage the capabilities of DiceDB.

<!-- Optional -->

## Subcommands

<!-- if the command you are working on has subcommands -->
<!-- please mention them here and add links to the pages -->
<!-- please see below example for COMMAND docs -->
<!-- follow below bullet structure -->

- **subcommand**: Optional. Available subcommands include:
  - `COUNT` : Returns the total number of commands in the DiceDB server.
  - `GETKEYS` : Returns the keys from the provided command and arguments.
  - `LIST` : Returns the list of all the commands in the DiceDB server.
  - `INFO` : Returns details about the specified commands.
  - `HELP` : Displays the help section for `COMMAND`, providing information about each available subcommand.
