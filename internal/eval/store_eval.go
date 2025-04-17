// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package eval

import (
	"encoding/base64"
	"errors"
	"fmt"
	"math"
	"math/bits"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unsafe"

	"github.com/axiomhq/hyperloglog"
	"github.com/bytedance/sonic"
	"github.com/dicedb/dice/internal/cmd"
	diceerrors "github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/eval/geo"
	"github.com/dicedb/dice/internal/eval/sortedset"
	"github.com/dicedb/dice/internal/object"
	"github.com/dicedb/dice/internal/server/utils"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/gobwas/glob"
	"github.com/ohler55/ojg/jp"
	"github.com/rs/xid"
)

// evalHEXISTS returns if field is an existing field in the hash stored at key.
//
// This command returns 0, if the specified field doesn't exist in the key and 1 if it exists.
//
// If key doesn't exist, it returns 0.
//
// Usage: HEXISTS key field
func evalHEXISTS(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 2 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("HEXISTS"),
		}
	}

	key := args[0]
	hmKey := args[1]
	obj := store.Get(key)

	var hashMap HashMap

	if obj == nil {
		return &EvalResponse{
			Result: IntegerZero,
			Error:  nil,
		}
	}
	if err := object.AssertType(obj.Type, object.ObjTypeSSMap); err != nil {
		return &EvalResponse{
			Error:  diceerrors.ErrGeneral(diceerrors.WrongTypeErr),
			Result: nil,
		}
	}

	hashMap = obj.Value.(HashMap)

	_, ok := hashMap.Get(hmKey)
	if ok {
		return &EvalResponse{
			Result: IntegerOne,
			Error:  nil,
		}
	}
	// Return 0, if specified field doesn't exist in the HashMap.
	return &EvalResponse{
		Result: IntegerZero,
		Error:  nil,
	}
}

// evalHKEYS is used to retrieve all the keys(or field names) within a hash.
//
// This command returns empty array, if the specified key doesn't exist.
//
// Complexity is O(n) where n is the size of the hash.
//
// Usage: HKEYS key
func evalHKEYS(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 1 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("HKEYS"),
		}
	}

	key := args[0]
	obj := store.Get(key)

	var hashMap HashMap
	var result []string

	if obj != nil {
		if err := object.AssertType(obj.Type, object.ObjTypeSSMap); err != nil {
			return &EvalResponse{
				Error:  diceerrors.ErrGeneral(diceerrors.WrongTypeErr),
				Result: nil,
			}
		}
		hashMap = obj.Value.(HashMap)
	} else {
		return &EvalResponse{
			Result: EmptyArray,
			Error:  nil,
		}
	}

	for hmKey := range hashMap {
		result = append(result, hmKey)
	}

	return &EvalResponse{
		Result: result,
		Error:  nil,
	}
}

// evalHKEYS is used to retrieve all the values within a hash.
//
// This command returns empty array, if the specified key doesn't exist.
//
// Complexity is O(n) where n is the size of the hash.
//
// Usage: HVALS key
func evalHVALS(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 1 {
		return &EvalResponse{Error: diceerrors.ErrWrongArgumentCount("HVALS"), Result: nil}
	}

	key := args[0]
	obj := store.Get(key)

	if obj == nil {
		// Return an empty array for non-existent keys
		return &EvalResponse{
			Result: EmptyArray,
			Error:  nil,
		}
	}

	if err := object.AssertType(obj.Type, object.ObjTypeSSMap); err != nil {
		return &EvalResponse{
			Error:  diceerrors.ErrGeneral(diceerrors.WrongTypeErr),
			Result: nil,
		}
	}

	hashMap := obj.Value.(HashMap)
	results := make([]string, 0, len(hashMap))

	for _, value := range hashMap {
		results = append(results, value)
	}

	return &EvalResponse{
		Result: results,
		Error:  nil,
	}
}

// Key, start and end are mandatory args.
// Returns a substring from the key(if it's a string) from start -> end.
// Returns ""(empty string) if key is not present and if start > end.
func evalGETRANGE(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 3 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("GETRANGE"),
		}
	}

	key := args[0]
	obj := store.Get(key)
	if obj == nil {
		return &EvalResponse{
			Result: string(""),
			Error:  nil,
		}
	}

	start, err := strconv.Atoi(args[1])
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrIntegerOutOfRange,
		}
	}
	end, err := strconv.Atoi(args[2])
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrIntegerOutOfRange,
		}
	}

	var str string
	switch oType := obj.Type; oType {
	case object.ObjTypeString:
		if val, ok := obj.Value.(string); ok {
			str = val
		} else {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrGeneral("expected string but got another type"),
			}
		}
	case object.ObjTypeInt:
		str = strconv.FormatInt(obj.Value.(int64), 10)
	case object.ObjTypeByteArray:
		if val, ok := obj.Value.(*ByteArray); ok {
			str = string(val.data)
		} else {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongTypeOperation,
			}
		}
	default:
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	if str == "" {
		return &EvalResponse{
			Result: string(""),
			Error:  nil,
		}
	}

	if start < 0 {
		start = len(str) + start
	}

	if end < 0 {
		end = len(str) + end
	}

	if start >= len(str) || end < 0 || start > end {
		return &EvalResponse{
			Result: string(""),
			Error:  nil,
		}
	}

	if start < 0 {
		start = 0
	}

	if end >= len(str) {
		end = len(str) - 1
	}

	return &EvalResponse{
		Result: str[start : end+1],
		Error:  nil,
	}
}

// evalZADD adds all the specified members with the specified scores to the sorted set stored at key.
// If a specified member is already a member of the sorted set, the score is updated and the element
// reinserted at the right position to ensure the correct ordering.
// If key does not exist, a new sorted set with the specified members as sole members is created.
func evalZADD(args []string, store *dstore.Store) *EvalResponse {
	// if length of command is 3, throw error as it is not possible
	if len(args) < 3 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("ZADD"),
		}
	}
	key := args[0]
	sortedSet, err := getOrCreateSortedSet(store, key)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  err,
		}
	}
	// flags parsing
	flags, nextIndex := parseFlags(args[1:])
	if nextIndex >= len(args) || (len(args)-nextIndex)%2 != 0 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("ZADD"),
		}
	}
	// only valid flags works
	if err := validateFlagsAndArgs(args[nextIndex:], flags); err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  err,
		}
	}
	// all processing takes place here
	return processMembersWithFlags(args[nextIndex:], sortedSet, store, key, flags)
}

// parseFlags identifies and parses the flags used in ZADD.
func parseFlags(args []string) (parsedFlags map[string]bool, nextIndex int) {
	parsedFlags = map[string]bool{
		NX:   false,
		XX:   false,
		LT:   false,
		GT:   false,
		CH:   false,
		INCR: false,
	}
	for i := 0; i < len(args); i++ {
		switch strings.ToUpper(args[i]) {
		case NX:
			parsedFlags[NX] = true
		case XX:
			parsedFlags[XX] = true
		case LT:
			parsedFlags[LT] = true
		case GT:
			parsedFlags[GT] = true
		case CH:
			parsedFlags[CH] = true
		case INCR:
			parsedFlags[INCR] = true
		default:
			return parsedFlags, i + 1
		}
	}

	return parsedFlags, len(args) + 1
}

// only valid combination of options works
func validateFlagsAndArgs(args []string, flags map[string]bool) error {
	if len(args)%2 != 0 {
		return diceerrors.ErrGeneral("syntax error")
	}
	if flags[NX] && flags[XX] {
		return diceerrors.ErrGeneral("XX and NX options at the same time are not compatible")
	}
	if (flags[GT] && flags[NX]) || (flags[LT] && flags[NX]) || (flags[GT] && flags[LT]) {
		return diceerrors.ErrGeneral("GT, LT, and/or NX options at the same time are not compatible")
	}
	if flags[INCR] && len(args)/2 > 1 {
		return diceerrors.ErrGeneral("INCR option supports a single increment-element pair")
	}
	return nil
}

// processMembersWithFlags processes the members and scores while handling flags.
func processMembersWithFlags(args []string, sortedSet *sortedset.Set, store *dstore.Store, key string, flags map[string]bool) *EvalResponse {
	added, updated := 0, 0

	for i := 0; i < len(args); i += 2 {
		scoreStr := args[i]
		member := args[i+1]

		score, err := strconv.ParseFloat(scoreStr, 64)
		if err != nil || math.IsNaN(score) {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrInvalidNumberFormat,
			}
		}

		currentScore, exists := sortedSet.Get(member)

		// If INCR is used, increment the score first
		if flags[INCR] {
			if exists {
				score += currentScore
			} else {
				score = 0.0 + score
			}

			// Now check GT and LT conditions based on the incremented score and return accordingly
			if (flags[GT] && exists && score <= currentScore) ||
				(flags[LT] && exists && score >= currentScore) {
				return &EvalResponse{
					Result: nil,
					Error:  nil,
				}
			}
		}

		// Check if the member should be skipped based on NX or XX flags
		if shouldSkipMember(score, currentScore, exists, flags) {
			continue
		}

		// Insert or update the member in the sorted set
		wasInserted := sortedSet.Upsert(score, member)

		if wasInserted && !exists {
			added++
		} else if exists && score != currentScore {
			updated++
		}

		// If INCR is used, exit after processing one score-member pair
		if flags[INCR] {
			return &EvalResponse{
				Result: score,
				Error:  nil,
			}
		}
	}

	// Store the updated sorted set in the store
	storeUpdatedSet(store, key, sortedSet)

	if flags[CH] {
		return &EvalResponse{
			Result: added + updated,
			Error:  nil,
		}
	}

	// Return only the count of added members
	return &EvalResponse{
		Result: added,
		Error:  nil,
	}
}

// shouldSkipMember determines if a member should be skipped based on flags.
func shouldSkipMember(score, currentScore float64, exists bool, flags map[string]bool) bool {
	useNX, useXX, useLT, useGT := flags[NX], flags[XX], flags[LT], flags[GT]

	return (useNX && exists) || (useXX && !exists) ||
		(exists && useLT && score >= currentScore) ||
		(exists && useGT && score <= currentScore)
}

// storeUpdatedSet stores the updated sorted set in the store.
func storeUpdatedSet(store *dstore.Store, key string, sortedSet *sortedset.Set) {
	store.Put(key, store.NewObj(sortedSet, -1, object.ObjTypeSortedSet), dstore.WithPutCmd(dstore.ZAdd))
}

// getOrCreateSortedSet fetches the sorted set if it exists, otherwise creates a new one.
func getOrCreateSortedSet(store *dstore.Store, key string) (*sortedset.Set, error) {
	obj := store.Get(key)
	if obj != nil {
		sortedSet, err := sortedset.FromObject(obj)
		if err != nil {
			return nil, diceerrors.ErrWrongTypeOperation
		}
		return sortedSet, nil
	}
	return sortedset.New(), nil
}

// The ZCOUNT command in DiceDB counts the number of members in a sorted set at the specified key
// whose scores fall within a given range. The command takes three arguments: the key of the sorted set
// the minimum score, and the maximum score.
func evalZCOUNT(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 3 {
		// 1. Check no of arguments
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("ZCOUNT"),
		}
	}

	key := args[0]
	minArg := args[1]
	maxArg := args[2]

	// 2. Parse the min and max score arguments
	minValue, errMin := strconv.ParseFloat(minArg, 64)
	maxValue, errMax := strconv.ParseFloat(maxArg, 64)
	if errMin != nil || errMax != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrInvalidNumberFormat,
		}
	}

	// 3. Retrieve the object from the store
	obj := store.Get(key)
	if obj == nil {
		// If the key does not exist, return 0 (no error)
		return &EvalResponse{
			Result: 0,
			Error:  nil,
		}
	}

	// 4. Ensure the object is a valid sorted set
	var sortedSet *sortedset.Set
	var err []byte
	sortedSet, err = sortedset.FromObject(obj)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	count := sortedSet.CountInRange(minValue, maxValue)

	return &EvalResponse{
		Result: count,
		Error:  nil,
	}
}

// evalZRANGE returns the specified range of elements in the sorted set stored at key.
// The elements are considered to be ordered from the lowest to the highest score.
func evalZRANGE(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 3 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("ZRANGE"),
		}
	}

	key := args[0]
	startStr := args[1]
	stopStr := args[2]

	withScores := false
	reverse := false
	for i := 3; i < len(args); i++ {
		arg := strings.ToUpper(args[i])
		if arg == WithScores {
			withScores = true
		} else if arg == REV {
			reverse = true
		} else {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrInvalidSyntax("ZRANGE"),
			}
		}
	}

	start, err := strconv.Atoi(startStr)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrInvalidNumberFormat,
		}
	}

	stop, err := strconv.Atoi(stopStr)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrInvalidNumberFormat,
		}
	}

	obj := store.Get(key)
	if obj == nil {
		return &EvalResponse{
			Result: []string{},
			Error:  nil,
		}
	}

	sortedSet, errMsg := sortedset.FromObject(obj)

	if errMsg != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	result := sortedSet.GetRange(start, stop, withScores, reverse)

	return &EvalResponse{
		Result: result,
		Error:  nil,
	}
}

// evalZREM removes the specified members from the sorted set stored at key.
// Non-existing members are ignored.
// Returns the number of members removed.
func evalZREM(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 2 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("ZREM"),
		}
	}

	key := args[0]
	obj := store.Get(key)

	if obj == nil {
		return &EvalResponse{
			Result: IntegerZero,
			Error:  nil,
		}
	}

	sortedSet, err := sortedset.FromObject(obj)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	countRem := 0
	for i := 1; i < len(args); i++ {
		if sortedSet.Remove(args[i]) {
			countRem += 1
		}
	}

	return &EvalResponse{
		Result: int64(countRem),
		Error:  nil,
	}
}

// evalAPPEND takes two arguments: the key and the value to append to the key's current value.
// If the key does not exist, it creates a new key with the given value (so APPEND will be similar to SET in this special case)
// If key already exists and is a string (or integers stored as strings), this command appends the value at the end of the string
func evalAPPEND(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 2 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("APPEND"),
		}
	}

	key, value := args[0], args[1]
	obj := store.Get(key)

	// Get the current expiry time
	var exDurationMs int64 = -1 // -1 indicates no expiry time
	expiryTStampMs, hasExpiry := dstore.GetExpiry(obj, store)

	// Set the new expiry time
	if hasExpiry {
		// get the new expiry time in milliseconds
		exDurationMs = expiryTStampMs - time.Now().UnixMilli()
		if exDurationMs < 0 {
			// set expiry time to 0
			exDurationMs = 0
		}
	}

	// Key does not exist, create a new key
	if obj == nil {
		storedValue, oType := getRawStringOrInt(value)
		store.Put(key, store.NewObj(storedValue, exDurationMs, oType))
		return &EvalResponse{
			Result: len(value),
			Error:  nil,
		}
	}

	// Key exists path
	if _, ok := obj.Value.(*sortedset.Set); ok {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}
	oType := obj.Type

	// Transform the value based on the current encoding
	currentValue, err := convertValueToString(obj, oType)
	if err != nil {
		// If the encoding is neither integer nor string, return a "wrong type" error
		return &EvalResponse{
			Result: nil,
			Error:  err,
		}
	}

	// Append the value
	newValue := currentValue + value

	// We need to store the new appended value as a string
	// Even if append is performed on integers, the result will be stored as a string
	store.Put(key, store.NewObj(newValue, exDurationMs, object.ObjTypeString))
	return &EvalResponse{
		Result: len(newValue),
		Error:  nil,
	}
}

// evalZRANK returns the rank of the member in the sorted set stored at key.
// The rank (or index) is 0-based, which means that the member with the lowest score has rank 0.
// If the 'WITHSCORE' option is specified, it returns both the rank and the score of the member.
// Returns nil if the key does not exist or the member is not a member of the sorted set.
func evalZRANK(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 2 || len(args) > 3 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("ZRANK"),
		}
	}

	key := args[0]
	member := args[1]
	withScore := false

	if len(args) == 3 {
		if !strings.EqualFold(args[2], WithScore) {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrInvalidSyntax("ZRANK"),
			}
		}
		withScore = true
	}

	obj := store.Get(key)
	if obj == nil {
		return &EvalResponse{
			Result: NIL,
			Error:  nil,
		}
	}

	sortedSet, err := sortedset.FromObject(obj)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}
	rank, score := sortedSet.RankWithScore(member, false)
	if rank == -1 {
		return &EvalResponse{
			Result: NIL,
			Error:  nil,
		}
	}

	if withScore {
		scoreStr := strconv.FormatFloat(score, 'f', -1, 64)
		return &EvalResponse{
			Result: []interface{}{rank, scoreStr},
			Error:  nil,
		}
	}

	return &EvalResponse{
		Result: rank,
		Error:  nil,
	}
}

// evalZCARD returns the cardinality (number of elements) of the sorted set stored at key.
// Returns 0 if the key does not exist.
func evalZCARD(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 1 || len(args) > 1 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("ZCARD"),
		}
	}

	key := args[0]
	obj := store.Get(key)

	if obj == nil {
		return &EvalResponse{
			Result: IntegerZero,
			Error:  nil,
		}
	}

	sortedSet, err := sortedset.FromObject(obj)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}
	return &EvalResponse{
		Result: int64(sortedSet.Len()),
		Error:  nil,
	}
}

