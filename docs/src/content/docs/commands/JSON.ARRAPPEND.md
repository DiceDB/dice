---
title: JSON.ARRAPPEND
description: Documentation for the DiceDB command JSON.ARRAPPEND
---

The `JSON.ARRAPPEND` command in DiceDB is used to append one or more JSON values to the end of a JSON array located at a specified path within a JSON document. This command is part of the DiceDBJSON module, which provides native JSON capabilities in DiceDB.

## Syntax

```plaintext
JSON.ARRAPPEND <key> <path> <json_value> [<json_value> ...]
```

## Parameters

- `key`: (String) The key under which the JSON document is stored.
- `path`: (String) The JSONPath expression that specifies the location within the JSON document where the array is located.
- `json_value`: (JSON) One or more JSON values to be appended to the array. These values must be valid JSON data types (e.g., string, number, object, array, boolean, or null).

## Return Value

- `Integer`: The length of the array after the append operation.

## Behaviour

When the `JSON.ARRAPPEND` command is executed, the specified JSON values are appended to the end of the array located at the given path within the JSON document stored under the specified key. If the path does not exist or does not point to an array, an error will be raised.

## Error Handling

- `(error) ERR key does not exist`: Raised if the specified key does not exist in the DiceDB database.
- `(error) ERR path does not exist`: Raised if the specified path does not exist within the JSON document.
- `(error) ERR path is not an array`: Raised if the specified path does not point to a JSON array.
- `(error) ERR invalid JSON`: Raised if any of the provided JSON values are not valid JSON.

## Example Usage

### Example 1: Appending a single value to an array

```plaintext
127.0.0.1:6379> JSON.SET myjson . '{"numbers": [1, 2, 3]}'
OK
127.0.0.1:6379> JSON.ARRAPPEND myjson .numbers 4
(integer) 4
127.0.0.1:6379> JSON.GET myjson
"{\"numbers\":[1,2,3,4]}"
```

### Example 2: Appending multiple values to an array

```plaintext
127.0.0.1:6379> JSON.SET myjson . '{"fruits": ["apple", "banana"]}'
OK
127.0.0.1:6379> JSON.ARRAPPEND myjson .fruits "cherry" "date"
(integer) 4
127.0.0.1:6379> JSON.GET myjson
"{\"fruits\":[\"apple\",\"banana\",\"cherry\",\"date\"]}"
```

### Example 3: Error when key does not exist

```plaintext
127.0.0.1:6379> JSON.ARRAPPEND nonexistingkey .array 1
(error) ERR key does not exist
```

### Example 4: Error when path does not exist

```plaintext
127.0.0.1:6379> JSON.SET myjson . '{"numbers": [1, 2, 3]}'
OK
127.0.0.1:6379> JSON.ARRAPPEND myjson .nonexistingpath 4
(error) ERR path does not exist
```

### Example 5: Error when path is not an array

```plaintext
127.0.0.1:6379> JSON.SET myjson . '{"object": {"key": "value"}}'
OK
127.0.0.1:6379> JSON.ARRAPPEND myjson .object 4
(error) ERR path is not an array
```

### Example 6: Error when invalid JSON is provided

```plaintext
127.0.0.1:6379> JSON.SET myjson . '{"numbers": [1, 2, 3]}'
OK
127.0.0.1:6379> JSON.ARRAPPEND myjson .numbers invalidjson
(error) ERR invalid JSON
```

## Notes

- Ensure that the DiceDBJSON module is loaded in your DiceDB instance to use the `JSON.ARRAPPEND` command.
- JSONPath expressions are used to navigate and specify the location within the JSON document. Familiarity with JSONPath syntax is beneficial for effective use of this command.

By following this documentation, users can effectively utilize the `JSON.ARRAPPEND` command to manipulate JSON arrays within DiceDB.

