---
title: COMMAND LIST
description: Documentation for the DiceDB command COMMAND LIST.
---

## Introduction

The `COMMAND LIST` command retrieves a list of all commands supported by the DiceDB server. This allows users to discover available commands for various operations, making it easier to understand the capabilities of the database.

## Syntax

```
COMMAND LIST
```

## Parameters

This command does not accept any parameters.

## Return values

The command returns an array of strings, where each string represents a command name available in the DiceDB server. If no commands are available (which is unlikely), an empty array is returned.

## Behavior

When executed, the `COMMAND LIST` command scans the DiceDB server's command registry and compiles a list of command names. The operation is efficient, leveraging the server's internal command registration system to provide results quickly.

## Errors

- **Error: unknown sucommand**: This error may occur if the subcommand is misspelled or not recognized by the server.
  - `(error) ERR unknown subcommand 'sucommand-name'. Try COMMAND HELP.`

## Examples

```bash
127.0.0.1:7379> COMMAND LIST
  1) "SLEEP"
  2) "SMEMBERS"
  3) "BFINIT"
  4) "FLUSHDB"
  5) "HINCRBYFLOAT"
  6) "RENAME"
  7) "SET"
  8) "DECRBY"
  9) "GETEX"
 10) "HGET"
 11) "LPOP"
 12) "HELLO"
 13) "HMGET"
 14) "BITPOS"
 15) "COMMAND"
 16) "COMMAND|COUNT"
 17) "COMMAND|GETKEYS"
 18) "COMMAND|LIST"
 19) "COMMAND|HELP"
 20) "COMMAND|INFO"
 21) "PFMERGE"
 22) "PTTL"
 23) "TOUCH"
 24) "ECHO"
 25) "TYPE"
 26) "ZADD"
 27) "DUMP"
 28) "JSON.OBJKEYS"
 29) "QWATCH"
 30) "DBSIZE"
 31) "GEOADD"
 32) "HSET"
 33) "JSON.ARRTRIM"
 34) "LATENCY"
 35) "INCR"
 36) "ZRANGE"
 37) "GEODIST"
 38) "LRU"
 39) "BGREWRITEAOF"
 40) "MULTI"
 41) "DEL"
 42) "INCRBY"
 43) "GET"
 44) "HINCRBY"
 45) "HRANDFIELD"
 46) "JSON.OBJLEN"
 47) "GETRANGE"
 48) "JSON.CLEAR"
 49) "RPOP"
 50) "PFADD"
 51) "GETBIT"
 52) "GETSET"
 53) "HMSET"
 54) "HSETNX"
 55) "JSON.STRLEN"
 56) "OBJECT"
 57) "SCARD"
 58) "APPEND"
 59) "BFADD"
 60) "BFEXISTS"
 61) "DISCARD"
 62) "KEYS"
 63) "BITCOUNT"
 64) "RPUSH"
 65) "BITFIELD"
 66) "LLEN"
 67) "SETEX"
 68) "HEXISTS"
 69) "INCRBYFLOAT"
 70) "JSON.DEL"
 71) "JSON.NUMINCRBY"
 72) "JSON.NUMMULTBY"
 73) "GETDEL"
 74) "HKEYS"
 75) "JSON.RESP"
 76) "MSET"
 77) "PFCOUNT"
 78) "DECR"
 79) "EXEC"
 80) "EXISTS"
 81) "JSON.ARRLEN"
 82) "JSON.SET"
 83) "ABORT"
 84) "JSON.ARRINSERT"
 85) "SETBIT"
 86) "EXPIREAT"
 87) "HLEN"
 88) "SINTER"
 89) "TTL"
 90) "BFINFO"
 91) "JSON.MGET"
 92) "LPUSH"
 93) "SDIFF"
 94) "JSON.TOGGLE"
 95) "HGETALL"
 96) "SADD"
 97) "SREM"
 98) "COPY"
 99) "HSTRLEN"
100) "MGET"
101) "JSON.ARRPOP"
102) "JSON.TYPE"
103) "JSON.DEBUG"
104) "QUNWATCH"
105) "SUBSCRIBE"
106) "BITOP"
107) "EXPIRE"
108) "HDEL"
109) "HVALS"
110) "INFO"
111) "EXPIRETIME"
112) "JSON.FORGET"
113) "SELECT"
114) "AUTH"
115) "HSCAN"
116) "PERSIST"
117) "PING"
118) "CLIENT"
119) "JSON.ARRAPPEND"
120) "JSON.GET"
121) "JSON.INGEST"
122) "RESTORE"
127.0.0.1:7379>
```
