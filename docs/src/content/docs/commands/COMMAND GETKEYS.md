---
title: COMMAND GETKEYS
description: Documentation for the DiceDB command COMMAND GETKEYS.
---

## Introduction

The `COMMAND GETKEYS` command is used to extract the keys from a given command and its arguments in DiceDB. This command is particularly useful for analyzing the keys involved in a multi-key operation, such as MSET or DEL.

## Syntax

```bash
COMMAND GETKEYS command arg [arg ...]
```

## Parameters

- **`command`**: The DiceDB command from which the keys will be extracted (e.g., MSET, DEL, etc.).
- **`arg [arg ...]`**: The arguments for the command, which may include keys, values, or other command parameters.

## Return values

- **Array**: Returns an array of keys found in the provided command and its arguments.
  - For example, if the command is `MSET key1 value1 key2 value2`, the return value will be:
    ```bash
    1) "key1"
    2) "key2"
    ```

## Behavior

The `COMMAND GETKEYS` command parses the provided command and its arguments to extract the keys that are involved. It ensures that the correct keys are identified, regardless of the specific operation being performed. Internally, DiceDB understands the structure of each command using keyspecs and identifies which parameters are keys.

## Errors

- **Error: Arity Error for `COMMAND GETKEYS`**: Occurs when an incorrect number of arguments is provided for the `COMMAND GETKEYS` command.
  ```bash
    (error) ERR wrong number of arguments for 'command|getkeys' command
  ```
- **Error: Arity Error**: Occurs when invalid number of arguments provided for command.
  ```bash
    (error) ERR invalid number of arguments specified for command
  ```
- **Error: Invalid command specified**: If the provided command is not a recognized DiceDB command.
  ```bash
    (error) ERR invalid command specified
  ```
- **Error: No keys arguments**: If the provided command does not accept any key arguments (ex. `FLUSHDB`).
  ```bash
    (error) ERR the command has no key arguments
  ```

## Examples

### Extracting keys from MSET command

```bash
127.0.0.1:7379> COMMAND GETKEYS MSET key1 value1 key2 value2
1) "key1"
2) "key2"
```

### Extracting keys from DEL command

```bash
127.0.0.1:7379> COMMAND GETKEYS DEL key1 key2 key3
1) "key1"
2) "key2"
3) "key3"
```

### Arity error due to incorrect number of arguments for `COMMAND GETKEYS`

```bash
127.0.0.1:7379> COMMAND GETKEYS
(error) ERR wrong number of arguments for 'command|getkeys' command
```

### Arity error due to invalid number of arguments for command

```bash
127.0.0.1:7379> COMMAND GETKEYS MSET key1
(error) ERR invalid number of arguments specified for command
```

### Error when specified command is not supported.

```bash
127.0.0.1:7379> COMMAND GETKEYS UNKNOWNCOMMAND key1
(error) ERR invalid command specified
```

### Error when specified command does not accept any key arguments

```bash
127.0.0.1:7379> COMMAND GETKEYS FLUSHDB
(error) ERR The command has no key arguments
```
