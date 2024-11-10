---
title: MSET
description: The `MSET` command in DiceDB is used to set multiple key-value pairs in a single atomic operation. This command is particularly useful for reducing the number of round-trip times between the client and the server when you need to set multiple keys at once.
---

The `MSET` command in DiceDB is used to set multiple key-value pairs in a single atomic operation. This command is particularly useful for reducing the number of round-trip times between the client and the server when you need to set multiple keys at once.

## Syntax

```bash
MSET key1 value1 [key2 value2 ...]
```

## Parameters

| Parameter             | Description                                           | Type   | Required |
| --------------------- | ----------------------------------------------------- | ------ | -------- |
| `key1, key2, ...`     | The keys to be set.                                   | String | Yes      |
| `value1, value2, ...` | The values to be associated with the respective keys. | String | Yes      |

## Return values

| Condition                                                | Return Value   |
| -------------------------------------------------------- | -------------- |
| The command returns `OK` if the operation is successful. | A string value |

## Behaviour

When the `MSET` command is executed:

- DiceDB sets the specified keys to their respective values.
- This operation is atomic, meaning that either all the keys are set, or none of them are.
- This ensures data consistency and integrity.
- Any pre-existing keys are overwritten and their respective TTL (if set) are reset.

## Errors

The `MSET` command can raise errors in the following scenarios:

- `Wrong number of arguments`: If the number of arguments is not even (i.e., there is a key without a corresponding value), DiceDB will return an error:
  ```bash
  (error) ERROR wrong number of arguments for 'mset' command
  ```
- `Non-string keys or values`: If any of the keys or values are not strings, DiceDB will return an error:
  ```bash
  (error) ERROR value is not a valid string
  ```

## Example Usage

### Basic Example

To set multiple key-value pairs in DiceDB:

```bash
127.0.0.1:7379> MSET key1 "value1" key2 "value2" key3 "value3"
OK
```

### Example with Retrieval

To set multiple key-value pairs and then retrieve them:

```bash
127.0.0.1:7379> MSET name "Alice" age "30" city "Wonderland"
OK
127.0.0.1:7379> GET name
"Alice"
127.0.0.1:7379> GET age
"30"
127.0.0.1:7379> GET city
"Wonderland"
```

### Error Example

Attempting to set an odd number of arguments:

```bash
127.0.0.1:7379> MSET key1 "value1" key2
(error) ERROR wrong number of arguments for 'mset' command
```
