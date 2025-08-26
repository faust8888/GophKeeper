package server

import (
	"context"
	"errors"
	"fmt"
	pb "github.com/faust8888/GophKeeper/internal/pkg/proto/gophkeeperpb"
	"github.com/faust8888/GophKeeper/internal/server/config"
	"github.com/faust8888/GophKeeper/internal/server/logger"
	"github.com/faust8888/GophKeeper/internal/server/migration"
	"github.com/faust8888/GophKeeper/internal/server/repository/postgres"
	"github.com/faust8888/GophKeeper/internal/server/service"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type GRPCServer struct {
	Server *grpc.Server
	Cfg    *config.Config
}

func (srv *GRPCServer) Run() error {
	// --- Setup Signal Handling ---
	ctx, stop := createNotifyContext()
	defer stop()
	group, groupContext := errgroup.WithContext(ctx)

	group.Go(func() error {
		return srv.runGRPCWithContext(groupContext)
	})

	// Wait for either server to fail or signal to terminate
	if err := group.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("application error: %w", err)
	}

	logger.Log.Info("Application stopped successfully")
	return nil
}

func (srv *GRPCServer) runGRPCWithContext(ctx context.Context) error {
	lis, err := net.Listen("tcp", srv.Cfg.ServerGRPCAddress)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", srv.Cfg.ServerGRPCAddress, err)
	}

	// Serve gRPC in a goroutine
	go func() {
		logger.Log.Info("Starting gRPC server", zap.String("address", srv.Cfg.ServerGRPCAddress))
		if err = srv.Server.Serve(lis); err != nil && err != grpc.ErrServerStopped {
			logger.Log.Error("gRPC server error", zap.Error(err))
		}
	}()

	// Wait for shutdown signal
	<-ctx.Done()

	logger.Log.Info("Shutting down gRPC server gracefully...")
	shutdownDone := make(chan struct{})
	go func() {
		srv.Server.GracefulStop()
		close(shutdownDone)
	}()

	select {
	case <-shutdownDone:
		return nil
	case <-time.After(10 * time.Second):
		srv.Server.Stop() // Force stop after timeout
		return context.DeadlineExceeded
	}
}

func NewGRPCServer() *GRPCServer {
	cfg := config.Create()
	if err := logger.Initialize(cfg.LoggingLevel); err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize logger: %v", err)
		os.Exit(1)
	}

	err := migration.Run(cfg.DataSourceName)
	if err != nil {
		logger.Log.Error("migration error", zap.Error(err))
		os.Exit(1)
	}

	repo, err := postgres.NewPostgresRepository(cfg)
	if err != nil {
		logger.Log.Error("failed to initialize repository", zap.Error(err))
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()
	gophKeeperService := service.NewGophKeeperService(cfg, repo)
	pb.RegisterGophKeeperServer(grpcServer, gophKeeperService)

	return &GRPCServer{
		Server: grpcServer,
		Cfg:    cfg,
	}
}

func createNotifyContext() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, os.Interrupt)
}
