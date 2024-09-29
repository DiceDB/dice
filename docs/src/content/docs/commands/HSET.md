---
title: HSET
description: Documentation for the DiceDB command HSET
---

The `HSET` command in DiceDB is used to set the value of a field in a hash. If the hash does not exist, a new hash is created. If the field already exists in the hash, the value is updated. This command is useful for managing and storing key-value pairs within a hash data structure.

## Syntax

```
HSET key field value [field value ...]
```

## Parameters

- `key`: The name of the hash.
- `field`: The field within the hash to set the value for.
- `value`: The value to set for the specified field.
- `[field value ...]`: Optional additional field-value pairs to set in the hash.

## Return Value

- `Integer`: The number of fields that were added to the hash, not including fields that were already present and updated.

## Behaviour

When the `HSET` command is executed, the following actions occur:

1. If the specified hash does not exist, a new hash is created.
2. The specified field(s) and value(s) are set in the hash.
3. If a field already exists, its value is updated with the new value provided.
4. The command returns the number of fields that were newly added to the hash.

## Error Handling

The `HSET` command can raise errors in the following scenarios:

1. `WRONGTYPE Operation against a key holding the wrong kind of value`: This error occurs if the key exists but is not a hash.
2. `ERR wrong number of arguments for 'hset' command`: This error occurs if the command is not provided with the correct number of arguments (i.e., an even number of arguments after the key).

## Example Usage

### Basic Example

```DiceDB
HSET user:1000 name "John Doe" age 30
```

This command sets the `name` field to "John Doe" and the `age` field to 30 in the hash stored at `user:1000`. If the hash does not exist, it will be created.

### Multiple Field-Value Pairs

```DiceDB
HSET user:1000 name "John Doe" age 30 email "john.doe@example.com"
```

This command sets the `name`, `age`, and `email` fields in the hash stored at `user:1000`. If the hash does not exist, it will be created.

### Updating Existing Fields

```DiceDB
HSET user:1000 age 31
```

This command updates the `age` field to 31 in the hash stored at `user:1000`. If the `age` field already exists, its value is updated.

## Detailed Example

### Creating a New Hash

```DiceDB
HSET product:2000 name "Laptop" price 999.99 stock 50
```

- `Behaviour`: A new hash is created with the key `product:2000`. The fields `name`, `price`, and `stock` are set with the respective values.
- `Return Value`: `3` (since three new fields were added).

### Updating an Existing Hash

```DiceDB
HSET product:2000 price 899.99 stock 45
```

- `Behaviour`: The `price` and `stock` fields in the hash `product:2000` are updated with the new values.
- `Return Value`: `0` (since no new fields were added, only existing fields were updated).

### Error Handling Example

```DiceDB
SET product:2000 "This is a string"
HSET product:2000 name "Laptop"
```

- `Behaviour`: The `SET` command sets the key `product:2000` to a string value.
- `Error`: The `HSET` command will raise a `WRONGTYPE` error because `product:2000` is not a hash.
