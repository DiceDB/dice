---
title: OBJECT
description: Documentation for the DiceDB command OBJECT
---

The `OBJECT` command in DiceDB is used to inspect the internals of DiceDB objects. It provides various subcommands that allow you to retrieve information about the encoding, reference count, and idle time of a key. This command is particularly useful for debugging and understanding the memory usage and performance characteristics of your DiceDB instance.

## Syntax

```bash
OBJECT <subcommand> <key>
```

## Parameters

- `<subcommand>`: The specific operation you want to perform on the key. The available subcommands are:

  - `REFCOUNT`: Returns the number of references of the value associated with the specified key.
  - `IDLETIME`: Returns the number of seconds since the object was last accessed.
  - `FREQ`: Returns the access frequency of a key, if the LFU (Least Frequently Used) eviction policy is enabled.

- `<key>`: The key for which you want to retrieve the information.

## Return Value

The return value depends on the subcommand used:

- `REFCOUNT`: Returns an integer representing the reference count of the key.
- `IDLETIME`: Returns an integer representing the idle time in seconds.
- `FREQ`: Returns an integer representing the access frequency of the key.

## Behaviour

When the `OBJECT` command is executed, DiceDB inspects the specified key and returns the requested information based on the subcommand. The command does not modify the key or its value; it only retrieves metadata about the key.

### Subcommand Behaviours

- `REFCOUNT`: This subcommand returns the number of references to the key's value. A higher reference count indicates that the value is being shared among multiple keys or clients.
- `IDLETIME`: This subcommand provides the time in seconds since the key was last accessed. It is useful for identifying stale keys.
- `FREQ`: This subcommand returns the access frequency of the key, which is useful when using the LFU eviction policy.

## Errors

The `OBJECT` command can raise errors in the following scenarios:

- `ERR wrong number of arguments for 'object' command`: This error occurs if the command is not provided with the correct number of arguments.
- `ERR no such key`: This error occurs if the specified key does not exist in the database.
- `ERR syntax error`: This error occurs if an invalid subcommand is provided.

## Example Usage

### Using the `REFCOUNT` Subcommand

```bash
OBJECT REFCOUNT mykey
(integer) 1
```

This response indicates that the value associated with `mykey` has a reference count of 1.

### Using the `IDLETIME` Subcommand

```bash
OBJECT IDLETIME mykey
(integer) 120
```

This response indicates that `mykey` has been idle for 120 seconds.

### Using the `FREQ` Subcommand

```bash
OBJECT FREQ mykey
(integer) 5
```

This response indicates that the access frequency of `mykey` is 5.
