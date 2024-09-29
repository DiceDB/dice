package iohandler

import (
	"context"
)

type IOHandler interface {
	Read(ctx context.Context) ([]byte, error)
	Write(ctx context.Context, response interface{}) error
	Close() error
}
