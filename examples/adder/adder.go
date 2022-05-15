package adder

import (
	"context"
)

//go:generate mocky -i Adder

type Adder interface {
	Add(ctx context.Context, value string) error
}
