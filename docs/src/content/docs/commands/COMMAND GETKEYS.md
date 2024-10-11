---
title: COMMAND GETKEYS
description: Documentation for the DiceDB command COMMAND GETKEYS.
---

## Introduction

The `COMMAND GETKEYS` command is used to extract the keys from a given command and its arguments in DiceDB. This command is particularly useful for analyzing the keys involved in a multi-key operation, such as MSET or DEL.

## Syntax

```
COMMAND GETKEYS command arg [arg ...]
```

## Parameters

- **`command`**: The DiceDB command from which the keys will be extracted (e.g., MSET, DEL, etc.).
- **`arg [arg ...]`**: The arguments for the command, which may include keys, values, or other command parameters.

## Return values

- **Array**: Returns an array of keys found in the provided command and its arguments.
  - For example, if the command is `MSET key1 value1 key2 value2`, the return value will be:
    ```
    ["key1", "key2"]
    ```

## Behavior

The `COMMAND GETKEYS` command parses the provided command and its arguments to extract the keys that are involved. It ensures that the correct keys are identified, regardless of the specific operation being performed. Internally, DiceDB understands the structure of each command using keyspecs and identifies which parameters are keys.

## Errors

- **Error: Invalid number of arguments**: Returned when the number of arguments provided is insufficient or incorrect.
  - `(error) ERR invalid number of arguments specified for command`
- **Error: Invalid command specified**: If the provided command is not a recognized DiceDB command.
  - `(error) ERR invalid command specified`

## Examples

### Example 1: Extracting keys from MSET command

```bash
127.0.0.1:7379> COMMAND GETKEYS MSET key1 value1 key2 value2
["key1", "key2"]
```

### Example 2: Extracting keys from DEL command

```bash
127.0.0.1:7379> COMMAND GETKEYS DEL key1 key2 key3
["key1", "key2", "key3"]

```

### Example 3: Error due to insufficient arguments

```bash
127.0.0.1:7379> COMMAND GETKEYS MSET key1
(error) ERR invalid number of arguments specified for command
```

### Example 4: Error when speficied command is not supported

```bash
127.0.0.1:7379> COMMAND GETKEYSs MSET key1
(error) ERR invalid command specified
```