// evalJSONCLEAR Clear container values (arrays/objects) and set numeric values to 0,
// Already cleared values are ignored for empty containers and zero numbers
// args must contain at least the key;  (path unused in this implementation)
// Returns encoded error if key is expired, or it does not exist
// Returns encoded error response if incorrect number of arguments
// Returns an integer reply specifying the number of matching JSON arrays
// and objects cleared + number of matching JSON numerical values zeroed.
func evalJSONCLEAR(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 1 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("JSON.CLEAR"),
		}
	}
	key := args[0]

	// Default path is root if not specified
	path := defaultRootPath
	if len(args) > 1 {
		path = args[1]
	}

	// Retrieve the object from the database
	obj := store.Get(key)
	if obj == nil {
		return &EvalResponse{
			Result: nil,
			Error:  nil,
		}
	}

	if errWithMessage := object.AssertType(obj.Type, object.ObjTypeJSON); errWithMessage != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	jsonData := obj.Value

	_, err := sonic.Marshal(jsonData)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	var countClear int64 = 0
	if len(args) == 1 || path == defaultRootPath {
		if jsonData != struct{}{} {
			// If path is root and len(args) == 1, return it instantly
			newObj := store.NewObj(struct{}{}, -1, object.ObjTypeJSON)
			store.Put(key, newObj)
			countClear++
			return &EvalResponse{
				Result: countClear,
				Error:  nil,
			}
		}
	}

	expr, err := jp.ParseString(path)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrJSONPathNotFound(path),
		}
	}

	newData, err := expr.Modify(jsonData, func(element any) (altered any, changed bool) {
		switch utils.GetJSONFieldType(element) {
		case utils.IntegerType, utils.NumberType:
			if element != utils.NumberZeroValue {
				countClear++
				return utils.NumberZeroValue, true
			}
		case utils.ArrayType:
			if len(element.([]interface{})) != 0 {
				countClear++
				return []interface{}{}, true
			}
		case utils.ObjectType:
			if element != struct{}{} {
				countClear++
				return struct{}{}, true
			}
		default:
			return element, false
		}
		return
	})
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral(err.Error()),
		}
	}

	jsonData = newData
	obj.Value = jsonData
	return &EvalResponse{
		Result: countClear,
		Error:  nil,
	}
}

// evalJSONGET retrieves a JSON value stored at the specified key
// args must contain at least the key;  (path unused in this implementation)
// Returns response.RespNIL if key is expired, or it does not exist
// Returns encoded error response if incorrect number of arguments
// The RESP value of the key is encoded and then returned
func evalJSONGET(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 1 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("JSON.GET"),
		}
	}

	key := args[0]
	// Default path is root if not specified
	path := defaultRootPath
	if len(args) > 1 {
		path = args[1]
	}
	return jsonGETHelper(store, path, key)
}

// helper function used by evalJSONGET and evalJSONMGET to prepare the results
func jsonGETHelper(store *dstore.Store, path, key string) *EvalResponse {
	// Retrieve the object from the database
	obj := store.Get(key)
	if obj == nil {
		return &EvalResponse{
			Result: NIL,
			Error:  nil,
		}
	}

	// Check if the object is of JSON type
	if errWithMessage := object.AssertType(obj.Type, object.ObjTypeJSON); errWithMessage != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	jsonData := obj.Value

	// If path is root, return the entire JSON
	if path == defaultRootPath {
		resultBytes, err := sonic.Marshal(jsonData)
		if err != nil {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrGeneral("could not serialize result"),
			}
		}

		return &EvalResponse{
			Result: string(resultBytes),
			Error:  nil,
		}
	}

	// Parse the JSONPath expression
	expr, err := jp.ParseString(path)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral("invalid JSONPath"),
		}
	}

	// Execute the JSONPath query
	results := expr.Get(jsonData)
	if len(results) == 0 {
		return &EvalResponse{
			Result: NIL,
			Error:  nil,
		}
	}

	// Serialize the result
	var resultBytes []byte
	if len(results) == 1 {
		resultBytes, err = sonic.Marshal(results[0])
	} else {
		resultBytes, err = sonic.Marshal(results)
	}
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral("could not serialize result"),
		}
	}
	return &EvalResponse{
		Result: string(resultBytes),
		Error:  nil,
	}
}

// evalJSONSET stores a JSON value at the specified key
// args must contain at least the key, path (unused in this implementation), and JSON string
// Returns encoded error response if incorrect number of arguments
// Returns encoded error if the JSON string is invalid
// Returns response.RespOK if the JSON value is successfully stored
func evalJSONSET(args []string, store *dstore.Store) *EvalResponse {
	// Check if there are enough arguments
	if len(args) < 3 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("JSON.SET"),
		}
	}

	key := args[0]
	path := args[1]
	jsonStr := args[2]
	for i := 3; i < len(args); i++ {
		switch strings.ToUpper(args[i]) {
		case NX:
			if i != len(args)-1 {
				return &EvalResponse{
					Result: nil,
					Error:  diceerrors.ErrInvalidSyntax("JSONSET"),
				}
			}
			obj := store.Get(key)
			if obj != nil {
				return &EvalResponse{
					Result: NIL,
					Error:  nil,
				}
			}
		case XX:
			if i != len(args)-1 {
				return &EvalResponse{
					Result: nil,
					Error:  diceerrors.ErrInvalidSyntax("JSONSET"),
				}
			}
			obj := store.Get(key)
			if obj == nil {
				return &EvalResponse{
					Result: NIL,
					Error:  nil,
				}
			}

		default:
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrInvalidSyntax("JSONSET"),
			}
		}
	}

	// Parse the JSON string
	var jsonValue interface{}
	if err := sonic.UnmarshalString(jsonStr, &jsonValue); err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral("invalid JSON"),
		}
	}

	// Retrieve existing object or create new one
	obj := store.Get(key)
	var rootData interface{}

	if obj == nil {
		// If the key doesn't exist, create a new object
		if path != defaultRootPath {
			rootData = make(map[string]interface{})
		} else {
			rootData = jsonValue
		}
	} else {
		// If the key exists, check if it's a JSON object
		if err := object.AssertType(obj.Type, object.ObjTypeJSON); err != nil {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongTypeOperation,
			}
		}
		rootData = obj.Value
	}

	// If path is not root, use JSONPath to set the value
	if path != defaultRootPath {
		expr, err := jp.ParseString(path)
		if err != nil {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrGeneral("invalid JSONPath"),
			}
		}

		err = expr.Set(rootData, jsonValue)
		if err != nil {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrGeneral("failed to set value"),
			}
		}
	} else {
		// If path is root, replace the entire JSON
		rootData = jsonValue
	}

	// Create a new object with the updated JSON data
	newObj := store.NewObj(rootData, -1, object.ObjTypeJSON)
	store.Put(key, newObj)
	return &EvalResponse{
		Result: OK,
		Error:  nil,
	}
}

// evalJSONINGEST stores a value at a dynamically generated key
// The key is created using a provided key prefix combined with a unique identifier
// args must contains key_prefix and path and json value
// It will call to evalJSONSET internally.
// Returns encoded error response if incorrect number of arguments
// Returns encoded error if the JSON string is invalid
// Returns unique identifier if the JSON value is successfully stored
func evalJSONINGEST(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 3 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("JSON.INGEST"),
		}
	}

	keyPrefix := args[0]

	uniqueID := xid.New()
	uniqueKey := keyPrefix + uniqueID.String()

	var setArgs []string
	setArgs = append(setArgs, uniqueKey)
	setArgs = append(setArgs, args[1:]...)

	result := evalJSONSET(setArgs, store)
	if resultValue, ok := result.Result.(RespType); ok {
		// If Result is of type RespType, check equality
		if resultValue == OK {
			return &EvalResponse{
				Result: uniqueID.String(),
				Error:  nil,
			}
		}
	}
	return result
}

// evalJSONTYPE retrieves a JSON value type stored at the specified key
// args must contain at least the key;  (path unused in this implementation)
// Returns response.RespNIL if key is expired, or it does not exist
// Returns encoded error response if incorrect number of arguments
// The RESP value of the key's value type is encoded and then returned
func evalJSONTYPE(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 1 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("JSON.TYPE"),
		}
	}
	key := args[0]

	// Default path is root if not specified
	path := defaultRootPath
	if len(args) > 1 {
		path = args[1]
	}
	// Retrieve the object from the database
	obj := store.Get(key)
	if obj == nil {
		return &EvalResponse{
			Result: NIL,
			Error:  nil,
		}
	}

	if errWithMessage := object.AssertType(obj.Type, object.ObjTypeJSON); errWithMessage != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	jsonData := obj.Value

	if path == defaultRootPath {
		_, err := sonic.Marshal(jsonData)
		if err != nil {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrGeneral("could not serialize result"),
			}
		}
		// If path is root and len(args) == 1, return "object" instantly
		if len(args) == 1 {
			return &EvalResponse{
				Result: utils.ObjectType,
				Error:  nil,
			}
		}
	}

	// Parse the JSONPath expression
	expr, err := jp.ParseString(path)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral("invalid JSONPath"),
		}
	}

	results := expr.Get(jsonData)
	if len(results) == 0 {
		return &EvalResponse{
			Result: EmptyArray,
			Error:  nil,
		}
	}

	typeList := make([]string, 0, len(results))
	for _, result := range results {
		jsonType := utils.GetJSONFieldType(result)
		typeList = append(typeList, jsonType)
	}
	return &EvalResponse{
		Result: typeList,
		Error:  nil,
	}
}

// PFADD Adds all the element arguments to the HyperLogLog data structure stored at the variable
// name specified as first argument.
//
// Returns:
// If the approximated cardinality estimated by the HyperLogLog changed after executing the command,
// returns 1, otherwise 0 is returned.
func evalPFADD(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 1 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("PFADD"),
		}
	}

	key := args[0]
	obj := store.Get(key)

	// If key doesn't exist prior initial cardinality changes hence return 1
	if obj == nil {
		hll := hyperloglog.New()
		for _, arg := range args[1:] {
			hll.Insert([]byte(arg))
		}

		obj = store.NewObj(hll, -1, object.ObjTypeHLL)

		store.Put(key, obj, dstore.WithPutCmd(dstore.PFADD))
		return &EvalResponse{
			Result: int64(1),
			Error:  nil,
		}
	}

	existingHll, ok := obj.Value.(*hyperloglog.Sketch)
	if !ok {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrInvalidHyperLogLogKey,
		}
	}
	initialCardinality := existingHll.Estimate()
	for _, arg := range args[1:] {
		existingHll.Insert([]byte(arg))
	}

	obj = store.NewObj(existingHll, -1, object.ObjTypeHLL)
	store.Put(key, obj, dstore.WithPutCmd(dstore.PFADD))

	if newCardinality := existingHll.Estimate(); initialCardinality != newCardinality {
		return &EvalResponse{
			Result: int64(1),
			Error:  nil,
		}
	}

	return &EvalResponse{
		Result: int64(0),
		Error:  nil,
	}
}

// evalJSONSTRLEN Report the length of the JSON String at path in key
// Returns by recursive descent an array of integer replies for each path,
// the string's length, or nil, if the matching JSON value is not a string.
func evalJSONSTRLEN(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 1 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("JSON.STRLEN"),
		}
	}

	key := args[0]

	obj := store.Get(key)

	if obj == nil {
		return &EvalResponse{
			Result: nil,
			Error:  nil,
		}
	}

	if len(args) < 2 {
		// no recursive
		// making consistent with arrlen
		// to-do parsing
		jsonData := obj.Value

		jsonDataType := strings.ToLower(utils.GetJSONFieldType(jsonData))
		if jsonDataType == "number" {
			jsonDataFloat := jsonData.(float64)
			if jsonDataFloat == float64(int64(jsonDataFloat)) {
				jsonDataType = "integer"
			}
		}
		if jsonDataType != utils.StringType {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrUnexpectedJSONPathType("string", jsonDataType),
			}
		}
		return &EvalResponse{
			Result: int64(len(jsonData.(string))),
			Error:  nil,
		}
	}

	path := args[1]

	// Check if the object is of JSON type
	if errWithMessage := object.AssertType(obj.Type, object.ObjTypeJSON); errWithMessage != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	jsonData := obj.Value
	if path == defaultRootPath {
		defaultStringResult := make([]interface{}, 0, 1)
		if utils.GetJSONFieldType(jsonData) == utils.StringType {
			defaultStringResult = append(defaultStringResult, int64(len(jsonData.(string))))
		} else {
			defaultStringResult = append(defaultStringResult, nil)
		}

		return &EvalResponse{
			Result: defaultStringResult,
			Error:  nil,
		}
	}

	// Parse the JSONPath expression
	expr, err := jp.ParseString(path)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrJSONPathNotFound(path),
		}
	}
	// Execute the JSONPath query
	results := expr.Get(jsonData)
	if len(results) == 0 {
		return &EvalResponse{
			Result: []interface{}{},
			Error:  nil,
		}
	}
	strLenResults := make([]interface{}, 0, len(results))
	for _, result := range results {
		switch utils.GetJSONFieldType(result) {
		case utils.StringType:
			strLenResults = append(strLenResults, int64(len(result.(string))))
		default:
			strLenResults = append(strLenResults, nil)
		}
	}
	return &EvalResponse{
		Result: strLenResults,
		Error:  nil,
	}
}

func evalPFCOUNT(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 1 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("PFCOUNT"),
		}
	}

	unionHll := hyperloglog.New()

	for _, arg := range args {
		obj := store.Get(arg)
		if obj != nil {
			currKeyHll, ok := obj.Value.(*hyperloglog.Sketch)
			if !ok {
				return &EvalResponse{
					Result: nil,
					Error:  diceerrors.ErrInvalidHyperLogLogKey,
				}
			}
			err := unionHll.Merge(currKeyHll)
			if err != nil {
				return &EvalResponse{
					Result: nil,
					Error:  diceerrors.ErrCorruptedHyperLogLogObject,
				}
			}
		}
	}

	return &EvalResponse{
		Result: unionHll.Estimate(),
		Error:  nil,
	}
}

// evalJSONOBJLEN return the number of keys in the JSON object at path in key.
// Returns an array of integer replies, an integer for each matching value,
// which is the json objects length, or nil, if the matching value is not a json.
// Returns encoded error if the key doesn't exist or key is expired or the matching value is not an array.
// Returns encoded error response if incorrect number of arguments
func evalJSONOBJLEN(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 1 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("JSON.OBJLEN"),
		}
	}

	key := args[0]

	// Retrieve the object from the database
	obj := store.Get(key)
	if obj == nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrKeyDoesNotExist,
		}
	}

	// check if the object is json
	if errWithMessage := object.AssertType(obj.Type, object.ObjTypeJSON); errWithMessage != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	// get the value & check for marsheling error
	jsonData := obj.Value
	_, err := sonic.Marshal(jsonData)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}
	if len(args) == 1 {
		// check if the value is of json type
		if utils.GetJSONFieldType(jsonData) == utils.ObjectType {
			if castedData, ok := jsonData.(map[string]interface{}); ok {
				return &EvalResponse{
					Result: int64(len(castedData)),
					Error:  nil,
				}
			}
			return &EvalResponse{
				Result: nil,
				Error:  nil,
			}
		}
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	path, isDefinitePath := utils.ParseInputJSONPath(args[1])

	expr, err := jp.ParseString(path)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrJSONPathNotFound(path),
		}
	}

	// get all values for matching paths
	results := expr.Get(jsonData)

	objectLen := make([]interface{}, 0, len(results))

	for _, result := range results {
		switch utils.GetJSONFieldType(result) {
		case utils.ObjectType:
			if castedResult, ok := result.(map[string]interface{}); ok {
				objectLen = append(objectLen, int64(len(castedResult)))
			} else {
				objectLen = append(objectLen, nil)
			}
		default:
			// If it is a definitePath, and the only value is not JSON, throw wrong type error
			if isDefinitePath {
				return &EvalResponse{
					Result: nil,
					Error:  diceerrors.ErrWrongTypeOperation,
				}
			}
			objectLen = append(objectLen, nil)
		}
	}

	// Must return a single integer if it is a definite Path
	if isDefinitePath {
		if len(objectLen) == 0 {
			return &EvalResponse{
				Result: nil,
				Error:  nil,
			}
		}
		return &EvalResponse{
			Result: objectLen[0],
			Error:  nil,
		}
	}

	return &EvalResponse{
		Result: objectLen,
		Error:  nil,
	}
}

// evalPTTL returns Time-to-Live in millisecs for the queried key in args
// The key should be the only param in args else returns with an error
// Returns	RESP encoded time (in secs) remaining for the key to expire
//
//	RESP encoded -2 stating key doesn't exist or key is expired
//	RESP encoded -1 in case no expiration is set on the key
func evalPTTL(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 1 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("PTTL"),
		}
	}

	key := args[0]

	obj := store.Get(key)

	if obj == nil {
		return &EvalResponse{
			Result: IntegerNegativeTwo,
			Error:  nil,
		}
	}

	exp, isExpirySet := dstore.GetExpiry(obj, store)

	if !isExpirySet {
		return &EvalResponse{
			Result: IntegerNegativeOne,
			Error:  nil,
		}
	}

	// compute the time remaining for the key to expire and
	// return the RESP encoded form of it
	durationMs := exp - time.Now().UnixMilli()
	return &EvalResponse{
		Result: durationMs,
		Error:  nil,
	}
}

