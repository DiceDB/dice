// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package iohandler

import (
	"context"
)

type IOHandler interface {
	Read(ctx context.Context) ([]byte, error)
	Write(ctx context.Context, response interface{}) error
	Close() error
}
