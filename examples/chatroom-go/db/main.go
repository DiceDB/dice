package db

import "github.com/dicedb/dicedb-go"

var (
	Client *dicedb.Client
)

func init() {
	var err error
	Client, err = dicedb.NewClient("localhost", 7379, dicedb.WithWatch())
	if err != nil {
		panic(err)
	}
}
