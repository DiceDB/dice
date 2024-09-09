package parser

import (
	"fmt"
	"testing"
)

func TestParsing(t *testing.T) {
	// Example SQL query
	query := "SELECT * FROM users*dkfj:gdf???dfg WHERE $value.id = 1 AND $value.score > 50.5 OR $value.name = 'some_name' ORDER BY $value.score[0].name, $value.name LIMIT 512;"

	result, err := Parse(query)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("%+v\n", result)
}