// Increments the number stored at field in the hash stored at key by increment.
//
// If key does not exist, a new key holding a hash is created.
// If field does not exist the value is set to the increment value passed
//
// The range of values supported by HINCRBY is limited to 64-bit signed integers.
//
// Usage: HINCRBY key field increment
func evalHINCRBY(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 3 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("HINCRBY"),
		}
	}

	increment, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrIntegerOutOfRange,
		}
	}
	var hashmap HashMap
	key := args[0]
	obj := store.Get(key)
	if obj != nil {
		if err := object.AssertType(obj.Type, object.ObjTypeSSMap); err != nil {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongTypeOperation,
			}
		}
		hashmap = obj.Value.(HashMap)
	}

	if hashmap == nil {
		hashmap = make(HashMap)
	}

	field := args[1]
	numkey, err := hashmap.incrementValue(field, increment)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral(err.Error()),
		}
	}

	obj = store.NewObj(hashmap, -1, object.ObjTypeSSMap)
	store.Put(key, obj)

	return &EvalResponse{
		Result: numkey,
		Error:  nil,
	}
}

// Increments the number stored at field in the hash stored at key by the specified floating point increment.
//
// If key does not exist, a new key holding a hash is created.
// If field does not exist, the value is set to the increment passed before the operation is performed.
//
// The precision of the increment is not restricted to integers, allowing for floating point values.
//
// Usage: HINCRBYFLOAT key field increment
func evalHINCRBYFLOAT(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 3 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("HINCRBYFLOAT"),
		}
	}

	increment, err := strconv.ParseFloat(strings.TrimSpace(args[2]), 64)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrInvalidNumberFormat,
		}
	}

	key := args[0]
	obj := store.Get(key)
	var hashmap HashMap
	if obj != nil {
		if err := object.AssertType(obj.Type, object.ObjTypeSSMap); err != nil {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongTypeOperation,
			}
		}
		hashmap = obj.Value.(HashMap)
	}

	if hashmap == nil {
		hashmap = make(HashMap)
	}

	field := args[1]
	numkey, err := hashmap.incrementFloatValue(field, increment)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral(err.Error()),
		}
	}

	obj = store.NewObj(hashmap, -1, object.ObjTypeSSMap)
	store.Put(key, obj)

	return &EvalResponse{
		Result: numkey,
		Error:  nil,
	}
}

// evalHRANDFIELD returns random fields from a hash stored at key.
// If only the key is provided, one random field is returned.
// If count is provided, it returns that many unique random fields. A negative count allows repeated selections.
// The "WITHVALUES" option returns both fields and values.
// Returns nil if the key doesn't exist or the hash is empty.
// Errors: arity error, type error for non-hash, syntax error for "WITHVALUES", or count format error.
func evalHRANDFIELD(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 1 || len(args) > 3 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("HRANDFIELD"),
		}
	}

	key := args[0]
	obj := store.Get(key)
	if obj == nil {
		return &EvalResponse{
			Result: NIL,
			Error:  nil,
		}
	}

	if err := object.AssertType(obj.Type, object.ObjTypeSSMap); err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	hashMap := obj.Value.(HashMap)
	if len(hashMap) == 0 {
		return &EvalResponse{
			Result: EmptyArray,
			Error:  nil,
		}
	}

	count := 1
	withValues := false

	if len(args) > 1 {
		var err error
		// The second argument is the count.
		count, err = strconv.Atoi(args[1])
		if err != nil {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrIntegerOutOfRange,
			}
		}

		// The third argument is the "WITHVALUES" option.
		if len(args) == 3 {
			if !strings.EqualFold(args[2], WithValues) {
				return &EvalResponse{
					Result: nil,
					Error:  diceerrors.ErrInvalidSyntax("HRANDFIELD"),
				}
			}
			withValues = true
		}
	}

	return selectRandomFields(hashMap, count, withValues)
}

// evalINCRBYFLOAT increments the value of the  key in args by the specified increment,
// if the key exists and the value is a number.
// The key should be the first parameter in args, and the increment should be the second parameter.
// If the key does not exist, a new key is created with increment's value.
// If the value at the key is a string, it should be parsable to float64,
// if not evalINCRBYFLOAT returns an  error response.
// evalINCRBYFLOAT returns the incremented value for the key after applying the specified increment if there are no errors.
func evalINCRBYFLOAT(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 2 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("INCRBYFLOAT"),
		}
	}
	incr, err := strconv.ParseFloat(strings.TrimSpace(args[1]), 64)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral("value is not a valid float"),
		}
	}
	return incrByFloatCmd(args, incr, store)
}

func incrByFloatCmd(args []string, incr float64, store *dstore.Store) *EvalResponse {
	key := args[0]
	obj := store.Get(key)

	if obj == nil {
		strValue := formatFloat(incr, false)
		_, oType := getRawStringOrInt(strValue)
		obj = store.NewObj(strValue, -1, oType)
		store.Put(key, obj)
		return &EvalResponse{
			Result: strValue,
			Error:  nil,
		}
	}

	errString := object.AssertType(obj.Type, object.ObjTypeString)
	errInt := object.AssertType(obj.Type, object.ObjTypeInt)
	if errString != nil && errInt != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	value, err := floatValue(obj.Value)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral("value is not a valid float"),
		}
	}
	value += incr
	if math.IsInf(value, 0) {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrValueOutOfRange,
		}
	}
	strValue := formatFloat(value, true)

	_, oType := getRawStringOrInt(strValue)

	strValue = strings.TrimSuffix(strValue, ".0")

	obj.Value = strValue
	obj.Type = oType

	return &EvalResponse{
		Result: strValue,
		Error:  nil,
	}
}

// floatValue returns the float64 value for an interface which
// contains either a string or an int.
func floatValue(value interface{}) (float64, error) {
	switch raw := value.(type) {
	case string:
		parsed, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return 0, err
		}
		return parsed, nil
	case int64:
		return float64(raw), nil
	}

	return 0, fmt.Errorf(diceerrors.IntOrFloatErr)
}

// ZPOPMIN Removes and returns the member with the lowest score from the sorted set at the specified key.
// If multiple members have the same score, the one that comes first alphabetically is returned.
// You can also specify a count to remove and return multiple members at once.
// If the set is empty, it returns an empty result.
func evalZPOPMIN(args []string, store *dstore.Store) *EvalResponse {
	// Incorrect number of arguments should return error
	if len(args) < 1 || len(args) > 2 {
		return &EvalResponse{
			Result: NIL,
			Error:  diceerrors.ErrWrongArgumentCount("ZPOPMIN"),
		}
	}

	key := args[0]        // Key argument
	obj := store.Get(key) // Getting sortedSet object from store

	// If the sortedSet is nil, return an empty list
	if obj == nil {
		return &EvalResponse{
			Result: []string{},
			Error:  nil,
		}
	}

	sortedSet, err := sortedset.FromObject(obj)
	if err != nil {
		return &EvalResponse{
			Result: NIL,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	count := 1
	// Check if the count argument is provided.
	if len(args) == 2 {
		countArg, err := strconv.Atoi(args[1])
		if err != nil {
			// Return an error if the argument is not a valid integer
			return &EvalResponse{
				Result: NIL,
				Error:  diceerrors.ErrIntegerOutOfRange,
			}
		}
		count = countArg
	}

	// If count is less than 1, empty array is returned
	if count < 1 {
		return &EvalResponse{
			Result: []string{},
			Error:  nil,
		}
	}

	// If the count argument is present, return all the members with lowest score sorted in ascending order.
	// If there are multiple lowest scores with same score value, it sorts the members in lexographical order of member name
	results := sortedSet.GetMin(count)

	return &EvalResponse{
		Result: results,
		Error:  nil,
	}
}

func evalDUMP(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 1 {
		return makeEvalError(diceerrors.ErrWrongArgumentCount("DUMP"))
	}
	key := args[0]
	obj := store.Get(key)
	if obj == nil {
		return makeEvalResult(NIL)
	}

	serializedValue, err := rdbSerialize(obj)
	if err != nil {
		fmt.Println("error", err)
		return makeEvalError(diceerrors.ErrGeneral("serialization failed"))
	}
	encodedResult := base64.StdEncoding.EncodeToString(serializedValue)

	return makeEvalResult(encodedResult)
}

func evalRestore(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 3 {
		return makeEvalError(diceerrors.ErrWrongArgumentCount("RESTORE"))
	}

	key := args[0]
	ttlStr := args[1]
	ttl, _ := strconv.ParseInt(ttlStr, 10, 64)

	encodedValue := args[2]
	serializedData, err := base64.StdEncoding.DecodeString(encodedValue)
	if err != nil {
		return makeEvalError(diceerrors.ErrGeneral("failed to decode base64 value"))
	}
	obj, err := rdbDeserialize(serializedData)
	if err != nil {
		return makeEvalError(diceerrors.ErrGeneral("deserialization failed"))
	}

	newobj := store.NewObj(obj.Value, ttl, obj.Type)
	var keepttl = true

	if ttl > 0 {
		store.Put(key, newobj, dstore.WithKeepTTL(keepttl))
	} else {
		store.Put(key, obj)
	}

	return makeEvalResult(OK)
}

// evalHLEN returns the number of fields contained in the hash stored at key.
//
// If key doesn't exist, it returns 0.
//
// Usage: HLEN key
func evalHLEN(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 1 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("HLEN"),
		}
	}

	key := args[0]

	obj := store.Get(key)

	if obj == nil {
		return &EvalResponse{
			Result: IntegerZero,
			Error:  nil,
		}
	}

	if err := object.AssertType(obj.Type, object.ObjTypeSSMap); err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	hashMap := obj.Value.(HashMap)
	return &EvalResponse{
		Result: len(hashMap),
		Error:  nil,
	}
}

// evalHSTRLEN returns the length of value associated with field in the hash stored at key.
//
// This command returns 0, if the specified field doesn't exist in the key
//
// If key doesn't exist, it returns 0.
//
// Usage: HSTRLEN key field value
func evalHSTRLEN(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 2 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("HSTRLEN"),
		}
	}

	key := args[0]
	hmKey := args[1]
	obj := store.Get(key)

	var hashMap HashMap

	if obj != nil {
		if err := object.AssertType(obj.Type, object.ObjTypeSSMap); err != nil {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongTypeOperation,
			}
		}
		hashMap = obj.Value.(HashMap)
	} else {
		return &EvalResponse{
			Result: IntegerZero,
			Error:  nil,
		}
	}

	val, ok := hashMap.Get(hmKey)
	// Return 0, if specified field doesn't exist in the HashMap.
	if ok {
		return &EvalResponse{
			Result: len(*val),
			Error:  nil,
		}
	}
	return &EvalResponse{
		Result: IntegerZero,
		Error:  nil,
	}
}

// evalHSCAN return a two element multi-bulk reply, where the first element is a string representing the cursor,
// and the second element is a multi-bulk with an array of elements.
//
// The array of elements contain two elements, a field and a value, for every returned element of the Hash.
//
// If key doesn't exist, it returns an array containing 0 and empty array.
//
// Usage: HSCAN key cursor [MATCH pattern] [COUNT count]
func evalHSCAN(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 2 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("HSCAN"),
		}
	}

	key := args[0]
	cursor, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrIntegerOutOfRange,
		}
	}

	obj := store.Get(key)
	if obj == nil {
		return &EvalResponse{
			Result: []interface{}{"0", []string{}},
			Error:  nil,
		}
	}

	if err := object.AssertType(obj.Type, object.ObjTypeSSMap); err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	hashMap := obj.Value.(HashMap)
	pattern := "*"
	count := 10

	// Parse optional arguments
	for i := 2; i < len(args); i += 2 {
		switch strings.ToUpper(args[i]) {
		case "MATCH":
			if i+1 < len(args) {
				pattern = args[i+1]
			}
		case CountConst:
			if i+1 < len(args) {
				parsedCount, err := strconv.Atoi(args[i+1])
				if err != nil || parsedCount < 1 {
					return &EvalResponse{
						Result: nil,
						Error:  diceerrors.ErrIntegerOutOfRange,
					}
				}
				count = parsedCount
			}
		}
	}

	// Note that this implementation has a time complexity of O(N), where N is the number of keys in 'hashMap'.
	keys := make([]string, 0, len(hashMap))
	for k := range hashMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	matched := 0
	results := make([]string, 0, count*2)
	newCursor := 0

	g, err := glob.Compile(pattern)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral(fmt.Sprintf("Invalid glob pattern: %s", err)),
		}
	}

	// Scan the keys and add them to the results if they match the pattern
	for i := int(cursor); i < len(keys); i++ {
		if g.Match(keys[i]) {
			results = append(results, keys[i], hashMap[keys[i]])
			matched++
			if matched >= count {
				newCursor = i + 1
				break
			}
		}
	}

	// If we've scanned all keys, reset cursor to 0
	if newCursor >= len(keys) {
		newCursor = 0
	}

	return &EvalResponse{
		Result: []interface{}{strconv.Itoa(newCursor), results},
		Error:  nil,
	}
}

// evalBF.RESERVE evaluates the BF.RESERVE command responsible for initializing a
// new bloom filter and allocation it's relevant parameters based on given inputs.
// If no params are provided, it uses defaults.
func evalBFRESERVE(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 3 {
		return makeEvalError(diceerrors.ErrWrongArgumentCount("BF.RESERVE"))
	}

	opts, err := newBloomOpts(args[1:])
	if err != nil {
		return makeEvalError(err)
	}

	key := args[0]

	bf, err := GetBloomFilter(key, store)
	if err != nil && err != diceerrors.ErrKeyNotFound { // bloom filter does not exist
		return makeEvalError(err)
	} else if err != nil && err == diceerrors.ErrKeyNotFound { // key does not exists
		CreateOrReplaceBloomFilter(key, opts, store)
		return makeEvalResult(OK)
	} else if bf != nil { // bloom filter already exists
		return makeEvalError(diceerrors.ErrKeyExists)
	}
	return makeEvalResult(OK)
}

// evalBFADD evaluates the BF.ADD command responsible for adding an element to a bloom filter. If the filter does not
// exist, it will create a new one with default parameters.
func evalBFADD(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 2 {
		return makeEvalError(diceerrors.ErrWrongArgumentCount("BF.ADD"))
	}

	bf, err := GetOrCreateBloomFilter(args[0], store, nil)
	if err != nil {
		return makeEvalError(err)
	}

	result, err := bf.add(args[1])
	if err != nil {
		return makeEvalError(err)
	}

	return makeEvalResult(result)
}

// evalBFEXISTS evaluates the BF.EXISTS command responsible for checking existence of an element in a bloom filter.
func evalBFEXISTS(args []string, store *dstore.Store) *EvalResponse {
	// todo must work with objects of
	if len(args) != 2 {
		return makeEvalError(diceerrors.ErrWrongArgumentCount("BF.EXISTS"))
	}

	bf, err := GetBloomFilter(args[0], store)
	if err != nil && err != diceerrors.ErrKeyNotFound {
		return makeEvalError(err)
	} else if err != nil && err == diceerrors.ErrKeyNotFound {
		return makeEvalResult(IntegerZero)
	}

	result, err := bf.exists(args[1])
	if err != nil {
		return makeEvalError(err)
	}
	return makeEvalResult(result)
}

// evalBFINFO evaluates the BF.INFO command responsible for returning the
// parameters and metadata of an existing bloom filter.
func evalBFINFO(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 1 || len(args) > 2 {
		return makeEvalError(diceerrors.ErrWrongArgumentCount("BF.INFO"))
	}

	bf, err := GetBloomFilter(args[0], store)
	if err != nil {
		return makeEvalError(err)
	}

	opt := ""
	if len(args) == 2 {
		opt = args[1]
	}

	result, err := bf.info(opt)
	if err != nil {
		return makeEvalError(err)
	}
	return makeEvalResult(result)
}

// This command removes the element with the maximum score from the sorted set.
// If two elements have the same score then the members are aligned in lexicographically and the lexicographically greater element is removed.
// There is a second optional element called count which specifies the number of element to be removed.
// Returns the removed elements from the sorted set.
func evalZPOPMAX(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 1 || len(args) > 2 {
		return &EvalResponse{
			Result: NIL,
			Error:  diceerrors.ErrWrongArgumentCount("ZPOPMAX"),
		}
	}

	key := args[0]
	obj := store.Get(key)

	count := 1
	if len(args) > 1 {
		ops, err := strconv.Atoi(args[1])
		if err != nil {
			return &EvalResponse{
				Result: NIL,
				Error:  diceerrors.ErrGeneral("value is out of range, must be positive"), // This error is thrown when then count argument is not an integer
			}
		}
		if ops <= 0 {
			return &EvalResponse{
				Result: []string{}, // Returns empty array when the count is less than or equal to  0
				Error:  nil,
			}
		}
		count = ops
	}

	if obj == nil {
		return &EvalResponse{
			Result: []string{}, // Returns empty array when the object with given key is not present in the store
			Error:  nil,
		}
	}

	var sortedSet *sortedset.Set
	sortedSet, err := sortedset.FromObject(obj)
	if err != nil {
		return &EvalResponse{
			Result: NIL,
			Error:  diceerrors.ErrWrongTypeOperation, // Returns this error when a key is present in the store but is not of type sortedset.Set
		}
	}

	var res []string = sortedSet.PopMax(count)

	return &EvalResponse{
		Result: res,
		Error:  nil,
	}
}

