package lib

import "context"

type Server interface {
	Serve(ctx context.Context) error
	Closable
}
