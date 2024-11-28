---
title: BITFIELD_RO
description: Read-only variant of the BITFIELD command. It is like the original BITFIELD but only accepts GET subcommand.
---

## Syntax

```bash
BITFIELD_RO key [GET type offset [GET type offset ...]]
```

## Parameters

| Parameter         | Description                                                                                                               | Type   | Required |
| ----------------- | ------------------------------------------------------------------------------------------------------------------------- | ------ | -------- |
| `key`             | The name of the key containing the bitfield.                                                                              | String | Yes      |
| `GET type offset` | Retrieves bits starting at the specified offset with the specified type. Type defines the signed/unsigned integer format. | String | Optional |

## Return values

| Condition                                   | Return Value                                         |
| ------------------------------------------- | ---------------------------------------------------- |
| Command is successful                       | Array of results corresponding to each `GET` command |
| Syntax or specified constraints are invalid | error                                                |

## Behaviour

- Read-only variant of the BITFIELD command. It is like the BITFIELD but only accepts GET subcommand.
- See [BITFIELD](/commands/bitfield) for more details.

## Example Usage

### Basic Usage

```bash
127.0.0.1:7379> SET hello "Hello World"
OK
127.0.0.1:7379> BITFIELD_RO hello GET i8 16
1) "108"
```