// evalJSONARRTRIM trim an array so that it contains only the specified inclusive range of elements
// an array of integer replies for each path, the array's new size, or nil, if the matching JSON value is not an array.
func evalJSONARRTRIM(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 4 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("JSON.ARRTRIM"),
		}
	}
	var err error

	start := args[2]
	stop := args[3]
	var startIdx, stopIdx int
	startIdx, err = strconv.Atoi(start)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrIntegerOutOfRange,
		}
	}
	stopIdx, err = strconv.Atoi(stop)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrIntegerOutOfRange,
		}
	}

	key := args[0]
	obj := store.Get(key)
	if obj == nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral("key does not exist"),
		}
	}

	if errWithMessage := object.AssertType(obj.Type, object.ObjTypeJSON); errWithMessage != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral(string(errWithMessage)),
		}
	}

	jsonData := obj.Value

	_, err = sonic.Marshal(jsonData)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral("Existing key has wrong Dice type"),
		}
	}

	path := args[1]
	expr, err := jp.ParseString(path)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrJSONPathNotFound(path),
		}
	}

	results := expr.Get(jsonData)
	if len(results) == 0 {
		return &EvalResponse{
			Result: RespEmptyArray,
			Error:  nil,
		}
	}

	var resultsArray []interface{}
	// Capture the modified data when modifying the root path
	newData, modifyErr := expr.Modify(jsonData, func(data any) (interface{}, bool) {
		arr, ok := data.([]interface{})
		if !ok {
			// Not an array
			resultsArray = append(resultsArray, nil)
			return data, false
		}

		updatedArray := trimElementAndUpdateArray(arr, startIdx, stopIdx)

		resultsArray = append(resultsArray, len(updatedArray))
		return updatedArray, true
	})
	if modifyErr != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral(fmt.Sprintf("ERR failed to modify JSON data: %v", modifyErr)),
		}
	}

	jsonData = newData
	obj.Value = jsonData

	return &EvalResponse{
		Result: resultsArray,
		Error:  nil,
	}
}

// evalLPUSH inserts value(s) one by one at the head of of the list
//
// # Returns list length after command execution
//
// Usage: LPUSH key value [value...]
func evalLPUSH(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 2 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("LPUSH"),
		}
	}

	obj := store.Get(args[0])
	if obj == nil {
		obj = store.NewObj(NewDeque(), -1, object.ObjTypeDequeue)
	}

	if err := object.AssertType(obj.Type, object.ObjTypeDequeue); err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	store.Put(args[0], obj)
	for i := 1; i < len(args); i++ {
		obj.Value.(*Deque).LPush(args[i])
	}

	deq := obj.Value.(*Deque)

	return &EvalResponse{
		Result: deq.Length,
		Error:  nil,
	}
}

// evalRPUSH inserts value(s) one by one at the tail of of the list
//
// # Returns list length after command execution
//
// Usage: RPUSH key value [value...]
func evalRPUSH(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 2 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("RPUSH"),
		}
	}

	obj := store.Get(args[0])
	if obj == nil {
		obj = store.NewObj(NewDeque(), -1, object.ObjTypeDequeue)
	}

	if err := object.AssertType(obj.Type, object.ObjTypeDequeue); err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	store.Put(args[0], obj)
	for i := 1; i < len(args); i++ {
		obj.Value.(*Deque).RPush(args[i])
	}

	deq := obj.Value.(*Deque)

	return &EvalResponse{
		Result: deq.Length,
		Error:  nil,
	}
}

// evalLPOP pops the element at the head of the list and returns it
//
// # Returns the element at the head of the list
//
// Usage: LPOP key
func evalLPOP(args []string, store *dstore.Store) *EvalResponse {
	// By default we pop only 1
	popNumber := 1

	// LPOP accepts 1 or 2 arguments only - LPOP key [count]

	if len(args) < 1 || len(args) > 2 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("LPOP"),
		}
	}

	if len(args) == 2 {
		nos, err := strconv.Atoi(args[1])
		if err != nil {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrInvalidNumberFormat,
			}
		}
		if nos == 0 {
			// returns empty string if count given is 0
			return &EvalResponse{
				Result: NIL,
				Error:  nil,
			}
		}
		if nos < 0 {
			// returns an out of range err if count is negetive
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrIntegerOutOfRange,
			}
		}
		popNumber = nos
	}

	obj := store.Get(args[0])
	if obj == nil {
		return &EvalResponse{
			Result: NIL,
			Error:  nil,
		}
	}

	if err := object.AssertType(obj.Type, object.ObjTypeDequeue); err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	deq := obj.Value.(*Deque)

	// holds the elements popped
	var elements []string
	for iter := 0; iter < popNumber; iter++ {
		x, err := deq.LPop()
		if err != nil {
			if errors.Is(err, ErrDequeEmpty) {
				break
			}
		}
		elements = append(elements, x)
	}

	if len(elements) == 0 {
		return &EvalResponse{
			Result: NIL,
			Error:  nil,
		}
	}

	if len(elements) == 1 {
		return &EvalResponse{
			Result: elements[0],
			Error:  nil,
		}
	}

	return &EvalResponse{
		Result: elements,
		Error:  nil,
	}
}

// evalRPOP pops the element at the tail of the list and returns it
//
// # Returns the element at the tail of the list
//
// Usage: RPOP key
func evalRPOP(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 1 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("RPOP"),
		}
	}

	obj := store.Get(args[0])
	if obj == nil {
		return &EvalResponse{
			Result: NIL,
			Error:  nil,
		}
	}

	if err := object.AssertType(obj.Type, object.ObjTypeDequeue); err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	deq := obj.Value.(*Deque)
	x, err := deq.RPop()
	if err != nil {
		if errors.Is(err, ErrDequeEmpty) {
			return &EvalResponse{
				Result: NIL,
				Error:  nil,
			}
		}
	}

	return &EvalResponse{
		Result: x,
		Error:  nil,
	}
}

// evalLLEN returns the number of elements in the list
//
// Usage: LLEN key
func evalLLEN(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 1 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("LLEN"),
		}
	}

	obj := store.Get(args[0])
	if obj == nil {
		return &EvalResponse{
			Result: IntegerZero,
			Error:  nil,
		}
	}

	if err := object.AssertType(obj.Type, object.ObjTypeDequeue); err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	deq := obj.Value.(*Deque)
	return &EvalResponse{
		Result: deq.Length,
		Error:  nil,
	}
}

// evalJSONARRAPPEND appends the value(s) provided in the args to the given array path
// in the JSON object saved at key in arguments.
// Args must contain atleast a key, path and value.
// If the key does not exist or is expired, it returns response.NIL.
// If the object at given path is not an array, it returns response.NIL.
// Returns the new length of the array at path.
func evalJSONARRAPPEND(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 3 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("JSON.ARRAPPEND"),
		}
	}

	key := args[0]
	path := args[1]
	values := args[2:]

	obj := store.Get(key)
	if obj == nil {
		return &EvalResponse{
			Result: NIL,
			Error:  nil,
		}
	}
	if errWithMessage := object.AssertType(obj.Type, object.ObjTypeJSON); errWithMessage != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}
	jsonData := obj.Value

	expr, err := jp.ParseString(path)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrJSONPathNotFound(path),
		}
	}

	// Parse the input values as JSON
	parsedValues := make([]interface{}, len(values))
	for i, v := range values {
		var parsedValue interface{}
		err := sonic.UnmarshalString(v, &parsedValue)
		if err != nil {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrGeneral(err.Error()),
			}
		}
		parsedValues[i] = parsedValue
	}

	var resultsArray []interface{}
	modified := false

	// Capture the modified data when modifying the root path
	var newData interface{}
	var modifyErr error

	newData, modifyErr = expr.Modify(jsonData, func(data any) (interface{}, bool) {
		arr, ok := data.([]interface{})
		if !ok {
			// Not an array
			resultsArray = append(resultsArray, NIL)
			return data, false
		}

		// Append the parsed values to the array
		arr = append(arr, parsedValues...)

		resultsArray = append(resultsArray, int64(len(arr)))
		modified = true
		return arr, modified
	})

	if modifyErr != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral(modifyErr.Error()),
		}
	}

	if !modified {
		// If no modification was made, it means the path did not exist or was not an array
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrJSONPathNotFound(path),
		}
	}

	jsonData = newData
	obj.Value = jsonData

	return &EvalResponse{
		Result: resultsArray,
		Error:  nil,
	}
}

// evalJSONARRLEN return the length of the JSON array at path in key
// Returns an array of integer replies, an integer for each matching value,
// each is the array's length, or nil, if the matching value is not an array.
// Returns encoded error if the key doesn't exist or key is expired or the matching value is not an array.
// Returns encoded error response if incorrect number of arguments
func evalJSONARRLEN(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 1 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("JSON.ARRLEN"),
		}
	}
	key := args[0]

	// Retrieve the object from the database
	obj := store.Get(key)

	// If the object is not present in the store or if its nil, then we should simply return nil.
	if obj == nil {
		return &EvalResponse{
			Result: NIL,
			Error:  nil,
		}
	}

	if errWithMessage := object.AssertType(obj.Type, object.ObjTypeJSON); errWithMessage != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	jsonData := obj.Value

	_, err := sonic.Marshal(jsonData)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	// This is the case if only argument passed to JSON.ARRLEN is the key itself.
	// This is valid only if the key holds an array; otherwise, an error should be returned.
	if len(args) == 1 {
		if utils.GetJSONFieldType(jsonData) == utils.ArrayType {
			return &EvalResponse{
				Result: len(jsonData.([]interface{})),
				Error:  nil,
			}
		}
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	path := args[1] // Getting the path to find the length of the array
	expr, err := jp.ParseString(path)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrJSONPathNotFound(path),
		}
	}

	results := expr.Get(jsonData)

	// If there are no results, that means the JSONPath does not exist
	if len(results) == 0 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrJSONPathNotFound(path),
		}
	}

	// If the results are greater than one, we need to print them as a list
	// This condition should be updated in future when supporting Complex JSONPaths
	if len(results) > 1 {
		arrlenList := make([]interface{}, 0, len(results))
		for _, result := range results {
			switch utils.GetJSONFieldType(result) {
			case utils.ArrayType:
				arrlenList = append(arrlenList, len(result.([]interface{})))
			default:
				arrlenList = append(arrlenList, NIL)
			}
		}

		return &EvalResponse{
			Result: arrlenList,
			Error:  nil,
		}
	}

	// Single result should be printed as single integer instead of list
	jsonValue := results[0]

	if utils.GetJSONFieldType(jsonValue) == utils.ArrayType {
		return &EvalResponse{
			Result: len(jsonValue.([]interface{})),
			Error:  nil,
		}
	}

	// If execution reaches this point, the provided path either does not exist.
	return &EvalResponse{
		Result: nil,
		Error:  diceerrors.ErrJSONPathNotFound(path),
	}
}

// popElementAndUpdateArray removes an element at the given index
// Returns popped element, remaining array and error
func popElementAndUpdateArray(arr []any, index string) (popElem any, updatedArray []any, err error) {
	if len(arr) == 0 {
		return nil, nil, nil
	}

	var idx int
	// if index is empty, pop last element
	if index == "" {
		idx = len(arr) - 1
	} else {
		var err error
		idx, err = strconv.Atoi(index)
		if err != nil {
			return nil, nil, err
		}
		// convert index to a valid index
		idx = adjustIndex(idx, arr)
	}

	popElem = arr[idx]
	arr = append(arr[:idx], arr[idx+1:]...)

	return popElem, arr, nil
}

func evalJSONARRPOP(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 1 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("JSON.ARRPOP"),
		}
	}
	key := args[0]

	var path = defaultRootPath
	if len(args) >= 2 {
		path = args[1]
	}

	var index string
	if len(args) >= 3 {
		index = args[2]
	}

	// Retrieve the object from the database
	obj := store.Get(key)
	if obj == nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrKeyNotFound,
		}
	}

	if errWithMessage := object.AssertType(obj.Type, object.ObjTypeJSON); errWithMessage != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	jsonData := obj.Value
	_, err := sonic.Marshal(jsonData)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	if path == defaultRootPath {
		arr, ok := jsonData.([]any)
		// if value can not be converted to array, it is of another type
		// returns nil
		// also, return nil if array is empty
		if !ok || len(arr) == 0 {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongTypeOperation,
			}
		}
		popElem, arr, err := popElementAndUpdateArray(arr, index)
		if err != nil {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrGeneral(err.Error()),
			}
		}

		// save the remaining array
		newObj := store.NewObj(arr, -1, object.ObjTypeJSON)
		store.Put(key, newObj)

		return &EvalResponse{
			Result: popElem,
			Error:  nil,
		}
	}

	// if path is not root then extract value at path
	expr, err := jp.ParseString(path)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrJSONPathNotFound(path),
		}
	}
	results := expr.Get(jsonData)

	// process value at each path
	popArr := make([]any, 0, len(results))
	for _, result := range results {
		arr, ok := result.([]any)
		// if value can not be converted to array, it is of another type
		// returns nil
		// also, return nil if array is empty
		if !ok || len(arr) == 0 {
			popArr = append(popArr, NIL)
			continue
		}

		popElem, arr, err := popElementAndUpdateArray(arr, index)
		if err != nil {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrGeneral(err.Error()),
			}
		}

		// update array in place in the json object
		err = expr.Set(jsonData, arr)
		if err != nil {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrGeneral(err.Error()),
			}
		}

		popArr = append(popArr, popElem)
	}
	return &EvalResponse{
		Result: popArr,
		Error:  nil,
	}
}

// evalJSONARRINSERT insert the json values into the array at path before the index (shifts to the right)
// returns an array of integer replies for each path, the array's new size, or nil.
func evalJSONARRINSERT(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 4 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("JSON.ARRINSERT"),
		}
	}
	key := args[0]
	obj := store.Get(key)
	if obj == nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral("key does not exist"),
		}
	}

	if errWithMessage := object.AssertType(obj.Type, object.ObjTypeJSON); errWithMessage != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral(string(errWithMessage)),
		}
	}

	jsonData := obj.Value
	var err error
	_, err = sonic.Marshal(jsonData)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral("Existing key has wrong Dice type"),
		}
	}

	path := args[1]
	expr, err := jp.ParseString(path)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrJSONPathNotFound(path),
		}
	}

	results := expr.Get(jsonData)
	if len(results) == 0 {
		return &EvalResponse{
			Result: RespEmptyArray,
			Error:  nil,
		}
	}
	index := args[2]
	var idx int
	idx, err = strconv.Atoi(index)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrIntegerOutOfRange,
		}
	}

	values := args[3:]
	// Parse the input values as JSON
	parsedValues := make([]interface{}, len(values))
	for i, v := range values {
		var parsedValue interface{}
		err := sonic.UnmarshalString(v, &parsedValue)
		if err != nil {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrGeneral(err.Error()),
			}
		}
		parsedValues[i] = parsedValue
	}

	var resultsArray []interface{}
	// Capture the modified data when modifying the root path
	modified := false
	newData, modifyErr := expr.Modify(jsonData, func(data any) (interface{}, bool) {
		arr, ok := data.([]interface{})
		if !ok {
			// Not an array
			resultsArray = append(resultsArray, nil)
			return data, false
		}

		// Append the parsed values to the array
		updatedArray, insertErr := insertElementAndUpdateArray(arr, idx, parsedValues)
		if insertErr != nil {
			err = insertErr
			return data, false
		}
		modified = true
		resultsArray = append(resultsArray, len(updatedArray))
		return updatedArray, true
	})
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral(err.Error()),
		}
	}

	if modifyErr != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral(fmt.Sprintf("ERR failed to modify JSON data: %v", modifyErr)),
		}
	}

	if !modified {
		return &EvalResponse{
			Result: resultsArray,
			Error:  nil,
		}
	}

	jsonData = newData
	obj.Value = jsonData
	return &EvalResponse{
		Result: resultsArray,
		Error:  nil,
	}
}

// evalJSONOBJKEYS retrieves the keys of a JSON object stored at path specified.
// It takes two arguments: the key where the JSON document is stored, and an optional JSON path.
// It returns a list of keys from the object at the specified path or an error if the path is invalid.
func evalJSONOBJKEYS(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 1 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("JSON.OBJKEYS"),
		}
	}

	key := args[0]
	// Default path is root if not specified
	path := defaultRootPath
	if len(args) > 1 {
		path = args[1]
	}

	// Retrieve the object from the database
	obj := store.Get(key)
	if obj == nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral("could not perform this operation on a key that doesn't exist"),
		}
	}

	// Check if the object is of JSON type
	if errWithMessage := object.AssertType(obj.Type, object.ObjTypeJSON); errWithMessage != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral(string(errWithMessage)),
		}
	}

	jsonData := obj.Value
	_, err := sonic.Marshal(jsonData)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral("Existing key has wrong Dice type"),
		}
	}

	// If path is root, return all keys of the entire JSON
	if len(args) == 1 {
		if utils.GetJSONFieldType(jsonData) == utils.ObjectType {
			keys := make([]string, 0)
			for key := range jsonData.(map[string]interface{}) {
				keys = append(keys, key)
			}
			return &EvalResponse{
				Result: keys,
				Error:  nil,
			}
		}
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	// Parse the JSONPath expression
	expr, err := jp.ParseString(path)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral(err.Error()),
		}
	}

	// Execute the JSONPath query
	results := expr.Get(jsonData)
	if len(results) == 0 {
		return &EvalResponse{
			Result: RespEmptyArray,
			Error:  nil,
		}
	}

	keysList := make([]interface{}, 0, len(results))

	for _, result := range results {
		switch utils.GetJSONFieldType(result) {
		case utils.ObjectType:
			keys := make([]string, 0)
			for key := range result.(map[string]interface{}) {
				keys = append(keys, key)
			}
			keysList = append(keysList, keys)
		default:
			keysList = append(keysList, nil)
		}
	}

	return &EvalResponse{
		Result: keysList,
		Error:  nil,
	}
}

