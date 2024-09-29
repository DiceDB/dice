---
title: COMMAND
description: Documentation for the DiceDB command COMMAND
---

The `COMMAND` command in DiceDB is a powerful introspection tool that provides detailed information about all the DiceDB commands supported by the server. This command can be used to retrieve metadata about commands, such as their arity, flags, first key, last key, and key step. It is particularly useful for clients and developers who need to understand the capabilities and constraints of the DiceDB commands available in their environment.

## Parameters

The `COMMAND` command can be used in several forms, each with different parameters:

1. `COMMAND`: Returns details about all DiceDB commands.
2. `COMMAND COUNT`: Returns the total number of commands in the DiceDB server.
3. `COMMAND INFO command-name [command-name ...]`: Returns details about the specified commands.
4. `COMMAND GETKEYS command arg [arg ...]`: Returns the keys from the provided command and arguments.

### Detailed Parameter Descriptions

- `COMMAND`: No parameters. This form returns a list of all commands supported by the DiceDB server.
- `COMMAND COUNT`: No parameters. This form returns the total number of commands.
- `COMMAND INFO command-name [command-name ...]`:
  - `command-name`: One or more command names for which information is requested.
- `COMMAND GETKEYS command arg [arg ...]`:
  - `command`: The command to analyze.
  - `arg [arg ...]`: The arguments for the command.

## Return Value

The return value of the `COMMAND` command varies based on the form used:

1. `COMMAND`: Returns an array where each element is an array describing a command.
2. `COMMAND COUNT`: Returns an integer representing the total number of commands.
3. `COMMAND INFO command-name [command-name ...]`: Returns an array of arrays, each containing information about the specified commands.
4. `COMMAND GETKEYS command arg [arg ...]`: Returns an array of keys extracted from the provided command and arguments.

### Detailed Return Value Descriptions

- `COMMAND`:
  ```plaintext
  [
    [
      "command-name",
      arity,
      [
        "flag1",
        "flag2",
        ...
      ],
      first-key,
      last-key,
      key-step
    ],
    ...
  ]
  ```
- `COMMAND COUNT`:
  ```plaintext
  (integer) number_of_commands
  ```
- `COMMAND INFO command-name [command-name ...]`:
  ```plaintext
  [
    [
      "command-name",
      arity,
      [
        "flag1",
        "flag2",
        ...
      ],
      first-key,
      last-key,
      key-step
    ],
    ...
  ]
  ```
- `COMMAND GETKEYS command arg [arg ...]`:
  ```plaintext
  [
    "key1",
    "key2",
    ...
  ]
  ```

## Example Usage

### Example 1: Retrieving All Commands

```plaintext
> COMMAND
1) 1) "get"
   2) (integer) 2
   3) 1) "readonly"
      2) "fast"
   4) (integer) 1
   5) (integer) 1
   6) (integer) 1
2) 1) "set"
   2) (integer) -3
   3) 1) "write"
      2) "denyoom"
   4) (integer) 1
   5) (integer) 1
   6) (integer) 1
...
```

### Example 2: Counting Commands

```plaintext
> COMMAND COUNT
(integer) 200
```

### Example 3: Retrieving Information About Specific Commands

```plaintext
> COMMAND INFO get set
1) 1) "get"
   2) (integer) 2
   3) 1) "readonly"
      2) "fast"
   4) (integer) 1
   5) (integer) 1
   6) (integer) 1
2) 1) "set"
   2) (integer) -3
   3) 1) "write"
      2) "denyoom"
   4) (integer) 1
   5) (integer) 1
   6) (integer) 1
```

### Example 4: Extracting Keys from a Command

```plaintext
> COMMAND GETKEYS del key1 key2 key3
1) "key1"
2) "key2"
3) "key3"
```

## Behaviour

When the `COMMAND` command is executed, DiceDB inspects its internal command table and returns the requested information. The behavior varies slightly depending on the form of the command used:

- `COMMAND`: Returns a comprehensive list of all commands and their metadata.
- `COMMAND COUNT`: Returns the total number of commands.
- `COMMAND INFO`: Returns detailed information about the specified commands.
- `COMMAND GETKEYS`: Analyzes the provided command and arguments to extract the keys involved.

## Error Handling

The `COMMAND` command can raise errors in the following scenarios:

1. `Invalid Subcommand`: If an unrecognized subcommand is provided, DiceDB will return an error.

   - `Error Message`: `(error) ERR unknown subcommand`

2. `Invalid Command Name`: If a non-existent command name is provided in the `COMMAND INFO` subcommand, DiceDB will return an error.

   - `Error Message`: `(error) ERR unknown command 'command-name'`

3. `Invalid Arguments`: If the arguments provided to the `COMMAND GETKEYS` subcommand do not match the expected format, DiceDB will return an error.

   - `Error Message`: `(error) ERR wrong number of arguments for 'command' command`

### Example of Error Handling

#### Invalid Subcommand

```plaintext
> COMMAND INVALID
(error) ERR unknown subcommand
```

#### Invalid Command Name

```plaintext
> COMMAND INFO non_existent_command
(error) ERR unknown command 'non_existent_command'
```

#### Invalid Arguments for `COMMAND GETKEYS`

```plaintext
> COMMAND GETKEYS set
(error) ERR wrong number of arguments for 'set' command
```
