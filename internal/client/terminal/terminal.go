package terminal

import (
	"context"
	"fmt"
	"github.com/faust8888/GophKeeper/internal/client/service"
	"github.com/spf13/cobra"
)

func New(srv *service.ClientService, version, buildDate string) *cobra.Command {
	commands := &cobra.Command{
		Use:     "GophKeeper",
		Short:   "Password manager",
		Version: fmt.Sprintf("%s (built at %s)", version, buildDate),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			ctx = context.WithValue(ctx, "gophkeeper_client", srv)
			cmd.SetContext(ctx)
			return nil
		},
	}
	commands.AddCommand(
		RegisterCommand(),
		LoginCommand(),
		LoadSecretCommand(),
		GetSecretCommand(),
	)
	return commands
}

func getService(cmd *cobra.Command) (*service.ClientService, error) {
	ctx := cmd.Context()
	cli, ok := ctx.Value("gophkeeper_client").(*service.ClientService)
	if !ok || cli == nil {
		return nil, fmt.Errorf("client not found in context")
	}
	return cli, nil
}
