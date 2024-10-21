package abstractserver

import "context"

type AbstractServer interface {
	Run(ctx context.Context) error
}
