---
title: JSON.ARRTRIM
description: The `JSON.ARRTRIM` command in DiceDB is used to trim a JSON array to only include elements within a specified range. It removes elements outside of this range, modifying the array in place. This command is part of the DiceDBJSON module, which provides native JSON capabilities in DiceDB.
---

The `JSON.ARRTRIM` command in DiceDB modifies a JSON array by removing elements outside a specified range, effectively trimming it to contain only desired elements. This in-place operation is part of the DiceDBJSON module, which enables seamless handling and manipulation of JSON data within DiceDB.

## Syntax

```bash
JSON.ARRTRIM <key> <path> <start> <stop>
```

## Parameters

| Parameter | Description                                                            | Type    | Required |
| --------- | ---------------------------------------------------------------------- | ------- | -------- |
| `key`     | The name of the key holding the JSON document.                         | String  | Yes      |
| `path`    | JSONPath pointing to an array within the JSON document.                | String  | Yes      |
| `index`   | The index of the first element to retain in the array (0-based index). | Integer | Yes      |
| `value`   | The index of the last element to retain in the array (inclusive).      | Integer | Yes      |

## Return Values

| Condition                 | Return Value                                                                     |
| ------------------------- | -------------------------------------------------------------------------------- |
| Success                   | (Int) `Integer representing the length of the array after the append operation.` |
| Key does not exist        | Error: `(error) ERR key does not exist`                                          |
| Wrong number of arguments | Error: `(error) ERR wrong number of arguments for JSON.ARRTRIM command`          |
| Key has wrong type        | Error: `(error) ERR Existing key has wrong Dice type`                            |
| Path doesn't exist        | Error: `(error) ERR Path <path> does not exist`                                  |
| Operation failed          | Error: `(error) ERR failed to modify JSON data`                                  |

## Behaviour

- The command trims the JSON array located at the specified path within the document stored under the given key to only include elements within the specified start and stop indices.
- Elements outside of the specified range are removed from the array, modifying it in place.
- If the start or stop indices are out of bounds, they are adjusted to the nearest valid index within the array.
- An error is returned if the path does not point to an array or if the key does not exist in the DiceDB database.

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

### Basic usage

Trimming an array to a specific range

```bash
127.0.0.1:7379> JSON.SET b $ '[1, 2, 3, 4, 5]'
"OK"
127.0.0.1:7379> JSON.ARRTRIM b $ 1 3
1) "3"
127.0.0.1:7379> JSON.GET b
"[2,3,4]"
127.0.0.1:7379>
```

### Trim array down to 1 element

Trimming an array to a single element

```bash
127.0.0.1:7379> JSON.SET a $ '[0,1,2]'
"OK"
127.0.0.1:7379> JSON.ARRTRIM a $ 1 1
"1"
127.0.0.1:7379> JSON.GET a
"[1]"
127.0.0.1:7379>
```

### Trimming array with out of bound index

Trimming an array with out-of-bounds indices

```bash
127.0.0.1:7379> JSON.SET c $ '[1, 2, 3, 4, 5]'
"OK"
127.0.0.1:7379> JSON.ARRTRIM c $ -10 10
1) "5"
127.0.0.1:7379> JSON.GET c
"[1,2,3,4,5]"
```

### When path doesn't exist

Error when path does not exist

```bash
127.0.0.1:7379> JSON.SET d $ '[1, 2, 3, 4, 5]'
"OK"
127.0.0.1:7379> JSON.ARRTRIM d . -10 10
(error) ERROR Path '.' does not exist
```

### When key doesn't exist

Error when key does not exist

```bash
127.0.0.1:7379> JSON.SET a $ '[1, 2, 3, 4, 5]'
"OK"
127.0.0.1:7379> JSON.ARRTRIM aa . -10 10
(error) ERROR key does not exist

```
