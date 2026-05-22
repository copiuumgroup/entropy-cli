package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/copiuumgroup/entropy-cli/internal/api"
	"github.com/copiuumgroup/entropy-cli/internal/config"
	"github.com/copiuumgroup/entropy-cli/internal/daemon"
	"github.com/copiuumgroup/entropy-cli/internal/database"
)

func main() {
	port := flag.Int("port", 8080, "API server port")
	workers := flag.Int("workers", 3, "Number of download workers")
	flag.Parse()

	// Load configuration
	if err := config.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not load config: %v\n", err)
	}

	// Initialize database
	if err := database.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to initialize database: %v\n", err)
		os.Exit(1)
	}

	// Create daemon
	d := daemon.NewDaemon(*workers)
	if err := d.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to start daemon: %v\n", err)
		os.Exit(1)
	}

	// Create API server
	addr := fmt.Sprintf(":%d", *port)
	server := api.NewServer(addr)

	// Start server in a goroutine
	go func() {
		if err := server.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "error: server failed: %v\n", err)
		}
	}()

	// Wait for interrupt signal
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch

	fmt.Println("\nShutting down...")
	d.Stop()
	fmt.Println("Goodbye!")
}
