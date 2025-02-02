---
title: JSON.RESP
description: Documentation for the DiceDB command JSON.RESP
---

The `JSON.RESP` command in DiceDB is part of the DiceDBJSON module, which returns the JSON in the specified key in RESP.

## Syntax

```bash
JSON.RESP <key> [path]
```

## Parameters

| Parameter | Description                                                                                         | Type   | Required |
| --------- | --------------------------------------------------------------------------------------------------- | ------ | -------- |
| `key`     | The key under which the JSON data is stored.                                                        | String | Yes      |
| `path`    | The JSON path to the specific part of the JSON data to fetch. Defaults to the root if not provided. | String | No       |

## Return Value

| Condition             | Return Value                                   |
| --------------------- | ---------------------------------------------- |
| if `path` is provided | JSON value at the specified path in RESP form. |

## Behaviour

- If the path is not provided, it defaults to the root of the JSON data.

## Errors

1. `Wrong Type`:
   - Error Message: `WRONGTYPE Operation against a key holding the wrong kind of value`
   - If the key exists but does not hold JSON data, DiceDB will return an error.

## Example Usage

### JSON.RESP on array

The `JSON.RESP` command returns the entire JSON data stored under the key `arrayjson` in RESP form.

```bash
127.0.0.1:7379> JSON.SET arrayjson $ '["dice",10,10.5,true,null]'
OK
127.0.0.1:7379> JSON.RESP
1) [
2) "dice"
3) (integer) 10
4) "10.5"
5) true
6) (nil)
```

### JSON.RESP on nested JSON

The `JSON.RESP` command returns the JSON data stored under the key `myjson` at `$.b` in RESP form.

```bash
127.0.0.1:7379> JSON.SET myjson $ '{"a":100,"b":["dice",10,10.5,true,null]}'
OK
127.0.0.1:7379> JSON.RESP myjson $.b
1) 1) [
   2) "dice"
   3) (integer) 10
   4) "10.5"
   5) true
   6) (nil)
```

## Notes

- JSONPath expressions are used to navigate and specify the location within the JSON document. Familiarity with JSONPath syntax is beneficial for effective use of this command.
