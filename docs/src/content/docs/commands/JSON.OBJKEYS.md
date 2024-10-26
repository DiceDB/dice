---
title: JSON.OBJKEYS
description: Documentation for the DiceDB command JSON.OBJKEYS
---

The `JSON.OBJKEYS` command command in DiceDB retrieves the keys of a JSON object located at a specified path within the document stored under the given key. This command is useful when you want to list the fields within a JSON object stored in a database.

## Syntax

```plaintext
JSON.OBJKEYS key [path]
```

## Parameters
- `key`: (String) The key under which the JSON document is stored.
- `path`: (String) The JSONPath expression that specifies the location within the JSON document where the array is located.

## Return Value

- `[]String`: It returns an array of strings. This array contains the keys present within the JSON object at the specified path. Each entry in the array represents a key found in the JSON object.

## Behaviour

- Root Path (Default): If no path is provided, JSON.OBJKEYS retrieves keys from the root object of the JSON document.
- Path Validation: If the specified path does not point to an object (e.g., if it points to a scalar, array, or does not exist), the command returns nil for that path.
- Non-existing Key: If the specified key does not exist in the database, an error is returned.
- Invalid JSON Path: If the provided JSONPath expression is invalid, an error message with the details of the parse error is returned.


## Error Handling

- `(error) ERR wrong number of arguments for JSON.OBJKEYS command`: Raised if the number of arguments are less or more than expected.
- `(error) ERR could not perform this operation on a key that doesn't exist`: Raised if the specified key does not exist in the DiceDB database.
- `(error) ERR Existing key has wrong Dice type`:Raised if thevalue of the specified key doesn't match the specified value in DIceDb
- `(error) ERR WRONGTYPE Operation against a key holding the wrong kind of value`: Raised if an operation attempted on a key with an incompatible type.

## Example Usage

### Example 1: Retrieving Keys of the Root Object

```plaintext
127.0.0.1:6379> JSON.SET a $ '{"name": "Alice", "age": 30, "address": {"city": "Wonderland", "zipcode": "12345"}}'
"OK"
127.0.0.1:6379> JSON.OBJKEYS a $
1) "name"
2) "age"
3) "address"
```

### Example 2: Retrieving Keys of a Nested Object

```plaintext
127.0.0.1:6379> JSON.SET b $ '{"name": "Alice", "partner": {"name": "Bob", "age": 28}}'
"OK"
127.0.0.1:6379> JSON.OBJKEYS b $.partner
1) "name"
2) "age"
```

### Example 3: Error When Path Points to a Non-Object Type

```plaintext
127.0.0.1:6379> JSON.SET c $ '{"name": "Alice", "age": 30}'
"OK"
127.0.0.1:6379> JSON.OBJKEYS c $.age
(nil)
```

### Example 4: Error When Path Does Not Exist

```plaintext
127.0.0.1:6379> JSON.SET d $ '{"name": "Alice", "address": {"city": "Wonderland"}}'
"OK"
127.0.0.1:6379> JSON.OBJKEYS d $.nonexistentPath
(empty list or set)
```

### Example 5: Error When Key Does Not Exist

```plaintext
127.0.0.1:6379> JSON.OBJKEYS nonexistent_key $
(error) ERROR could not perform this operation on a key that doesn't exist
```

## Notes

- Ensure that the DiceDBJSON module is loaded in your DiceDB instance to use the `JSON.OBJKEYS` command.
- JSONPath expressions are used to navigate and specify the location within the JSON document. Familiarity with JSONPath syntax is beneficial for effective use of this command.

By following this documentation, users can effectively utilize the `JSON.OBJKEYS` command to manipulate JSON arrays within DiceDB.

