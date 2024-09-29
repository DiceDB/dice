---
title: SMEMBERS
description: Documentation for the DiceDB command SMEMBERS
---

The `SMEMBERS` command in DiceDB is used to retrieve all the members of a set stored at a specified key. Sets in DiceDB are unordered collections of unique strings. This command is useful for obtaining the entire set of elements for further processing or inspection.

## Syntax

```
SMEMBERS key
```

## Parameters

- `key`: The key of the set from which you want to retrieve all members. This parameter is required and must be a valid string.

## Return Value

The `SMEMBERS` command returns an array of all the members in the set stored at the specified key. If the key does not exist, an empty array is returned.

## Behaviour

When the `SMEMBERS` command is executed:

1. DiceDB checks if the key exists.
2. If the key exists and is of type set, DiceDB retrieves all the members of the set.
3. If the key does not exist, DiceDB returns an empty array.
4. If the key exists but is not of type set, an error is returned.

## Error Handling

The `SMEMBERS` command can raise the following errors:

- `WRONGTYPE Operation against a key holding the wrong kind of value`: This error occurs if the key exists but is not associated with a set. DiceDB expects the key to be of type set, and if it is not, this error is raised.

## Example Usage

### Example 1: Retrieving Members from an Existing Set

```DiceDB
SADD myset "apple" "banana" "cherry"
SMEMBERS myset
```

`Output:`

```
1) "apple"
2) "banana"
3) "cherry"
```

### Example 2: Retrieving Members from a Non-Existent Set

```DiceDB
SMEMBERS nonexistentset
```

`Output:`

```
(empty array)
```

### Example 3: Error Case - Key Exists but is Not a Set

```DiceDB
SET mystring "hello"
SMEMBERS mystring
```

`Output:`

```
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```

## Notes

- The order of elements returned by `SMEMBERS` is not guaranteed to be consistent. Sets in DiceDB are unordered collections, so the order of elements may vary.
- For large sets, consider using the `SSCAN` command to iterate over the set incrementally to avoid blocking the DiceDB server for a long time.

## Best Practices

- Always ensure that the key you are querying with `SMEMBERS` is of type set to avoid type errors.
- Use `EXISTS` command to check if a key exists before using `SMEMBERS` if you are unsure about the key's existence.
- For very large sets, prefer using `SSCAN` to avoid performance issues.

By following this documentation, you should be able to effectively use the `SMEMBERS` command in DiceDB to retrieve all members of a set.

