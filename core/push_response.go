package core

import "github.com/dicedb/dice/internal/constants"

func CreatePushResponse(query *DSQLQuery, result *[]DSQLQueryResultRow) (response []interface{}) {
	response = make([]interface{}, 3)
	response[0] = constants.Qwatch
	response[1] = query.String()
	response[2] = *result
	return
}
