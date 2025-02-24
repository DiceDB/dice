---
title: GEOPOS
description: The `GEOPOS` command in DiceDB is used to return the longitude, latitude to a specified key, as stored in the sorted set.
---

The `GEOPOS` command in DiceDB is used to return the longitude, latitude to a specified key which is stored in a sorted set. When elements are added via `GEOADD` then they are stored in 52 bit geohash hence the values returned by `GEOPOS` might have small margins of error.

## Syntax

```bash
GEOPOS key [member [member ...]]
```

## Parameters

| Parameter | Description                                                                  | Type   | Required |
| --------- | ---------------------------------------------------------------------------- | ------ | -------- |
| key       | The name of the sorted set key whose member's coordinates are to be returned | string | Yes      |
| member    | A unique identifier for the location.                                        | string | Yes      |

## Return Values

| Condition                                            | Return Value                                                                           |
| ---------------------------------------------------- | -------------------------------------------------------------------------------------- |
| Coordinates exist for the specified member(s)        | Returns an ordered list of coordinates (longitude, latitude) for each specified member |
| Coordinates do not exist for the specified member(s) | Returns `(nil)` for each member without coordinates                                    |
| Incorrect Argument Count                             | `ERR wrong number of arguments for 'geopos' command`                                   |
| Key does not exist in the sorted set                 | `Error: nil`                                                                           |

## Behaviour

When the GEOPOS command is issued, DiceDB performs the following steps:

1. It checks if argument count is valid or not. If not an error is thrown.
2. It checks the validity of the key.
3. If the key is invalid then an error is returned.
4. Else it checks the members provided after the key.
5. For each member it checks the coordinates of the member.
6. If the coordinates exist then it is returned in an ordered list of latitude, longitude for the particular member.
7. If the coordinates do not exist then a `(nil)` is returned for that member.

## Errors

1. `Wrong number of arguments for 'GEOPOS' command`
   - Error Message: (error) ERR wrong number of arguments for 'geoadd' command.
   - Occurs when the command is executed with an incorrect number of arguments.
2. `Wrong key for 'GEOPOS' command`
   - Error Message: Error: nil
   - Occurs when the command is executed with a key that does not exist in the sorted set.

## Example Usage

Here are a few examples demonstrating the usage of the GEOPOS command:

### Example: Fetching the latitude, longitude of an existing member of the set

```bash
127.0.0.1:7379> GEOADD Sicily 13.361389 38.115556 "Palermo" 15.087269 37.502669 "Catania"
2
127.0.0.1:7379> GEOPOS Sicily "Palermo"
1) 1) 13.361387
2) 38.115556
```

### Example: Fetching the latitude, longitude of a member not in the set

```bash
127.0.0.1:7379> GEOPOS Sicily "Agrigento"
1) (nil)
```
