---
title: COMMAND
description: Documentation for the DiceDB command COMMAND
---

The `COMMAND` command in DiceDB is a powerful introspection tool that provides detailed information about all the DiceDB commands supported by the server. This command can be used to retrieve metadata about commands, such as their arity, flags, first key, last key, and key step. It is particularly useful for clients and developers who need to understand the capabilities and constraints of the DiceDB commands available in their environment.

The `COMMAND` command can be used in multiple forms, supporting various subcommands, each with its own set of parameters. However, the default implementation, which does not require any subcommand, is as follows.

## Syntax

```bash
COMMAND
```

## Parameters

This command does not accept any parameters.

## Return values

| Condition             | Return Value                                                                             |
| --------------------- | ---------------------------------------------------------------------------------------- |
| Command is successful | Array containing detailed information about all commands supported by the DiceDB server. |
| Error                 | An error is returned if the command fails.                                               |

- **Command Name**: The name of the command.
- **Arity**: An integer representing the number of arguments the command expects.
  - A positive number indicates the exact number of arguments.
  - A negative number indicates that the command accepts a variable number of arguments.
- **Flags** (_Note_: Not supported currently) : An array of flags that describe the command's properties (e.g., `readonly`, `fast`).
- **First Key**: The position of the first key in the argument list (0-based index).
- **Last Key**: The position of the last key in the argument list.
- **Key Step**: The step between keys in the argument list, useful for commands with multiple keys.

```bash
127.0.0.1:7379> COMMAND
  1)  1) "command-name"
      2) (integer) arity
      3) 1) "flag1"       # Optional
         2) "flag2"      # Optional
         ...
      4) (integer) first-key
      5) (integer) last-key
      6) (integer) key-step
  .
  .
  .
```

## Behavior

When no subcommand is provided, this command functions as the default implementation of the `COMMAND INFO` command in the absence of a specified command name. It iterates through the list of registered commands and subcommands, returning an array containing detailed metadata for each command.

## Errors

No error is thrown in the default implementation of the `COMMAND` command when no subcommand is provided.

## Example Usage

### Retrieve Detailed Information for Each Command Supported by the DiceDB Server

```bash
127.0.0.1:7379> COMMAND
  1) 1) "AUTH"
     2) (integer) 0
     3) (integer) 0
     4) (integer) 0
     5) (integer) 0
  2) 1) "HSCAN"
     2) (integer) -3
     3) (integer) 1
     4) (integer) 0
     5) (integer) 0
  3) 1) "PERSIST"
     2) (integer) 0
     3) (integer) 0
     4) (integer) 0
     5) (integer) 0
  4) 1) "PING"
     2) (integer) -1
     3) (integer) 0
     4) (integer) 0
     5) (integer) 0
   .
   .
   .
127.0.0.1:7379>
```

## Subcommands

### Syntax

```bash
COMMAND <subcommand>
```

### Parameters

- **subcommand**: Optional. Available subcommands include:
  - [`COUNT`](/commands/command-count) : Returns the total number of commands in the DiceDB server.
  - [`GETKEYS`](/commands/command-getkeys) : Returns the keys from the provided command and arguments.
  - [`LIST`](/commands/command-list): Returns the list of all the commands in the DiceDB server.
  - [`INFO`](/commands/command-info): Returns details about the specified commands.
  - [`HELP`](/commands/command-help) : Displays the help section for `COMMAND`, providing information about each available subcommand.

**For more details on each subcommand, please refer to their respective documentation pages.**

## Errors

1.  `Unknown subcommand`
    - Error Message: ` (error) ERR unknown subcommand 'sucommand-name'. Try COMMAND HELP.`
    - This error may occur if the subcommand is misspelled or not recognized by the DiceDB server.

## Example Usage

### Invalid usage

An error is thrown when an incorrect or unsupported subcommand name is provided.

```bash
127.0.0.1:7379> COMMAND UNKNOWNSUBCOMMAND
(error) ERR unknown subcommand 'UNKNOWNSUBCOMMAND'. Try COMMAND HELP.
```
