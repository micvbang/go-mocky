package main

import (
	"log"

	"github.com/micvbang/go-mocky"
)

func main() {
	flags, err := mocky.ParseFlags()
	if err != nil {
		log.Fatalf("failed to parse flags: %s", err)
	}

	err = mocky.Run(flags)
	if err != nil {
		log.Fatalf("failed to generate mocks: %s", err)
	}
}
