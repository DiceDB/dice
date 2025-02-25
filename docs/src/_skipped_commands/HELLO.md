---
title: HELLO
description: The `HELLO` command in DiceDB is used to initiate a connection handshake with the server. This command is particularly useful for clients to negotiate the protocol version and authentication details. It can also be used to switch the connection to a different protocol version.
---

The `HELLO` command in DiceDB is used to initiate a connection handshake with the server. This command is particularly useful for clients to negotiate the protocol version and authentication details. It can also be used to switch the connection to a different protocol version.

## Parameters

### Required Parameters

- `protover`: (integer) The protocol version to use. DiceDB supports protocol versions 2 and 3.

### Optional Parameters

- `AUTH`: (string) The keyword to specify authentication details.
  - `username`: (string) The username for authentication. This is required if the `AUTH` keyword is used.
  - `password`: (string) The password for authentication. This is required if the `AUTH` keyword is used.
- `SETNAME`: (string) The keyword to specify a client name.
  - `clientname`: (string) The name to assign to the client connection.

## Return Value

The `HELLO` command returns a map with the following fields:

- `server`: (string) The server type, typically "DiceDB".
- `version`: (string) The DiceDB server version.
- `proto`: (integer) The protocol version in use.
- `id`: (integer) The client ID.
- `mode`: (string) The server mode, typically "standalone".
- `role`: (string) The role of the server, either "master" or "slave".
- `modules`: (array) A list of loaded modules.

## Behaviour

When the `HELLO` command is issued:

1. The server will switch to the specified protocol version.
2. If authentication details are provided, the server will attempt to authenticate the client.
3. If a client name is provided, the server will set the client name.
4. The server will return a map containing details about the server and the connection.

## Error Handling

The `HELLO` command can raise the following errors:

- `ERR wrong number of arguments for 'hello' command`: This error occurs if the command is issued with an incorrect number of arguments.
- `ERR invalid protocol version`: This error occurs if the specified protocol version is not supported.
- `WRONGPASS invalid username-password pair`: This error occurs if the provided authentication details are incorrect.
- `ERR Client sent AUTH, but no password is set`: This error occurs if the `AUTH` keyword is used but no password is set on the server.

## Example Usage

### Basic Usage

Switch to protocol version 3:

```bash
127.0.0.1:7379> HELLO 3
```

### Usage with Authentication

Switch to protocol version 3 and authenticate with username and password:

```bash
127.0.0.1:7379> HELLO 3 AUTH myusername mypassword
```

### Usage with Client Name

Switch to protocol version 3 and set the client name:

```bash
127.0.0.1:7379> HELLO 3 SETNAME myclientname
```

### Combined Usage

Switch to protocol version 3, authenticate, and set the client name:

```bash
127.0.0.1:7379> HELLO 3 AUTH myusername mypassword SETNAME myclientname
```

## Notes

- The `HELLO` command is particularly useful for clients that need to ensure they are using a specific protocol version.
- It is recommended to use the `HELLO` command at the beginning of a connection to set the desired protocol version and authentication details.
- The `HELLO` command can be used to switch between RESP2 and RESP3 protocols dynamically.

By understanding and utilizing the `HELLO` command, clients can effectively manage their connection settings and ensure proper communication with the DiceDB server.
