---
title: COMMAND
description: Documentation for the DiceDB command COMMAND
---

The `COMMAND` command in DiceDB is a powerful introspection tool that provides detailed information about all the DiceDB commands supported by the server. This command can be used to retrieve metadata about commands, such as their arity, flags, first key, last key, and key step. It is particularly useful for clients and developers who need to understand the capabilities and constraints of the DiceDB commands available in their environment.

## Syntex

The `COMMAND` command can be used in multiple forms, supporting various subcommands, each with its own set of parameters:

```
COMMAND <subcommand>
```

## Parameters

- **subcommand**: Optional. Available subcommands include:
  - `COUNT` : Returns the total number of commands in the DiceDB server.
  - `GETKEYS` : Returns the keys from the provided command and arguments.
  - `LIST` : Returns the list of all the commands in the DiceDB server.
  - `INFO` : Returns details about the specified commands.
  - `HELP` : Displays the help section for `COMMAND`, providing information about each available subcommand.

##### For more details on each subcommand, please refer to their respective documentation pages.

## COMMAND (No subcommands)

### Parameters

- **COMMAND**: This form takes no parameters and returns a list of all commands and their metadata supported by the DiceDB server.

### Behavior

This command serves as the default implementation of the `COMMAND INFO` command when no command name is specified.

### Return Value

Returns an array, where each element is a nested array containing the following details for each command

- **Command Name**: The name of the command.
- **Arity**: An integer representing the number of arguments the command expects.
  - A positive number indicates the exact number of arguments.
  - A negative number indicates that the command accepts a variable number of arguments.
- **Flags** (_Note_: Not supported currently) : An array of flags that describe the command's properties (e.g., `readonly`, `fast`).
- **First Key**: The position of the first key in the argument list (0-based index).
- **Last Key**: The position of the last key in the argument list.
- **Key Step**: The step between keys in the argument list, useful for commands with multiple keys.

### Detailed Return Value Descriptions

