package main

import (
	"fmt"

	"github.com/caarlos0/env/v6"
	"github.com/rs/zerolog/log"
)

var (
	cfgEnv Config
	cfg    Config

	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func printStartMessage() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}

func main() {
	printStartMessage()

	err := env.Parse(&cfgEnv)
	if err != nil {
		log.Err(err).Send()
		return
	}

	err = root.Execute()
	if err != nil {
		log.Err(err).Send()
		return
	}
}
