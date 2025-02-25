---
title: JSON.TOGGLE
description: Documentation for the DiceDB command JSON.TOGGLE
---

The `JSON.TOGGLE` command is part of the DiceDBJSON module, which allows you to manipulate JSON data stored in DiceDB. This command is used to toggle the boolean value at a specified path within a JSON document. If the value at the specified path is `true`, it will be changed to `false`, and vice versa.

## Parameters

| Parameter | Description                                                               | Type   | Required |
| --------- | ------------------------------------------------------------------------- | ------ | -------- |
| `key`     | The key under which the JSON document is stored.                          | String | Yes      |
| `path`    | The JSONPath expression specifying the location within the JSON document. | String | Yes      |

## Return Values

| Condition             | Return Value                            |
| --------------------- | --------------------------------------- |
| Command is successful | The number of values that were toggled. |
| Path does not exist   | `0`                                     |

## Behaviour

When the `JSON.TOGGLE` command is executed, it will locate the boolean value at the specified path within the JSON document stored under the given key. If the value is `true`, it will be changed to `false`, and if it is `false`, it will be changed to `true`. If the path does not exist or the value at the path is not a boolean, the command will not perform any toggling and will return `0`.

## Errors

The `JSON.TOGGLE` command can raise the following errors:

1. `(error) ERR wrong number of arguments for 'JSON.TOGGLE' command`:

   - `Cause`: This error occurs if the number of arguments provided to the command is incorrect.
   - `Solution`: Ensure that you provide exactly two arguments: the key and the path.

2. `(error) ERR key does not exist`:

   - `Cause`: This error occurs if the specified key does not exist in the DiceDB database.
   - `Solution`: Verify that the key exists before attempting to toggle a value within it.

3. `(error) ERR path does not exist`:

   - `Cause`: This error occurs if the specified path does not exist within the JSON document.
   - `Solution`: Ensure that the path is correct and exists within the JSON document.

4. `(error) ERR value at path is not a boolean`:

   - `Cause`: This error occurs if the value at the specified path is not a boolean.
   - `Solution`: Ensure that the value at the specified path is a boolean before attempting to toggle it.

## Example Usage

### Toggling a Boolean Value

JSON Document

```json
{
  "name": "John Doe",
  "active": true,
  "settings": {
    "notifications": true
  }
}
```

```bash
127.0.0.1:7379> JSON.TOGGLE user:1001 $.active
(integer) 1
```

Updated JSON Document

```json
{
  "name": "John Doe",
  "active": false,
  "settings": {
    "notifications": true
  }
}
```

### Toggling a Nested Boolean Value

```bash
127.0.0.1:7379> JSON.TOGGLE user:1001 $.settings.notifications
(integer) 1
```

Updated JSON Document

```json
{
  "name": "John Doe",
  "active": false,
  "settings": {
    "notifications": false
  }
}
```

### Path Does Not Exist

```bash
127.0.0.1:7379> JSON.TOGGLE user:1001 $.nonexistent
(integer) 0
```

### Value at Path is Not a Boolean

```bash
127.0.0.1:7379> JSON.TOGGLE user:1001 $.name
(error) ERR value at path is not a boolean
```
