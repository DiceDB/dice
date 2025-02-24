---
title: JSON.ARRINSERT
description: The JSON.ARRINSERT command in DiceDB is used to insert one or more JSON values into an array at a specified path before a given index. This command shifts all existing elements in the array to the right, making room for the new elements.
---

The `JSON.ARRINSERT` command allows you to insert one or more JSON values into an array at a specified path before a given index. All existing elements in the array are shifted to the right to make room for the new elements.

This command returns an array of integer replies for each path, where each integer represents the new size of the array after the insertion. If the path does not exist or is not an array, it returns an error.

## Syntax

```bash
JSON.ARRINSERT <key> <path> <index> <value> [value ...]
```

## Parameters

| Parameter | Description                                                                                   | Type    | Required |
| --------- | --------------------------------------------------------------------------------------------- | ------- | -------- |
| `key`     | The name of the key holding the JSON document.                                                | String  | Yes      |
| `path`    | JSONPath pointing to an array within the JSON document.                                       | String  | Yes      |
| `index`   | Position where values will be inserted. Positive for start-to-end, negative for end-to-start. | Integer | Yes      |
| `value`   | JSON value(s) to be inserted into the array.                                                  | JSON    | Yes      |

## Return Values

| Condition                  | Return Value                                                         |
| -------------------------- | -------------------------------------------------------------------- |
| Array successfully updated | (Int) `Integer representing the new array size at each matched path` |
| Key does not exist         | Error: `(error) ERR key does not exist`                              |
| Path is not valid          | Error: `(error) ERR Existing key has wrong Dice type`                |
| Invalid JSON value         | Error: `(error) ERR Path <path> does not exist`                      |
| Invalid index              | Error: `(error) ERR index out of bounds`                             |

## Behaviour

- The command inserts specified JSON values into the array located at the provided path within the document stored under the given key.
- Values are inserted before the specified index, shifting existing elements to the right.
- If the index is out of bounds, it is adjusted to the nearest valid position within the array.
- Errors are returned if the path does not point to an array or if the index provided is not a valid integer.

## Errors

1. `Wrong number of arguments`:

   - Error Message: `(error) ERR wrong number of arguments for JSON.ARRINSERT command`
   - Raised if the number of arguments are less or more than expected.

2. `Key doesn't exist`:

   - Error Message: `(error) ERR key does not exist`
   - Raised if the specified key does not exist in the DiceDB database.

3. `Key has wrong Dice type`:

   - Error Message: `(error) ERR Existing key has wrong Dice type`
   - Raised if thevalue of the specified key doesn't match the specified value in DIceDb

4. `Path doesn't exist`:

   - Error Message: `(error) ERR Path <path> does not exist`
   - Raised if any of the provided JSON values are not valid JSON.

5. `failed to do the operation`:

   - Error Message: `(error) ERR failed to modify JSON data`
   - Raised if the argument specifying path contains an invalide path.

## Example Usage

### Basic Usage

Inserting at a valid index in the root path

```bash
127.0.0.1:7379> JSON.SET a $ '[1,2]'
OK
127.0.0.1:7379> JSON.ARRINSERT a $ 2 3 4 5
(integer) 5
127.0.0.1:7379> JSON.GET a
[1,2,3,4,5]
```

### Using negative index

Inserting at a negative index

```bash
127.0.0.1:7379> JSON.SET a $ '[1,2]'
OK
127.0.0.1:7379> JSON.ARRINSERT a $ -2 3 4 5
(integer) 5
127.0.0.1:7379> JSON.GET a
[3,4,5,1,2]
```

### Having nested arrays as value

Handling nested arrays

```bash
127.0.0.1:7379> JSON.SET b $ '{"name":"tom","score":[10,20],"partner2":{"score":[10,20]}}'
OK
127.0.0.1:7379> JSON.ARRINSERT b $..score 1 5 6 true
(integer) 5
127.0.0.1:7379> JSON.GET b
{"name":"tom","score":[10,5,6,true,20],"partner2":{"score":[10,5,6,true,20]}}
```

### Using out of bounds index

Inserting with an out-of-bounds index

```bash
127.0.0.1:7379> JSON.SET a $ '[1,2]'
OK
127.0.0.1:7379> JSON.ARRINSERT a $ 4 3
ERR index out of bounds
127.0.0.1:7379> JSON.GET a
[1,2]
```

### Invalid index type

Invalid index type

```bash
127.0.0.1:7379> JSON.SET a $ '[1,2]'
OK
127.0.0.1:7379> JSON.ARRINSERT a $ ss 3
ERR value is not an integer or out of range
127.0.0.1:7379> JSON.GET a
[1,2]
```
