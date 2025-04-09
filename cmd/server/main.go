// Package main is the entry point for the MCP server
// Responsibility: Initialization and launch of all system components
// Features: Supports two operating modes - daemon (HTTP server) and stdin/stdout
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/korchasa/speelka-agent-go/internal/agent"
	"github.com/korchasa/speelka-agent-go/internal/configuration"
	"github.com/korchasa/speelka-agent-go/internal/utils"
	"github.com/sirupsen/logrus"
)

// Global logger instance
// Responsibility: Providing access to the logger from anywhere in the program
// Features: Initialized at the start and used throughout the application
var log *logrus.Logger

// Command line parameters
// Responsibility: Determine the server operating mode
// Features: When true, the server runs as an HTTP daemon; otherwise, as a stdio server
var (
	daemonMode = flag.Bool("daemon", false, "Run as a daemon with HTTP SSE MCP server (default: false, runs as stdio MCP server)")
)

// panicHandler intercepts panics and logs them with a full call stack
// Responsibility: Providing panic information for debugging
// Features: Captures the panic, logs it, and then continues the panic
func panicHandler() {
	if r := recover(); r != nil {
		stackTrace := debug.Stack()
		log.WithFields(logrus.Fields{
			"panic": r,
			"stack": string(stackTrace),
		}).Error("PANIC OCCURRED")

		// Continue the panic after logging
		panic(r)
	}
}

// main - application entry point
// Responsibility: Starting the server and handling termination
// Features: Sets up signal handling for graceful shutdown
func main() {
	fp, err := os.Create("app2.log")
	if err != nil {
		log.Fatalf("Failed to create log file: %v", err)
	}
	defer fp.Close()

	fmt.Fprintf(fp, "CONFIG_JSON: %s\n", os.Getenv("CONFIG_JSON"))
	// Parse command line parameters
	flag.Parse()

	// Create base context for the application
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling for graceful shutdown
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-signalCh
		fmt.Printf("Received signal: %s. Starting graceful shutdown...\n", sig)
		cancel()
	}()

	// Start the server and handle errors
	if err := run(ctx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// run contains the main application logic
// Responsibility: Initialization and launch of all system components
// Features: Sequentially initializes all necessary components
// and launches the server in the required mode
func run(ctx context.Context) error {
	// Initialize logger with default configuration
	log = logrus.New()

	if *daemonMode {
		log.SetLevel(logrus.DebugLevel)
		log.SetOutput(os.Stdout)
	} else {
		log.SetLevel(logrus.DebugLevel)
		log.SetOutput(os.Stderr)
	}
	log.SetReportCaller(true)
	log.SetFormatter(utils.NewCustomLogFormatter())
	log.Info("Logger initialized with default configuration")

	// Set up panic handler
	defer panicHandler()

	// Load configuration from environment variables
	configManager := configuration.NewConfigurationManager(log)
	err := configManager.LoadConfiguration(ctx)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	log.Info("Configuration loaded successfully from environment variables")

	// Reconfigure the logger with parameters from configuration
	logConfig := configManager.GetLogConfig()
	log.SetLevel(logConfig.Level)
	log.SetOutput(logConfig.Output)
	log.Info("Logger reconfigured with settings from configuration")

	// Initialize system components
	agentApp, err := agent.NewAgent(configManager, log)
	if err != nil {
		return fmt.Errorf("failed to create agent: %w", err)
	}
	err = agentApp.Start(*daemonMode, ctx)
	if err != nil {
		return fmt.Errorf("failed to start agent: %w", err)
	}
	return nil
}
