---
title: LATENCY
description: The `LATENCY` command in DiceDB is used to measure and analyze the latency of various operations within the DiceDB server. This command provides insights into the performance and responsiveness of the DiceDB instance, helping administrators to identify and troubleshoot latency issues.
---

Note: `Not yet supported`

The `LATENCY` command in DiceDB is used to measure and analyze the latency of various operations within the DiceDB server. This command provides insights into the performance and responsiveness of the DiceDB instance, helping administrators to identify and troubleshoot latency issues.

## Command Syntax

```bash
LATENCY [SUBCOMMAND] [ARGUMENTS]
```

## Parameters

| Parameter    | Subcommand | Description                                                           | Return Type | Required |
| ------------ | ---------- | --------------------------------------------------------------------- | ----------- | -------- |
| `SUBCOMMAND` |            | Specifies the operation to perform. Available subcommands are:        | String      | Yes      |
|              | `LATEST`   | Returns the latest latency spikes.                                    | Array       | Yes      |
|              | `HISTORY`  | Returns the latency history for a specific event.                     | Array       | Yes      |
|              | `RESET`    | Resets the latency data for specific events or all events.            | Integer     | Yes      |
|              | `GRAPH`    | Returns a latency graph for a specific event.                         | String      | Yes      |
|              | `DOCTOR`   | Provides a human-readable report of latency issues.                   | String      | Yes      |
| `ARGUMENTS`  | N/A        | Additional arguments required by certain subcommands (if applicable). | Varies      | No       |

## Return Values

| Condition                                     | Return Value                                  |
| --------------------------------------------- | --------------------------------------------- |
| `LATEST`                                      | Array of latency spikes                       |
| `HISTORY`                                     | Array of latency samples (timestamp, latency) |
| `RESET`                                       | Integer (number of events reset)              |
| `GRAPH`                                       | String (graph representing latency data)      |
| `DOCTOR`                                      | String (detailed report with suggestions)     |
| `Syntax or specified constraints are invalid` | error                                         |

## Behaviour

When the `LATENCY` command is executed, it performs the specified subcommand operation. The behavior varies based on the subcommand:

- `LATEST`: Fetches and returns the most recent latency spikes recorded by the DiceDB server.
- `HISTORY`: Retrieves the historical latency data for a specified event, allowing for analysis of latency trends over time.
- `RESET`: Clears the latency data for specified events or all events, effectively resetting the latency monitoring.
- `GRAPH`: Generates and returns a visual representation of the latency data for a specified event.
- `DOCTOR`: Analyzes the latency data and provides a human-readable report with potential causes and suggestions for mitigation.

## Errors

- `Invalid Subcommand`:

  - Error Message: `ERR unknown subcommand 'subcommand'`
  - If an unrecognized subcommand is provided, DiceDB will return an error.

- `Missing Arguments`:

  - Error Message: `ERR wrong number of arguments for 'latency subcommand' command`
  - If required arguments for a subcommand are missing, DiceDB will return an error.

- `Invalid Event Name`:

  - Error Message: `ERR no such event 'event_name'`
  - If an invalid event name is provided for subcommands like `HISTORY`, `RESET`, or `GRAPH`, DiceDB will return an error.

## Example Usage

### LATEST Subcommand

Fetches the latest latency spikes recorded by the DiceDB server.

```bash
127.0.0.1:7379> LATENCY LATEST
1) 1) "command"
   2) (integer) 1633024800
   3) (integer) 15
2) 1) "fork"
   2) (integer) 1633024805
   3) (integer) 25
```

### HISTORY Subcommand

Retrieves the historical latency data for the `command` event.

```bash
127.0.0.1:7379> LATENCY HISTORY command
1) 1) (integer) 1633024800
   2) (integer) 15
2) 1) (integer) 1633024805
   2) (integer) 25
```

### RESET Subcommand

Resets the latency data for the `command` event.

```bash
127.0.0.1:7379> LATENCY RESET command
(integer) 1
```

### GRAPH Subcommand

Generates a visual representation of the latency data for the `command` event.

```bash
127.0.0.1:7379> LATENCY GRAPH command
| 15 | 25 |
```

### DOCTOR Subcommand

Provides a human-readable report of latency issues.

```bash
127.0.0.1:7379> LATENCY DOCTOR
Latency Doctor Report:
- Command latency spikes detected. Consider optimizing your commands.
- Fork latency spikes detected. Check your system's I/O performance.
```
