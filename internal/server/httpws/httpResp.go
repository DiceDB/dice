// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package httpws

const (
	HTTPStatusSuccess string = "success"
	HTTPStatusError   string = "error"
)

type HTTPResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
}
