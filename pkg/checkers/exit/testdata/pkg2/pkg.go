package main

import (
	"os"

	"github.com/caarlos0/env/v6"
	"github.com/rs/zerolog"
)

var cfgEnv struct{}

func foo() {
	os.Exit(100)
}

func main() {
	logger := zerolog.Logger{}

	err := env.Parse(&cfgEnv)
	if err != nil {
		logger.Println(err)
		os.Exit(1) // want "direct calling os.Exit in main func"
	}

	foo()
}
