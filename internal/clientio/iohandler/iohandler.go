package iohandler

import "context"

type IOHandler interface {
	ReadRequest(ctx context.Context) ([]byte, error)
	WriteResponse(ctx context.Context, response []byte) error
	Close() error
}
