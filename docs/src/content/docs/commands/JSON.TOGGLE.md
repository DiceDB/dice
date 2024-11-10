---
title: JSON.TOGGLE
description: Documentation for the DiceDB command JSON.TOGGLE
---

The `JSON.TOGGLE` command is part of the DiceDBJSON module, which allows you to manipulate JSON data stored in DiceDB. This command is used to toggle the boolean value at a specified path within a JSON document. If the value at the specified path is `true`, it will be changed to `false`, and vice versa.

## Parameters

- `key`: (String) The key under which the JSON document is stored.
- `path`: (String) The JSONPath expression specifying the location within the JSON document where the boolean value should be toggled.

### Detailed Parameter Description

1. `key`:

   - `Type`: String
   - `Description`: The unique identifier for the JSON document stored in DiceDB.
   - `Example`: `"user:1001"`

2. `path`:

   - `Type`: String
   - `Description`: A JSONPath expression that specifies the exact location of the boolean value within the JSON document.
   - `Example`: `$.active`

## Return Value

- `Type`: Integer
- `Description`: The number of values that were toggled.
- `Example`: `1` if one boolean value was toggled, `0` if no boolean values were toggled.

## Behaviour

When the `JSON.TOGGLE` command is executed, it will locate the boolean value at the specified path within the JSON document stored under the given key. If the value is `true`, it will be changed to `false`, and if it is `false`, it will be changed to `true`. If the path does not exist or the value at the path is not a boolean, the command will not perform any toggling and will return `0`.

## Error Handling

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

### Example 1: Toggling a Boolean Value

#### JSON Document

```json
{
  "name": "John Doe",
  "active": true,
  "settings": {
    "notifications": true
  }
}
```

#### Command

```bash
JSON.TOGGLE user:1001 $.active
```

#### Result

```bash
(integer) 1
```

#### Updated JSON Document

```json
{
  "name": "John Doe",
  "active": false,
  "settings": {
    "notifications": true
  }
}
```

### Example 2: Toggling a Nested Boolean Value

#### Command

```bash
JSON.TOGGLE user:1001 $.settings.notifications
```

#### Result

```bash
(integer) 1
```

#### Updated JSON Document

```json
{
  "name": "John Doe",
  "active": false,
  "settings": {
    "notifications": false
  }
}
```

### Example 3: Path Does Not Exist

#### Command

```bash
JSON.TOGGLE user:1001 $.nonexistent
```

#### Result

```bash
(integer) 0
```

### Example 4: Value at Path is Not a Boolean

#### Command

```bash
JSON.TOGGLE user:1001 $.name
```

#### Result

```bash
(error) ERR value at path is not a boolean
```