func evalJSONRESP(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 1 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("JSON.RESP"),
		}
	}
	key := args[0]

	path := defaultRootPath
	if len(args) > 1 {
		path = args[1]
	}

	obj := store.Get(key)
	if obj == nil {
		return &EvalResponse{
			Result: NIL,
			Error:  nil,
		}
	}

	// Check if the object is of JSON type
	if errWithMessage := object.AssertType(obj.Type, object.ObjTypeJSON); errWithMessage != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	jsonData := obj.Value
	if path == defaultRootPath {
		resp := parseJSONStructure(jsonData, false)

		return &EvalResponse{
			Result: resp,
			Error:  nil,
		}
	}

	// if path is not root then extract value at path
	expr, err := jp.ParseString(path)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrInvalidJSONPathType,
		}
	}
	results := expr.Get(jsonData)

	// process value at each path
	ret := make([]any, 0, len(results))

	for _, result := range results {
		resp := parseJSONStructure(result, false)
		ret = append(ret, resp)
	}

	return &EvalResponse{
		Result: ret,
		Error:  nil,
	}
}

func parseJSONStructure(jsonData interface{}, nested bool) (resp []any) {
	switch json := jsonData.(type) {
	case string, bool:
		resp = append(resp, json)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, nil:
		resp = append(resp, json)
	case map[string]interface{}:
		resp = append(resp, "{")
		for key, value := range json {
			resp = append(resp, key)
			resp = append(resp, parseJSONStructure(value, true)...)
		}
		// wrap in another array to offset print
		if nested {
			resp = []interface{}{resp}
		}
	case []interface{}:
		resp = append(resp, "[")
		for _, value := range json {
			resp = append(resp, parseJSONStructure(value, true)...)
		}
		// wrap in another array to offset print
		if nested {
			resp = []interface{}{resp}
		}
	default:
		resp = append(resp, []byte("(unsupported type)"))
	}
	return resp
}

// evalJSONDEBUG reports value's memory usage in bytes
// Returns arity error if subcommand is missing
// Supports only two subcommand as of now - HELP and MEMORY
func evalJSONDebug(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 1 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("JSON.DEBUG"),
		}
	}
	subcommand := strings.ToUpper(args[0])
	switch subcommand {
	case Help:
		return evalJSONDebugHelp()
	case Memory:
		return evalJSONDebugMemory(args[1:], store)
	default:
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral("unknown subcommand - try `JSON.DEBUG HELP`"),
		}
	}
}

// evalJSONDebugHelp implements HELP subcommand for evalJSONDebug
// It returns help text
// It ignore any other args
func evalJSONDebugHelp() *EvalResponse {
	memoryText := "MEMORY <key> [path] - reports memory usage"
	helpText := "HELP                - this message"
	message := []string{memoryText, helpText}
	return &EvalResponse{
		Result: message,
		Error:  nil,
	}
}

// evalJSONDebugMemory implements MEMORY subcommand for evalJSONDebug
// It returns value's memory usage in bytes
func evalJSONDebugMemory(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 1 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("JSON.DEBUG"),
		}
	}
	key := args[0]

	// default path is root if not specified
	path := defaultRootPath
	if len(args) > 1 {
		path = args[1] // anymore args are ignored for this command altogether
	}

	obj := store.Get(key)
	if obj == nil {
		return &EvalResponse{
			Result: IntegerZero,
			Error:  nil,
		}
	}

	// check if the object is a valid JSON
	if errWithMessage := object.AssertType(obj.Type, object.ObjTypeJSON); errWithMessage != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrInvalidJSONPathType,
		}
	}

	// handle root path
	if path == defaultRootPath {
		jsonData := obj.Value

		// memory used by json data
		size := calculateSizeInBytes(jsonData)
		if size == -1 {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongTypeOperation,
			}
		}
		// add memory used by storage object
		size += int(unsafe.Sizeof(obj)) + calculateSizeInBytes(obj.LastAccessedAt) + calculateSizeInBytes(obj.Type)

		return &EvalResponse{
			Result: size,
			Error:  nil,
		}
	}

	// handle nested paths
	var results []any
	if path != defaultRootPath {
		// check if path is valid
		expr, err := jp.ParseString(path)
		if err != nil {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrInvalidJSONPathType,
			}
		}

		results = expr.Get(obj.Value)

		// handle error cases
		if len(results) == 0 {
			// this block will return '[]' for out of bound index for an array json type
			isArray := utils.IsArray(obj.Value)
			if isArray {
				arr, ok := obj.Value.([]any)
				if !ok {
					return &EvalResponse{
						Result: nil,
						Error:  diceerrors.ErrGeneral("invalid array json"),
					}
				}
				// extract index from arg
				reg := regexp.MustCompile(`^\$\.?\[(\d+|\*)\]`)
				matches := reg.FindStringSubmatch(path)

				if len(matches) == 2 {
					// convert index to int
					index, err := strconv.Atoi(matches[1])
					if err != nil {
						return &EvalResponse{
							Result: nil,
							Error:  diceerrors.ErrGeneral("unable to extract index"),
						}
					}
					// if index is out of bound return empty array
					if index >= len(arr) {
						return &EvalResponse{
							Result: EmptyArray,
							Error:  nil,
						}
					}
				}
			}

			// for rest json types, throw error
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrFormatted("Path '$.%v' does not exist", path),
			}
		}
	}

	// get memory used by each path
	sizeList := make([]interface{}, 0, len(results))
	for _, result := range results {
		size := calculateSizeInBytes(result)
		sizeList = append(sizeList, size)
	}

	return &EvalResponse{
		Result: sizeList,
		Error:  nil,
	}
}

func calculateSizeInBytes(value interface{}) int {
	switch convertedValue := value.(type) {
	case string:
		return int(unsafe.Sizeof(value)) + len(convertedValue)

	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool, nil:
		return int(unsafe.Sizeof(value))

	// object
	case map[string]interface{}:
		size := int(unsafe.Sizeof(value))
		for k, v := range convertedValue {
			size += int(unsafe.Sizeof(k)) + len(k) + calculateSizeInBytes(v)
		}
		return size

	// array
	case []interface{}:
		size := int(unsafe.Sizeof(value))
		for _, elem := range convertedValue {
			size += calculateSizeInBytes(elem)
		}
		return size

	// unknown type
	default:
		return -1
	}
}

// helper function to insert key value in hashmap associated with the given hash
func insertInHashMap(args []string, store *dstore.Store) (int64, error) {
	key := args[0]

	obj := store.Get(key)

	var hashMap HashMap

	if obj != nil {
		if err := object.AssertType(obj.Type, object.ObjTypeSSMap); err != nil {
			return 0, diceerrors.ErrWrongTypeOperation
		}
		hashMap = obj.Value.(HashMap)
	}

	keyValuePairs := args[1:]

	hashMap, numKeys, err := hashMapBuilder(keyValuePairs, hashMap)
	if err != nil {
		return 0, err
	}

	obj = store.NewObj(hashMap, -1, object.ObjTypeSSMap)
	store.Put(key, obj)

	return numKeys, nil
}

// evalHSET sets the specified fields to their
// respective values in a hashmap stored at key
//
// This command overwrites the values of specified
// fields that exist in the hash.
//
// If key doesn't exist, a new key holding a hash is created.
//
// Usage: HSET key field value [field value ...]
func evalHSET(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 3 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("HSET"),
		}
	}

	numKeys, err := insertInHashMap(args, store)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  err,
		}
	}

	return &EvalResponse{
		Result: numKeys,
		Error:  nil,
	}
}

// evalHMSET sets the specified fields to their
// respective values in a hashmap stored at key
//
// This command overwrites the values of specified
// fields that exist in the hash.
//
// If key doesn't exist, a new key holding a hash is created.
//
// Usage: HMSET key field value [field value ...]
func evalHMSET(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 3 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("HMSET"),
		}
	}

	_, err := insertInHashMap(args, store)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  err,
		}
	}

	return &EvalResponse{
		Result: OK,
		Error:  nil,
	}
}

// evalHMGET returns an array of values associated with the given fields,
// in the same order as they are requested.
// If a field does not exist, returns a corresponding nil value in the array.
// If the key does not exist, returns an array of nil values.
func evalHMGET(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 2 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("HMGET"),
		}
	}
	key := args[0]

	// Fetch the object from the store using the key
	obj := store.Get(key)

	// Initialize the results slice
	results := make([]interface{}, len(args[1:]))

	// If the object is nil, return empty results for all requested fields
	if obj == nil {
		for i := range results {
			results[i] = nil // Return nil for non-existent fields
		}
		return &EvalResponse{
			Result: results,
			Error:  nil,
		}
	}

	// Assert that the object is of type HashMap
	if err := object.AssertType(obj.Type, object.ObjTypeSSMap); err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	hashMap := obj.Value.(HashMap)

	// Loop through the requested fields
	for i, hmKey := range args[1:] {
		hmValue, present := hashMap.Get(hmKey)
		if present {
			results[i] = *hmValue // Set the value if it exists
		} else {
			results[i] = nil // Set to nil if field does not exist
		}
	}

	// Return the results and no error
	return &EvalResponse{
		Result: results,
		Error:  nil,
	}
}

func evalHGET(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 2 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("HGET"),
		}
	}

	key := args[0]
	hmKey := args[1]

	response := getValueFromHashMap(key, hmKey, store)
	if response.Error != nil {
		return &EvalResponse{
			Result: nil,
			Error:  response.Error,
		}
	}

	return &EvalResponse{
		Result: response.Result,
		Error:  nil,
	}
}

func evalHGETALL(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 1 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("HGETALL"),
		}
	}

	key := args[0]

	obj := store.Get(key)

	var hashMap HashMap
	var results []string

	if obj != nil {
		if err := object.AssertType(obj.Type, object.ObjTypeSSMap); err != nil {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongTypeOperation,
			}
		}
		hashMap = obj.Value.(HashMap)
	}

	for hmKey, hmValue := range hashMap {
		results = append(results, hmKey, hmValue)
	}

	return &EvalResponse{
		Result: results,
		Error:  nil,
	}
}

func evalHSETNX(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 3 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("HSETNX"),
		}
	}

	key := args[0]
	hmKey := args[1]

	response := getValueFromHashMap(key, hmKey, store)
	if response.Error != nil {
		return &EvalResponse{
			Result: nil,
			Error:  response.Error,
		}
	}

	if response.Result != NIL {
		return &EvalResponse{
			Result: int64(0),
			Error:  nil,
		}
	}

	evalHSET(args, store)

	return &EvalResponse{
		Result: int64(1),
		Error:  nil,
	}
}

func evalHDEL(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 2 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("HDEL"),
		}
	}

	key := args[0]
	fields := args[1:]

	obj := store.Get(key)

	if obj == nil {
		return &EvalResponse{
			Result: int64(0),
			Error:  nil,
		}
	}

	if err := object.AssertType(obj.Type, object.ObjTypeSSMap); err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	hashMap := obj.Value.(HashMap)
	count := int64(0)
	for _, field := range fields {
		if _, ok := hashMap[field]; ok {
			delete(hashMap, field)
			count++
		}
	}

	if count > 0 {
		store.Put(key, obj)
	}

	return &EvalResponse{
		Result: count,
		Error:  nil,
	}
}

// evalSADD adds one or more members to a set
// args must contain a key and one or more members to add the set
// If the set does not exist, a new set is created and members are added to it
// An error response is returned if the command is used on a key that contains a non-set value(eg: string)
// Returns an integer which represents the number of members that were added to the set, not including
// the members that were already present
func evalSADD(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 2 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("SADD"),
		}
	}
	key := args[0]

	// Get the set object from the store.
	obj := store.Get(key)
	lengthOfItems := len(args[1:])

	var count = 0
	if obj == nil {
		var exDurationMs int64 = -1
		var keepttl = false
		// If the object does not exist, create a new set object.
		value := make(map[string]struct{}, lengthOfItems)
		// Create a new object.
		obj = store.NewObj(value, exDurationMs, object.ObjTypeSet)
		store.Put(key, obj, dstore.WithKeepTTL(keepttl))
	}

	if err := object.AssertType(obj.Type, object.ObjTypeSet); err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	// Get the set object.
	set := obj.Value.(map[string]struct{})

	for _, arg := range args[1:] {
		if _, ok := set[arg]; !ok {
			set[arg] = struct{}{}
			count++
		}
	}

	return &EvalResponse{
		Result: count,
		Error:  nil,
	}
}

// evalSREM removes one or more members from a set
// Members that are not member of this set are ignored
// Returns the number of members that are removed from set
// If set does not exist, 0 is returned
// An error response is returned if the command is used on a key that contains a non-set value(eg: string)
func evalSREM(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 2 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("SREM"),
		}
	}
	key := args[0]

	// Get the set object from the store.
	obj := store.Get(key)

	var count = 0
	if obj == nil {
		return &EvalResponse{
			Result: 0,
			Error:  nil,
		}
	}

	// If the object exists, check if it is a set object.
	if err := object.AssertType(obj.Type, object.ObjTypeSet); err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	// Get the set object.
	set := obj.Value.(map[string]struct{})

	for _, arg := range args[1:] {
		if _, ok := set[arg]; ok {
			delete(set, arg)
			count++
		}
	}

	return &EvalResponse{
		Result: count,
		Error:  nil,
	}
}

// evalSCARD returns the number of elements of the set stored at key
// Returns 0 if the key does not exist
// An error response is returned if the command is used on a key that contains a non-set value(eg: string)
func evalSCARD(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 1 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("SCARD"),
		}
	}

	key := args[0]

	// Get the set object from the store.
	obj := store.Get(key)

	if obj == nil {
		return &EvalResponse{
			Result: 0,
			Error:  nil,
		}
	}

	// If the object exists, check if it is a set object.
	if err := object.AssertType(obj.Type, object.ObjTypeSet); err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	// Get the set object.
	count := len(obj.Value.(map[string]struct{}))
	return &EvalResponse{
		Result: count,
		Error:  nil,
	}
}

// evalSMEMBERS returns all the members of a set
// An error response is returned if the command is used on a key that contains a non-set value(eg: string)
// An empty set is returned if no set exists for given key
func evalSMEMBERS(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 1 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("SMEMBERS"),
		}
	}
	key := args[0]

	// Get the set object from the store.
	obj := store.Get(key)

	if obj == nil {
		return &EvalResponse{
			Result: []string{},
			Error:  nil,
		}
	}

	// If the object exists, check if it is a set object.
	if err := object.AssertType(obj.Type, object.ObjTypeSet); err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	// Get the set object.
	set := obj.Value.(map[string]struct{})
	// Get the members of the set.
	members := make([]string, 0, len(set))
	for k := range set {
		members = append(members, k)
	}

	return &EvalResponse{
		Result: members,
		Error:  nil,
	}
}

// evalLRANGE returns the specified elements of the list stored at key.
//
// Returns Array reply: a list of elements in the specified range, or an empty array if the key doesn't exist.
//
// Usage: LRANGE key start stop
func evalLRANGE(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 3 {
		return makeEvalError(errors.New("-wrong number of arguments for LRANGE"))
	}
	key := args[0]
	start, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return makeEvalError(errors.New("-ERR value is not an integer or out of range"))
	}
	stop, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil {
		return makeEvalError(errors.New("-ERR value is not an integer or out of range"))
	}

	obj := store.Get(key)
	if obj == nil {
		return makeEvalResult([]string{})
	}

	if object.AssertType(obj.Type, object.ObjTypeDequeue) != nil {
		return makeEvalError(errors.New(diceerrors.WrongTypeErr))
	}

	q := obj.Value.(*Deque)
	res, err := q.LRange(start, stop)
	if err != nil {
		return makeEvalError(err)
	}
	return makeEvalResult(res)
}

// evalLINSERT command inserts the element at a key before / after the pivot element.
//
// Returns the list length (integer) after a successful insert operation, 0 when the key doesn't exist, -1 when the pivot wasn't found.
//
// Usage: LINSERT key <BEFORE | AFTER> pivot element
func evalLINSERT(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 4 {
		return makeEvalError(errors.New("-wrong number of arguments for LINSERT"))
	}

	key := args[0]
	beforeAfter := strings.ToLower(args[1])
	pivot := args[2]
	element := args[3]

	obj := store.Get(key)
	if obj == nil {
		return makeEvalResult(0)
	}

	// if object is a set type, return error
	if object.AssertType(obj.Type, object.ObjTypeDequeue) != nil {
		return makeEvalError(errors.New(diceerrors.WrongTypeErr))
	}

	q := obj.Value.(*Deque)
	res, err := q.LInsert(pivot, element, beforeAfter)
	if err != nil {
		return makeEvalError(err)
	}
	return makeEvalResult(res)
}

