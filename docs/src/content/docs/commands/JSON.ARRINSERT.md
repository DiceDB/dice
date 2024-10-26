---
title: JSON.ARRINSERT
description: Documentation for the DiceDB command JSON.ARRINSERT
---

The `JSON.ARRINSERT` command allows you to insert one or more JSON values into an array at a specified path before a given index. All existing elements in the array are shifted to the right to make room for the new elements.

This command returns an array of integer replies for each path, where each integer represents the new size of the array after the insertion. If the path does not exist or is not an array, it returns an error.

## Syntax

```plaintext
JSON.ARRINSERT <key> <path> <index> <value> [value ...]
```

## Parameters
- `key`: (String) The key in the Redis database that holds the JSON object.
- `path`: (String) the path is a valid JSON path pointing to an array inside the JSON object.
- `index`: The index is a position at which to insert the values. Positive values count from the beginning of the array, and negative values count from the end.
- `value`: The value is one or more JSON values to insert into the array.


## Return Value
- On success, it returns an `Integer` representing the new size of the array at each matched path.
- On failure, an `error` message is returned.

## Behaviour

When the JSON.ARRINSERT command is executed, it inserts one or more specified JSON values into an array located at the given path within the document stored under the specified key. The values are inserted before the provided index, shifting existing elements to the right to make room for the new elements. If the index is out of bounds, it is adjusted to the nearest valid position (either the beginning or end of the array).
If the path does not point to an array or does not exist, an error is returned. If the specified index is not a valid integer, an error will also be raised.

## Error Handling

- `(error) ERR wrong number of arguments for JSON.ARRINSERT command`: Raised if the number of arguments are less or more than expected.
- `(error) ERR key does not exist`: Raised if the specified key does not exist in the DiceDB database.
- `(error) ERR Existing key has wrong Dice type`:Raised if thevalue of the specified key doesn't match the specified value in DIceDb
- `(error) ERR Path <path> does not exist`: Raised if any of the provided JSON values are not valid JSON.
- `(error) ERR failed to modify JSON data`: Raised if the argument specifying path contains an invalide path.

## Example Usage

### Example 1: Inserting at a valid index in the root path

```plaintext
127.0.0.1:6379> JSON.SET a $ '[1,2]'
OK
127.0.0.1:6379> JSON.ARRINSERT a $ 2 3 4 5
(integer) 5
127.0.0.1:6379> JSON.GET a
[1,2,3,4,5]
```

### Example 2: Inserting at a negative index

```plaintext
127.0.0.1:6379> JSON.SET a $ '[1,2]'
OK
127.0.0.1:6379> JSON.ARRINSERT a $ -2 3 4 5
(integer) 5
127.0.0.1:6379> JSON.GET a
[3,4,5,1,2]
```

### Example 3: Handling nested arrays
```plaintext
127.0.0.1:6379> JSON.SET b $ '{"name":"tom","score":[10,20],"partner2":{"score":[10,20]}}'
OK
127.0.0.1:6379> JSON.ARRINSERT b $..score 1 5 6 true
(integer) 5
127.0.0.1:6379> JSON.GET b
{"name":"tom","score":[10,5,6,true,20],"partner2":{"score":[10,5,6,true,20]}}
```

### Example 4: Inserting with an out-of-bounds index

```plaintext
127.0.0.1:6379> JSON.SET a $ '[1,2]'
OK
127.0.0.1:6379> JSON.ARRINSERT a $ 4 3
ERR index out of bounds
127.0.0.1:6379> JSON.GET a
[1,2]
```

### Example 5: Invalid index type

```plaintext
127.0.0.1:6379> JSON.SET a $ '[1,2]'
OK
127.0.0.1:6379> JSON.ARRINSERT a $ ss 3
ERR value is not an integer or out of range
127.0.0.1:6379> JSON.GET a
[1,2]
```

## Notes

- Ensure that the DiceDBJSON module is loaded in your DiceDB instance to use the `JSON.ARRINSERT` command.
- JSONPath expressions are used to navigate and specify the location within the JSON document. Familiarity with JSONPath syntax is beneficial for effective use of this command.

By following this documentation, users can effectively utilize the `JSON.ARRINSERT` command to manipulate JSON arrays within DiceDB.

