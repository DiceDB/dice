---
title: HSCAN
description: Documentation for the DiceDB command HSCAN
---

The `HSCAN` command is used to incrementally iterate over the fields of a hash stored at a given key. It returns both the next cursor and the matching fields.

## Syntax

```bash
HSCAN key cursor [MATCH pattern] [COUNT count]
```

## Parameters

| Parameter       | Description                                                                                               | Type   | Required |
| --------------- | --------------------------------------------------------------------------------------------------------- | ------ | -------- |
| `key`           | The key of the hash to scan.                                                                              | String | Yes      |
| `cursor`        | The cursor indicating the starting position of the scan.                                                  | String | Yes      |
| `MATCH pattern` | Specifies a pattern to match against the fields. Only the fields that match the pattern will be returned. | String | No       |
| `COUNT count`   | Specifies the maximum number of fields to return.                                                         | String | Yes      |

## Return Value

The `HSCAN` command returns an array containing the next cursor and the matching fields. The format of the returned array is `[nextCursor, [field1, value1, field2, value2, ...]]`.

## Behaviour

- DiceDB checks if the specified key exists.
- If the key exists and is associated with a hash, DiceDB scans the fields of the hash and returns the next cursor and the matching fields.
- If the key does not exist, DiceDB returns an empty array.
- If the key exists but is not associated with a hash, an error is returned.
- If the key exists and all keys have been scanned, cursor is reset to 0.

## Error handling

1. `Wrong type of key`:

   - Error Message: `(error) WRONGTYPE Operation against a key holding the wrong kind of value`
   - Occurs when attempting to use the command on a key that contains a non-hash value.

2. `Invalid integer`:

   - Error Message: `(error) Invalid integer value for COUNT`
   - Occurs if the value provided for the `COUNT` option is not a valid integer or is out of range.

## Examples

### Basic Usage

Creating a hash `myhash` with two fields `field1` and `field2`. Getting `HSCAN` on `myhash` with valid cursors.

```bash
127.0.0.1:7379> HSET myhash field1 "value1" field2 "value2"
1) (integer) 2

127.0.0.1:7379> HSCAN myhash 0
1) "2"
2) 1) "field1"
   2) "value1"
   3) "field2"
   4) "value2"

127.0.0.1:7379> HSCAN myhash 0 MATCH field* COUNT 1
1) "1"
2) 1) "field1"
   2) "value1"

127.0.0.1:7379> HSCAN myhash 1 MATCH field* COUNT 1
1) "0"
2) 1) "field2"
   2) "value2"
```

### Invalid Usage on non-existent key

Getting `HSCAN` on `nonExistentHash`.

```bash
127.0.0.1:7379> HSCAN nonExistentHash 0
1) "0"
2) (empty array)
```

## Notes

- The `HSCAN` command has a time complexity of O(N), where N is the number of keys in the hash.
- The `HSCAN` command is particularly useful for iterating over the fields of a hash in a cursor-based manner, allowing for efficient processing of large hashes.
- The `MATCH` pattern allows for flexible filtering of fields based on their names, making it easy to target specific fields or groups of fields.
- The `COUNT` option enables limiting the number of fields returned, which can be beneficial for performance and memory usage considerations.
- The cursor returned by `HSCAN` can be used to resume the scan from the last position, making it suitable for use cases where the scan needs to be interrupted and resumed later.
