package clientio

import (
	"github.com/dicedb/dice/internal/sql"
	dstore "github.com/dicedb/dice/internal/store"
)

func CreatePushResponse(query *sql.DSQLQuery, result *[]dstore.DSQLQueryResultRow) (response []interface{}) {
	response = make([]interface{}, 3)
	response[0] = sql.Qwatch
	response[1] = query.String()
	response[2] = *result
	return
}