// SETBIT key offset value
func evalSETBIT(args []string, store *dstore.Store) *EvalResponse {
	var err error

	if len(args) != 3 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("SETBIT"),
		}
	}

	key := args[0]
	offset, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil || offset < 0 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral("bit offset is not an integer or out of range"),
		}
	}

	value, err := strconv.ParseBool(args[2])
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral("bit is not an integer or out of range"),
		}
	}

	obj := store.Get(key)
	requiredByteArraySize := offset>>3 + 1

	if obj == nil {
		obj = store.NewObj(NewByteArray(int(requiredByteArraySize)), -1, object.ObjTypeByteArray)
		store.Put(args[0], obj)
	}

	if object.AssertType(obj.Type, object.ObjTypeByteArray) == nil ||
		object.AssertType(obj.Type, object.ObjTypeString) == nil ||
		object.AssertType(obj.Type, object.ObjTypeInt) == nil {
		var byteArray *ByteArray
		oType := obj.Type

		switch oType {
		case object.ObjTypeByteArray:
			byteArray = obj.Value.(*ByteArray)
		case object.ObjTypeString, object.ObjTypeInt:
			byteArray, err = NewByteArrayFromObj(obj)
			if err != nil {
				return &EvalResponse{
					Result: nil,
					Error:  diceerrors.ErrWrongTypeOperation,
				}
			}
		default:
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongTypeOperation,
			}
		}

		// Perform the resizing check
		byteArrayLength := byteArray.Length

		// check whether resize required or not
		if requiredByteArraySize > byteArrayLength {
			// resize as per the offset
			byteArray = byteArray.IncreaseSize(int(requiredByteArraySize))
		}
		resp := byteArray.GetBit(int(offset))
		byteArray.SetBit(int(offset), value)

		// We are returning newObject here so it is thread-safe
		// Old will be removed by GC
		newObj, err := ByteSliceToObj(store, obj, byteArray.data, oType)
		if err != nil {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongTypeOperation,
			}
		}

		exp, ok := dstore.GetExpiry(obj, store)
		var exDurationMs int64 = -1
		if ok {
			exDurationMs = exp - time.Now().UnixMilli()
		}
		// newObj has bydefault expiry time -1 , we need to set it
		if exDurationMs > 0 {
			store.SetExpiry(newObj, exDurationMs)
		}

		store.Put(key, newObj)
		if resp {
			return &EvalResponse{
				Result: IntegerOne,
				Error:  nil,
			}
		}
		return &EvalResponse{
			Result: IntegerZero,
			Error:  nil,
		}
	}
	return &EvalResponse{
		Result: nil,
		Error:  diceerrors.ErrWrongTypeOperation,
	}
}

// GETBIT key offset
func evalGETBIT(args []string, store *dstore.Store) *EvalResponse {
	var err error

	if len(args) != 2 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("GETBIT"),
		}
	}

	key := args[0]
	offset, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrIntegerOutOfRange,
		}
	}

	obj := store.Get(key)
	if obj == nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	requiredByteArraySize := offset>>3 + 1
	switch oType := obj.Type; oType {
	case object.ObjTypeSet:
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	case object.ObjTypeByteArray:
		byteArray := obj.Value.(*ByteArray)
		byteArrayLength := byteArray.Length

		// check whether offset, length exists or not
		if requiredByteArraySize > byteArrayLength {
			return &EvalResponse{
				Result: IntegerZero,
				Error:  nil,
			}
		}
		value := byteArray.GetBit(int(offset))
		if value {
			return &EvalResponse{
				Result: IntegerOne,
				Error:  nil,
			}
		}
		return &EvalResponse{
			Result: IntegerZero,
			Error:  nil,
		}

	case object.ObjTypeString, object.ObjTypeInt:
		byteArray, err := NewByteArrayFromObj(obj)
		if err != nil {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongTypeOperation,
			}
		}
		if requiredByteArraySize > byteArray.Length {
			return &EvalResponse{
				Result: IntegerZero,
				Error:  nil,
			}
		}
		value := byteArray.GetBit(int(offset))
		if value {
			return &EvalResponse{
				Result: IntegerOne,
				Error:  nil,
			}
		}
		return &EvalResponse{
			Result: IntegerZero,
			Error:  nil,
		}

	default:
		return &EvalResponse{
			Result: IntegerZero,
			Error:  nil,
		}
	}
}

func evalBITCOUNT(args []string, store *dstore.Store) *EvalResponse {
	var err error

	// if no key is provided, return error
	if len(args) == 0 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("BITCOUNT"),
		}
	}

	// if more than 4 arguments are provided, return error
	if len(args) > 4 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrInvalidSyntax("GETBIT"),
		}
	}

	// fetching value of the key
	key := args[0]
	obj := store.Get(key)
	if obj == nil {
		return &EvalResponse{
			Result: IntegerZero,
			Error:  nil,
		}
	}

	var value []byte
	var valueLength int64

	switch {
	case object.AssertType(obj.Type, object.ObjTypeByteArray) == nil:
		byteArray := obj.Value.(*ByteArray)
		value = byteArray.data
		valueLength = byteArray.Length
	case object.AssertType(obj.Type, object.ObjTypeString) == nil:
		value = []byte(obj.Value.(string))
		valueLength = int64(len(value))
	case object.AssertType(obj.Type, object.ObjTypeInt) == nil:
		value = []byte(strconv.FormatInt(obj.Value.(int64), 10))
		valueLength = int64(len(value))
	default:
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	// defining constants of the function
	start, end := int64(0), valueLength-1
	unit := BYTE

	// checking which arguments are present and validating arguments
	if len(args) > 1 {
		start, err = strconv.ParseInt(args[1], 10, 64)
		if err != nil {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrIntegerOutOfRange,
			}
		}
		if len(args) <= 2 {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrInvalidSyntax("BITCOUNT"),
			}
		}
		end, err = strconv.ParseInt(args[2], 10, 64)
		if err != nil {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrIntegerOutOfRange,
			}
		}
	}
	if len(args) > 3 {
		unit = strings.ToUpper(args[3])
	}

	switch unit {
	case BYTE:
		if start < 0 {
			start += valueLength
		}
		if end < 0 {
			end += valueLength
		}
		if start > end || start >= valueLength {
			return &EvalResponse{
				Result: IntegerZero,
				Error:  nil,
			}
		}
		end = min(end, valueLength-1)
		bitCount := 0
		for i := start; i <= end; i++ {
			bitCount += bits.OnesCount8(value[i])
		}
		return &EvalResponse{
			Result: bitCount,
			Error:  nil,
		}
	case BIT:
		if start < 0 {
			start += valueLength * 8
		}
		if end < 0 {
			end += valueLength * 8
		}
		if start > end {
			return &EvalResponse{
				Result: IntegerZero,
				Error:  nil,
			}
		}
		startByte, endByte := start/8, min(end/8, valueLength-1)
		startBitOffset, endBitOffset := start%8, end%8

		if endByte == valueLength-1 {
			endBitOffset = 7
		}

		if startByte >= valueLength {
			return &EvalResponse{
				Result: IntegerZero,
				Error:  nil,
			}
		}

		bitCount := 0

		// Use bit masks to count the bits instead of a loop
		if startByte == endByte {
			mask := byte(0xFF >> startBitOffset)
			mask &= byte(0xFF << (7 - endBitOffset))
			bitCount = bits.OnesCount8(value[startByte] & mask)
		} else {
			// Handle first byte
			firstByteMask := byte(0xFF >> startBitOffset)
			bitCount += bits.OnesCount8(value[startByte] & firstByteMask)

			// Handle all the middle ones
			for i := startByte + 1; i < endByte; i++ {
				bitCount += bits.OnesCount8(value[i])
			}

			// Handle last byte
			lastByteMask := byte(0xFF << (7 - endBitOffset))
			bitCount += bits.OnesCount8(value[endByte] & lastByteMask)
		}
		return &EvalResponse{
			Result: bitCount,
			Error:  nil,
		}
	default:
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrInvalidSyntax("BITCOUNT"),
		}
	}
}

// Generic method for both BITFIELD and BITFIELD_RO.
// isReadOnly method is true for BITFIELD_RO command.
func bitfieldEvalGeneric(args []string, store *dstore.Store, isReadOnly bool) *EvalResponse {
	var ops []utils.BitFieldOp
	ops, err2 := utils.ParseBitfieldOps(args, isReadOnly)

	if err2 != nil {
		return &EvalResponse{
			Result: nil,
			Error:  err2,
		}
	}

	key := args[0]
	obj := store.Get(key)
	if obj == nil {
		obj = store.NewObj(NewByteArray(1), -1, object.ObjTypeByteArray)
		store.Put(args[0], obj)
	}
	var value *ByteArray
	var err error

	switch oType := obj.Type; oType {
	case object.ObjTypeByteArray:
		value = obj.Value.(*ByteArray)
	case object.ObjTypeString, object.ObjTypeInt:
		value, err = NewByteArrayFromObj(obj)
		if err != nil {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrGeneral("value is not a valid byte array"),
			}
		}
	default:
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	result := executeBitfieldOps(value, ops)
	return &EvalResponse{
		Result: result,
		Error:  nil,
	}
}

// evalBITFIELD evaluates BITFIELD operations on a key store string, int or bytearray types
// it returns an array of results depending on the subcommands
// it allows mutation using SET and INCRBY commands
// returns arity error, offset type error, overflow type error, encoding type error, integer error, syntax error
// GET <encoding> <offset> -- Returns the specified bit field.
// SET <encoding> <offset> <value> -- Set the specified bit field
// and returns its old value.
// INCRBY <encoding> <offset> <increment> -- Increments or decrements
// (if a negative increment is given) the specified bit field and returns the new value.
// There is another subcommand that only changes the behavior of successive
// INCRBY and SET subcommands calls by setting the overflow behavior:
// OVERFLOW [WRAP|SAT|FAIL]`
func evalBITFIELD(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 1 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("BITFIELD"),
		}
	}

	return bitfieldEvalGeneric(args, store, false)
}

// Read-only variant of the BITFIELD command. It is like the original BITFIELD but only accepts GET subcommand and can safely be used in read-only replicas.
func evalBITFIELDRO(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 1 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("BITFIELD_RO"),
		}
	}

	return bitfieldEvalGeneric(args, store, true)
}

func evalGetObject(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 1 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrInternalServer,
		}
	}

	key := args[0]

	obj := store.Get(key)

	// if key does not exist, return RESP encoded nil
	if obj == nil {
		return &EvalResponse{
			Result: IntegerZero,
			Error:  nil,
		}
	}

	exp, ok := dstore.GetExpiry(obj, store)
	var exDurationMs int64 = -1
	if ok {
		exDurationMs = exp - time.Now().UnixMilli()
	}

	exObj := &object.InternalObj{
		Obj:        obj,
		ExDuration: exDurationMs,
	}

	// Decode and return the value based on its encoding
	return &EvalResponse{
		Result: exObj,
		Error:  nil,
	}
}

// evalJSONSTRAPPEND appends a string value to the JSON string value at the specified path
// in the JSON object saved at the key in arguments.
// Args must contain at least a key and the string value to append.
// If the key does not exist or is expired, it returns an error response.
// If the value at the specified path is not a string, it returns an error response.
// Returns the new length of the string at the specified path if successful.
func evalJSONSTRAPPEND(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 3 {
		return makeEvalError(diceerrors.ErrWrongArgumentCount("JSON.STRAPPEND"))
	}

	key := args[0]
	path := args[1]
	value := args[2]

	obj := store.Get(key)
	if obj == nil {
		return makeEvalError(diceerrors.ErrKeyDoesNotExist)
	}

	if errWithMessage := object.AssertType(obj.Type, object.ObjTypeJSON); errWithMessage != nil {
		return makeEvalError(diceerrors.ErrWrongTypeOperation)
	}

	jsonData := obj.Value

	var resultsArray []interface{}

	if path == "$" {
		// Handle root-level string
		if str, ok := jsonData.(string); ok {
			unquotedValue := strings.Trim(value, "\"")
			newValue := str + unquotedValue
			resultsArray = append(resultsArray, len(newValue))
			jsonData = newValue
		} else {
			return makeEvalResult(EmptyArray)
		}
	} else {
		expr, err := jp.ParseString(path)
		if err != nil {
			return makeEvalResult(EmptyArray)
		}

		_, modifyErr := expr.Modify(jsonData, func(data any) (interface{}, bool) {
			switch v := data.(type) {
			case string:
				unquotedValue := strings.Trim(value, "\"")
				newValue := v + unquotedValue
				resultsArray = append([]interface{}{len(newValue)}, resultsArray...)
				return newValue, true
			default:
				resultsArray = append([]interface{}{RespNIL}, resultsArray...)
				return data, false
			}
		})

		if modifyErr != nil {
			return makeEvalResult(EmptyArray)
		}
	}

	if len(resultsArray) == 0 {
		return makeEvalResult(EmptyArray)
	}

	obj.Value = jsonData
	return makeEvalResult(resultsArray[0])
}

// evalJSONTOGGLE toggles a boolean value stored at the specified key and path.
// args must contain at least the key and path (where the boolean is located).
// If the key does not exist or is expired, it returns response.RespNIL.
// If the field at the specified path is not a boolean, it returns an encoded error response.
// If the boolean is `true`, it toggles to `false` (returns :0), and if `false`, it toggles to `true` (returns :1).
// Returns an encoded error response if the incorrect number of arguments is provided.
func evalJSONTOGGLE(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 2 {
		return makeEvalError(diceerrors.ErrWrongArgumentCount("JSON.TOGGLE"))
	}
	key := args[0]
	path := args[1]

	obj := store.Get(key)
	if obj == nil {
		return makeEvalError(diceerrors.ErrKeyDoesNotExist)
	}

	if errWithMessage := object.AssertType(obj.Type, object.ObjTypeJSON); errWithMessage != nil {
		return makeEvalError(diceerrors.ErrWrongTypeOperation)
	}

	jsonData := obj.Value
	_, err := sonic.Marshal(jsonData)

	if err != nil {
		return makeEvalError(diceerrors.ErrWrongTypeOperation)
	}

	expr, err := jp.ParseString(path)

	if err != nil {
		return makeEvalError(diceerrors.ErrJSONPathNotFound(path))
	}

	matches := expr.Get(jsonData)
	if len(matches) == 0 {
		return makeEvalError(diceerrors.ErrJSONPathNotFound(path))
	}

	var toggleResults []interface{}
	modified := false

	_, err = expr.Modify(jsonData, func(value interface{}) (interface{}, bool) {
		boolValue, ok := value.(bool)
		if !ok {
			toggleResults = append(toggleResults, nil)
			return value, false
		}
		newValue := !boolValue
		toggleResults = append(toggleResults, utils.BoolToInt(newValue))
		modified = true
		return newValue, true
	})

	if err != nil {
		return makeEvalError(diceerrors.ErrGeneral("failed to toggle values"))
	}

	if modified {
		obj.Value = jsonData
	}

	toggleResults = reverseSlice(toggleResults)
	return makeEvalResult(toggleResults)
}

// evaLJSONFORGET removes the field specified by the given JSONPath from the JSON document stored under the provided key.
// calls the evalJSONDEL() with the arguments passed
// If the specified key has expired or does not exist, it returns 0.
// Returns encoded error response if incorrect number of arguments
// If the JSONPath points to the root of the JSON document, the entire key is deleted from the store.
// Returns an integer reply specified as the number of paths deleted (0 or more)
func evalJSONFORGET(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 1 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("JSON.FORGET"),
		}
	}

	return evalJSONDEL(args, store)
}

// evalJSONDEL deletes a value specified by the given JSON path from the store.
// It returns an integer indicating the number of paths deleted (0 or more).
// If the specified key has expired or does not exist, it returns 0.
// If the number of arguments provided is incorrect, it returns an encoded error response.
func evalJSONDEL(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 1 {
		return makeEvalError(diceerrors.ErrWrongArgumentCount("JSON.DEL"))
	}
	key := args[0]

	// Default path is root if not specified
	path := defaultRootPath
	if len(args) > 1 {
		path = args[1]
	}

	// Retrieve the object from the database
	obj := store.Get(key)
	if obj == nil {
		return makeEvalResult(IntegerZero)
	}

	if errWithMessage := object.AssertType(obj.Type, object.ObjTypeJSON); errWithMessage != nil {
		return makeEvalError(diceerrors.ErrWrongTypeOperation)
	}

	jsonData := obj.Value

	_, err := sonic.Marshal(jsonData)
	if err != nil {
		return makeEvalError(diceerrors.ErrWrongTypeOperation)
	}

	if len(args) == 1 || path == defaultRootPath {
		store.Del(key)
		return makeEvalResult(1)
	}

	expr, err := jp.ParseString(path)
	if err != nil {
		return makeEvalError(diceerrors.ErrJSONPathNotFound(path))
	}
	results := expr.Get(jsonData)

	hasBrackets := strings.Contains(path, "[") && strings.Contains(path, "]")

	// If the command has square brackets then we have to delete an element inside an array
	if hasBrackets {
		_, err = expr.Remove(jsonData)
	} else {
		err = expr.Del(jsonData)
	}

	if err != nil {
		return makeEvalError(diceerrors.ErrInternalServer) // no need to send actual internal error
	}
	// Create a new object with the updated JSON data
	newObj := store.NewObj(jsonData, -1, object.ObjTypeJSON)
	store.Put(key, newObj)

	return makeEvalResult(len(results))
}

// Returns the new value after incrementing or multiplying the existing value
func incrMultValue(value any, multiplier interface{}, operation jsonOperation) (newVal interface{}, resultString string, isModified bool) {
	switch utils.GetJSONFieldType(value) {
	case utils.NumberType:
		oldVal := value.(float64)
		var newVal float64
		if v, ok := multiplier.(float64); ok {
			switch operation {
			case IncrBy:
				newVal = oldVal + v
			case MultBy:
				newVal = oldVal * v
			}
		} else {
			v, _ := multiplier.(int64)
			switch operation {
			case IncrBy:
				newVal = oldVal + float64(v)
			case MultBy:
				newVal = oldVal * float64(v)
			}
		}
		resultString := strconv.FormatFloat(newVal, 'f', -1, 64)
		return newVal, resultString, true
	case utils.IntegerType:
		if v, ok := multiplier.(float64); ok {
			oldVal := float64(value.(int64))
			var newVal float64
			switch operation {
			case IncrBy:
				newVal = oldVal + v
			case MultBy:
				newVal = oldVal * v
			}
			resultString := strconv.FormatFloat(newVal, 'f', -1, 64)
			return newVal, resultString, true
		}
		v, _ := multiplier.(int64)
		oldVal := value.(int64)
		var newVal int64
		switch operation {
		case IncrBy:
			newVal = oldVal + v
		case MultBy:
			newVal = oldVal * v
		}
		resultString := strconv.FormatInt(newVal, 10)
		return newVal, resultString, true
	default:
		return value, "null", false
	}
}

