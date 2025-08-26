package main

import (
	"fmt"
	"github.com/faust8888/GophKeeper/internal/client"
	"github.com/faust8888/GophKeeper/internal/client/config"
	"github.com/faust8888/GophKeeper/internal/client/repository/sqllite"
	"github.com/faust8888/GophKeeper/internal/client/service"
	"github.com/faust8888/GophKeeper/internal/client/terminal"
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

	clientGRPC, err := client.NewGRPCClient(cfg.ServerGRPCAddress)
	if err != nil {
		return fmt.Errorf("new grpc client error: %w\n", err)
	}

	repo, err := sqllite.NewRepository(cfg.DataSourceName)
	if err != nil {
		return fmt.Errorf("coldn't create client repository: %w", err)
	}

	srv := service.New(cfg, repo, clientGRPC)

	t := terminal.New(srv, version, buildDate)

	if err = t.Execute(); err != nil {
		return fmt.Errorf("couldn't run terminal %w\n", err)
	}
	return nil
}
