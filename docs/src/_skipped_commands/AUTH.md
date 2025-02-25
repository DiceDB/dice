---
title: AUTH
description: The `AUTH` command is a DiceDB command that enables you to authenticate a client connecting to the DiceDB server using a password. This command is particularly important in environments where the DiceDB instance is secured with access control measures, ensuring that only authorized users can interact with the server and its data.
---

The `AUTH` command is a DiceDB command that enables you to authenticate a client connecting to the DiceDB server using a password. This command is particularly important in environments where the DiceDB instance is secured with access control measures, ensuring that only authorized users can interact with the server and its data.

## Syntax

```bash
AUTH password
```

## Parameters

| Parameter  | Description                                           | Type   | Required |
| ---------- | ----------------------------------------------------- | ------ | -------- |
| `password` | The password that you have set for the DiceDB server. | String | Yes      |

## Return values

| Condition                                   | Return Value |
| ------------------------------------------- | ------------ |
| Authentication is successful                | `OK`         |
| Syntax or specified constraints are invalid | error        |

## Behaviour

When the `AUTH` command is executed, DiceDB will perform the following steps:

1. `Check Authentication Status:` DiceDB will check if the server is configured to require authentication (i.e., if a password is set in the DiceDB configuration).
2. `Validate Password:` If authentication is required, DiceDB will then verify the provided password against the configured password.
3. `Return Response:`
   - If the password is valid, DiceDB will store the authenticated session for that client.
   - If the password is invalid, an error message will be returned.

Once authenticated, the client can execute commands on the DiceDB server that require authentication. If a client attempts to send commands without being authenticated (and while authentication is required), DiceDB will respond with an error for unauthorized access.

## Errors

1. `(error) WRONGPASS invalid username-password pair or user is disabled`: This error occurs if the password provided to the `AUTH` command does not match the configured password of the DiceDB server. The client will not be granted access unless the correct password is provided.

2. `(error) NOAUTH Authentication required`: This error occurs when a client attempts to execute a command without authenticating first. The client must use the `AUTH` command to authenticate before executing any other commands.

3. `(error) ERR AUTH <password> called without any password configured for the default user. Are you sure your configuration is correct?`: This error will be raised if the DiceDB server is not configured to use a password (i.e., the `requirepass` directive in the DiceDB.conf file is empty or commented out), and a client attempts to authenticate using the `AUTH` command.

## Example Usage

### Successful Authentication

In this example, after successfully typing the correct password, the DiceDB server responds with `OK`, indicating that the client is now authenticated and can execute further commands.

```bash
127.0.0.1:7379> AUTH your_secret_password
OK
```

### Failed Authentication

In this example, the provided password is incorrect. As a result, the DiceDB server responds with an error indicating that the authentication has failed.

```bash
127.0.0.1:7379> AUTH incorrect_password
(error) WRONGPASS invalid username-password pair or user is disabled
```

### Invoking commands without AUTH

In this example, when some other command is fired without doing `AUTH` then the above error is thrown denoting that authentication is required.

```bash
127.0.0.1:7379> GET x
(error) NOAUTH Authentication required
```

### Invoking AUTH without configuring password

In this example, when `AUTH` is fired without actually configuring the password for the DiceDB server, then the server returns an error indicating that no password is set or the configuration is incorrect.

```bash
127.0.0.1:7379> AUTH incorrect_password
(error) ERR AUTH <password> called without any password configured for the default user. Are you sure your configuration is correct?
```

### Invalid usage

```bash
127.0.0.1:7379> AUTH your_secret_password foo bar
(error) ERR wrong number of arguments for 'auth' command
```

### Additional Notes

- The `AUTH` command should be used cautiously, especially over unsecured connections, as plaintext passwords may be transmitted. Using TLS or other secure methods is recommended for production environments.
- It’s possible to configure DiceDB to allow multiple clients to connect concurrently from different IP addresses. Each client needs to authenticate separately.
- If you change the password in the DiceDB configuration, all connected clients must re-authenticate with the new password.
