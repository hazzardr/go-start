package main

import (
	"flag"
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
	var cfg config
	flag.IntVar(&cfg.port, "port", 4000, "Port to listen on")
	flag.StringVar(&cfg.logFmt, "log-format", "text", "Log format (text|json)")
	//flag.StringVar(&cfg.dsn, "dsn", os.Getenv("DATABASE_URL"), "Database connection string")
	flag.Parse()

	logger := log.NewWithOptions(os.Stderr, log.Options{
		ReportCaller:    true,
		ReportTimestamp: true,
		TimeFormat:      time.Kitchen,
	})

	//if cfg.dsn == "" {
	//	slog.Error("database URL is required")
	//	os.Exit(1)
	//}

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
