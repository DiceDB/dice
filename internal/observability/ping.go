// This file is part of DiceDB.
// Copyright (C) 2025DiceDB (dicedb.io).
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package observability

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/dicedb/dice/config"
)

type PingPayload struct {
	Date           string         `json:"date"`
	HardwareConfig HardwareConfig `json:"hardware_config"`
	DBConfig       DBConfig       `json:"db_config"`
	Version        string         `json:"version"`
	InstanceID     string         `json:"instance_id"`
	Err            error          `json:"error"`
}

const (
	token = "p.eyJ1IjogIjhjNWQxMjdlLTczZmYtNGRjZS04Mzk5LTQyMDU0MThhYjc2OSIsI" +
		"CJpZCI6ICJhZjcxNGExNC0xZWQyLTQ3ZDktOTM0MS0xMzgwNWNiOWFhNDYiLCAiaG9zdCI6ICJ1cy1lYXN0LWF3cyJ9.o9LqZqTZ9YkhbcusZOltsm95RzVQUzJLQOHV2YA7L0E"
	url = "https://api.us-east.aws.tinybird.co/v0/events?name=ping2"
)

type DBConfig struct {
}

func Ping() {
	hwConfig, err := GetHardwareMeta()
	if err != nil {
		return
	}

	payload := &PingPayload{
		HardwareConfig: hwConfig,
		InstanceID:     config.DiceConfig.InstanceID,
		Version:        config.DiceConfig.Version,
		Err:            err,
		Date:           time.Now().UTC().Format("2006-01-02 15:04:05"),
		DBConfig:       DBConfig{},
	}

	b, _ := json.Marshal(payload)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(b))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: time.Second * 5}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("Error reporting observability metrics.", slog.Any("error", err))
		return
	}

	_ = resp.Body.Close()
}
