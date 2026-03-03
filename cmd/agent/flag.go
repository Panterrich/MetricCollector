package main

import (
	"fmt"
	"net"

	"github.com/spf13/cobra"
)

var (
	DefaultEndPoint            = "localhost:8080"
	DefaultReportInterval uint = 10
	DefaultPollInterval   uint = 2
	DefaultKeyHash             = ""
	DefaultRateLimit      uint = 1
)

type Config struct {
	EndPoint       string `env:"ADDRESS"`
	ReportInterval uint   `env:"REPORT_INTERVAL"`
	PollInterval   uint   `env:"POLL_INTERVAL"`
	KeyHash        string `env:"KEY"`
	RateLimit      uint   `env:"RATE_LIMIT"`
}

var (
	root = &cobra.Command{
		Use:   "server",
		Short: "Server for storing metrics",
		Long:  "Server for storing metrics",
		Args: func(cmd *cobra.Command, args []string) error {

			if err := cobra.ExactArgs(0)(cmd, args); err != nil {
				return err
			}

			if _, _, err := net.SplitHostPort(cfg.EndPoint); err != nil {
				return fmt.Errorf("invalid end-point for HTTP-server: %w", err)
			}

			if cfg.PollInterval == 0 || cfg.ReportInterval == 0 {
				return fmt.Errorf("zero interval")
			}

			if cfg.RateLimit == 0 {
				return fmt.Errorf("zero rate limit")
			}

			return nil
		},
		PreRun: preRun,
		RunE:   run,
	}
)

func init() {
	root.Flags().StringVarP(&cfg.EndPoint, "a", "a", DefaultEndPoint, "end-point for HTTP-server")
	root.Flags().UintVarP(&cfg.ReportInterval, "r", "r", DefaultReportInterval, "report interval")
	root.Flags().UintVarP(&cfg.PollInterval, "p", "p", DefaultPollInterval, "poll interval")
	root.Flags().StringVarP(&cfg.KeyHash, "key", "k", DefaultKeyHash, "key for hash sha256")
	root.Flags().UintVarP(&cfg.RateLimit, "l", "l", DefaultRateLimit, "rate limit")
}

func preRun(_ *cobra.Command, _ []string) {
	if cfgEnv.EndPoint != "" {
		cfg.EndPoint = cfgEnv.EndPoint
	}

	if cfgEnv.ReportInterval != 0 {
		cfg.ReportInterval = cfgEnv.ReportInterval
	}

	if cfgEnv.PollInterval != 0 {
		cfg.PollInterval = cfgEnv.PollInterval
	}

	if cfgEnv.KeyHash != "" {
		cfg.KeyHash = cfgEnv.KeyHash
	}
}
