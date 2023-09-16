package tests

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"testing"
)


func TestMain(m *testing.M) {
	var wg sync.WaitGroup
	go runTestServer(&wg)
	go func(){
		conn := getLocalConnection()
		fireCommand(conn, "ABORT")
		conn.Close()
		wg.Wait()
	}()
	m.Run()
}

func TestSetAndGet(t *testing.T) {
	
	conn := getLocalConnection()
	t.Run("SingleCommand", func(t *testing.T) {
		key := "k1"
		value := "value1"

		fireCommand(conn, fmt.Sprintf("SET %s %s", key, value))
		retValue := fireCommand(conn, fmt.Sprintf("GET %s", key)).(string)
		if retValue != value {
			t.Errorf("Expected value %s, got %s", value, retValue)
		}

	})

	t.Run("MultipleCommands", func(t *testing.T) {
		for i := 1; i <= 10; i++ {
			key := fmt.Sprintf("k%d", i)
			value := fmt.Sprintf("value%d", i)
			fireCommand(conn, fmt.Sprintf("SET %s %s", key, value))
		}
		for i := 1; i <= 10; i++ {
			key := fmt.Sprintf("k%d", i)
			expectedValue := fmt.Sprintf("value%d", i)
			retValue := fireCommand(conn, fmt.Sprintf("GET %s", key)).(string)
			if retValue != expectedValue {
				t.Errorf("Expected value %s for key %s, got %s", expectedValue, key, retValue)
			}
		}

	})
	t.Run("InvalidCommands", func(t *testing.T) {

		invalidCommands := []string{
			"SET key",       // No value
			"GET",           // No key
			"SET k1",        // No value
			"GET k_invalid", // Non-existent key
		}
		for _, cmd := range invalidCommands {
			response := fireCommand(conn, cmd).(string)
			if response != "ERROR" {
				t.Errorf("Expected ERROR response for invalid command %s, got %s", cmd, response)
			}
		}
	})

	t.Run("LargeInput", func(t *testing.T) {
		key := "large_key"
		largeValue := strings.Repeat("a", 1024*2024)
		log.Println(largeValue)
		fireCommand(conn, fmt.Sprintf("SET %s %s", key, largeValue))
		retValue := fireCommand(conn, fmt.Sprintf("GET %s", key)).(string)
		if retValue != largeValue {
			t.Errorf("Expected large value to be retrieved correctly")
		}
	})
}
