package global

import "context"

type IOThread interface {
	ID() string
	Start(context.Context) error
	Stop() error
	String() string
}
