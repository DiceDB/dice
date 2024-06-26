package core

import "sort"

type DSQLQueryResultRow struct {
	Key   string
	Value *Obj
}

// TODO: Implement thread-safe access to the keypool and store.
func ExecuteQuery(query DSQLQuery) ([]DSQLQueryResultRow, error) {
	var result []DSQLQueryResultRow

	for key, ptr := range keypool {
		if RegexMatch(query.KeyRegex, key) {
			row := DSQLQueryResultRow{}
			if query.Selection.KeySelection {
				row.Key = key
			}
			if query.Selection.ValueSelection {
				row.Value = store[ptr]
			}
			result = append(result, row)
		}
	}

	sortResults(query, result)

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
