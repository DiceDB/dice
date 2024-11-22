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

| Parameter | Description                                      | Type   | Required |
| --------- | ------------------------------------------------ | ------ | -------- |
| `command` | The command for which keys need to be extracted. | String | Yes      |
| `arg`     | Arguments for the specified command.             | String | Yes      |

## Return values

| Condition             | Return Value                               |
| --------------------- | ------------------------------------------ |
| Command is successful | Array of keys                              |
| Error                 | An error is returned if the command fails. |

## Behavior

The `COMMAND GETKEYS` command parses the provided command and its arguments to extract the keys that are involved. It ensures that the correct keys are identified, regardless of the specific operation being performed. Internally, DiceDB understands the structure of each command using keyspecs and identifies which parameters are keys.

## Errors

1.  `Arity Error for COMMAND GETKEYS`

    - Error Message: `(error) ERR wrong number of arguments for 'command|getkeys' command`
    - Occurs when an incorrect number of arguments is provided for the `COMMAND GETKEYS` command.

2.  `Arity Error for command`

    - Error Message: `(error) ERR invalid number of arguments specified for command`
    - Occurs when invalid number of arguments provided for command.

3.  `Invalid command specified`

    - Error Message: `(error) ERR invalid command specified`
    - Occurs when the provided command is not a recognized DiceDB command.

4.  `No keys arguments`
    - Error Message: `(error) ERR the command has no key arguments`
    - Occurs when the provided command does not accept any key arguments (ex. `FLUSHDB`).

## Example Usage

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

### Error: Arity Error for `COMMAND GETKEYS`

An arity error is thrown when the incorrect number of arguments is provided to the `COMMAND GETKEYS` command.

```bash
127.0.0.1:7379> COMMAND GETKEYS
(error) ERR wrong number of arguments for 'command|getkeys' command
```

### Error: Arity Error for command name passed to `COMMAND GETKEYS`

An arity error is thrown when the incorrect number of arguments is provided to the command passed to `COMMAND GETKEYS` command.

```bash
127.0.0.1:7379> COMMAND GETKEYS MSET key1
(error) ERR invalid number of arguments specified for command
```

### Error: Command Not Supported

An error is thrown when the specified command is not supported or recognized by the DiceDB server.

```bash
127.0.0.1:7379> COMMAND GETKEYS UNKNOWNCOMMAND key1
(error) ERR invalid command specified
```

### Error: Command Does Not Accept Key Arguments

An error is thrown when attempting to retrieve keys for a command that does not accept any key arguments.

```bash
127.0.0.1:7379> COMMAND GETKEYS FLUSHDB
(error) ERR The command has no key arguments
```
