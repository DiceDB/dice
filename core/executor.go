package core

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/xwb1989/sqlparser"
)

type DSQLQueryResultRow struct {
	Key   string
	Value *Obj
}

func ExecuteQuery(query DSQLQuery) ([]DSQLQueryResultRow, error) {
	var result []DSQLQueryResultRow

	var err error
	withLocks(DefaultLockIdentifier, func() {
		for key, ptr := range keypool {
			if WildCardMatch(query.KeyRegex, key) {
				row := DSQLQueryResultRow{
					Key:   key,
					Value: store[ptr],
				}

				if query.Where != nil {
					match, evalErr := evaluateWhereClause(query.Where, row)
					if evalErr != nil {
						err = evalErr
						return
					}
					if !match {
						continue
					}
				}

				result = append(result, row)
			}
		}
	}, WithStoreRLock(), WithKeypoolRLock())

	if err != nil {
		return nil, err
	}

	sortResults(query, result)

	if !query.Selection.KeySelection {
		for i := range result {
			result[i].Key = ""
		}
	}

	if !query.Selection.ValueSelection {
		for i := range result {
			result[i].Value = nil
		}
	}

	if query.Limit > 0 && query.Limit < len(result) {
		result = result[:query.Limit]
	}

	return result, nil
}

func sortResults(query DSQLQuery, result []DSQLQueryResultRow) {
	if query.OrderBy.OrderBy == "" {
		return
	}

	switch query.OrderBy.OrderBy {
	case CustomKey:
		sort.Slice(result, func(i, j int) bool {
			if query.OrderBy.Order == "desc" {
				return result[i].Key > result[j].Key
			}
			return result[i].Key < result[j].Key
		})
	case CustomValue:
		sort.Slice(result, func(i, j int) bool {
			return compareValues(query.OrderBy.Order, result[i].Value, result[j].Value)
		})
	}
}

func compareValues(order string, valI, valJ *Obj) bool {
	if valI == nil || valJ == nil {
		return handleNilForOrder(order, valI, valJ)
	}

	typeI, encodingI := ExtractTypeEncoding(valI)
	typeJ, encodingJ := ExtractTypeEncoding(valJ)

	if typeI != typeJ || encodingI != encodingJ {
		return handleMismatch()
	}

	switch kindI := valI.Value.(type) {
	case string:
		kindJ, ok := valJ.Value.(string)
		if !ok {
			return handleMismatch()
		}
		return compareStringValues(order, kindI, kindJ)
	case int:
		kindJ, ok := valJ.Value.(int)
		if !ok {
			return handleMismatch()
		}
		return compareIntValues(order, uint8(kindI), uint8(kindJ))
	default:
		return handleUnsupportedType()
	}
}

func handleNilForOrder(order string, valI, valJ *Obj) bool {
	if order == "asc" {
		return valI == nil
	}
	return valJ == nil
}

func handleMismatch() bool {
	// undefined behavior
	return true
}

func handleUnsupportedType() bool {
	// undefined behavior
	return true
}

func compareStringValues(order string, valI, valJ string) bool {
	if order == "asc" {
		return valI < valJ
	}
	return valI > valJ
}

func compareIntValues(order string, valI, valJ uint8) bool {
	if order == "asc" {
		return valI < valJ
	}
	return valI > valJ
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

func evaluateComparison(expr *sqlparser.ComparisonExpr, row DSQLQueryResultRow) (bool, error) {
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
	case "string":
		return compareStrings(left.(string), right.(string), expr.Operator)
	case "int":
		return compareInts(left.(int), right.(int), expr.Operator)
	case "float":
		return compareFloats(left.(float64), right.(float64), expr.Operator)
	default:
		return false, fmt.Errorf("unsupported type for comparison: %s", leftType)
	}
}

func getExprValueAndType(expr sqlparser.Expr, row DSQLQueryResultRow) (interface{}, string, error) {
	switch expr := expr.(type) {
	case *sqlparser.ColName:
		switch expr.Name.String() {
		case TempKey:
			return row.Key, "string", nil
		case TempValue:
			return getValueAndType(row.Value)
		default:
			return nil, "", fmt.Errorf("unknown column: %s", expr.Name.String())
		}
	case *sqlparser.SQLVal:
		return sqlValToGoValue(expr)
	default:
		return nil, "", fmt.Errorf("unsupported expression type: %T", expr)
	}
}

func getValueAndType(obj *Obj) (interface{}, string, error) {
	switch v := obj.Value.(type) {
	case string:
		return v, "string", nil
	case int:
		return v, "int", nil
	case float64:
		return v, "float", nil
	default:
		return nil, "", fmt.Errorf("unsupported value type: %T", v)
	}
}

func sqlValToGoValue(sqlVal *sqlparser.SQLVal) (interface{}, string, error) {
	switch sqlVal.Type {
	case sqlparser.StrVal:
		return string(sqlVal.Val), "string", nil
	case sqlparser.IntVal:
		i, err := strconv.Atoi(string(sqlVal.Val))
		if err != nil {
			return nil, "", err
		}
		return i, "int", nil
	case sqlparser.FloatVal:
		f, err := strconv.ParseFloat(string(sqlVal.Val), 64)
		if err != nil {
			return nil, "", err
		}
		return f, "float", nil
	default:
		return nil, "", fmt.Errorf("unsupported SQLVal type: %v", sqlVal.Type)
	}
}

func compareStrings(left, right string, operator string) (bool, error) {
	switch operator {
	case "=":
		return left == right, nil
	case "!=", "<>":
		return left != right, nil
	case "<":
		return left < right, nil
	case "<=":
		return left <= right, nil
	case ">":
		return left > right, nil
	case ">=":
		return left >= right, nil
	default:
		return false, fmt.Errorf("unsupported operator for strings: %s", operator)
	}
}

func compareInts(left, right int, operator string) (bool, error) {
	switch operator {
	case "=":
		return left == right, nil
	case "!=", "<>":
		return left != right, nil
	case "<":
		return left < right, nil
	case "<=":
		return left <= right, nil
	case ">":
		return left > right, nil
	case ">=":
		return left >= right, nil
	default:
		return false, fmt.Errorf("unsupported operator for integers: %s", operator)
	}
}

func compareFloats(left, right float64, operator string) (bool, error) {
	switch operator {
	case "=":
		return left == right, nil
	case "!=", "<>":
		return left != right, nil
	case "<":
		return left < right, nil
	case "<=":
		return left <= right, nil
	case ">":
		return left > right, nil
	case ">=":
		return left >= right, nil
	default:
		return false, fmt.Errorf("unsupported operator for floats: %s", operator)
	}
}
