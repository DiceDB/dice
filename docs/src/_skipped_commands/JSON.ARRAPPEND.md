---
title: JSON.ARRAPPEND
description: Documentation for the DiceDB command JSON.ARRAPPEND
---

The `JSON.ARRAPPEND` command in DiceDB is used to append one or more JSON values to the end of a JSON array located at a specified path within a JSON document. This command is part of the DiceDBJSON module, which provides native JSON capabilities in DiceDB.

## Syntax

```bash
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

1. `Wrong type of value or key`:

   - Error Message: `(error) WRONGTYPE Operation against a key holding the wrong kind of value`
   - Occurs when attempting to use the command on a key that contains a non-string value.

2. `Invalid Key`:

   - Error Message: `(error) ERR key does not exist`
   - Occurs when attempting to use the command on a key that does not exist.

3. `Invalid Path`:

   - Error Message: `(error) ERR path %s does not exist`
   - Occurs when attempting to use the command on a path that does not exist in the JSON document.

4. `Non Array Value at Path`:
   - Error Message: `(error) ERR path is not an array`
   - Occurs when attempting to use the command on a path that contains a non-array value.

## Example Usage

### Appending a single value to an array

```bash
127.0.0.1:7379> JSON.SET myjson . '{"numbers": [1, 2, 3]}'
OK
127.0.0.1:7379> JSON.ARRAPPEND myjson .numbers 4
(integer) 4
127.0.0.1:7379> JSON.GET myjson
"{\"numbers\":[1,2,3,4]}"
```

### Appending multiple values to an array

```bash
127.0.0.1:7379> JSON.SET myjson . '{"fruits": ["apple", "banana"]}'
OK
127.0.0.1:7379> JSON.ARRAPPEND myjson .fruits "cherry" "date"
(integer) 4
127.0.0.1:7379> JSON.GET myjson
"{\"fruits\":[\"apple\",\"banana\",\"cherry\",\"date\"]}"
```

### Error when key does not exist

```bash
127.0.0.1:7379> JSON.ARRAPPEND nonexistingkey .array 1
(error) ERR key does not exist
```

### Error when path does not exist

```bash
127.0.0.1:7379> JSON.SET myjson . '{"numbers": [1, 2, 3]}'
OK
127.0.0.1:7379> JSON.ARRAPPEND myjson .nonexistingpath 4
(error) ERR path .nonexistingpath does not exist
```

### Error when path is not an array

```bash
127.0.0.1:7379> JSON.SET myjson . '{"object": {"key": "value"}}'
OK
127.0.0.1:7379> JSON.ARRAPPEND myjson .object 4
(error) ERR path is not an array
```

### Error when invalid JSON is provided

```bash
127.0.0.1:7379> JSON.SET myjson . '{"numbers": [1, 2, 3]}'
OK
127.0.0.1:7379> JSON.ARRAPPEND myjson .numbers invalidjson
(error) ERR invalid JSON
```

## Notes

- Ensure that the DiceDBJSON module is loaded in your DiceDB instance to use the `JSON.ARRAPPEND` command.
- JSONPath expressions are used to navigate and specify the location within the JSON document. Familiarity with JSONPath syntax is beneficial for effective use of this command.

By following this documentation, users can effectively utilize the `JSON.ARRAPPEND` command to manipulate JSON arrays within DiceDB.
