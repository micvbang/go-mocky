package adder

//go:generate mocky -i Adder

type Adder interface {
	Add() error
}
