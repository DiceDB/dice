---
title: GET
description: The `GET` command in DiceDB is used to retrieve the value of a specified key. If the key exists, the value is written and if it does not then the command returns `nil`. This is one of the most fundamental operations in DiceDB.
---

The `GET` command in DiceDB is used to retrieve the value of a specified key. If the key exists, the value is written and
if it does not then the command returns `nil` and an error is returned if the value stored at key is not a string. This is one of the most fundamental operations in DiceDB.

## Syntax

```bash
GET key
```

## Parameters

| Parameter | Description                                                              | Type   | Required |
|-----------|--------------------------------------------------------------------------|--------|----------|
| key       | The name of the key whose value you want to retrieve. The key is a string.| string | Yes      |

## Return Values

| Condition                                              | Return Value                                                                                       |
|--------------------------------------------------------|---------------------------------------------------------------------------------------------------|
| The specified key exists and holds a string value      | The string value stored at the key.                                                               |
| The specified key does not exist                       | `nil`                                                                                             |
| The specified key exists but is not a string, or multiple keys are passed | error                                                                                         |

## Behaviour

When the GET command is issued, DiceDB checks the existence of the specified key:

1. `Key Exists and Holds a String`: The value associated with the key is retrieved and returned.
2. `Key Does Not Exist`: The command returns `nil`
3. `Key Exists but Holds a Non-string Value`: An error is raised indicating that the operation against that key is not permitted.
4. `Multiple Keys Passed`: If multiple keys are passed to the GET command, an error is raised, as the command only accepts a single key at a time.

The GET command is a read-only operation and does not modify the state of the DiceDB database.

## Errors
1. **Expected string but got another type:**
    - Error Message: (error) ERR expected string but got another type
    - Occurs when the specified key holds a value that is not a string (e.g., a list, set, hash, or zset). DiceDB uses strict type checking to ensure that the correct type of operation is performed on the appropriate data type.

2. **Wrong number of arguments for 'GET' command:**
    - Error Message: (error) ERR wrong number of arguments for 'get' command
    - Occurs when multiple keys are passed as parameters.
	

## Example Usage

Here are a few examples demonstrating the usage of the GET command:

### Example 1: Key Exists and Holds a String Value

```bash
127.0.0.1:7379> SET mykey "Hello, DiceDB!"
127.0.0.1:7379> GET mykey
"Hello, DiceDB!"
```

### Example 2: Key Does Not Exist

```bash
127.0.0.1:7379> GET nonexistingkey
(nil)
```

## Additional Notes

- `Memory Usage`: The value returned by the GET command occupies memory only during the command execution and is handed back to the client.
- `Performance`: The GET command is an O(1) operation, which means it has a constant time complexity regardless of the size of the value associated with the key.
- `Atomicity`: Like other individual DiceDB commands, GET is executed atomically. No other command can alter the key's state during the execution of GET.
