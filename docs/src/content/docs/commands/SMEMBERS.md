---
title: SMEMBERS
description: The SMEMBERS command retrieves all members from a set stored at a specified key in DiceDB. Sets are unordered collections of unique strings, and this command provides a way to access all elements within a set for inspection or processing.
---

The `SMEMBERS` command retrieves all members from a set stored at a specified key in DiceDB. Sets are unordered collections of unique strings, and this command provides a way to access all elements within a set for inspection or processing.

## Syntax

```bash
SMEMBERS key
```

## Parameters

| Parameter | Description                                          | Type   | Required |
| --------- | ---------------------------------------------------- | ------ | -------- |
| key       | The key identifying the set to retrieve members from | string | Yes      |

## Return values

| Condition                   | Return Value                            |
| --------------------------- | --------------------------------------- |
| Key exists and is a set     | Array containing all members of the set |
| Key does not exist          | Empty array                             |
| Key exists but is not a set | Error message indicating wrong type     |

## Behaviour

- Checks if the provided key exists in the database
- Verifies if the key is associated with a set data structure
- If key exists and is a set, retrieves all members from the set
- Returns an empty array if the key doesn't exist
- Returns error if key exists but holds a different data type
- The order of elements in the returned array is not guaranteed to be consistent

## Errors

1. `Wrong type of key`:
   - Error Message: `(error) WRONGTYPE Operation against a key holding the wrong kind of value`
   - Occurs when attempting to use SMEMBERS on a key that contains a non-set value

## Example Usage

### Basic Usage

```bash
127.0.0.1:7379> SADD myset "apple" "banana" "cherry"
(integer) 3
127.0.0.1:7379> SMEMBERS myset
1) "apple"
2) "banana"
3) "cherry"
```

### Empty Set

```bash
127.0.0.1:7379> SMEMBERS nonexistentset
(empty array)
```

### Invalid Usage

```bash
127.0.0.1:7379> SMEMBERS
(error) ERR wrong number of arguments for 'smembers' command
```

### Wrong Type Error

```bash
127.0.0.1:7379> SET mystring "hello"
OK
127.0.0.1:7379> SMEMBERS mystring
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```

## Best Practices

1. Always verify the key type before using SMEMBERS to avoid type errors
2. Implement proper error handling in your application for WRONGTYPE errors
3. Use [`EXISTS`](/commands/exists) command to check for key presence before SMEMBERS if key existence is uncertain
   <!--  TODO: uncomment when SSCAN is implemented -->
   <!-- 4. For large sets, consider using SSCAN instead of SMEMBERS to avoid blocking operations -->

## Notes

1. The command returns all members at once, which may impact performance for very large sets
2. The order of returned elements is not guaranteed and may vary between calls
3. Memory usage scales linearly with the size of the set
4. This command has O(N) time complexity where N is the set size