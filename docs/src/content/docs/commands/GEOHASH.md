---
title: GEOHASH
description: The GEOHASH command in DiceDB returns the Geohash representation of specified members stored in a geospatial dataset under a specified key. This allows efficient encoding and querying of geographic points.
---

## Syntax

```bash
GEOHASH key member [member ...]
```

## Parameters

| Parameter | Description                                                      | Type   | Required |
| --------- | ---------------------------------------------------------------- | ------ | -------- |
| key       | The name of the sorted set containing geospatial data.           | string | Yes      |
| member    | One or more members whose Geohash representations are requested. | string | Yes      |

## Return Values

| Condition                                 | Return Value                                                        |
| ----------------------------------------- | ------------------------------------------------------------------- |
| Geohash representation for valid members. | A string representing the Geohash of each member.                   |
| Non-existent member in the specified key. | `(nil)` for that member.                                            |
| Incorrect Argument Count                  | `ERR wrong number of arguments for 'geohash' command`               |
| Non-existent or invalid key type.         | `WRONGTYPE Operation against a key holding the wrong kind of value` |

## Behaviour

When the GEOHASH command is issued, DiceDB performs the following steps:

1. Checks if the key exists and corresponds to a valid geospatial dataset.
2. Verifies the presence of each specified member within the dataset.
3. For each member:
   - If the member exists, its Geohash representation is returned.
   - If the member does not exist, `(nil)` is returned.
4. Returns the results as an array of strings or nil values, maintaining the order of input members.

## Errors

1. **Wrong Number of Arguments**

   - **Error Message:** (error) ERR wrong number of arguments for 'geohash' command
   - Occurs when the command is executed without a key or member(s).

2. **Key Does Not Exist or Is of Wrong Type**

   - **Error Message:** (error) WRONGTYPE Operation against a key holding the wrong kind of value
   - Occurs when the specified key does not exist or is not a sorted set.

3. **Member Does Not Exist**
   - **Error Message:** Returns `(nil)` for non-existent members.
   - Occurs when a member is specified but not found under the given key.

## Example Usage

Here are a few examples demonstrating the usage of the GEOHASH command:

### Example: Retrieve Geohash for Existing Members

```bash
127.0.0.1:7379> GEOADD locations 13.361389 38.115556 "Palermo" 15.087269 37.502669 "Catania"
2
127.0.0.1:7379> GEOHASH locations Palermo Catania
1) "sqc8b49rny"
2) "sq9sm17147"
```

### Example: Retrieve Geohash for a Non-Existent Member

```bash
127.0.0.1:7379> GEOHASH locations Venice
(nil)
```

### Example: Retrieve Geohash with a Non-Existent Key

```bash
127.0.0.1:7379> GEOHASH points Palermo
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```

### Example: Retrieve Geohash with Missing Arguments

```bash
127.0.0.1:7379> GEOHASH
(error) ERR wrong number of arguments for 'geohash' command
```

## Notes

- The returned Geohash strings are encoded representations of the geographic locations stored in the sorted set.
- Geohash precision is influenced by the DiceDB implementation and is typically sufficient for most spatial queries.
