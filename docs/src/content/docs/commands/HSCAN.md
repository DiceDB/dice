---
title: HSCAN
description: Documentation for the DiceDB command HSCAN
---

The `HSCAN` command is used to incrementally iterate over the fields of a hash stored at a given key. It returns both the next cursor and the matching fields.

## Syntax

```
HSCAN key cursor [MATCH pattern] [COUNT count]
```

## Parameters

- `key`: The key of the hash to scan.
- `cursor`: The cursor indicating the starting position of the scan.
- `MATCH pattern` (optional): Specifies a pattern to match against the fields. Only the fields that match the pattern will be returned.
- `COUNT count` (optional): Specifies the maximum number of fields to return.

## Return Value

The `HSCAN` command returns an array containing the next cursor and the matching fields. The format of the returned array is `[nextCursor, [field1, value1, field2, value2, ...]]`.

## Behaviour
When the `HSCAN` command is executed:

1. DiceDB checks if the specified key exists.
2. If the key exists and is associated with a hash, DiceDB scans the fields of the hash and returns the next cursor and the matching fields.
3. If the key does not exist, DiceDB returns an empty array.
4. If the key exists but is not associated with a hash, an error is returned.
5. If the key exists and all keys have been scanned, cursor is reset to 0.

## Error handling
The `HSCAN` command can raise the following errors:

- `WRONGTYPE Operation against a key holding the wrong kind of value`: This error occurs if the specified key exists but is not associated with a hash. For example, if the key is associated with a string, list, set, or any other data type, this error will be raised.
- `Invalid integer value for COUNT`: This error occurs if the value provided for the `COUNT` option is not a valid integer or is out of range.



## Examples

```bash
> HSET myhash field1 "value1" field2 "value2"
1) (integer) 2

> HSCAN myhash 0
1) "2"
2) 1) "field1"
   2) "value1"
   3) "field2"
   4) "value2"

> HSCAN myhash 0 MATCH field* COUNT 1
1) "1"
2) 1) "field1"
   2) "value1"

> HSCAN myhash 1 MATCH field* COUNT 1
1) "0"
2) 1) "field2"
   2) "value2"
```


## Additional Notes

- The `HSCAN` command has a time complexity of O(N), where N is the number of keys in the hash. This is in contrast to Redis, which implements `HSCAN` in O(1) time complexity by maintaining a cursor.
- The `HSCAN` command is particularly useful for iterating over the fields of a hash in a cursor-based manner, allowing for efficient processing of large hashes.
- The `MATCH` pattern allows for flexible filtering of fields based on their names, making it easy to target specific fields or groups of fields.
- The `COUNT` option enables limiting the number of fields returned, which can be beneficial for performance and memory usage considerations.
- The cursor returned by `HSCAN` can be used to resume the scan from the last position, making it suitable for use cases where the scan needs to be interrupted and resumed later.
