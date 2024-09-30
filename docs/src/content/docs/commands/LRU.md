---
title: LRU
description: The `LRU` (Least Recently Used) command in DiceDB is used to manage and retrieve information about the least recently used keys in the DiceDB database. This command is particularly useful for cache management and optimization, allowing users to understand and control the eviction of keys based on their usage patterns.
---

The `LRU` (Least Recently Used) command in DiceDB is used to manage and retrieve information about the least recently used keys in the DiceDB database. This command is particularly useful for cache management and optimization, allowing users to understand and control the eviction of keys based on their usage patterns.

## Parameters

The `LRU` command does not take any parameters. It is a standalone command that operates on the entire DiceDB database to provide information about the least recently used keys.

## Return Value

The `LRU` command returns a list of keys that are the least recently used in the DiceDB database. The exact format and content of the return value may vary depending on the DiceDB version and configuration.

## Example Usage

Here is an example of how to use the `LRU` command in DiceDB:

```bash
127.0.0.1:7379> LRU
1) "key1"
2) "key2"
3) "key3"
```

In this example, the `LRU` command returns a list of keys (`key1`, `key2`, `key3`) that are the least recently used in the DiceDB database.

## Behaviour

When the `LRU` command is executed, DiceDB performs the following actions:

1. Scans the entire database to identify keys based on their usage patterns.
2. Determines the least recently used keys.
3. Returns a list of these keys to the user.

This command is useful for understanding which keys are candidates for eviction when the DiceDB memory limit is reached and the `maxmemory-policy` is set to `allkeys-lru` or `volatile-lru`.

## Error Handling

The `LRU` command may raise errors in the following scenarios:

1. `Command Not Found`: If the `LRU` command is not recognized by the DiceDB server, an error will be raised.

   - `Error Message`: `(error) ERR unknown command 'LRU'`
   - `Cause`: This error occurs if the DiceDB server version does not support the `LRU` command or if there is a typo in the command.

2. `Permission Denied`: If the user does not have the necessary permissions to execute the `LRU` command, an error will be raised.

   - `Error Message`: `(error) NOAUTH Authentication required.`
   - `Cause`: This error occurs if the DiceDB server requires authentication and the user has not authenticated.

3. `Memory Limit Exceeded`: If the DiceDB server is under heavy memory pressure, executing the `LRU` command may result in an error.

   - `Error Message`: `(error) OOM command not allowed when used memory > 'maxmemory'.`
   - `Cause`: This error occurs if the DiceDB server has exceeded its configured memory limit and is unable to allocate additional memory for the `LRU` command.

## Notes

- The `LRU` command is particularly useful for cache management and optimization.
- Ensure that your DiceDB server version supports the `LRU` command before using it.
- Proper authentication and permissions are required to execute the `LRU` command successfully.
