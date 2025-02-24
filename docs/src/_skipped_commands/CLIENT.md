---
title: CLIENT
description: The `CLIENT` command in DiceDB is a versatile command used to manage and inspect client connections to the DiceDB server. It provides various subcommands to perform different operations related to client connections, such as listing connected clients, killing client connections, getting and setting client name, and more.
---

The `CLIENT` command in DiceDB is a versatile command used to manage and inspect client connections to the DiceDB server. It provides various subcommands to perform different operations related to client connections, such as listing connected clients, killing client connections, getting and setting client name, and more.

## Parameters

Each subcommand has its own set of parameters.

## Return Value

The return value of the `CLIENT` command depends on the subcommand used:

## Behaviour

When the `CLIENT` command is executed, it performs the action specified by the subcommand. For example, `CLIENT LIST` will list all connected clients, `CLIENT KILL` will terminate a specified client connection, and `CLIENT SETNAME` will set the name for the current connection. The command's behavior is determined by the subcommand and its parameters.

## Errors

Errors may be raised in the following scenarios:

- `Invalid Subcommand`:
  - `Error Message`: `(error) ERR unknown subcommand`
  - Occurs when an invalid subcommand is provided to the `CLIENT` command.

## Example Usage

### CLIENT LIST

```sh
127.0.0.1:7379> CLIENT LIST
id=3 addr=127.0.0.1:6379 fd=6 name= age=3 idle=0 flags=N db=0 sub=0 psub=0 multi=-1 qbuf=0 qbuf-free=0 obl=0 oll=0 omem=0 events=r cmd=client
```

### CLIENT KILL

```sh
127.0.0.1:7379> CLIENT KILL 127.0.0.1:6379
OK
```

### CLIENT GETNAME

```sh
127.0.0.1:7379> CLIENT GETNAME
"my-client"
```

### CLIENT SETNAME

```sh
127.0.0.1:7379> CLIENT SETNAME my-client
OK
```

### CLIENT PAUSE

```sh
127.0.0.1:7379> CLIENT PAUSE 5000
OK
```

### CLIENT REPLY

```sh
127.0.0.1:7379> CLIENT REPLY ON
OK
```

### CLIENT ID

```sh
127.0.0.1:7379> CLIENT ID
3
```

### CLIENT UNPAUSE

```sh
127.0.0.1:7379> CLIENT UNPAUSE
OK
```

### CLIENT TRACKING

```sh
127.0.0.1:7379> CLIENT TRACKING ON
OK
```

### CLIENT CACHING

```sh
127.0.0.1:7379> CLIENT CACHING YES
OK
```

### CLIENT NO-EVICT

```sh
127.0.0.1:7379> CLIENT NO-EVICT ON
OK
```

## Additional Information

The `CLIENT` command has several subcommands, each serving a specific purpose. The available subcommands are:

- `CLIENT LIST`
- `CLIENT KILL`
- `CLIENT GETNAME`
- `CLIENT SETNAME`
- `CLIENT PAUSE`
- `CLIENT REPLY`
- `CLIENT ID`
- `CLIENT UNPAUSE`
- `CLIENT TRACKING`
- `CLIENT CACHING`
- `CLIENT NO-EVICT`

### CLIENT LIST

- `Syntax`: `CLIENT LIST`
- `Description`: Returns information and statistics about the client connections server in a human-readable format.
- Returns a bulk string with the list of clients.

### CLIENT KILL

- `Syntax`: `CLIENT KILL [ip:port]`
- `Description`: Closes a client connection identified by `ip:port`.
- `Optional Parameters`:
  - `ID <client-id>`: Kill the client with the specified ID.
  - `TYPE <normal|master|replica|pubsub>`: Kill clients by type.
  - `ADDR <ip:port>`: Kill the client at the specified address.
  - `SKIPME <yes|no>`: Skip killing the current connection if set to `yes`.
- Returns `OK` if the client was successfully

### CLIENT GETNAME

- `Syntax`: `CLIENT GETNAME`
- `Description`: Returns the name of the current connection as set by `CLIENT SETNAME`.
- Returns the name of the current connection or a null bulk reply if no name is set.

### CLIENT SETNAME

- `Syntax`: `CLIENT SETNAME <name>`
- `Description`: Sets the name of the current connection.
- Returns `OK` if the name was successfully set.

### CLIENT PAUSE

- `Syntax`: `CLIENT PAUSE <timeout>`
- `Description`: Suspends all the DiceDB clients for the specified amount of time (in milliseconds).
- Returns `OK` if the clients were successfully paused.

### CLIENT REPLY

- `Syntax`: `CLIENT REPLY <ON|OFF|SKIP>`
- `Description`: Controls the replies sent to the client.
- Returns `OK` if the reply mode was successfully set.

### CLIENT ID

- `Syntax`: `CLIENT ID`
- `Description`: Returns the ID of the current connection.
- Returns the ID of the current connection.

### CLIENT UNPAUSE

- `Syntax`: `CLIENT UNPAUSE`
- `Description`: Resumes the clients that were paused by `CLIENT PAUSE`.
- Returns `OK` if the clients were successfully unpaused.

### CLIENT TRACKING

- `Syntax`: `CLIENT TRACKING <ON|OFF> [REDIRECT <id>] [BCAST] [PREFIX <prefix>] [OPTION] [OPTOUT] [NOLOOP]`
- `Description`: Enables or disables server-assisted client-side caching.
- Returns `OK` if tracking was successfully enabled or disabled.

### CLIENT CACHING

- `Syntax`: `CLIENT CACHING <YES|NO>`
- `Description`: Enables or disables tracking of the keys for the next command executed by the connection.
- Returns `OK` if caching was successfully enabled or disabled.

### CLIENT NO-EVICT

- `Syntax`: `CLIENT NO-EVICT <ON|OFF>`
- `Description`: Enables or disables the no-eviction mode for the current connection.
- Returns `OK` if no-eviction mode was successfully enabled or disabled.
