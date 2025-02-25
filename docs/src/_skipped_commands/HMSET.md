---
title: HMSET
description: The `HMSET` command in DiceDB is used to set multiple field-value pairs in a hash at once. If the hash does not exist, a new hash is created. This command is efficient for setting multiple fields at once within a hash data structure.
sidebar:
  badge:
    text: Deprecated
    variant: danger
---

The `HMSET` command in DiceDB is used to set multiple field-value pairs in a hash at once. If the hash does not exist, a new hash is created. This command is efficient for setting multiple fields at once within a hash data structure.

## Syntax

```bash
HMSET key field value [field value ...]
```

## Parameters

| Parameter           | Description                                      | Type   | Required |
| ------------------- | ------------------------------------------------ | ------ | -------- |
| `key`               | The name of the hash.                            | String | Yes      |
| `field`             | The field within the hash to set the value for.  | String | Yes      |
| `value`             | The value to set for the specified field.        | String | Yes      |
| `[field value ...]` | Additional field-value pairs to set in the hash. | String | No       |

## Return Values

| Condition                        | Return Value                                                                |
| -------------------------------- | --------------------------------------------------------------------------- |
| A new field added                | `OK`                                                                        |
| Existing field updated           | `OK`                                                                        |
| Multiple fields added            | `OK`                                                                        |
| Non-hash type or wrong data type | `(error) WRONGTYPE Operation against a key holding the wrong kind of value` |
| Incorrect Argument Count         | `(error) ERR wrong number of arguments for 'hmset' command`                 |

## Behaviour

When the `HMSET` command is executed, the following actions occur:

- If the specified hash does not exist, a new hash is created.
- The specified fields and values are set in the hash.
- If any field already exists, its value is updated with the new value provided.
- The command returns `OK` to indicate successful execution.

## Errors

The `HMSET` command can raise errors in the following scenarios:

1. `Non-hash type or wrong data type`:

   - Error Message: `(error) WRONGTYPE Operation against a key holding the wrong kind of value`
   - Occurs if `key` holds a non-hash data structure.

2. `Incorrect Argument Count`:

   - Error Message: `(error) ERR wrong number of arguments for 'hmset' command`
   - Occurs if the command is not provided with the correct number of arguments (i.e., an even number of arguments after the key).

## Example Usage

### Creating a New Hash with Multiple Fields

```bash
127.0.0.1:7379> HMSET product:4000 name "Tablet" price 299.99 stock 30
OK
```

### Updating an Existing Hash with Multiple Fields

```bash
127.0.0.1:7379> HMSET product:4000 price 279.99 stock 25
OK
```

### Invalid Usage

Trying to set fields in a key that is not a hash.

```bash
127.0.0.1:7379> SET product:4000 "This is a string"
OK
127.0.0.1:7379> HMSET product:4000 name "Tablet"
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```

Wrong Number of Arguments for HMSET Command

```bash
127.0.0.1:7379> HMSET product:4000
(error) ERR wrong number of arguments for 'hmset' command

127.0.0.1:7379> HMSET product:4000 name
(error) ERR wrong number of arguments for 'hmset' command
```

## Best Practices

- **Use HMSET for Batch Updates**: Utilize `HMSET` when you need to set multiple fields at once in a hash to reduce command overhead and improve performance.