- `COMMAND`:
  ```bash
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

### Example Usage

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
  5) 1) "CLIENT"
     2) (integer) -2
     3) (integer) 0
     4) (integer) 0
     5) (integer) 0
  6) 1) "JSON.ARRAPPEND"
     2) (integer) -3
     3) (integer) 0
     4) (integer) 0
     5) (integer) 0
  7) 1) "JSON.GET"
     2) (integer) 2
     3) (integer) 1
     4) (integer) 0
     5) (integer) 0
  8) 1) "JSON.INGEST"
     2) (integer) -3
     3) (integer) 1
     4) (integer) 0
     5) (integer) 0
  9) 1) "RESTORE"
     2) (integer) 2
     3) (integer) 1
     4) (integer) 0
     5) (integer) 0
 10) 1) "SLEEP"
     2) (integer) 1
     3) (integer) 0
     4) (integer) 0
     5) (integer) 0
 11) 1) "SMEMBERS"
     2) (integer) 2
     3) (integer) 1
     4) (integer) 0
     5) (integer) 0
 12) 1) "BFINIT"
     2) (integer) -2
     3) (integer) 1
     4) (integer) 0
     5) (integer) 1
 13) 1) "FLUSHDB"
     2) (integer) -1
     3) (integer) 0
     4) (integer) 0
     5) (integer) 0
 14) 1) "HINCRBYFLOAT"
     2) (integer) -4
     3) (integer) 1
     4) (integer) 0
     5) (integer) 0
 15) 1) "RENAME"
     2) (integer) 3
     3) (integer) 0
     4) (integer) 0
     5) (integer) 0
 16) 1) "SET"
     2) (integer) -3
     3) (integer) 1
     4) (integer) 0
     5) (integer) 0
 17) 1) "DECRBY"
     2) (integer) 3
     3) (integer) 1
     4) (integer) 0
     5) (integer) 1
 18) 1) "GETEX"
     2) (integer) -2
     3) (integer) 1
     4) (integer) 0
     5) (integer) 0
 19) 1) "HGET"
     2) (integer) -3
     3) (integer) 1
     4) (integer) 0
     5) (integer) 0
 20) 1) "LPOP"
     2) (integer) 2
     3) (integer) 0
     4) (integer) 0
     5) (integer) 0
 21) 1) "HELLO"
     2) (integer) -1
     3) (integer) 0
     4) (integer) 0
     5) (integer) 0
 22) 1) "HMGET"
     2) (integer) -2
     3) (integer) 1
     4) (integer) 0
     5) (integer) 0
 23) 1) "BITPOS"
     2) (integer) -2
     3) (integer) 0
     4) (integer) 0
     5) (integer) 0
 24) 1) "COMMAND <subcommand>"
     2) (integer) -1
     3) (integer) 0
     4) (integer) 0
     5) (integer) 0
 25) 1) "PFMERGE"
     2) (integer) -2
     3) (integer) 1
     4) (integer) 0
     5) (integer) 0
 26) 1) "PTTL"
     2) (integer) 2
     3) (integer) 1
     4) (integer) 0
     5) (integer) 0
 27) 1) "TOUCH"
     2) (integer) -2
     3) (integer) 1
     4) (integer) 0
     5) (integer) 0
 28) 1) "ECHO"
     2) (integer) 1
     3) (integer) 0
     4) (integer) 0
     5) (integer) 0
 29) 1) "TYPE"
     2) (integer) 1
     3) (integer) 1
     4) (integer) 0
     5) (integer) 0
 30) 1) "ZADD"
     2) (integer) -4
     3) (integer) 1
     4) (integer) 0
     5) (integer) 0
 31) 1) "DUMP"
     2) (integer) 1
     3) (integer) 1
     4) (integer) 0
     5) (integer) 0
 32) 1) "JSON.OBJKEYS"
     2) (integer) 2
     3) (integer) 0
     4) (integer) 0
     5) (integer) 0
 33) 1) "QWATCH"
     2) (integer) 1
     3) (integer) 0
     4) (integer) 0
     5) (integer) 0
 34) 1) "DBSIZE"
     2) (integer) 1
     3) (integer) 0
     4) (integer) 0
     5) (integer) 0
 35) 1) "GEOADD"
     2) (integer) -5
     3) (integer) 1
     4) (integer) 0
     5) (integer) 0
 36) 1) "HSET"
     2) (integer) -4
     3) (integer) 1
     4) (integer) 0
     5) (integer) 0
 37) 1) "JSON.ARRTRIM"
     2) (integer) -5
     3) (integer) 0
     4) (integer) 0
     5) (integer) 0
 38) 1) "LATENCY"
     2) (integer) -2
     3) (integer) 0
     4) (integer) 0
     5) (integer) 0
 39) 1) "INCR"
     2) (integer) 2
     3) (integer) 1
     4) (integer) 0
     5) (integer) 1
 40) 1) "ZRANGE"
     2) (integer) -4
     3) (integer) 1
     4) (integer) 0
     5) (integer) 0
 41) 1) "GEODIST"
     2) (integer) -4
     3) (integer) 1
     4) (integer) 0
     5) (integer) 0
 42) 1) "LRU"
     2) (integer) 1
     3) (integer) 0
     4) (integer) 0
     5) (integer) 0
 43) 1) "BGREWRITEAOF"
     2) (integer) 1
     3) (integer) 0
     4) (integer) 0
     5) (integer) 0
 44) 1) "MULTI"
     2) (integer) 1
     3) (integer) 0
     4) (integer) 0
     5) (integer) 0
 45) 1) "DEL"
     2) (integer) -2
     3) (integer) 1
     4) (integer) -1
     5) (integer) 1
 46) 1) "INCRBY"
     2) (integer) 2
     3) (integer) 1
     4) (integer) 0
     5) (integer) 1
 47) 1) "GET"
     2) (integer) 2
     3) (integer) 1
     4) (integer) 0
     5) (integer) 0
 48) 1) "HINCRBY"
     2) (integer) -4
     3) (integer) 1
     4) (integer) 0
     5) (integer) 0
 49) 1) "HRANDFIELD"
     2) (integer) -2
     3) (integer) 1
     4) (integer) 0
     5) (integer) 0
 50) 1) "JSON.OBJLEN"
     2) (integer) -2
     3) (integer) 1
     4) (integer) 0
     5) (integer) 0
 51) 1) "GETRANGE"
     2) (integer) 4
     3) (integer) 1
     4) (integer) 0
     5) (integer) 0
 52) 1) "JSON.CLEAR"
     2) (integer) -2
     3) (integer) 1
     4) (integer) 0
     5) (integer) 0
 53) 1) "RPOP"
     2) (integer) 2
     3) (integer) 0
     4) (integer) 0
     5) (integer) 0
 54) 1) "PFADD"
     2) (integer) -2
     3) (integer) 1
     4) (integer) 0
     5) (integer) 0
 55) 1) "GETBIT"
     2) (integer) 0
     3) (integer) 0
     4) (integer) 0
     5) (integer) 0
 56) 1) "GETSET"
     2) (integer) 2
     3) (integer) 0
     4) (integer) 0
     5) (integer) 0
 57) 1) "HMSET"
     2) (integer) -4
     3) (integer) 1
     4) (integer) 0
     5) (integer) 0
 58) 1) "HSETNX"
     2) (integer) 4
     3) (integer) 1
     4) (integer) 0
     5) (integer) 0
 59) 1) "JSON.STRLEN"
     2) (integer) -2
     3) (integer) 1
     4) (integer) 0
     5) (integer) 0
 60) 1) "OBJECT"
     2) (integer) -2
     3) (integer) 2
     4) (integer) 0
     5) (integer) 0
 61) 1) "SCARD"
     2) (integer) 2
     3) (integer) 1
     4) (integer) 0
     5) (integer) 0
 62) 1) "APPEND"
     2) (integer) 3
     3) (integer) 0
     4) (integer) 0
     5) (integer) 0
 63) 1) "BFADD"
     2) (integer) 3
     3) (integer) 1
     4) (integer) 0
     5) (integer) 1
 64) 1) "BFEXISTS"
     2) (integer) 3
     3) (integer) 1
     4) (integer) 0
     5) (integer) 1
 65) 1) "DISCARD"
     2) (integer) 1
     3) (integer) 0
     4) (integer) 0
     5) (integer) 0
 66) 1) "KEYS"
     2) (integer) 0
     3) (integer) 0
     4) (integer) 0
     5) (integer) 0
 67) 1) "BITCOUNT"
     2) (integer) -1
     3) (integer) 0
     4) (integer) 0
     5) (integer) 0
 68) 1) "RPUSH"
     2) (integer) -3
     3) (integer) 0
     4) (integer) 0
     5) (integer) 0
 69) 1) "BITFIELD"
     2) (integer) -1
     3) (integer) 1
     4) (integer) 0
     5) (integer) 0
 70) 1) "LLEN"
     2) (integer) 1
     3) (integer) 0
     4) (integer) 0
     5) (integer) 0
 71) 1) "SETEX"
     2) (integer) 3
     3) (integer) 1
     4) (integer) 0
     5) (integer) 0
 72) 1) "HEXISTS"
     2) (integer) -3
     3) (integer) 1
     4) (integer) 0
     5) (integer) 0
 73) 1) "INCRBYFLOAT"
     2) (integer) 2
     3) (integer) 0
     4) (integer) 0
     5) (integer) 0
 74) 1) "JSON.DEL"
     2) (integer) -2
     3) (integer) 1
     4) (integer) 0
     5) (integer) 0
 75) 1) "JSON.NUMINCRBY"
     2) (integer) 3
     3) (integer) 1
     4) (integer) 0
     5) (integer) 0
 76) 1) "JSON.NUMMULTBY"
     2) (integer) 3
     3) (integer) 1
     4) (integer) 0
     5) (integer) 0
 77) 1) "GETDEL"
     2) (integer) 2
     3) (integer) 1
     4) (integer) 0
     5) (integer) 0
 78) 1) "HKEYS"
     2) (integer) 1
     3) (integer) 1
     4) (integer) 0
     5) (integer) 0
 79) 1) "JSON.RESP"
     2) (integer) -2
     3) (integer) 1
     4) (integer) 0
     5) (integer) 0
 80) 1) "MSET"
     2) (integer) -3
     3) (integer) 1
     4) (integer) -1
     5) (integer) 2
 81) 1) "PFCOUNT"
     2) (integer) -2
     3) (integer) 1
     4) (integer) 0
     5) (integer) 0
 82) 1) "DECR"
     2) (integer) 2
     3) (integer) 1
     4) (integer) 0
     5) (integer) 1
 83) 1) "EXEC"
     2) (integer) 1
     3) (integer) 0
     4) (integer) 0
     5) (integer) 0
 84) 1) "EXISTS"
     2) (integer) 0
     3) (integer) 0
     4) (integer) 0
     5) (integer) 0
 85) 1) "JSON.ARRLEN"
     2) (integer) -2
     3) (integer) 1
     4) (integer) 0
     5) (integer) 0
 86) 1) "JSON.SET"
     2) (integer) -3
     3) (integer) 1
     4) (integer) 0
     5) (integer) 0
 87) 1) "ABORT"
     2) (integer) 1
     3) (integer) 0
     4) (integer) 0
     5) (integer) 0
 88) 1) "JSON.ARRINSERT"
     2) (integer) -5
     3) (integer) 1
     4) (integer) 0
     5) (integer) 0
 89) 1) "SETBIT"
     2) (integer) 0
     3) (integer) 0
     4) (integer) 0
     5) (integer) 0
 90) 1) "EXPIREAT"
     2) (integer) -3
     3) (integer) 1
     4) (integer) 0
     5) (integer) 1
 91) 1) "HLEN"
     2) (integer) 2
     3) (integer) 0
     4) (integer) 0
     5) (integer) 0
 92) 1) "SINTER"
     2) (integer) -2
     3) (integer) 1
     4) (integer) 0
     5) (integer) 0
 93) 1) "TTL"
     2) (integer) 2
     3) (integer) 1
     4) (integer) 0
     5) (integer) 0
 94) 1) "BFINFO"
     2) (integer) 2
     3) (integer) 0
     4) (integer) 0
     5) (integer) 0
 95) 1) "JSON.MGET"
     2) (integer) 2
     3) (integer) 1
     4) (integer) 0
     5) (integer) 0
 96) 1) "LPUSH"
     2) (integer) -3
     3) (integer) 0
     4) (integer) 0
     5) (integer) 0
 97) 1) "SDIFF"
     2) (integer) -2
     3) (integer) 1
     4) (integer) 0
     5) (integer) 0
 98) 1) "JSON.TOGGLE"
     2) (integer) 2
     3) (integer) 1
     4) (integer) 0
     5) (integer) 0
 99) 1) "HGETALL"
     2) (integer) -2
     3) (integer) 1
     4) (integer) 0
     5) (integer) 0
100) 1) "SADD"
     2) (integer) -3
     3) (integer) 1
     4) (integer) 0
     5) (integer) 0
101) 1) "SREM"
     2) (integer) -3
     3) (integer) 1
     4) (integer) 0
     5) (integer) 0
102) 1) "COPY"
     2) (integer) -2
     3) (integer) 0
     4) (integer) 0
     5) (integer) 0
103) 1) "HSTRLEN"
     2) (integer) -3
     3) (integer) 1
     4) (integer) 0
     5) (integer) 0
104) 1) "MGET"
     2) (integer) -2
     3) (integer) 1
     4) (integer) -1
     5) (integer) 1
105) 1) "JSON.ARRPOP"
     2) (integer) -2
     3) (integer) 0
     4) (integer) 0
     5) (integer) 0
106) 1) "JSON.TYPE"
     2) (integer) -2
     3) (integer) 1
     4) (integer) 0
     5) (integer) 0
107) 1) "JSON.DEBUG"
     2) (integer) 2
     3) (integer) 1
     4) (integer) 0
     5) (integer) 0
108) 1) "QUNWATCH"
     2) (integer) 1
     3) (integer) 0
     4) (integer) 0
     5) (integer) 0
109) 1) "SUBSCRIBE"
     2) (integer) 1
     3) (integer) 0
     4) (integer) 0
     5) (integer) 0
110) 1) "BITOP"
     2) (integer) 0
     3) (integer) 0
     4) (integer) 0
     5) (integer) 0
111) 1) "EXPIRE"
     2) (integer) -3
     3) (integer) 1
     4) (integer) 0
     5) (integer) 1
112) 1) "HDEL"
     2) (integer) -3
     3) (integer) 1
     4) (integer) 0
     5) (integer) 0
113) 1) "HVALS"
     2) (integer) -2
     3) (integer) 1
     4) (integer) 0
     5) (integer) 0
114) 1) "INFO"
     2) (integer) -1
     3) (integer) 0
     4) (integer) 0
     5) (integer) 0
115) 1) "EXPIRETIME"
     2) (integer) -2
     3) (integer) 1
     4) (integer) 0
     5) (integer) 1
116) 1) "JSON.FORGET"
     2) (integer) -2
     3) (integer) 1
     4) (integer) 0
     5) (integer) 0
117) 1) "SELECT"
     2) (integer) 1
     3) (integer) 0
     4) (integer) 0
     5) (integer) 0
127.0.0.1:7379>
```
