package sql

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/dicedb/dice/internal/object"

	"github.com/bytedance/sonic"
	"github.com/dicedb/dice/internal/common"
	"github.com/dicedb/dice/internal/regex"
	"github.com/dicedb/dice/internal/server/utils"
	"github.com/ohler55/ojg/jp"
)

var ErrNoResultsFound = errors.New("ERR No results found")
var ErrInvalidJSONPath = errors.New("ERR invalid JSONPath")

type QueryResultRow struct {
	Key   string
	Value object.Obj // Use pointer to avoid copying
}

type QueryResultRowWithOrder struct {
	Row          QueryResultRow
	OrderByValue interface{}
	OrderByType  string
}

func ExecuteQuery(query *DSQLQuery, store common.ITable[string, *object.Obj]) ([]QueryResultRow, error) {
	var result []QueryResultRow
	var err error
	jsonPathCache := make(map[string]jp.Expr) // Cache for parsed JSON paths

	store.All(func(key string, value *object.Obj) bool {
		row := QueryResultRow{
			Key:   key,
			Value: *value, // Use pointer to avoid copying
		}

		if query.Where != nil {
			match, evalErr := EvaluateWhereClause(query.Where, row, jsonPathCache)
			if errors.Is(evalErr, ErrNoResultsFound) {
				// if no result found error
				// and continue with the next iteration
				return true
			}
			if evalErr != nil {
				err = evalErr
				// stop iteration if any other error
				return false
			}
			if !match {
				// if did not match, continue the iteration
				return true
			}
		}

		result = append(result, row)

		// Early termination if limit is reached and no sorting is required
		if query.Limit > 0 && len(result) >= query.Limit && query.OrderBy.OrderBy.Value == utils.EmptyStr {
			return false
		}

		return true
	})

	if err != nil {
		return nil, err
	}

	// Precompute order-by values and sort if necessary
	if query.OrderBy.OrderBy.Value != utils.EmptyStr {
		resultWithOrder, err := precomputeOrderByValues(&query.OrderBy.OrderBy, result, jsonPathCache)
		if err != nil {
			return nil, err
		}
		sortResults(query.OrderBy.Order, resultWithOrder)
		// Extract sorted rows
		for i := range resultWithOrder {
			result[i] = resultWithOrder[i].Row
		}
	}

	// Apply limit after sorting
	if query.Limit > 0 && query.Limit < len(result) {
		result = result[:query.Limit]
	}

	// Marshal JSON results if necessary
	for i := range result {
		if err := MarshalResultIfJSON(&result[i]); err != nil {
			return nil, err
		}
	}

	if !query.Selection.KeySelection {
		for i := range result {
			result[i].Key = utils.EmptyStr
		}
	}

	if !query.Selection.ValueSelection {
		for i := range result {
			result[i].Value = object.Obj{}
		}
	}

	return result, nil
}

func MarshalResultIfJSON(row *QueryResultRow) error {
	// if the row contains JSON field then convert the json object into string representation so it can be encoded
	// before being returned to the client
	if object.GetEncoding(row.Value.TypeEncoding) == object.ObjEncodingJSON && object.GetType(row.Value.TypeEncoding) == object.ObjTypeJSON {
		marshaledData, err := sonic.MarshalString(row.Value.Value)
		if err != nil {
			return err
		}

		row.Value.Value = marshaledData
	}
	return nil
}

func precomputeOrderByValues(orderBy *Value, result []QueryResultRow, jsonPathCache map[string]jp.Expr) ([]QueryResultRowWithOrder, error) {
	resultWithOrder := make([]QueryResultRowWithOrder, len(result))
	for i, row := range result {
		val, valType, err := getOrderByValue(orderBy, row, jsonPathCache)
		if err != nil {
			return nil, err
		}
		resultWithOrder[i] = QueryResultRowWithOrder{
			Row:          row,
			OrderByValue: val,
			OrderByType:  valType,
		}
	}
	return resultWithOrder, nil
}

func sortResults(order string, result []QueryResultRowWithOrder) {
	sort.Slice(result, func(i, j int) bool {
		valI := result[i].OrderByValue
		valJ := result[j].OrderByValue
		typeI := result[i].OrderByType
		typeJ := result[j].OrderByType

		// Handle Nil types
		if typeI == Nil && typeJ == Nil {
			return false // They are equal
		}
		if typeI == Nil {
			return order == Desc // Place Nil values at the beginning in Descending order
		}
		if typeJ == Nil {
			return order == Asc // Place Nil values at the end in Ascending order
		}

		// Only compare values if types are the same
		if typeI != typeJ {
			return false // Types differ; cannot compare
		}

		comparison, err := compareOrderByValues(valI, valJ, typeI, typeJ, order)
		if err != nil {
			return false // Cannot compare; treat as equal
		}

		return comparison
	})
}

