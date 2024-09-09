package querywatcher

import (
	dstore "github.com/dicedb/dice/internal/store"
)

func CreatePushResponse(query *DSQLQuery, result *[]dstore.DSQLQueryResultRow) (response []interface{}) {
	response = make([]interface{}, 3)
	response[0] = Qwatch
	response[1] = query.String()
	response[2] = *result
	return
}
