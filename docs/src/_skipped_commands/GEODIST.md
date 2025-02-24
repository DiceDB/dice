---
title: GEODIST
description: The `GEODIST` command in DiceDB is used to calculate the distance between two members (geospatial points) stored in a geospatial index(set).
---

The `GEODIST` command in DiceDB is used to calculate the distance between two members (geospatial points) stored in a geospatial index(set).

## Syntax

```bash
GEODIST key member1 member2 [m | km | ft | mi]
```

## Parameters

| Parameter | Description                                                          | Type   | Required |
| --------- | -------------------------------------------------------------------- | ------ | -------- |
| key       | The name of the sorted set where the geospatial data is stored.      | string | Yes      |
| member1   | The name of the member1 from where you want to measure the distance. | string | Yes      |
| member2   | The name of the member2 to where you want to measure the distance.   | string | Yes      |
| m         | The distance to be measured in meters.                               | NONE   | NO       |
| km        | The distance to be measured in kilometers.                           | NONE   | NO       |
| ft        | The distance to be measured in feet.                                 | NONE   | NO       |
| mi        | The distance to be measured in miles.                                | NONE   | NO       |

## Return Values

| Condition                                       | Return Value                                          |
| ----------------------------------------------- | ----------------------------------------------------- |
| If both members exist in the set with no option | distance b/w them in meters                           |
| If both members exist in the set with option km | distance b/w them in kilometers                       |
| If both members exist in the set with option ft | distance b/w them in feet                             |
| If both members exist in the set with option mi | distance b/w them in miles                            |
| If any member doesn't exist in Set              | nil                                                   |
| Incorrect Argument Count                        | `ERR wrong number of arguments for 'geodist' command` |

## Behaviour

When the GEODIST command is issued, DiceDB performs the following steps:

1. It gets the sorted set(key).
2. It gets the scores(geohashes) from the sorted sets for both the members.
3. It calculates the distance bw them and returns it.

## Errors

1.`Wrong number of arguments for 'GEODIST' command`

- Error Message: (error) ERR wrong number of arguments for 'geodist' command.
- Occurs when the command is executed with an incorrect number of arguments.

## Example Usage

Here are a few examples demonstrating the usage of the GEODIST command:

### Example : Adding new member to a set

```bash
127.0.0.1:7379> GEOADD cities -74.0060 40.7128 "New York"
1
127.0.0.1:7379> GEOADD cities -79.3470 43.6510 "Toronto"
1
127.0.0.1:7379> GEODIST cities "New York" "Toronto"
"548064.1868"
127.0.0.1:7379> GEODIST cities "New York" "Toronto km"
"548.0642"
127.0.0.1:7379> GEODIST cities "New York" "Toronto mi"
"340.5521"
```
