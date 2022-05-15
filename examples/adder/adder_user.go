package adder

import (
	"context"
	"fmt"
)

func GiveValueToAdder(a Adder, v string) error {
	ctx := context.Background()
	err := a.Add(ctx, v)
	if err != nil {
		return fmt.Errorf("failed to add: %w", err)
	}

	return nil
}
