---
title: JSON.ARRTRIM
description: Documentation for the DiceDB command JSON.ARRTRIM
---

The `JSON.ARRTRIM` command in DiceDB is used to trim a JSON array to only include elements within a specified range. It removes elements outside of this range, modifying the array in place. This command is part of the DiceDBJSON module, which provides native JSON capabilities in DiceDB.

## Syntax

```plaintext
JSON.ARRTRIM <key> <path> <start> <stop>
```

## Parameters
- `key`: (String) The key under which the JSON document is stored.
- `path`: (String) The JSONPath expression that specifies the location within the JSON document where the array is located.
- `start`: The index of the first element to retain in the array (0-based index).
- `stop`: The index of the last element to retain in the array (inclusive).

## Return Value

- `Integer`: The length of the array after the append operation.

## Behaviour

When the `JSON.ARRTRIM` command is executed, the JSON array located at the specified path within the document stored under the given key is trimmed to retain only the elements within the provided range. The array is modified in place, removing all elements outside of the specified start and stop indices. If the start or stop indices are out of bounds, they are adjusted to the nearest valid index. If the path does not point to an array or does not exist, an error is returned.

## Error Handling

- `(error) ERR wrong number of arguments for JSON.ARRTRIM command`: Raised if the number of arguments are less or more than expected.
- `(error) ERR key does not exist`: Raised if the specified key does not exist in the DiceDB database.
- `(error) ERR Existing key has wrong Dice type`:Raised if thevalue of the specified key doesn't match the specified value in DIceDb
- `(error) ERR Path <path> does not exist`: Raised if any of the provided JSON values are not valid JSON.
- `(error) ERR failed to modify JSON data`: Raised if the argument specifying path contains an invalide path.

## Example Usage

### Example 1: Trimming an array to a single element

```plaintext
127.0.0.1:6379> JSON.SET a $ '[0,1,2]'
"OK"
127.0.0.1:6379> JSON.ARRTRIM a $ 1 1
"1"
127.0.0.1:6379> JSON.GET a
"[1]"
127.0.0.1:6379>
```

### Example 2: Trimming an array to a specific range

```plaintext
127.0.0.1:6379> JSON.SET b $ '[1, 2, 3, 4, 5]'
"OK"
127.0.0.1:6379> JSON.ARRTRIM b $ 1 3
1) "3"
127.0.0.1:6379> JSON.GET b
"[2,3,4]"
127.0.0.1:6379>
```

### Example 3: Trimming an array with out-of-bounds indices

```plaintext
127.0.0.1:6379> JSON.SET c $ '[1, 2, 3, 4, 5]'
"OK"
127.0.0.1:6379> JSON.ARRTRIM c $ -10 10
1) "5"
127.0.0.1:6379> JSON.GET c
"[1,2,3,4,5]"
```

### Example 4: Error when path does not exist

```plaintext
127.0.0.1:6379> JSON.SET d $ '[1, 2, 3, 4, 5]'
"OK"
127.0.0.1:6379> JSON.ARRTRIM d . -10 10
(error) ERROR Path '.' does not exist
```

### Example 5: Error when key does not exist

```plaintext
127.0.0.1:6379> JSON.SET a $ '[1, 2, 3, 4, 5]'
"OK"
127.0.0.1:6379> JSON.ARRTRIM aa . -10 10
(error) ERROR key does not exist

```

## Notes

- Ensure that the DiceDBJSON module is loaded in your DiceDB instance to use the `JSON.ARRTRIM` command.
- JSONPath expressions are used to navigate and specify the location within the JSON document. Familiarity with JSONPath syntax is beneficial for effective use of this command.

By following this documentation, users can effectively utilize the `JSON.ARRTRIM` command to manipulate JSON arrays within DiceDB.

