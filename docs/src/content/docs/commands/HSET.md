---
title: HSET
description: The `HSET` command in DiceDB is used to set the value of a field in a hash. If the hash does not exist, a new hash is created. If the field already exists in the hash, the value is updated. This command is useful for managing and storing key-value pairs within a hash data structure.
---

The `HSET` command in DiceDB is used to set the value of a field in a hash. If the hash does not exist, a new hash is created. If the field already exists in the hash, the value is updated. This command is useful for managing and storing key-value pairs within a hash data structure.

## Syntax

```bash
HSET key field value [field value ...]
```

## Parameters

| Parameter           | Description                                                   | Type    | Required |
|---------------------|---------------------------------------------------------------|---------|----------|
| `key`               | The name of the hash.                                         | String  | Yes      |
| `field`             | The field within the hash to set the value for.               | String  | Yes      |
| `value`             | The value to set for the specified field.                     | String  | Yes      |
| `[field value ...]` | Optional additional field-value pairs to set in the hash.     | String  | No       |

## Return Values

| Condition                                   | Return Value                                                                |
|---------------------------------------------|-----------------------------------------------------------------------------|
| A new field added                           | `1`                                                                         |
| Existing field updated                      | `0`                                                                         |
| Multiple fields added                       | `Integer` (count of new fields)                                             |
| Wrong data type                             | `(error) WRONGTYPE Operation against a key holding the wrong kind of value` |
| Incorrect Argument Count                    | `(error) ERR wrong number of arguments for 'hset' command`                  |

## Behaviour

When the `HSET` command is executed, the following actions occur:

- If the specified hash does not exist, a new hash is created.
- The specified field(s) and value(s) are set in the hash.
- If a field already exists, its value is updated with the new value provided.
- The command returns the number of fields that were newly added to the hash.

## Errors

The `HSET` command can raise errors in the following scenarios:

1. `Non-hash type or wrong data type`:

   - Error Message: `(error) WRONGTYPE Operation against a key holding the wrong kind of value`
   - Occurs if `key` holds a non-hash data structure.

2. `Incorrect Argument Count`:

   - Error Message: `(error) ERR wrong number of arguments for 'hset' command`
   - occurs if the command is not provided with the correct number of arguments (i.e., an even number of arguments after the key).

## Example Usage

### Basic Usage

#### Creating a New Hash

```bash
127.0.0.1:7379> HSET product:2000 name "Laptop" price 999.99 stock 50
3
```
- **Behaviour**: A new hash is created with the key `product:2000`. The fields `name`, `price`, and `stock` are set with the respective values.
- **Return Value**: `3` (since three new fields were added).

#### Updating an Existing Hash

Updating existing fields in a hash `product:2000`.

```bash
127.0.0.1:7379> HSET product:2000 price 899.99 stock 45
```

- **Behavior**: The `price` and `stock` fields in the hash `product:2000` are updated with the new values.
- **Return Value**: `0` (since no new fields were added, only existing fields were updated).

#### Setting Multiple Field-Value Pairs

Setting multiple fields in a hash `user:1000`.

```bash
127.0.0.1:7379> HSET user:1000 name "John Doe" age 30 email "john.doe@example.com"
```

- **Behavior**: This command sets the `name`, `age`, and `email` fields in the hash stored at `user:1000`. If the hash does not exist, it will be created.
- **Return Value**: `3` (if all three fields were added).

#### Updating Existing Fields

Updating a field in an existing hash `user:1000`.

```bash
127.0.0.1:7379> HSET user:1000 age 31
```

- **Behavior**: This command updates the `age` field to 31 in the hash stored at `user:1000`. If the `age` field already exists, its value is updated.
- **Return Value**: `0` (if the field was already present and only updated).

### Invalid Usage

Trying to set a field in a key that is not a hash.

```bash
127.0.0.1:7379> SET product:2000 "This is a string"
OK
127.0.0.1:7379> HSET product:2000 name "Laptop"
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```

- **Behavior**: The `SET` command sets the key `product:2000` to a string value.
- **Error**: The `HSET` command will raise a `WRONGTYPE` error because `product:2000` is not a hash.

Wrong Number of Arguments for HSET Command

```bash
127.0.0.1:7379> HSET product:2000
(error) ERR wrong number of arguments for 'hset' command

127.0.0.1:7379> HSET product:2000 name
(error) ERR wrong number of arguments for 'hset' command
```
- **Behavior**: The `HSET` command requires atleast three arguments: the key, the field name, and the field value.
- **Error**: The command fails because it requires at least one field-value pair in addition to the key. If insufficient arguments are provided, DiceDB raises an error indicating that the number of arguments is incorrect.

### Best Practices

- **Check for Existence**: Before updating fields, consider checking if the hash exists to avoid unnecessary updates.

## Notes

- `HSET` can also accept multiple field-value pairs, making it efficient for adding or updating multiple fields in a single command.
By understanding the `HSET` command, you can effectively update or expand hashes in your DiceDB database, allowing for quick modifications and optimizations when handling key-value pairs within hashes.
