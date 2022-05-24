package adder

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type MockAdder struct {
	T testing.TB

	AddMock  func() error
	AddCalls []addCall
}

func NewMockAdder(t testing.TB) *MockAdder {
	return &MockAdder{T: t}
}

type addCall struct {
	Out0 error
}

func (v *MockAdder) Add() error {
	if v.AddMock == nil {
		msg := "call to Add, but MockAdd is not set"
		if v.T == nil {
			panic(msg)
		}
		require.Fail(v.T, msg)
	}

	v.AddCalls = append(v.AddCalls, addCall{})
	out0 := v.AddMock()
	v.AddCalls[len(v.AddCalls)-1].Out0 = out0
	return out0
}
