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
 1) "JSON.CLEAR"
 2) "JSON.OBJLEN"
 3) "OBJECT"
 4) "DBSIZE"
 5) "SREM"
 6) "PING"
 7) "MSET"
 8) "MULTI"
 9) "SETBIT"
10) "HSET"
11) "GETSET"
12) "FLUSHDB"
13) "PFMERGE"
14) "EXPIREAT"
15) "DISCARD"
16) "DECR"
17) "PTTL"
18) "JSON.MGET"
19) "TTL"
20) "INFO"
21) "LPUSH"
22) "TYPE"
23) "SET"
24) "LRU"
25) "SLEEP"
26) "BFEXISTS"
27) "ABORT"
28) "HLEN"
29) "LPOP"
30) "GET"
31) "JSON.TOGGLE"
32) "BGREWRITEAOF"
33) "CLIENT"
34) "BFADD"
35) "COMMAND"
36) "MGET"
37) "SCARD"
38) "PFADD"
39) "PFCOUNT"
40) "SELECT"
41) "JSON.ARRPOP"
42) "BITPOS"
43) "SADD"
44) "SINTER"
45) "HGETALL"
46) "HELLO"
47) "BITOP"
48) "JSON.NUMINCRBY"
49) "JSON.TYPE"
50) "BFINFO"
51) "EXISTS"
52) "GETEX"
53) "EXPIRE"
54) "EXPIRETIME"
55) "EXEC"
56) "RENAME"
57) "JSON.STRLEN"
58) "JSON.SET"
59) "DEL"
60) "BFINIT"
61) "SUBSCRIBE"
62) "QWATCH"
63) "BITCOUNT"
64) "TOUCH"
65) "JSON.ARRAPPEND"
66) "JSON.DEBUG"
67) "GETBIT"
68) "DECRBY"
69) "RPOP"
70) "LLEN"
71) "AUTH"
72) "JSON.GET"
73) "INCR"
74) "PERSIST"
75) "RPUSH"
76) "JSON.DEL"
77) "JSON.INGEST"
78) "GETDEL"
79) "SMEMBERS"
80) "HGET"
81) "LATENCY"
82) "KEYS"
83) "JSON.FORGET"
84) "JSON.ARRLEN"
85) "QUNWATCH"
86) "COPY"
87) "SDIFF"
127.0.0.1:7379>
```
