---
title: GET
description: The `GET` command in DiceDB is used to retrieve the value of a specified key. If the key exists, the value is written and if it does not then the command returns `nil`. This is one of the most fundamental operations in DiceDB.
---

The `GET` command in DiceDB is used to retrieve the value of a specified key. If the key exists, the value is written and
if it does not then the command returns `nil`. This is one of the most fundamental operations in DiceDB.

## Synopsis

```bash
GET key
```

## Parameters

- `key`: The name of the key whose value you want to retrieve. The key is a string, and it is required for the GET command to execute.

## Return Value

- `String`: If the specified key exists and holds a string value, the GET command returns the value stored at the key.
- `nil`: If the specified key does not exist, the command returns `nil`.
- `Error`: If the specified key exists but is not a string, an error is returned.

## Behaviour

When the GET command is issued, DiceDB checks the existence of the specified key:

1. `Key Exists and Holds a String`: The value associated with the key is retrieved and returned.
2. `Key Does Not Exist`: The command returns `nil`.
3. `Key Exists but Holds a Non-string Value`: An error is raised indicating that the operation against that key is not permitted.

The GET command is a read-only operation and does not modify the state of the DiceDB database.

## Error Handling

The GET command can raise the following errors:

- `WRONGTYPE Operation against a key holding the wrong kind of value`: This error is returned if the specified key holds a value that is not a string (e.g., a list, set, hash, or zset). DiceDB uses strict type checking to ensure that the correct type of operation is performed on the appropriate data type.

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
