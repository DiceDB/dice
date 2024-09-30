---
title: HGETALL
description: Documentation for the DiceDB command HGETALL
---

The `HGETALL` command in DiceDB is used to retrieve all the fields and values of a hash stored at a specified key. This command is particularly useful when you need to fetch the entire hash in one go, rather than fetching individual fields one by one.

## Syntax

```
HGETALL key
```

## Parameters

- `key`: The key of the hash from which you want to retrieve all fields and values. This parameter is mandatory.

## Return Value

The `HGETALL` command returns an array of strings in the form of field-value pairs. If the key does not exist, an empty array is returned.

## Behaviour

When the `HGETALL` command is executed:

1. DiceDB checks if the specified key exists.
2. If the key exists and is of type hash, DiceDB retrieves all the fields and their corresponding values.
3. If the key does not exist, DiceDB returns an empty array.
4. If the key exists but is not of type hash, an error is returned.

## Error Handling

The `HGETALL` command can raise the following errors:

- `WRONGTYPE Operation against a key holding the wrong kind of value`: This error is raised if the specified key exists but is not a hash. DiceDB expects the key to be associated with a hash data structure, and if it is associated with a different data type (e.g., string, list, set, etc.), this error will be triggered.

## Example Usage

### Example 1: Retrieving all fields and values from an existing hash

```DiceDB
HSET user:1000 name "John Doe" age "30" country "USA"
HGETALL user:1000
```

`Output:`

```
1) "name"
2) "John Doe"
3) "age"
4) "30"
5) "country"
6) "USA"
```

### Example 2: Retrieving from a non-existing key

```DiceDB
HGETALL user:2000
```

`Output:`

```
(empty array)
```

### Example 3: Error when key is not a hash

```DiceDB
SET user:3000 "This is a string"
HGETALL user:3000
```

`Output:`

```
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```

## Notes

- The `HGETALL` command is a blocking command and can be slow if the hash contains a large number of fields.
- It is recommended to use this command with caution in production environments where performance is critical, especially if the hash is expected to grow large.

## Best Practices

- Ensure that the key you are querying with `HGETALL` is indeed a hash to avoid unnecessary errors.
- Consider using other hash commands like `HSCAN` for large hashes to iterate over fields and values incrementally.

By understanding the `HGETALL` command and its behavior, you can effectively manage and retrieve data from hash structures in DiceDB.