func getOrderByValue(orderBy *Value, row QueryResultRow, jsonPathCache map[string]jp.Expr) (value interface{}, valueType string, err error) {
	switch orderBy.Type {
	case FieldKey:
		return row.Key, String, nil
	case FieldVal:
		return getValueAndType(&row.Value)
	case FieldJSON:
		if isJSON(&row.Value) {
			return retrieveValueFromJSON(orderBy.Value, &row.Value, jsonPathCache)
		}
	}
	return nil, "", fmt.Errorf("invalid ORDER BY clause: %s", orderBy)
}

func compareOrderByValues(valI, valJ interface{}, typeI, typeJ, order string) (bool, error) {
	// If types differ, define consistent ordering based on type names
	if typeI != typeJ {
		if order == Asc {
			return typeI < typeJ, nil // Ascending order based on type names
		}
		return typeI > typeJ, nil // Descending order based on type names
	}

	// Types are the same, proceed with comparison
	switch typeI {
	case String:
		return compareStringValues(order, valI.(string), valJ.(string)), nil
	case Int64:
		return compareInt64Values(order, valI.(int64), valJ.(int64)), nil
	case Float:
		return compareFloatValues(order, valI.(float64), valJ.(float64)), nil
	case Bool:
		return compareBoolValues(order, valI.(bool), valJ.(bool)), nil
	default:
		return false, fmt.Errorf("unsupported type for comparison: %s", typeI)
	}
}

func compareStringValues(order, valI, valJ string) bool {
	if order == Asc {
		return valI < valJ
	}
	return valI > valJ
}

func compareInt64Values(order string, valI, valJ int64) bool {
	if order == Asc {
		return valI < valJ
	}
	return valI > valJ
}

func compareFloatValues(order string, valI, valJ float64) bool {
	if order == Asc {
		return valI < valJ
	}
	return valI > valJ
}

func compareBoolValues(order string, valI, valJ bool) bool {
	if order == Asc {
		return !valI && valJ
	}
	return valI && !valJ
}

func EvaluateWhereClause(node ConditionNode, row QueryResultRow, jsonPathCache map[string]jp.Expr) (bool, error) {
	if node == nil {
		return false, fmt.Errorf("condition node is nil")
	}

	switch node := node.(type) {
	case *AndNode:
		left, err := EvaluateWhereClause(node.Left, row, jsonPathCache)
		if err != nil || !left {
			return false, err
		}
		return EvaluateWhereClause(node.Right, row, jsonPathCache)
	case *OrNode:
		left, err := EvaluateWhereClause(node.Left, row, jsonPathCache)
		if err != nil {
			return false, err
		}
		if left {
			return true, nil
		}
		return EvaluateWhereClause(node.Right, row, jsonPathCache)
	case *ComparisonNode:
		return evaluateComparisonExpr(node, row, jsonPathCache)
	}

	return false, fmt.Errorf("unsupported condition node structure")
}

// Function to evaluate a ComparisonExpr node
func evaluateComparisonExpr(expr *ComparisonNode, row QueryResultRow, jsonPathCache map[string]jp.Expr) (bool, error) {
	left, leftType, err := getExprValueAndType(expr.Left, row, jsonPathCache)
	if err != nil {
		if errors.Is(err, ErrNoResultsFound) {
			return false, nil
		}
		return false, err
	}

	right, rightType, err := getExprValueAndType(expr.Right, row, jsonPathCache)
	if err != nil {
		if errors.Is(err, ErrNoResultsFound) {
			return false, nil
		}
		return false, err
	}

	// Handle Nil types
	if leftType == Nil || rightType == Nil {
		// Comparisons with NULL result in FALSE
		return false, nil
	}

	// Check if types are the same
	if leftType != rightType {
		return false, fmt.Errorf("incompatible types in comparison: %s and %s", leftType, rightType)
	}

	switch leftType {
	case String:
		return compareStrings(left.(string), right.(string), expr.Operator)
	case Int64:
		return compareInt64s(left.(int64), right.(int64), expr.Operator)
	case Float:
		return compareFloats(left.(float64), right.(float64), expr.Operator)
	default:
		return false, fmt.Errorf("unsupported type for comparison: %s", leftType)
	}
}

// Function to extract value and type from an expression
func getExprValueAndType(expr Value, row QueryResultRow, jsonPathCache map[string]jp.Expr) (value interface{}, valueType string, err error) {
	switch expr.Type {
	case FieldKey:
		return row.Key, String, nil
	case FieldVal:
		return getValueAndType(&row.Value)
	case FieldFloat:
		f, _ := strconv.ParseFloat(expr.Value, 64)
		return f, Float, nil
	case FieldInt:
		i, _ := strconv.Atoi(expr.Value)
		return int64(i), Int64, nil
	case FieldString:
		return expr.Value, String, nil
	case FieldJSON:
		return retrieveValueFromJSON(expr.Value, &row.Value, jsonPathCache) // FIX: ensure row.Value is json
	default:
		return nil, utils.EmptyStr, fmt.Errorf("unsupported expression type: %T", expr)
	}
}

