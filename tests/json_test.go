package tests

import (
	"testing"

	"github.com/kinbiko/jsonassert"
	"gotest.tools/v3/assert"
)

func TestJSONOperations(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()
	ja := jsonassert.New(t)
	t.Run("Set and Get Simple JSON", func(t *testing.T) {
		setCmd := `JSON.SET user $ {"name":"John","age":30}`
		result := fireCommand(conn, setCmd)
		assert.Equal(t, "OK", result)

		getCmd := `JSON.GET user`
		result = fireCommand(conn, getCmd)

		ja.Assertf(result.(string), `{"age":30,"name":"John"}`)
	})

	t.Run("Set and Get Nested JSON", func(t *testing.T) {
		setCmd := `JSON.SET user:2 $ {"name":"Alice","address":{"city":"New York","zip":"10001"},"array":[1,2,3,4,5]}`
		result := fireCommand(conn, setCmd)
		assert.Equal(t, "OK", result)

		getCmd := `JSON.GET user:2`
		result = fireCommand(conn, getCmd)
		ja.Assertf(result.(string), `{"name":"Alice","address":{"city":"New York","zip":"10001"},"array":[1,2,3,4,5]}`)
	})

	t.Run("Set and Get JSON Array", func(t *testing.T) {
		setCmd := `JSON.SET numbers $ [1,2,3,4,5]`
		result := fireCommand(conn, setCmd)
		assert.Equal(t, "OK", result)

		getCmd := `JSON.GET numbers`
		result = fireCommand(conn, getCmd)
		ja.Assertf(result.(string), `[1,2,3,4,5]`)
	})

	t.Run("Set and Get JSON with Special Characters", func(t *testing.T) {
		setCmd := `JSON.SET special $ {"key":"value with spaces","emoji":"😀"}`
		result := fireCommand(conn, setCmd)
		assert.Equal(t, "OK", result)

		getCmd := `JSON.GET special`
		result = fireCommand(conn, getCmd)
		ja.Assertf(result.(string), `{"key":"value with spaces","emoji":"😀"}`)
	})

	t.Run("Set Invalid JSON", func(t *testing.T) {
		setCmd := `JSON.SET invalid $ {invalid:json}`
		result := fireCommand(conn, setCmd)
		assert.Equal(t, "ERR invalid JSON", result)
	})

	t.Run("Set JSON with Wrong Number of Arguments", func(t *testing.T) {
		setCmd := `JSON.SET`
		result := fireCommand(conn, setCmd)
		assert.Equal(t, "ERR wrong number of arguments for 'JSON.SET' command", result)
	})

	t.Run("Get JSON with Wrong Number of Arguments", func(t *testing.T) {
		getCmd := `JSON.GET`
		result := fireCommand(conn, getCmd)
		assert.Equal(t, "ERR wrong number of arguments for 'JSON.GET' command", result)
	})

	t.Run("Set Non-JSON Value", func(t *testing.T) {
		setCmd := `SET nonJson "not a json"`
		result := fireCommand(conn, setCmd)
		assert.Equal(t, "OK", result)
		getCmd := `JSON.GET nonJson`
		result = fireCommand(conn, getCmd)

		assert.Equal(t, "WRONGTYPE Operation against a key holding the wrong kind of value", result)
	})

	t.Run("Set Empty JSON Object", func(t *testing.T) {
		setCmd := `JSON.SET empty $ {}`
		result := fireCommand(conn, setCmd)
		assert.Equal(t, "OK", result)
		getCmd := `JSON.GET empty`
		result = fireCommand(conn, getCmd)
		ja.Assertf(result.(string), `{}`)
	})

	t.Run("Set Empty JSON Array", func(t *testing.T) {
		setCmd := `JSON.SET emptyArray $ []`
		result := fireCommand(conn, setCmd)
		assert.Equal(t, "OK", result)

		getCmd := `JSON.GET emptyArray`
		result = fireCommand(conn, getCmd)
		ja.Assertf(result.(string), `[]`)
	})
	t.Run("Set JSON with Unicode", func(t *testing.T) {
		setCmd := `JSON.SET unicode $ {"unicode":"こんにちは世界"}`
		result := fireCommand(conn, setCmd)
		assert.Equal(t, "OK", result)
		getCmd := `JSON.GET unicode`
		result = fireCommand(conn, getCmd)

		ja.Assertf(result.(string), `{"unicode":"こんにちは世界"}`)
	})

	t.Run("Set JSON with Escaped Characters", func(t *testing.T) {
		setCmd := `JSON.SET escaped $ {"escaped":"\"quoted\", \\backslash\\ and \/forward\/slash"}`
		result := fireCommand(conn, setCmd)
		assert.Equal(t, "OK", result)
		getCmd := `JSON.GET escaped`
		result = fireCommand(conn, getCmd)

		ja.Assertf(result.(string), `{"escaped":"\"quoted\", \\backslash\\ and /forward/slash"}`)
	})

}
