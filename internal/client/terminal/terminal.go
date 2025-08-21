package terminal

import (
	"context"
	"fmt"
	"github.com/faust8888/GophKeeper/internal/client/service"
	pb "github.com/faust8888/GophKeeper/internal/pkg/proto/gophkeeperpb"
	"github.com/spf13/cobra"
	"os"
)

func New(srv *service.Service, version, buildDate string) *cobra.Command {
	rootCmd := &cobra.Command{
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

	rootCmd.AddCommand(
		RegisterCommand(),
		LoginCommand(),
		LoadSecretCommand(),
		GetSecretCommand(),
	)
	return rootCmd
}

func RegisterCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "register [login] [password]",
		Short: "Register new user",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			srv, err := getService(cmd)
			if err != nil {
				fmt.Printf("cleint error: %v\n", err)
				return
			}

			if err = srv.SignUpUser(cmd.Context(), &pb.RegisterUserRequest{Login: args[0], Password: args[1]}); err != nil {
				fmt.Printf("Registration failed: %v\n", err)
				return
			}
			fmt.Println("Registration completed!")
		},
	}
}

func LoginCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "login [login] [password]",
		Short: "Authenticate user",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			srv, err := getService(cmd)
			if err != nil {
				fmt.Printf("cleint error: %v\n", err)
				return
			}

			if err = srv.SignInUser(cmd.Context(), &pb.LoginRequest{Login: args[0], Password: args[1]}); err != nil {
				fmt.Printf("Login failed: %v\n", err)
				return
			}
			fmt.Println("Successfully logged in")
		},
	}
}

func LoadSecretCommand() *cobra.Command {
	var (
		secretType  string
		description string
		secretFile  string
	)

	secretCmd := &cobra.Command{
		Use:   "add-secret [data]",
		Short: "Add new secret (for text: [data], for card: [number] [expiry] [cvv], for binary: use --file)",
		Args:  cobra.ArbitraryArgs,
		Run: func(cmd *cobra.Command, args []string) {

			// Чтение данных в зависимости от типа
			var data []byte
			var err error

			srv, err := getService(cmd)
			if err != nil {
				fmt.Printf("cleint error: %v\n", err)
				return
			}

			switch secretType {
			case "text":
				if len(args) < 1 {
					fmt.Println("Text secret requires at least 1 argument")
					return
				}
				fmt.Println(args[0])
				data = []byte(args[0])
			case "binary":
				if secretFile == "" {
					fmt.Println("Binary secret requires --file flag")
					return
				}
				data, err = os.ReadFile(secretFile)
				if err != nil {
					fmt.Printf("Error reading file: %v\n", err)
					return
				}
			case "card":
				if len(args) < 3 {
					fmt.Println("Card secret requires 3 arguments: number, expiry, cvv")
					return
				}
				data = make([]byte, 0, 128)
				data = fmt.Appendf(data, `{"number":"%s","expiry":"%s","cvv":"%s"}`, args[0], args[1], args[2])
			default:
				fmt.Println("Invalid secret type")
				return
			}

			if err := srv.AddSecret(cmd.Context(), secretType, description, data); err != nil {
				fmt.Printf("Error adding secret: %v\n", err)
				return
			}

			fmt.Println("Secret added successfully")
		},
	}

	secretCmd.Flags().StringVarP(&secretType, "type", "t", "", "Secret type (text|binary|card)")
	secretCmd.Flags().StringVarP(&description, "desc", "d", "", "Secret description")
	secretCmd.Flags().StringVarP(&secretFile, "file", "f", "", "File path for binary data")

	return secretCmd
}

func GetSecretCommand() *cobra.Command {
	getCmd := &cobra.Command{
		Use:   "get-secret [secret-id]",
		Short: "Get secret data by ID",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			srv, err := getService(cmd)
			if err != nil {
				fmt.Printf("cleint error: %v\n", err)
				return
			}

			// Получаем секрет
			secret, err := srv.GetSecret(cmd.Context(), args[0])
			if err != nil {
				fmt.Printf("Error getting secret: %v\n", err)
				os.Exit(1)
			}

			// Выводим результат в зависимости от типа
			switch secret.Type {
			case "text":
				fmt.Printf("Text content: %s\n", string(secret.Data))
			case "card":
				fmt.Printf("Card: %s",
					string(secret.Data),
				)
			default:
				fmt.Printf("Secret ID: %s\nType: %s\nMetadata: %v\nData: [%d bytes]\n",
					secret.ID, secret.Type, secret.Metadata, len(secret.Data))
			}
		},
	}
	return getCmd
}

func getService(cmd *cobra.Command) (*service.Service, error) {
	ctx := cmd.Context()
	cli, ok := ctx.Value("gophkeeper_client").(*service.Service)
	if !ok || cli == nil {
		return nil, fmt.Errorf("client not found in context")
	}
	return cli, nil
}
