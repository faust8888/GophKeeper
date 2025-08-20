package main

import (
	"fmt"
	"github.com/faust8888/GophKeeper/internal/server"
	"os"
)

var (
	// buildVersion is the semantic version of the build (e.g., v1.0.0).
	// This value is typically injected at compile time via -ldflags.
	buildVersion string = "N/A"

	// buildDate is the timestamp when the binary was built, usually in ISO format.
	// This value is typically injected at compile time via -ldflags.
	buildDate string = "N/A"

	// buildCommit is the git commit hash used to build the binary.
	// This value is typically injected at compile time via -ldflags.
	buildCommit string = "N/A"
)

// main starts the GophKeeper server.
//
// It initializes logging, configures the repository based on configuration,
// sets up routing, and starts the HTTP server with graceful shutdown.
func main() {
	printBuildInfo()
	if err := server.NewGRPCServer().Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Application error: %v\n", err)
		os.Exit(1)
	}
}

func printBuildInfo() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}
