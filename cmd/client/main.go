package main

import (
	"fmt"
	"github.com/faust8888/GophKeeper/internal/client"
	"github.com/faust8888/GophKeeper/internal/client/config"
	"github.com/faust8888/GophKeeper/internal/client/repository/sqllite"
	"github.com/faust8888/GophKeeper/internal/client/service"
	"github.com/faust8888/GophKeeper/internal/client/terminal"
	"github.com/faust8888/GophKeeper/internal/server/logger"
	"os"
)

var (
	version   = "N/A"
	buildDate = "N/A"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("run agent error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg := config.Create()
	if err := logger.Initialize(cfg.LoggingLevel); err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize logger: %v", err)
		os.Exit(1)
	}

	clnt, err := client.NewGRPCClient(cfg)
	if err != nil {
		return fmt.Errorf("new grpc client error: %w\n", err)
	}

	repo, err := sqllite.NewRepository(cfg.DataSourceName)
	if err != nil {
		return fmt.Errorf("coldn't create client repository: %w", err)
	}

	srv, err := service.New(cfg, repo, clnt)
	if err != nil {
		return fmt.Errorf("new user service error: %w\n", err)
	}

	t := terminal.New(srv, version, buildDate)

	if err = t.Execute(); err != nil {
		return fmt.Errorf("couldn't run terminal %w\n", err)
	}
	return nil
}
