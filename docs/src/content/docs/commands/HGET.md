---
title: HGET
description: Documentation for the DiceDB command HGET
---

The `HGET` command in DiceDB is used to retrieve the value associated with a specified field within a hash stored at a given key. If the key or the field does not exist, the command returns a `nil` value.

## Syntax

```
HGET key field
```

## Parameters

- `key`: The key of the hash from which the field's value is to be retrieved. This is a string.
- `field`: The field within the hash whose value is to be retrieved. This is also a string.

## Return Value

- `String`: The value associated with the specified field within the hash.
- `nil`: If the key does not exist or the field is not present in the hash.

## Behaviour

When the `HGET` command is executed, DiceDB performs the following steps:

1. It checks if the key exists in the database.
2. If the key exists and is of type hash, it then checks if the specified field exists within the hash.
3. If the field exists, it retrieves and returns the value associated with the field.
4. If the key does not exist or the field is not present in the hash, it returns `nil`.

## Error Handling

- `WRONGTYPE Operation against a key holding the wrong kind of value`: This error is raised if the key exists but is not of type hash. DiceDB expects the key to be associated with a hash data structure for the `HGET` command to work correctly.

## Example Usage

### Example 1: Retrieving a value from a hash

```DiceDB
HSET user:1000 name "John Doe"
HSET user:1000 age "30"
HGET user:1000 name
```

`Output:`

```
"John Doe"
```

### Example 2: Field does not exist

```DiceDB
HGET user:1000 email
```

`Output:`

```
(nil)
```

### Example 3: Key does not exist

```DiceDB
HGET user:2000 name
```

`Output:`

```
(nil)
```

### Example 4: Key is not a hash

```DiceDB
SET user:3000 "Not a hash"
HGET user:3000 name
```

`Output:`

```
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```

## Notes

- The `HGET` command is a read-only command and does not modify the hash or any other data in the DiceDB database.
- It is a constant time operation, O(1), meaning it executes in the same amount of time regardless of the size of the hash.

By understanding the `HGET` command, you can efficiently retrieve values from hashes stored in your DiceDB database, ensuring that your application can access the necessary data quickly and reliably.

