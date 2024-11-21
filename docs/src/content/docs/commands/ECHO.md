---
title: ECHO
description: The `ECHO` command in DiceDB is used to print a message
---

### Syntax

```bash
ECHO message

```

## Parameters

| Parameter | Description                                                | Type            | Required |
| --------- | ---------------------------------------------------------- | --------------- | -------- |
| `message` | A string of characters, numbers, or a mix of both to print | String / Number | Yes      |

## Return values

| Condition                                   | Return Value                      |
| ------------------------------------------- | --------------------------------- |
| Command is successful                       | The message passed as a parameter |
| Syntax or specified constraints are invalid | error                             |

## Errors
1. `Syntax Error`:
   - Error Message: `(error) ERR wrong number of arguments for 'echo' command`
   - Occurs if the command is called with additional or fewer parameters than required

## Example Usage

### Valid usage

```bash
127.0.0.1:7379> ECHO "DiceDB is very efficient"
"DiceDB is very efficient"
```

### Invalid Usage

```bash
127.0.0.1:7379> ECHO
`(error) ERROR wrong number of arguments for 'echo' command`
```


```bash
127.0.0.1:7379> ECHO "DiceDB is" "very efficient"
(error) ERROR wrong number of arguments for 'echo' command`
```

## Conclusion

In DiceDB, the ECHO command accepts only one message string and prints it. If no message or more than one message is provided, it results in an error
