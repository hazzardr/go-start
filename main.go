package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/charmbracelet/log"
)

type config struct {
	port   int
	logFmt string
	dsn    string
}

func main() {
	//TODO: remove after project init
	initCmd := flag.NewFlagSet("init", flag.ExitOnError)

	// Check if "init" subcommand is provided
	if len(os.Args) > 1 && os.Args[1] == "init" {
		initCmd.Parse(os.Args[2:])
		if err := initTemplate(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}
	//TODO: End init

	var cfg config
	flag.IntVar(&cfg.port, "port", 4000, "Port to listen on")
	flag.StringVar(&cfg.logFmt, "log-format", "text", "Log format (text|json)")
	flag.Parse()

	logger := log.NewWithOptions(os.Stderr, log.Options{
		ReportCaller:    true,
		ReportTimestamp: true,
		TimeFormat:      time.Kitchen,
	})

	slog.SetDefault(slog.New(logger))
	slog.Info("hello world!")

	errs := make(chan error)
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
		sig := <-quit
		slog.Info("shutting down...", "signal", sig)

		// TODO: Add shutdown logic

		errs <- nil
	}()

	os.Exit(0)
}
