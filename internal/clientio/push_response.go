package clientio

import (
	"github.com/dicedb/dice/internal/sql"
)

// CreatePushResponse creates a push response. Push responses refer to messages that the server sends to clients without
// the client explicitly requesting them. These are typically seen in scenarios where the client has subscribed to some
// kind of event or data feed and is notified in real-time when changes occur.
// `key` is the unique key that identifies the push response.
func CreatePushResponse[T any](key string, result T) (response []interface{}) {
	response = make([]interface{}, 3)
	response[0] = sql.Qwatch
	response[1] = key
	response[2] = result
	return
}
