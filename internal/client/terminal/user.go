package terminal

import (
	"fmt"
	pb "github.com/faust8888/GophKeeper/internal/pkg/proto/gophkeeperpb"
	"github.com/spf13/cobra"
)

func RegisterCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "register [login] [password]",
		Short: "Register new user in the GophKeeper",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			srv, err := getService(cmd)
			if err != nil {
				fmt.Printf("couldn't get service for registration: %v\n", err)
				return
			}

			registrationRequest := &pb.RegisterUserRequest{Login: args[0], Password: args[1]}
			if err = srv.Register(cmd.Context(), registrationRequest); err != nil {
				fmt.Printf("registration failed: %v\n", err)
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
				fmt.Printf("couldn't get service to log in: %v\n", err)
				return
			}

			loginRequest := &pb.LoginRequest{Login: args[0], Password: args[1]}
			if err = srv.Login(cmd.Context(), loginRequest); err != nil {
				fmt.Printf("registration failed: %v\n", err)
				return
			}
			fmt.Println("Successfully logged in")
		},
	}
}
