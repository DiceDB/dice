---
title: HRANDFIELD
description: The `HRANDFIELD` command in DiceDB is used to return one or more random fields from a hash stored at a specified key. It can also return the values associated with those fields if specified.
---

The `HRANDFIELD` command in DiceDB is used to return one or more random fields from a hash stored at a specified key. It can also return the values associated with those fields if specified.

## Syntax

```bash
HRANDFIELD key [count [WITHVALUES]]
```

## Parameters

| Parameter    | Description                                                             | Type    | Required |
| ------------ | ----------------------------------------------------------------------- | ------- | -------- |
| `key`        | The key of the hash from which random fields are to be returned         | String  | Yes      |
| `count`      | The number of random fields to retrieve. If negative, allows repetition | Integer | No       |
| `WITHVALUES` | Option to include the values associated with the returned fields        | Flag    | No       |

## Return values

| Condition                             | Return Value                                                          |
| ------------------------------------- | --------------------------------------------------------------------- |
| Key exists and count is not specified | `(String)`                                                            |
| Key exists and count is specified     | Array of random fields (or field-value pairs if `WITHVALUES` is used) |
| Key does not exist                    | `nil`                                                                 |
| Key exists but is not a hash          | `error`                                                               |

## Behaviour

- DiceDB checks if the specified `key` exists.
- If the key does not exist, the command returns `nil`.
- If the key exists but is not a hash, an error is returned.
- If the key does not have any fields, an empty array is returned.
- If no `count` parameter is passed, it is defaulted to `1`
- If the `count` or `WITHVALUES` parameters are passed,they are checked for typechecks and syntax errors.
- If the `count` parameter is negative, the command allows repeated fields in the result.
- The command will return the random field(s) based on the specified `count`.
- If the `WITHVALUES` option is provided, the command returns the fields along with their associated values.

## Errors

1. `Wrong type of value or key`:
   - Error Message: `(error) ERROR WRONGTYPE Operation against a key holding the wrong kind of value`
   - Occurs when attempting to use the command on a key that is not a hash.
2. `Invalid count type`:
   - Error Message: `(error) ERROR value is not an integer or out of range`
   - Occurs when a non-integer count parameter is passed.
3. `Invalid number of arguments`
   - Error Message: `(error) ERROR wrong number of arguments for 'hrandfield' command`
   - Occurs when an invalid number of arguments are passed to the command.

## Example Usage

### Basic Usage

Executing `HRANDFIELD` on a key without any parameters

```bash
127.0.0.1:7379> HSET keys field1 value1 field2 value2 field3 value3
(integer) 3
127.0.0.1:7379> HRANDFIELD keys
"field1"
```

### Usage with `count` parameter

Executing `HRANDFIELD` on a key with a `count` parameter of 2

```bash
127.0.0.1:7379> HRANDFIELD keys 2
1) "field2"
2) "field1"
```

### Usage with `WITHVALUES` parameter

Executing `HRANDFIELD` with the `WITHVALUES` parameter

```bash
127.0.0.1:7379> HRANDFIELD keys 2 WITHVALUES
1) "field2"
2) "value2"
3) "field1"
4) "value1"
```

### Invalid key usage

Executing `hrandfield` on a non-hash key

```bash
127.0.0.1:7379> SET key "not a hash"
OK
127.0.0.1:7379> HRANDFIELD key
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```

### Invalid count parameter

Non-integer value passed as `count`

```bash
127.0.0.1:7379> HRANDFIELD keys hello
(error) ERROR value is not an integer or out of range
```

### Invalid number of arguments

Passing invalid number of arguments to the `hrandfield` command

```bash
127.0.0.1:7379> HRANDFIELD
(error) ERR wrong number of arguments for 'hrandfield' command
```

## Best Practices

- `Check Key Existence`: Before using `HRANDFIELD`, ensure that the hash key exists to avoid unnecessary errors. You can use the [`HEXISTS`](/commands/hexists) command to verify the presence of the key.

## Alternatives

- [`HKEYS`](/commands/hkeys): Use `HKEYS` to retrieve all fields in a hash. If you need to work with all fields rather than a random selection, this command provides a complete view.
- `Sampling with HSCAN`: For larger hashes, consider using [`HSCAN`](/commands/hscan) to iterate through the fields and randomly select a subset. This approach can help manage memory usage when dealing with large datasets.

## Notes

- The `HRANDFIELD` command is useful for scenarios where random selection from hash fields is required, such as in games, lotteries, or randomized surveys.
- The command can return multiple fields at once, allowing for efficient random sampling without the need for multiple calls. This can be particularly advantageous when working with larger hashes.