package main

// @Title MetricCollector API
// @Description Metric collector service.
// @Version 1.0

import (
	"fmt"

	"github.com/caarlos0/env"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	cfg    = Config{LogLevel: int(zerolog.TraceLevel) - 1}
	cfgEnv = Config{LogLevel: int(zerolog.TraceLevel) - 1}

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
