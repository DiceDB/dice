package observability

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/dicedb/dice/config"
)

type PingPayload struct {
	Date           string         `json:"date"`
	HardwareConfig HardwareConfig `json:"hardware_config"`
	Version        string         `json:"version"`
	InstanceID     string         `json:"instance_id"`
	Err            error          `json:"error"`
}

func Ping() {
	hwConfig, err := GetHardwareMeta()
	if err != nil {
		return
	}

	payload := &PingPayload{
		HardwareConfig: hwConfig,
		InstanceID:     config.DiceConfig.InstanceID,
		Version:        config.DiceConfig.Server.Version,
		Err:            err,
		Date:           time.Now().UTC().Format("2006-01-02 15:04:05"),
	}

	url := "https://api.us-east.aws.tinybird.co/v0/events?name=test"
	b, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(b))
	req.Header.Set("Authorization", "Bearer p.eyJ1IjogIjhjNWQxMjdlLTczZmYtNGRjZS04Mzk5LTQyMDU0MThhYjc2OSIsICJpZCI6ICJhZjcxNGExNC0xZWQyLTQ3ZDktOTM0MS0xMzgwNWNiOWFhNDYiLCAiaG9zdCI6ICJ1cy1lYXN0LWF3cyJ9.o9LqZqTZ9YkhbcusZOltsm95RzVQUzJLQOHV2YA7L0E")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: time.Second * 5}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
}