func isJSON(obj *object.Obj) bool {
	if err := object.AssertEncoding(obj.TypeEncoding, object.ObjEncodingJSON); err != nil {
		return false
	}

	if err := object.AssertType(obj.TypeEncoding, object.ObjTypeJSON); err != nil {
		return false
	}

	return true
}

func retrieveValueFromJSON(path string, jsonData *object.Obj, jsonPathCache map[string]jp.Expr) (value interface{}, valueType string, err error) {
	if !isJSON(jsonData) {
		return nil, Nil, nil
	}

	// path is in the format '_value.field1.field2'. We need to remove _value reference from the prefix to get the json
	// path.
	jsonPath := strings.Split(path, ".")
	if len(jsonPath) < 2 {
		return nil, "", ErrInvalidJSONPath
	}

	pathKey := "$." + strings.Join(jsonPath[1:], ".")

	expr, exists := jsonPathCache[pathKey]
	if !exists {
		exprParsed, err := jp.ParseString(pathKey)
		if err != nil {
			return nil, "", ErrInvalidJSONPath
		}
		jsonPathCache[pathKey] = exprParsed
		expr = exprParsed
	}

	results := expr.Get(jsonData.Value)
	if len(results) == 0 {
		return nil, Nil, nil // Return nil value with Nil type
	}

	return inferTypeAndConvert(results[0])
}

// inferTypeAndConvert infers the type of the value and converts it to the appropriate type. currently the only data
// types we support are strings, floats, ints, booleans and nil.
func inferTypeAndConvert(val interface{}) (value interface{}, valueType string, err error) {
	switch v := val.(type) {
	case string:
		return v, String, nil
	case float64:
		if isInt64(v) {
			return int64(v), Int64, nil
		}
		return v, Float, nil
	case bool:
		return v, Bool, nil
	case nil:
		return nil, Nil, nil
	default:
		return nil, utils.EmptyStr, fmt.Errorf("unsupported JSONPath result type: %T", v)
	}
}

// isInt64 checks if a float is an integer. When we unmarshal JSON data into an interface it sets all numbers as
// floats, https://forum.golangbridge.org/t/type-problem-in-json-conversion/19420.
// This function does not handle the edge case where user enters a floating point number with trailing zeros in the
// fractional part of a decimal number (e.g 10.0) then our code treats that as an integer rather than float.
// One way to solve this is as follows (credit https://github.com/YahyaHaq):
// floatStr := strconv.FormatFloat(val, 'f', -1, 64)
// if strings.Contains(floatStr, ".") {
// return val, "float", nil
// }
// return int(val), "int", nil
//
// However, the string conversion is expensive and we are trying to avoid it. We can assume this to be a limitation of
// using the JSON data type.
// TODO: handle the edge case where the integer is too large for float64.
func isInt64(f float64) bool {
	return f == float64(int64(f))
}

// getValueAndType returns the type-casted value and type of the object
func getValueAndType(obj *object.Obj) (val interface{}, s string, e error) {
	switch v := obj.Value.(type) {
	case string:
		return v, String, nil
	case int64:
		return v, Int64, nil
	case float64:
		return v, Float, nil
	default:
		return nil, utils.EmptyStr, fmt.Errorf("unsupported value type: %T", v)
	}
}

func compareStrings(left, right string, operator ComparisonOp) (bool, error) {
	switch operator {
	case OpEq:
		return left == right, nil
	case OpNeq:
		return left != right, nil
	case OpLt:
		return left < right, nil
	case OpLte:
		return left <= right, nil
	case OpGt:
		return left > right, nil
	case OpGte:
		return left >= right, nil
	case OpLike:
		return regex.WildCardMatch(right, left), nil
	case OpNotLike:
		return !regex.WildCardMatch(right, left), nil
	default:
		return false, fmt.Errorf("unsupported operator for strings: %s", operator)
	}
}

func compareInt64s(left, right int64, operator ComparisonOp) (bool, error) {
	switch operator {
	case OpEq:
		return left == right, nil
	case OpNeq:
		return left != right, nil
	case OpLt:
		return left < right, nil
	case OpLte:
		return left <= right, nil
	case OpGt:
		return left > right, nil
	case OpGte:
		return left >= right, nil
	default:
		return false, fmt.Errorf("unsupported operator for integers: %s", operator)
	}
}

func compareFloats(left, right float64, operator ComparisonOp) (bool, error) {
	switch operator {
	case OpEq:
		return left == right, nil
	case OpNeq:
		return left != right, nil
	case OpLt:
		return left < right, nil
	case OpLte:
		return left <= right, nil
	case OpGt:
		return left > right, nil
	case OpGte:
		return left >= right, nil
	default:
		return false, fmt.Errorf("unsupported operator for floats: %s", operator)
	}
}
