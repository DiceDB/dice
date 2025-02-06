package parsers

import (
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
)

func FireWSCommandAndReadResponse(conn *websocket.Conn, cmd string) (interface{}, error) {
	err := FireWSCommand(conn, cmd)
	if err != nil {
		return nil, err
	}

	// read the response
	_, resp, err := conn.ReadMessage()
	if err != nil {
		return nil, err
	}

	// marshal to json
	var respJSON interface{}
	if err = json.Unmarshal(resp, &respJSON); err != nil {
		return nil, fmt.Errorf("error unmarshaling response")
	}
	respJSON = ParseResponse(respJSON)
	return respJSON, nil
}

func FireWSCommand(conn *websocket.Conn, cmd string) error {
	// send request
	err := conn.WriteMessage(websocket.TextMessage, []byte(cmd))
	if err != nil {
		return err
	}

	return nil
}
