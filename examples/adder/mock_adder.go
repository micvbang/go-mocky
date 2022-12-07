package adder

import (
	"fmt"
	"testing"
)

type MockAdder struct {
	T testing.TB

	AddMock  func() error
	AddCalls []adderAddCall
}

func NewMockAdder(t testing.TB) *MockAdder {
	return &MockAdder{T: t}
}

type adderAddCall struct {
	Out0 error
}

func (_v *MockAdder) Add() error {
	if _v.AddMock == nil {
		msg := fmt.Sprintf("call to %T.Add, but MockAdd is not set", _v)
		panic(msg)
	}

	_v.AddCalls = append(_v.AddCalls, adderAddCall{})
	out0 := _v.AddMock()
	_v.AddCalls[len(_v.AddCalls)-1].Out0 = out0
	return out0
}
