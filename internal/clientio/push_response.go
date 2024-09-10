package clientio

import (
	"github.com/dicedb/dice/internal/sql"
)

func CreatePushResponse(query *sql.DSQLQuery, result *[]sql.QueryResultRow) (response []interface{}) {
	response = make([]interface{}, 3)
	response[0] = sql.Qwatch
	response[1] = query.String()
	response[2] = *result
	return
}