// evalJSONNUMMULTBY multiplies the JSON fields matching the specified JSON path at the specified key
// args must contain key, JSON path and the multiplier value
// Returns encoded error response if incorrect number of arguments
// Returns encoded error if the JSON path or key is invalid
// Returns bulk string reply specified as a stringified updated values for each path
// Returns null if matching field is non-numerical
func evalJSONNUMMULTBY(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 3 {
		return makeEvalError(diceerrors.ErrWrongArgumentCount("JSON.NUMMULTBY"))
	}
	key := args[0]

	// Retrieve the object from the database
	obj := store.Get(key)
	if obj == nil {
		return makeEvalError(diceerrors.ErrKeyDoesNotExist)
	}

	// Check if the object is of JSON type
	if errWithMessage := object.AssertType(obj.Type, object.ObjTypeJSON); errWithMessage != nil {
		return makeEvalError(diceerrors.ErrWrongTypeOperation)
	}
	path := args[1]
	// Parse the JSONPath expression
	expr, err := jp.ParseString(path)

	if err != nil {
		return makeEvalError(diceerrors.ErrJSONPathNotFound(path))
	}

	// Get json matching expression
	jsonData := obj.Value
	results := expr.Get(jsonData)
	if len(results) == 0 {
		return makeEvalResult("[]")
	}

	for i, r := range args[2] {
		if !unicode.IsDigit(r) && r != '.' && r != '-' && r != 'e' && r != 'E' {
			if i == 0 {
				return makeEvalError(diceerrors.ErrGeneral(fmt.Sprintf("expected value at line 1 column %d", i+1)))
			}

			return makeEvalError(diceerrors.ErrGeneral(fmt.Sprintf("trailing characters at line 1 column %d", i+1)))
		}
	}

	// Parse the mulplier value
	multiplier, err := parseFloatInt(args[2])
	if err != nil {
		return makeEvalError(diceerrors.ErrIntegerOutOfRange)
	}

	// Update matching values using Modify
	resultArray := make([]string, 0, len(results))
	if path == defaultRootPath {
		newValue, resultString, modified := incrMultValue(jsonData, multiplier, MultBy)
		if modified {
			jsonData = newValue
		}
		resultArray = append(resultArray, resultString)
	} else {
		_, err := expr.Modify(jsonData, func(value any) (interface{}, bool) {
			newValue, resultString, modified := incrMultValue(value, multiplier, MultBy)
			resultArray = append(resultArray, resultString)
			return newValue, modified
		})
		if err != nil {
			return makeEvalError(diceerrors.ErrJSONPathNotFound(path))
		}
	}

	// Stringified updated values
	resultString := `[` + strings.Join(resultArray, ",") + `]`

	newObj := &object.Obj{
		Value: jsonData,
		Type:  object.ObjTypeJSON,
	}
	exp, ok := dstore.GetExpiry(obj, store)

	var exDurationMs int64 = -1
	if ok {
		exDurationMs = exp - time.Now().UnixMilli()
	}
	// newObj has default expiry time of -1 , we need to set it
	if exDurationMs > 0 {
		store.SetExpiry(newObj, exDurationMs)
	}

	store.Put(key, newObj)
	return makeEvalResult(resultString)
}

// evalSetObject stores an object in the store with a given key and optional expiry.
// If an object with the same key exists, it is replaced.
// This function is usually specifc to multishard multi-op commands
func evalCOPYObject(cd *cmd.DiceDBCmd, store *dstore.Store) *EvalResponse {
	args := cd.Args

	var isReplace bool
	if len(cd.Args) > 1 {
		if cd.Args[1] == "REPLACE" {
			isReplace = true
		}
	}

	key := args[0]

	obj := store.Get(key)

	if !isReplace && obj != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral("ERR target key already exists"),
		}
	}

	store.Del(key)

	if len(cd.InternalObjs) < 1 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral("ERR dest key not present"),
		}
	}

	copyObj := cd.InternalObjs[0].Obj.DeepCopy()
	if copyObj == nil {
		return &EvalResponse{
			Result: IntegerZero,
			Error:  nil,
		}
	}

	exDurationMs := cd.InternalObjs[0].ExDuration

	store.Put(key, copyObj)

	if exDurationMs > 0 {
		store.SetExpiry(copyObj, exDurationMs)
	}

	return &EvalResponse{
		Result: IntegerOne,
		Error:  nil,
	}
}

func evalPFMERGE(cd *cmd.DiceDBCmd, store *dstore.Store) *EvalResponse {
	if len(cd.Args) < 1 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("PFMERGE"),
		}
	}

	var mergedHll *hyperloglog.Sketch
	destKey := cd.Args[0]
	obj := store.Get(destKey)

	// If destKey doesn't exist, create a new HLL, else fetch the existing
	if obj == nil {
		mergedHll = hyperloglog.New()
	} else {
		var ok bool
		mergedHll, ok = obj.Value.(*hyperloglog.Sketch)
		if !ok {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrInvalidHyperLogLogKey,
			}
		}
	}

	for _, obj := range cd.InternalObjs {
		if obj != nil {
			currKeyHll, ok := obj.Obj.Value.(*hyperloglog.Sketch)
			if !ok {
				return &EvalResponse{
					Result: nil,
					Error:  diceerrors.ErrInvalidHyperLogLogKey,
				}
			}

			err := mergedHll.Merge(currKeyHll)
			if err != nil {
				return &EvalResponse{
					Result: nil,
					Error:  diceerrors.ErrCorruptedHyperLogLogObject,
				}
			}
		}
	}

	// Save the mergedHll
	obj = store.NewObj(mergedHll, -1, object.ObjTypeHLL)
	store.Put(destKey, obj, dstore.WithPutCmd(dstore.PFMERGE))

	return &EvalResponse{
		Result: OK,
		Error:  nil,
	}
}

// takes original value, increment values (float or int), a flag representing if increment is float
// returns new value, string representation, a boolean representing if the value was modified
func incrementValue(value any, isIncrFloat bool, incrFloat float64, incrInt int64) (newVal interface{}, stringRepresentation string, isModified bool) {
	switch utils.GetJSONFieldType(value) {
	case utils.NumberType:
		oldVal := value.(float64)
		var newVal float64
		if isIncrFloat {
			newVal = oldVal + incrFloat
		} else {
			newVal = oldVal + float64(incrInt)
		}
		resultString := formatFloat(newVal, isIncrFloat)
		return newVal, resultString, true
	case utils.IntegerType:
		if isIncrFloat {
			oldVal := float64(value.(int64))
			newVal := oldVal + incrFloat
			resultString := formatFloat(newVal, isIncrFloat)
			return newVal, resultString, true
		}
		oldVal := value.(int64)
		newVal := oldVal + incrInt
		resultString := fmt.Sprintf("%d", newVal)
		return newVal, resultString, true
	default:
		return value, null, false
	}
}

func evalJSONNUMINCRBY(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 3 {
		return makeEvalError(diceerrors.ErrWrongArgumentCount("JSON.NUMINCRBY"))
	}
	key := args[0]
	obj := store.Get(key)

	if obj == nil {
		return makeEvalError(diceerrors.ErrKeyDoesNotExist)
	}

	// Check if the object is of JSON type
	if errWithMessage := object.AssertType(obj.Type, object.ObjTypeJSON); errWithMessage != nil {
		return makeEvalError(diceerrors.ErrWrongTypeOperation)
	}

	path := args[1]

	jsonData := obj.Value
	// Parse the JSONPath expression
	expr, err := jp.ParseString(path)

	if err != nil {
		return makeEvalError(diceerrors.ErrJSONPathNotFound(path))
	}

	isIncrFloat := false

	for i, r := range args[2] {
		if !unicode.IsDigit(r) && r != '.' && r != '-' {
			if i == 0 {
				return makeEvalError(diceerrors.ErrGeneral(fmt.Sprintf("expected value at line 1 column %d", i+1)))
			}
			return makeEvalError(diceerrors.ErrGeneral(fmt.Sprintf("trailing characters at line 1 column %d", i+1)))
		}
		if r == '.' {
			isIncrFloat = true
		}
	}
	var incrFloat float64
	var incrInt int64
	if isIncrFloat {
		incrFloat, err = strconv.ParseFloat(args[2], 64)
		if err != nil {
			return makeEvalError(diceerrors.ErrIntegerOutOfRange)
		}
	} else {
		incrInt, err = strconv.ParseInt(args[2], 10, 64)

		if err != nil {
			return makeEvalError(diceerrors.ErrIntegerOutOfRange)
		}
	}
	results := expr.Get(jsonData)

	if len(results) == 0 {
		respString := "[]"
		return makeEvalResult(respString)
	}

	resultArray := make([]string, 0, len(results))

	if path == defaultRootPath {
		newValue, resultString, isModified := incrementValue(jsonData, isIncrFloat, incrFloat, incrInt)
		if isModified {
			jsonData = newValue
		}
		resultArray = append(resultArray, resultString)
	} else {
		// Execute the JSONPath query
		_, err := expr.Modify(jsonData, func(value any) (interface{}, bool) {
			newValue, resultString, isModified := incrementValue(value, isIncrFloat, incrFloat, incrInt)
			resultArray = append(resultArray, resultString)
			return newValue, isModified
		})
		if err != nil {
			return makeEvalError(diceerrors.ErrJSONPathNotFound(path))
		}
	}

	resultString := `[` + strings.Join(resultArray, ",") + `]`

	obj.Value = jsonData

	return makeEvalResult(resultString)
}

func evalGEOADD(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 4 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("GEOADD"),
		}
	}

	key := args[0]
	var nx, xx bool
	startIdx := 1

	// Parse options
	for startIdx < len(args) {
		option := strings.ToUpper(args[startIdx])
		if option == "NX" {
			nx = true
			startIdx++
		} else if option == "XX" {
			xx = true
			startIdx++
		} else {
			break
		}
	}

	// Check if we have the correct number of arguments after parsing options
	if (len(args)-startIdx)%3 != 0 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("GEOADD"),
		}
	}

	if xx && nx {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral("XX and NX options at the same time are not compatible"),
		}
	}

	// Get or create sorted set
	obj := store.Get(key)
	var ss *sortedset.Set
	if obj != nil {
		var err []byte
		ss, err = sortedset.FromObject(obj)
		if err != nil {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongTypeOperation,
			}
		}
	} else {
		ss = sortedset.New()
	}

	added := 0
	for i := startIdx; i < len(args); i += 3 {
		longitude, err := strconv.ParseFloat(args[i], 64)
		if err != nil || math.IsNaN(longitude) || longitude < -180 || longitude > 180 {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrGeneral("invalid longitude"),
			}
		}

		latitude, err := strconv.ParseFloat(args[i+1], 64)
		if err != nil || math.IsNaN(latitude) || latitude < -85.05112878 || latitude > 85.05112878 {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrGeneral("invalid latitude"),
			}
		}

		member := args[i+2]
		_, exists := ss.Get(member)

		// Handle XX option: Only update existing elements
		if xx && !exists {
			continue
		}

		// Handle NX option: Only add new elements
		if nx && exists {
			continue
		}

		hash := geo.EncodeInt(latitude, longitude)

		wasInserted := ss.Upsert(hash, member)
		if wasInserted {
			added++
		}
	}

	obj = store.NewObj(ss, -1, object.ObjTypeSortedSet)
	store.Put(key, obj)

	return &EvalResponse{
		Result: added,
		Error:  nil,
	}
}

func evalGEODIST(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 3 || len(args) > 4 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("GEODIST"),
		}
	}

	key := args[0]
	member1 := args[1]
	member2 := args[2]
	unit := "m"
	if len(args) == 4 {
		unit = strings.ToLower(args[3])
	}

	// Get the sorted set
	obj := store.Get(key)
	if obj == nil {
		return &EvalResponse{
			Result: NIL,
			Error:  nil,
		}
	}
	ss, err := sortedset.FromObject(obj)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	// Get the scores (geohashes) for both members
	score1, ok := ss.Get(member1)
	if !ok {
		return &EvalResponse{
			Result: nil,
			Error:  nil,
		}
	}
	score2, ok := ss.Get(member2)
	if !ok {
		return &EvalResponse{
			Result: nil,
			Error:  nil,
		}
	}

	lat1, lon1 := geo.DecodeInt(score1)
	lat2, lon2 := geo.DecodeInt(score2)

	distance := geo.GetDistance(lon1, lat1, lon2, lat2)

	result, err := geo.ConvertDistance(distance, unit)

	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	return &EvalResponse{
		Result: utils.RoundToDecimals(result, 4),
		Error:  nil,
	}
}

func evalGEOPOS(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 2 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("GEOPOS"),
		}
	}

	key := args[0]
	obj := store.Get(key)

	if obj == nil {
		return &EvalResponse{
			Result: NIL,
			Error:  nil,
		}
	}

	ss, err := sortedset.FromObject(obj)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	results := make([]interface{}, len(args)-1)

	for index := 1; index < len(args); index++ {
		member := args[index]
		hash, ok := ss.Get(member)

		if !ok {
			results[index-1] = (nil)
			continue
		}

		lat, lon := geo.DecodeInt(hash)

		latFloat, _ := strconv.ParseFloat(fmt.Sprintf("%f", lat), 64)
		lonFloat, _ := strconv.ParseFloat(fmt.Sprintf("%f", lon), 64)

		results[index-1] = []interface{}{lonFloat, latFloat}
	}

	return &EvalResponse{
		Result: results,
		Error:  nil,
	}
}

func evalGEOHASH(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 2 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("GEOHASH"),
		}
	}

	key := args[0]
	members := args[1:]

	obj := store.Get(key)
	if obj == nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrKeyNotFound,
		}
	}

	ss, err := sortedset.FromObject(obj)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	// Prepare the response for each member
	result := make([]interface{}, 0)
	for _, member := range members {
		entry, exists := ss.Get(member)
		if !exists {
			result = append(result, nil) // RespNIL
			continue
		}

		lat, lon := geo.DecodeInt(entry)
		result = append(result, geo.EncodeString(lat, lon))
	}

	return &EvalResponse{
		Result: result,
		Error:  nil,
	}
}

func evalTouch(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 1 {
		return makeEvalError(diceerrors.ErrWrongArgumentCount("TOUCH"))
	}

	if store.Get(args[0]) != nil {
		return makeEvalResult(1)
	}

	return makeEvalResult(0)
}

func evalDBSize(args []string, store *dstore.Store) *EvalResponse {
	if len(args) > 0 {
		return makeEvalError(diceerrors.ErrWrongArgumentCount("DBSIZE"))
	}

	// Expired keys must be explicitly deleted since the cronFrequency for cleanup is configurable.
	// A longer delay may prevent timely cleanup, leading to incorrect DBSIZE results.
	dstore.DeleteExpiredKeys(store)

	return makeEvalResult(store.GetDBSize())
}

func evalKEYS(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 1 {
		return makeEvalError(diceerrors.ErrWrongArgumentCount("KEYS"))
	}

	pattern := args[0]
	keys, err := store.Keys(pattern)
	if err != nil {
		return makeEvalError(diceerrors.ErrGeneral("bad pattern"))
	}

	return makeEvalResult(keys)
}

// TODO: Placeholder to support monitoring
func evalCLIENT(args []string, store *dstore.Store) *EvalResponse {
	return makeEvalResult(OK)
}

// TODO: Placeholder to support monitoring
func evalLATENCY(args []string, store *dstore.Store) *EvalResponse {
	return makeEvalResult([]string{})
}

// evalPERSIST removes the expiry from the key
func evalPERSIST(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 1 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("PERSIST"),
		}
	}

	key := args[0]
	obj := store.Get(key)

	// If the key doesn't exist, return 0
	if obj == nil {
		return &EvalResponse{
			Result: IntegerZero,
			Error:  nil,
		}
	}

	// If the key has no expiry, return 0
	_, isExpirySet := dstore.GetExpiry(obj, store)
	if !isExpirySet {
		return &EvalResponse{
			Result: IntegerZero,
			Error:  nil,
		}
	}

	// Remove the expiry from the key
	dstore.DelExpiry(obj, store)

	return &EvalResponse{
		Result: IntegerOne,
		Error:  nil,
	}
}

// // BITOP <AND | OR | XOR | NOT> destkey key [key ...]
// func evalBITOP(args []string, store *dstore.Store) *EvalResponse {
// 	operation, destKey := args[0], args[1]
// 	operation = strings.ToUpper(operation)

// 	// get all the keys
// 	keys := args[2:]

// 	// validation of commands
// 	// if operation is not from enums, then error out
// 	if !(operation == AND || operation == OR || operation == XOR || operation == NOT) {
// 		return makeEvalError(diceerrors.ErrSyntax)
// 	}

// 	if operation == NOT {
// 		if len(keys) != 1 {
// 			return makeEvalError(diceerrors.ErrGeneral("BITOP NOT must be called with a single source key"))
// 		}
// 		key := keys[0]
// 		obj := store.Get(key)
// 		if obj == nil {
// 			return makeEvalResult(IntegerZero)
// 		}

// 		var value []byte

// 		switch oType, _ := obj.Type; oType {
// 		case object.ObjTypeByteArray:
// 			byteArray := obj.Value.(*ByteArray)
// 			byteArrayObject := *byteArray
// 			value = byteArrayObject.data
// 			// perform the operation
// 			result := make([]byte, len(value))
// 			for i := 0; i < len(value); i++ {
// 				result[i] = ^value[i]
// 			}

