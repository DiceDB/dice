---
title: LATENCY
description: The `LATENCY` command in DiceDB is used to measure and analyze the latency of various operations within the DiceDB server. This command provides insights into the performance and responsiveness of the DiceDB instance, helping administrators to identify and troubleshoot latency issues.
---

The `LATENCY` command in DiceDB is used to measure and analyze the latency of various operations within the DiceDB server. This command provides insights into the performance and responsiveness of the DiceDB instance, helping administrators to identify and troubleshoot latency issues.

## Command Syntax

```plaintext
LATENCY [SUBCOMMAND] [ARGUMENTS]
```

## Parameters

- `SUBCOMMAND`: The specific operation to perform. The available subcommands are:

  - `LATEST`: Returns the latest latency spikes.
  - `HISTORY`: Returns the latency history for a specific event.
  - `RESET`: Resets the latency data for specific events or all events.
  - `GRAPH`: Returns a latency graph for a specific event.
  - `DOCTOR`: Provides a human-readable report of latency issues.

- `ARGUMENTS`: Additional arguments required by certain subcommands. The arguments vary based on the subcommand used.

## Return Value

The return value of the `LATENCY` command depends on the subcommand used:

- `LATEST`: Returns an array of the latest latency spikes.
- `HISTORY`: Returns an array of latency samples for a specific event.
- `RESET`: Returns the number of events reset.
- `GRAPH`: Returns a string representing the latency graph.
- `DOCTOR`: Returns a string with a human-readable latency report.

## Behaviour

When the `LATENCY` command is executed, it performs the specified subcommand operation. The behavior varies based on the subcommand:

- `LATEST`: Fetches and returns the most recent latency spikes recorded by the DiceDB server.
- `HISTORY`: Retrieves the historical latency data for a specified event, allowing for analysis of latency trends over time.
- `RESET`: Clears the latency data for specified events or all events, effectively resetting the latency monitoring.
- `GRAPH`: Generates and returns a visual representation of the latency data for a specified event.
- `DOCTOR`: Analyzes the latency data and provides a human-readable report with potential causes and suggestions for mitigation.

## Error Handling

Errors may be raised in the following scenarios:

- `Invalid Subcommand`: If an unrecognized subcommand is provided, DiceDB will return an error.

  - Error Message: `ERR unknown subcommand 'subcommand'`

- `Missing Arguments`: If required arguments for a subcommand are missing, DiceDB will return an error.

  - Error Message: `ERR wrong number of arguments for 'latency subcommand' command`

- `Invalid Event Name`: If an invalid event name is provided for subcommands like `HISTORY`, `RESET`, or `GRAPH`, DiceDB will return an error.

  - Error Message: `ERR no such event 'event_name'`

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
