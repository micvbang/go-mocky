package adder

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type MockAdder struct {
	T testing.TB

	AddMock  func(ctx context.Context, value string) error
	AddCalls []addCall
}

func NewMockAdder(t testing.TB) *MockAdder {
	return &MockAdder{T: t}
}

type addCall struct {
	Ctx   context.Context
	Value string

	Out0 error
}

func (v *MockAdder) Add(ctx context.Context, value string) error {
	if v.AddMock == nil {
		msg := "call to Add, but MockAdd is not set"
		if v.T == nil {
			panic(msg)
		}
		require.Fail(v.T, msg)
	}

	v.AddCalls = append(v.AddCalls, addCall{
		Ctx:   ctx,
		Value: value,
	})
	out0 := v.AddMock(ctx, value)
	v.AddCalls[len(v.AddCalls)-1].Out0 = out0
	return out0
}
