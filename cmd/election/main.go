package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/commands"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		sig := <-sigs
		log.Printf("Received signal: %s, initiating shutdown...\n", sig)
		cancel()
	}()

	// Initialize and run the command
	rootCmd, err := commands.InitRunCommand(ctx)
	if err != nil {
		log.Printf("init run command: %v\n", err)
		os.Exit(1)
	}
	err = rootCmd.Execute()
	if err != nil {
		log.Printf("run command: %v\n", err)
		os.Exit(1)
	}
}
