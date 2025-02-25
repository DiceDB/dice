---
title: JSON.ARRINDEX
description: The `JSON.ARRINDEX` command in DiceDB searches for the first occurrence of a JSON value in an array.
---

The JSON.ARRINDEX command in DiceDB provides users with the ability to search for the position of a specific element within a JSON array stored at a specified path in a document identified by a given key. By executing this command, users can efficiently locate the index of an element that matches the provided value, enabling streamlined data access and manipulation.

This functionality is especially useful for developers dealing with large or nested JSON arrays who need to pinpoint the location of particular elements for further processing or validation. With support for specifying paths and flexible querying, JSON.ARRINDEX enhances the capability of managing and navigating complex JSON datasets within DiceDB.

## Syntax

```bash
JSON.ARRINDEX key path value [start [stop]]
```

## Parameters

| Parameter | Description                                                | Type    | Required |
| --------- | ---------------------------------------------------------- | ------- | -------- |
| `key`     | The name of the key holding the JSON document.             | String  | Yes      |
| `path`    | JSONPath pointing to an array within the JSON document.    | String  | Yes      |
| `value`   | The value to search for within the array in JSON document. | Mixed   | Yes      |
| `start`   | Optional index to start the search from. Defaults to 0.    | Integer | No       |
| `stop`    | Optional index to end the search. Defaults to 0.           | Integer | No       |

## Return Values

| Condition                 | Return Value                                                                                                                   |
| ------------------------- | ------------------------------------------------------------------------------------------------------------------------------ |
| Success                   | ([]integer) `array of integer replies for each path, the first position in the array of each JSON value that matches the path` |
| Wrong number of arguments | Error: `(error) ERR wrong number of arguments for JSON.ARRINDEX command`                                                       |
| Key has wrong type        | Error: `(error) ERR Existing key has wrong Dice type`                                                                          |
| Invalid integer arguments | Error: `(error) ERR Couldn't parse as integer`                                                                                 |
| Invalid jsonpath          | Error: `(error) ERR Path '<path>' does not exist`                                                                              |

## Behaviour

When the JSON.ARRINDEX command is issued, DiceDB performs the following steps:

1. It checks if argument count is valid or not. If not, an error is thrown.
2. It checks for the validity of value, start(optional) and stop(optional) argument passed. If not, an error is thrown.
3. If the jsonpath passed is invalid, an error is thrown.
4. It checks the type of value passed corresponding to the key. If it is not appropriate, error is thrown.
5. For each value matching the path, it checks if the value is JSON array.
6. If it is JSON array, it finds the first occurrence of the value.
7. If value is found, it adds to array the index where the value was found. Else, -1 is added.
8. If it is not JSON array, (nil) is added to resultant array.
9. The final array is returned.

## Errors

1. `Wrong number of arguments`:

   - Error Message: `(error) ERR wrong number of arguments for JSON.ARRINDEX command`
   - Raised if the number of arguments are less or more than expected.

2. `Couldn't parse as integer`:

   - Error Message: `(error) ERR Couldn't parse as integer`
   - Raised if the optional start and stop argument are non-integer strings.
   - Raised if the value is not a valid integer.

3. `Key has wrong Dice type`:

   - Error Message: `(error) ERR Existing key has wrong Dice type`
   - Raised if the value of the specified key doesn't match the specified value in DiceDb

4. `Path '<path>' does not exist`

   - Error Message: `(error) ERR Path '<path>' does not exist`
   - Raise if the path passed is not valid.

## Example Usage

### Basic usage

Searches for the first occurrence of a JSON value in an array

```bash
127.0.0.1:7379> JSON.SET a $ '{"name": "Alice", "age": 30, "mobile": [1902, 1903, 1904]}'
"OK"
127.0.0.1:7379> JSON.ARRINDEX a $.mobile 1903
1) (integer) 1
127.0.0.1:7379> JSON.ARRINDEX a $.mobile 1904
1) (integer) 2
```

### Finding the occurrence of value starting from given index

Searches for the first occurrence of a JSON value in an array starting from given index

```bash
127.0.0.1:7379> JSON.SET b $ '{"name": "Alice", "mobile": [1902, 1903, 1904]}'
"OK"
127.0.0.1:7379> JSON.ARRINDEX a $.mobile 1902 0
1) (integer) 0
127.0.0.1:7379> JSON.ARRINDEX a $.mobile 1902 1
1) (integer) -1
```

### Finding the occurrence of value starting from given index and ending at given index (exclusive)

Searches for the first occurrence of a JSON value in [start, stop) range

```bash
127.0.0.1:7379> JSON.SET b $ '{"name": "Alice", "mobile": [1902, 1903, 1904]}'
"OK"
127.0.0.1:7379> JSON.ARRINDEX a $.mobile 1902 0 2
1) (integer) 0
127.0.0.1:7379> JSON.ARRINDEX a $.mobile 1902 1 2
1) (integer) -1
127.0.0.1:7379> JSON.ARRINDEX a $.mobile 1904 0 1
1) (integer) -1
127.0.0.1:7379> JSON.ARRINDEX a $.mobile 1904 0 2
1) (integer) -1
127.0.0.1:7379> JSON.ARRINDEX a $.mobile 1904 0 3
1) (integer) 2
```

### When invalid start and stop argument is passed

Error When invalid start and stop argument is passed

```bash
127.0.0.1:7379> JSON.SET b $ '{"name": "Alice", "mobile": [1902, 1903, 1904]}'
"OK"
127.0.0.1:7379> JSON.ARRINDEX b $.mobile iamnotvalidinteger
(error) ERR Couldn't parse as integer
127.0.0.1:7379> JSON.ARRINDEX b $.mobile iamnotvalidinteger iamalsonotvalidinteger
(error) ERR Couldn't parse as integer
```

### When the jsonpath is not array object

Error When jsonpath is not array object

```bash
127.0.0.1:7379> set b '{"name":"Alice","mobile":[1902,1903,1904]}'
"OK"
127.0.0.1:7379> JSON.ARRINDEX b $.mobile 1902
(error) Existing key has wrong type
```

### When the jsonpath is not valid path

Error When jsonpath is not valid path

```bash
127.0.0.1:7379> JSON.SET b $ '{"name": "Alice", "mobile": [1902, 1903, 1904]}'
"OK"
127.0.0.1:7379> JSON.ARRINDEX b $invalid_path 3
(error) ERR Path '$invalid_path' does not exist
```
