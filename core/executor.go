package core

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/cockroachdb/swiss"

	"github.com/bytedance/sonic"
	"github.com/dicedb/dice/internal/constants"
	"github.com/ohler55/ojg/jp"
	"github.com/xwb1989/sqlparser"
)

var ErrNoResultsFound = errors.New("ERR No results found")
var ErrInvalidJSONPath = errors.New("ERR invalid JSONPath")

type DSQLQueryResultRow struct {
	Key   string
	Value Obj
}

func ExecuteQuery(query *DSQLQuery, store *swiss.Map[string, *Obj]) ([]DSQLQueryResultRow, error) {
	var result []DSQLQueryResultRow

	var err error
	store.All(func(key string, value *Obj) bool {
		// TODO: We can remove this check once we have scatter-gather algorithm for multi-threaded execution implemented.
		//  This is only required when we are running unit tests and passing the complete store to the function.
		if !WildCardMatch(query.KeyRegex, key) {
			return true
		}

		row := DSQLQueryResultRow{
			Key:   key,
			Value: *value,
		}

		if query.Where != nil {
			match, evalErr := evaluateWhereClause(query.Where, row)
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

		return true
	})

	if err != nil {
		return nil, err
	}

	sortResults(query, result)

	for i := range result {
		if err := MarshalResultIfJSON(&result[i]); err != nil {
			return nil, err
		}
	}

	if !query.Selection.KeySelection {
		for i := range result {
			result[i].Key = constants.EmptyStr
		}
	}

	if !query.Selection.ValueSelection {
		for i := range result {
			result[i].Value = Obj{}
		}
	}

	if query.Limit > 0 && query.Limit < len(result) {
		result = result[:query.Limit]
	}

	return result, nil
}

func MarshalResultIfJSON(row *DSQLQueryResultRow) error {
	// if the row contains JSON field then convert the json object into string representation so it can be encoded
	// before being returned to the client
	if getEncoding(row.Value.TypeEncoding) == ObjEncodingJSON && getType(row.Value.TypeEncoding) == ObjTypeJSON {
		marshaledData, err := sonic.MarshalString(row.Value.Value)
		if err != nil {
			return err
		}

		row.Value.Value = marshaledData
	}
	return nil
}

func sortResults(query *DSQLQuery, result []DSQLQueryResultRow) {
	if query.OrderBy.OrderBy == constants.EmptyStr {
		return
	}

	sort.Slice(result, func(i, j int) bool {
		valI, typeI, err := getOrderByValue(query.OrderBy.OrderBy, result[i])
		if err != nil {
			return false
		}
		valJ, typeJ, err := getOrderByValue(query.OrderBy.OrderBy, result[j])
		if err != nil {
			return false
		}

		if typeI != typeJ {
			return false
		}

		comparison, err := compareOrderByValues(valI, valJ, typeI, query.OrderBy.Order)
		if err != nil {
			return false
		}

		return comparison
	})
}

func getOrderByValue(orderBy string, row DSQLQueryResultRow) (value interface{}, valueType string, err error) {
	switch orderBy {
	case TempKey:
		return row.Key, constants.String, nil
	case TempValue:
		return getValueAndType(&row.Value)
	default:
		// Handle JSON field
		if isJSONField(&sqlparser.SQLVal{Val: []byte(orderBy)}, &row.Value) {
			return retrieveValueFromJSON(orderBy, &row.Value)
		}
	}
	return nil, "", fmt.Errorf("invalid ORDER BY clause: %s", orderBy)
}

func compareOrderByValues(valI, valJ interface{}, valueType, order string) (bool, error) {
	switch valueType {
	case constants.String:
		return compareStringValues(order, valI.(string), valJ.(string)), nil
	case constants.Int:
		switch v := valI.(type) {
		case int:
			return compareIntValues(order, v, valJ.(int)), nil
		case int64:
			return compareInt64Values(order, v, valJ.(int64)), nil
		default:
			return false, fmt.Errorf("unsupported type for order by comparison: %T", valI)
		}
	case constants.Float:
		return compareFloatValues(order, valI.(float64), valJ.(float64)), nil
	case constants.Bool:
		return compareBoolValues(order, valI.(bool), valJ.(bool)), nil
	default:
		return false, fmt.Errorf("unsupported type for comparison: %s", valueType)
	}
}

func compareStringValues(order, valI, valJ string) bool {
	if order == constants.Asc {
		return valI < valJ
	}
	return valI > valJ
}

func compareInt64Values(order string, valI, valJ int64) bool {
	if order == constants.Asc {
		return valI < valJ
	}
	return valI > valJ
}

func compareIntValues(order string, valI, valJ int) bool {
	if order == constants.Asc {
		return valI < valJ
	}
	return valI > valJ
}

func compareFloatValues(order string, valI, valJ float64) bool {
	if order == constants.Asc {
		return valI < valJ
	}
	return valI > valJ
}

func compareBoolValues(order string, valI, valJ bool) bool {
	if order == constants.Asc {
		return !valI && valJ
	}
	return valI && !valJ
}

func evaluateWhereClause(expr sqlparser.Expr, row DSQLQueryResultRow) (bool, error) {
	switch expr := expr.(type) {
	case *sqlparser.ComparisonExpr:
		return evaluateComparison(expr, row)
	case *sqlparser.AndExpr:
		left, err := evaluateWhereClause(expr.Left, row)
		if err != nil {
			return false, err
		}
		right, err := evaluateWhereClause(expr.Right, row)
		if err != nil {
			return false, err
		}
		return left && right, nil
	case *sqlparser.OrExpr:
		left, err := evaluateWhereClause(expr.Left, row)
		if err != nil {
			return false, err
		}
		right, err := evaluateWhereClause(expr.Right, row)
		if err != nil {
			return false, err
		}
		return left || right, nil
	default:
		return false, fmt.Errorf("unsupported expression type: %T", expr)
	}
}

func evaluateComparison(expr *sqlparser.ComparisonExpr, row DSQLQueryResultRow) (b bool, e error) {
	left, leftType, err := getExprValueAndType(expr.Left, row)
	if err != nil {
		return false, err
	}
	right, rightType, err := getExprValueAndType(expr.Right, row)
	if err != nil {
		return false, err
	}

	// Check if types are compatible
	if leftType != rightType {
		return false, fmt.Errorf("incompatible types in comparison: %s and %s", leftType, rightType)
	}

	switch leftType {
	case constants.String:
		return compareStrings(left.(string), right.(string), expr.Operator)
	case constants.Int:
		return compareInts(left.(int), right.(int), expr.Operator)
	case constants.Float:
		return compareFloats(left.(float64), right.(float64), expr.Operator)
	default:
		return false, fmt.Errorf("unsupported type for comparison: %s", leftType)
	}
}

func getExprValueAndType(expr sqlparser.Expr, row DSQLQueryResultRow) (value interface{}, valueType string, err error) {
	switch expr := expr.(type) {
	case *sqlparser.ColName:
		switch expr.Name.String() {
		case TempKey:
			return row.Key, "string", nil
		case TempValue:
			return getValueAndType(&row.Value)
		default:
			return nil, "", fmt.Errorf("unknown column: %s", expr.Name.String())
		}
	case *sqlparser.SQLVal:
		// we currently treat JSON query expression as a string value so we will need to differentiate between JSON and
		// SQL strings
		if isJSONField(expr, &row.Value) {
			return retrieveValueFromJSON(string(expr.Val), &row.Value)
		}
		return sqlValToGoValue(expr)
	default:
		return nil, "", fmt.Errorf("unsupported expression type: %T", expr)
	}
}

func isJSONField(expr *sqlparser.SQLVal, obj *Obj) bool {
	if err := assertEncoding(obj.TypeEncoding, ObjEncodingJSON); err != nil {
		return false
	}

	if err := assertType(obj.TypeEncoding, ObjTypeJSON); err != nil {
		return false
	}

	// We convert the $key and $value fields to _key, _value before querying. hence fields starting with _ are
	// considered to be stored values
	return expr.Type == sqlparser.StrVal &&
		strings.HasPrefix(string(expr.Val), TempPrefix)
}

func retrieveValueFromJSON(path string, jsonData *Obj) (value interface{}, valueType string, err error) {
	// path is in the format '_value.field1.field2'. We need to remove _value reference from the prefix to get the json
	// path.
	jsonPath := strings.Split(path, ".")
	if len(jsonPath) < 2 {
		return nil, "", ErrInvalidJSONPath
	}

	path = "$." + strings.Join(jsonPath[1:], ".")
	expr, err := jp.ParseString(path)
	if err != nil {
		return nil, "", ErrInvalidJSONPath
	}

	results := expr.Get(jsonData.Value)
	if len(results) == 0 {
		return nil, "", ErrNoResultsFound
	}

	return inferTypeAndConvert(results[0])
}

// inferTypeAndConvert infers the type of the value and converts it to the appropriate type. currently the only data
// types we support are strings, floats, ints, booleans and nil.
func inferTypeAndConvert(val interface{}) (value interface{}, valueType string, err error) {
	switch v := val.(type) {
	case string:
		return v, constants.String, nil
	case float64:
		if isInteger(v) {
			return int(v), constants.Int, nil
		}
		return v, constants.Float, nil
	case bool:
		return v, constants.Bool, nil
	case nil:
		return nil, constants.Nil, nil
	default:
		return nil, constants.EmptyStr, fmt.Errorf("unsupported JSONPath result type: %T", v)
	}
}

// isInteger checks if a float is an integer. When we unmarshal JSON data into an interface it sets all numbers as
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
func isInteger(f float64) bool {
	return f == float64(int(f))
}

// getValueAndType returns the type-casted value and type of the object
func getValueAndType(obj *Obj) (val interface{}, s string, e error) {
	switch v := obj.Value.(type) {
	case string:
		return v, constants.String, nil
	case int64:
		return v, constants.Int, nil
	case float64:
		return v, constants.Float, nil
	default:
		return nil, constants.EmptyStr, fmt.Errorf("unsupported value type: %T", v)
	}
}

// sqlValToGoValue converts SQLVal to Go value, and returns the type of the value.
func sqlValToGoValue(sqlVal *sqlparser.SQLVal) (val interface{}, s string, e error) {
	switch sqlVal.Type {
	case sqlparser.StrVal:
		return string(sqlVal.Val), constants.String, nil
	case sqlparser.IntVal:
		i, err := strconv.Atoi(string(sqlVal.Val))
		if err != nil {
			return nil, constants.EmptyStr, err
		}
		return i, constants.Int, nil
	case sqlparser.FloatVal:
		f, err := strconv.ParseFloat(string(sqlVal.Val), 64)
		if err != nil {
			return nil, constants.EmptyStr, err
		}
		return f, constants.Float, nil
	default:
		return nil, constants.EmptyStr, fmt.Errorf("unsupported SQLVal type: %v", sqlVal.Type)
	}
}

func compareStrings(left, right, operator string) (bool, error) {
	switch operator {
	case "=":
		return left == right, nil
	case constants.OperatorNotEquals, constants.OperatorNotEqualsTo:
		return left != right, nil
	case "<":
		return left < right, nil
	case constants.OperatorLessThanEqualsTo:
		return left <= right, nil
	case ">":
		return left > right, nil
	case constants.OperatorGreaterThanEqualsTo:
		return left >= right, nil
	default:
		return false, fmt.Errorf("unsupported operator for strings: %s", operator)
	}
}

func compareInts(left, right int, operator string) (bool, error) {
	switch operator {
	case "=":
		return left == right, nil
	case constants.OperatorNotEquals, constants.OperatorNotEqualsTo:
		return left != right, nil
	case "<":
		return left < right, nil
	case constants.OperatorLessThanEqualsTo:
		return left <= right, nil
	case ">":
		return left > right, nil
	case constants.OperatorGreaterThanEqualsTo:
		return left >= right, nil
	default:
		return false, fmt.Errorf("unsupported operator for integers: %s", operator)
	}
}

func compareFloats(left, right float64, operator string) (bool, error) {
	switch operator {
	case "=":
		return left == right, nil
	case constants.OperatorNotEquals, constants.OperatorNotEqualsTo:
		return left != right, nil
	case "<":
		return left < right, nil
	case constants.OperatorLessThanEqualsTo:
		return left <= right, nil
	case ">":
		return left > right, nil
	case constants.OperatorGreaterThanEqualsTo:
		return left >= right, nil
	default:
		return false, fmt.Errorf("unsupported operator for floats: %s", operator)
	}
}