// 			// initialize result with byteArray
// 			operationResult := NewByteArray(len(result))
// 			operationResult.data = result
// 			operationResult.Length = int64(len(result))

// 			// resize the byte array if necessary
// 			operationResult.ResizeIfNecessary()

// 			// create object related to result
// 			obj = store.NewObj(operationResult, -1, object.ObjTypeByteArray)

// 			// store the result in destKey
// 			store.Put(destKey, obj)
// 			return makeEvalResult(len(value))

// 		case object.ObjTypeString, object.ObjTypeInt:
// 			if oType == object.ObjTypeString {
// 				value = []byte(obj.Value.(string))
// 			} else {
// 				value = []byte(strconv.FormatInt(obj.Value.(int64), 10))
// 			}
// 			// perform the operation
// 			result := make([]byte, len(value))
// 			for i := 0; i < len(value); i++ {
// 				result[i] = ^value[i]
// 			}
// 			resOType, resOEnc := deduceType(string(result))
// 			var storedValue interface{}
// 			if resOType == object.ObjTypeInt {
// 				storedValue, _ = strconv.ParseInt(string(result), 10, 64)
// 			} else {
// 				storedValue = string(result)
// 			}
// 			store.Put(destKey, store.NewObj(storedValue, -1, resOType, resOEnc))
// 			return makeEvalResult(len(value))

// 		default:
// 			return makeEvalError(diceerrors.ErrWrongTypeOperation)
// 		}
// 	}
// 	// if operation is AND, OR, XOR
// 	values := make([][]byte, len(keys))

// 	// get the values of all keys
// 	for i, key := range keys {
// 		obj := store.Get(key)
// 		if obj == nil {
// 			values[i] = make([]byte, 0)
// 		} else {
// 			// handle the case when it is byte array
// 			switch oType, _ := obj.Type; oType {
// 			case object.ObjTypeByteArray:
// 				byteArray := obj.Value.(*ByteArray)
// 				byteArrayObject := *byteArray
// 				values[i] = byteArrayObject.data
// 			case object.ObjTypeString:
// 				value := obj.Value.(string)
// 				values[i] = []byte(value)
// 			case object.ObjTypeInt:
// 				value := strconv.FormatInt(obj.Value.(int64), 10)
// 				values[i] = []byte(value)
// 			default:
// 				return makeEvalError(diceerrors.ErrWrongTypeOperation)
// 			}
// 		}
// 	}
// 	// get the length of the largest value
// 	maxLength := 0
// 	minLength := len(values[0])
// 	maxKeyIterator := 0
// 	for keyIterator, value := range values {
// 		if len(value) > maxLength {
// 			maxLength = len(value)
// 			maxKeyIterator = keyIterator
// 		}
// 		minLength = min(minLength, len(value))
// 	}

// 	result := make([]byte, maxLength)
// 	if operation == AND {
// 		for i := 0; i < maxLength; i++ {
// 			result[i] = 0
// 			if i < minLength {
// 				result[i] = values[maxKeyIterator][i]
// 			}
// 		}
// 	} else {
// 		for i := 0; i < maxLength; i++ {
// 			result[i] = 0x00
// 		}
// 	}

// 	// perform the operation
// 	for _, value := range values {
// 		for i := 0; i < len(value); i++ {
// 			switch operation {
// 			case AND:
// 				result[i] &= value[i]
// 			case OR:
// 				result[i] |= value[i]
// 			case XOR:
// 				result[i] ^= value[i]
// 			}
// 		}
// 	}
// 	// initialize result with byteArray
// 	operationResult := NewByteArray(len(result))
// 	operationResult.data = result
// 	operationResult.Length = int64(len(result))

// 	// create object related to result
// 	operationResultObject := store.NewObj(operationResult, -1, object.ObjTypeByteArray)

// 	// store the result in destKey
// 	store.Put(destKey, operationResultObject)

// 	return makeEvalResult(len(result))
// }

func evalObjectIdleTime(key string, store *dstore.Store) *EvalResponse {
	obj := store.GetNoTouch(key)
	if obj == nil {
		return makeEvalResult(NIL)
	}

	return makeEvalResult(dstore.GetIdleTime(obj.LastAccessedAt))
}

func evalOBJECT(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 2 {
		return makeEvalError(diceerrors.ErrWrongArgumentCount("OBJECT"))
	}

	subcommand := strings.ToUpper(args[0])
	key := args[1]

	switch subcommand {
	case "IDLETIME":
		return evalObjectIdleTime(key, store)
	default:
		return makeEvalError(diceerrors.ErrInvalidSyntax("OBJECT"))
	}
}

// evalCommand evaluates COMMAND <subcommand> command based on subcommand
// COUNT: return total count of commands in Dice.
func evalCommand(args []string, store *dstore.Store) *EvalResponse {
	if len(args) == 0 {
		return evalCommandDefault()
	}
	subcommand := strings.ToUpper(args[0])
	switch subcommand {
	case Count:
		return evalCommandCount(args[1:])
	case GetKeys:
		return evalCommandGetKeys(args[1:])
	case List:
		return evalCommandList(args[1:])
	case Help:
		return evalCommandHelp(args[1:])
	case Info:
		return evalCommandInfo(args[1:])
	case Docs:
		return evalCommandDocs(args[1:])
	default:
		return makeEvalError(diceerrors.ErrGeneral(fmt.Sprintf("unknown subcommand '%s'. Try COMMAND HELP.", subcommand)))
	}
}

func evalCommandDefault() *EvalResponse {
	cmds := convertDiceCmdsMapToSlice()
	return makeEvalResult(cmds)
}

// evalCommandCount returns a number of commands supported by DiceDB
func evalCommandCount(args []string) *EvalResponse {
	if len(args) > 0 {
		return makeEvalError(diceerrors.ErrWrongArgumentCount("COMMAND|COUNT"))
	}

	return makeEvalResult(diceCommandsCount)
}

func evalCommandGetKeys(args []string) *EvalResponse {
	if len(args) == 0 {
		return makeEvalError(diceerrors.ErrWrongArgumentCount("COMMAND|GETKEYS"))
	}
	diceCmd, ok := DiceCmds[strings.ToUpper(args[0])]
	if !ok {
		return makeEvalError(diceerrors.ErrGeneral("invalid command specified"))
	}

	keySpecs := diceCmd.KeySpecs
	if keySpecs.BeginIndex == 0 {
		return makeEvalError(diceerrors.ErrGeneral("the command has no key arguments"))
	}

	arity := diceCmd.Arity
	if (arity < 0 && len(args) < -arity) ||
		(arity >= 0 && len(args) != arity) {
		return makeEvalError(diceerrors.ErrGeneral("invalid number of arguments specified for command"))
	}
	keys := make([]string, 0)
	step := max(keySpecs.Step, 1)
	lastIdx := keySpecs.BeginIndex
	if keySpecs.LastKey != 0 {
		lastIdx = len(args) + keySpecs.LastKey
	}
	for i := keySpecs.BeginIndex; i <= lastIdx; i += step {
		keys = append(keys, args[i])
	}

	return makeEvalResult(keys)
}

func evalCommandList(args []string) *EvalResponse {
	if len(args) > 0 {
		return makeEvalError(diceerrors.ErrWrongArgumentCount("COMMAND|LIST"))
	}

	cmds := make([]string, 0, diceCommandsCount)
	for k := range DiceCmds {
		cmds = append(cmds, k)
		for _, sc := range DiceCmds[k].SubCommands {
			cmds = append(cmds, fmt.Sprint(k, "|", sc))
		}
	}
	return makeEvalResult(cmds)
}

// evalCommandHelp prints help message
func evalCommandHelp(args []string) *EvalResponse {
	if len(args) > 0 {
		return makeEvalError(diceerrors.ErrWrongArgumentCount("COMMAND|HELP"))
	}

	format := "COMMAND <subcommand> [<arg> [value] [opt] ...]. Subcommands are:"
	noTitle := "(no subcommand)"
	noMessage := "     Return details about all DiceDB commands."
	countTitle := CountConst
	countMessage := "     Return the total number of commands in this DiceDB server."
	listTitle := "LIST"
	listMessage := "     Return a list of all commands in this DiceDB server."
	infoTitle := "INFO [<command-name> ...]"
	infoMessage := "     Return details about the specified DiceDB commands. If no command names are given, documentation details for all commands are returned."
	docsTitle := "DOCS [<command-name> ...]"
	docsMessage := "\tReturn documentation details about multiple diceDB commands.\n\tIf no command names are given, documentation details for all\n\tcommands are returned."
	getKeysTitle := "GETKEYS <full-command>"
	getKeysMessage := "     Return the keys from a full DiceDB command."
	helpTitle := "HELP"
	helpMessage := "     Print this help."
	message := []string{
		format,
		noTitle,
		noMessage,
		countTitle,
		countMessage,
		listTitle,
		listMessage,
		infoTitle,
		infoMessage,
		docsTitle,
		docsMessage,
		getKeysTitle,
		getKeysMessage,
		helpTitle,
		helpMessage,
	}

	return makeEvalResult(message)
}

func evalCommandDefaultDocs() *EvalResponse {
	cmds := convertDiceCmdsMapToDocs()
	return makeEvalResult(cmds)
}

func evalCommandInfo(args []string) *EvalResponse {
	if len(args) == 0 {
		return evalCommandDefault()
	}

	cmdMetaMap := make(map[string]interface{})
	for _, cmdMeta := range DiceCmds {
		cmdMetaMap[cmdMeta.Name] = convertCmdMetaToSlice(&cmdMeta)
	}

	var result []interface{}
	for _, arg := range args {
		arg = strings.ToUpper(arg)
		if cmdMeta, found := cmdMetaMap[arg]; found {
			result = append(result, cmdMeta)
		} else {
			result = append(result, RespNIL)
		}
	}

	return makeEvalResult(result)
}

func evalCommandDocs(args []string) *EvalResponse {
	if len(args) == 0 {
		return evalCommandDefaultDocs()
	}

	cmdMetaMap := make(map[string]interface{})
	for _, cmdMeta := range DiceCmds {
		cmdMetaMap[cmdMeta.Name] = convertCmdMetaToDocs(&cmdMeta)
	}

	var result []interface{}
	for _, arg := range args {
		arg = strings.ToUpper(arg)
		if cmdMeta, found := cmdMetaMap[arg]; found {
			result = append(result, cmdMeta)
		}
	}

	return makeEvalResult(result)
}

func evalJSONARRINDEX(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 3 || len(args) > 5 {
		return makeEvalError(diceerrors.ErrWrongArgumentCount("JSON.ARRINDEX"))
	}

	key := args[0]
	path := args[1]
	start := 0
	stop := 0

	var value interface{}
	var err error

	if strings.Contains(args[2], `"`) {
		// user has provided string argument
		value = args[2]
	} else {
		// parse it to float since default arg type would be string
		value, err = strconv.ParseFloat(args[2], 64)

		if err != nil {
			return makeEvalError(diceerrors.ErrGeneral("Couldn't parse as integer"))
		}
	}

	// Convert start to integer if provided
	if len(args) >= 4 {
		var err error
		start, err = strconv.Atoi(args[3])
		if err != nil {
			return makeEvalError(diceerrors.ErrGeneral("Couldn't parse as integer"))
		}
	}

	// Convert stop to integer if provided
	if len(args) == 5 {
		var err error
		stop, err = strconv.Atoi(args[4])
		if err != nil {
			return makeEvalError(diceerrors.ErrGeneral("Couldn't parse as integer"))
		}
	}

	// Check if the path specified is valid
	expr, err2 := jp.ParseString(path)
	if err2 != nil {
		return makeEvalError(diceerrors.ErrJSONPathNotFound(path))
	}

	obj := store.Get(key)
	if obj == nil {
		return makeEvalError(diceerrors.ErrKeyDoesNotExist)
	}

	if err2 := object.AssertType(obj.Type, object.ObjTypeJSON); err2 != nil {
		return makeEvalError(diceerrors.ErrGeneral("Existing key has wrong Dice type"))
	}

	jsonData := obj.Value

	// Check if the value stored is JSON type
	_, err = sonic.Marshal(jsonData)

	if err != nil {
		return makeEvalError(diceerrors.ErrGeneral("Existing key has wrong Dice type"))
	}

	results := expr.Get(jsonData)
	arrIndexList := make([]interface{}, 0, len(results))

	for _, result := range results {
		switch utils.GetJSONFieldType(result) {
		case utils.ArrayType:
			elementFound := false
			arr := result.([]interface{})
			length := len(arr)

			adjustedStart, adjustedStop := adjustIndices(start, stop, length)

			if adjustedStart == -1 {
				arrIndexList = append(arrIndexList, -1)
				continue
			}

			// Range [start, stop) : start is inclusive, stop is exclusive
			for i := adjustedStart; i < adjustedStop; i++ {
				if reflect.DeepEqual(arr[i], value) {
					arrIndexList = append(arrIndexList, i)
					elementFound = true
					break
				}
			}

			if !elementFound {
				arrIndexList = append(arrIndexList, -1)
			}
		default:
			arrIndexList = append(arrIndexList, nil)
		}
	}

	return makeEvalResult(arrIndexList)
}

// adjustIndices adjusts the start and stop indices for array traversal.
// It handles negative indices and ensures they are within the array bounds.
func adjustIndices(start, stop, length int) (adjustedStart, adjustedStop int) {
	if length == 0 {
		return -1, -1
	}
	if start < 0 {
		start += length
	}

	if stop <= 0 {
		stop += length
	}
	if start < 0 {
		start = 0
	}
	if stop < 0 {
		stop = 0
	}
	if start >= length {
		return -1, -1
	}
	if stop > length {
		stop = length
	}
	if start > stop {
		return -1, -1
	}
	return start, stop
}

// This method executes each operation, contained in ops array, based on commands used.
func executeBitfieldOps(value *ByteArray, ops []utils.BitFieldOp) []interface{} {
	overflowType := WRAP
	var result []interface{}
	for _, op := range ops {
		switch op.Kind {
		case GET:
			res := value.getBits(int(op.Offset), int(op.EVal), op.EType == SIGNED)
			result = append(result, res)
		case SET:
			prevValue := value.getBits(int(op.Offset), int(op.EVal), op.EType == SIGNED)
			value.setBits(int(op.Offset), int(op.EVal), op.Value)
			result = append(result, prevValue)
		case INCRBY:
			res, err := value.incrByBits(int(op.Offset), int(op.EVal), op.Value, overflowType, op.EType == SIGNED)
			if err != nil {
				result = append(result, nil)
			} else {
				result = append(result, res)
			}
		case OVERFLOW:
			overflowType = op.EType
		}
	}
	return result
}

// formatFloat formats float64 as string.
// Optionally appends a decimal (.0) for whole numbers,
// if b is true.
func formatFloat(f float64, b bool) string {
	formatted := strconv.FormatFloat(f, 'f', -1, 64)
	if b {
		parts := strings.Split(formatted, ".")
		if len(parts) == 1 {
			formatted += ".0"
		}
	}
	return formatted
}

// trimElementAndUpdateArray trim the array between the given start and stop index
// Returns trimmed array
func trimElementAndUpdateArray(arr []any, start, stop int) []any {
	updatedArray := make([]any, 0)
	length := len(arr)
	if len(arr) == 0 {
		return updatedArray
	}
	var startIdx, stopIdx int

	if start >= length {
		return updatedArray
	}

	startIdx = adjustIndex(start, arr)
	stopIdx = adjustIndex(stop, arr)

	if startIdx > stopIdx {
		return updatedArray
	}

	updatedArray = arr[startIdx : stopIdx+1]
	return updatedArray
}

// insertElementAndUpdateArray add an element at the given index
// Returns remaining array and error
func insertElementAndUpdateArray(arr []any, index int, elements []interface{}) (updatedArray []any, err error) {
	length := len(arr)
	var idx int
	if index >= -length && index <= length {
		idx = adjustIndex(index, arr)
	} else {
		return nil, errors.New("index out of bounds")
	}
	before := arr[:idx]
	after := arr[idx:]

	elements = append(elements, after...)
	before = append(before, elements...)
	updatedArray = append(updatedArray, before...)
	return updatedArray, nil
}

// adjustIndex will bound the array between 0 and len(arr) - 1
// It also handles negative indexes
func adjustIndex(idx int, arr []any) int {
	// if index is positive and out of bound, limit it to the last index
	if idx > len(arr) {
		idx = len(arr) - 1
	}

	// if index is negative, change it to equivalent positive index
	if idx < 0 {
		// if index is out of bound then limit it to the first index
		if idx < -len(arr) {
			idx = 0
		} else {
			idx = len(arr) + idx
		}
	}
	return idx
}

// reverseSlice takes a slice of any type and returns a new slice with the elements reversed.
func reverseSlice[T any](slice []T) []T {
	reversed := make([]T, len(slice))
	for i, v := range slice {
		reversed[len(slice)-1-i] = v
	}
	return reversed
}

// Parses and returns the input string as an int64 or float64
func parseFloatInt(input string) (result interface{}, err error) {
	// Try to parse as an integer
	if intValue, parseErr := strconv.ParseInt(input, 10, 64); parseErr == nil {
		result = intValue
		return
	}

	// Try to parse as a float
	if floatValue, parseErr := strconv.ParseFloat(input, 64); parseErr == nil {
		result = floatValue
		return
	}

	// If neither parsing succeeds, return an error
	err = errors.New("invalid input: not a valid int or float")
	return
}
