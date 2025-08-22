package terminal

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

func LoadSecretCommand() *cobra.Command {
	var (
		secretType  string
		description string
		secretFile  string
	)

	loadSecretCommand := &cobra.Command{
		Use:   "load-secret [data]",
		Short: "Add new secret (for text: [data], for card: [number] [expiry] [cvv], for binary: use --file)",
		Args:  cobra.ArbitraryArgs,
		Run: func(cmd *cobra.Command, args []string) {
			srv, err := getService(cmd)
			if err != nil {
				fmt.Printf("client error: %v\n", err)
				return
			}

			var data []byte
			switch secretType {
			case "text":
				if len(args) < 1 {
					fmt.Println("text secret requires at least 1 argument")
					return
				}
				fmt.Println(args[0])
				data = []byte(args[0])
			case "binary":
				if secretFile == "" {
					fmt.Println("binary secret requires --file flag")
					return
				}
				data, err = os.ReadFile(secretFile)
				if err != nil {
					fmt.Printf("Error reading file: %v\n", err)
					return
				}
			case "card":
				if len(args) < 3 {
					fmt.Println("card secret requires 3 arguments: number, expiry, cvv")
					return
				}
				data = make([]byte, 0, 128)
				data = fmt.Appendf(data, `{"number":"%s","expiry":"%s","cvv":"%s"}`, args[0], args[1], args[2])
			default:
				fmt.Println("invalid secret type")
				return
			}

			if err = srv.AddSecret(cmd.Context(), secretType, description, data); err != nil {
				fmt.Printf("error adding secret: %v\n", err)
				return
			}
			fmt.Println("secret added successfully")
		},
	}

	loadSecretCommand.Flags().StringVarP(&secretType, "type", "t", "", "secret type (text|binary|card)")
	loadSecretCommand.Flags().StringVarP(&description, "desc", "d", "", "secret description")
	loadSecretCommand.Flags().StringVarP(&secretFile, "file", "f", "", "File path for binary data")

	return loadSecretCommand
}

func GetSecretCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get-secret [secret-id]",
		Short: "Get secret data by ID",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			srv, err := getService(cmd)
			if err != nil {
				fmt.Printf("client error: %v\n", err)
				return
			}

			secret, err := srv.GetSecret(cmd.Context(), args[0])
			if err != nil {
				fmt.Printf("error getting secret: %v\n", err)
				os.Exit(1)
			}

			switch secret.Type {
			case "text":
				fmt.Printf("content: %s\n", string(secret.Data))
			case "card":
				fmt.Printf("card: %s",
					string(secret.Data),
				)
			default:
				fmt.Printf("secret ID: %s\ntype: %s\nmetadata: %v\ndata: [%d bytes]\n",
					secret.ID, secret.Type, secret.Metadata, len(secret.Data))
			}
		},
	}
}
