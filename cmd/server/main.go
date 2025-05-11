// Package main is the entry point for the MCP server
// Responsibility: Initialization and launch of all system components
// Features: Supports two operating modes - daemon (HTTP server) and stdin/stdout
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/korchasa/speelka-agent-go/internal/agent"
	"github.com/korchasa/speelka-agent-go/internal/configuration"
	"github.com/korchasa/speelka-agent-go/internal/logger"
	"github.com/sirupsen/logrus"
)

// Global logger instance
// Responsibility: Providing access to the logger from anywhere in the program
// Features: Initialized at the start and used throughout the application
var log *logger.Logger

// Command line parameters
// Responsibility: Determine the server operating mode
// Features: When true, the server runs as an HTTP daemon; otherwise, as a stdio server
var (
	daemonMode = flag.Bool("daemon", false, "Run as a daemon with HTTP SSE MCP server (default: false, runs as stdio MCP server)")
	configFile = flag.String("config", "", "Path to configuration file (YAML or JSON format)")
	callInput  = flag.String("call", "", "Run in direct call mode with the given user query (bypasses MCP server)")
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
	flag.Parse()

	log = logger.NewLogger()
	log.SetFormatter(logger.NewCustomLogFormatter())
	log.Info("Logger initialized")

	defer panicHandler()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if *callInput != "" {
		// Direct call mode
		configManager, err := loadConfiguration(ctx)
		if err != nil {
			outputDirectCallErrorAndExit("config", err)
		}
		app, err := agent.NewApp(log, configManager)
		if err != nil {
			outputDirectCallErrorAndExit("internal", err)
		}
		err = app.Initialize(ctx)
		if err != nil {
			outputDirectCallErrorAndExit("internal", err)
		}
		directApp := &agent.DirectApp{Agent: app.DirectAgent()}
		result := directApp.HandleCall(ctx, *callInput)
		if err := json.NewEncoder(os.Stdout).Encode(result); err != nil {
			fmt.Fprintf(os.Stderr, "{\"success\":false,\"result\":{},\"meta\":{},\"error\":{\"type\":\"internal\",\"message\":\"failed to encode result: %v\"}}\n", err)
			os.Exit(2)
		}
		if result.Success {
			os.Exit(0)
		}
		switch result.Error.Type {
		case "user", "config":
			os.Exit(1)
		default:
			os.Exit(2)
		}
	}

	// MCP server mode (default)
	// Set up signal handling for graceful shutdown
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-signalCh
		fmt.Printf("Received signal: %s. Starting graceful shutdown...\n", sig)
		cancel()
	}()

	configManager, err := loadConfiguration(ctx)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Start the server and handle errors
	if err := run(ctx, configManager); err != nil {
		log.Fatalf("Main application failed: %v", err)
	}
}

func loadConfiguration(ctx context.Context) (*configuration.Manager, error) {
	// Create configuration manager
	configManager := configuration.NewConfigurationManager(log)

	// Load configuration from file if specified, then from environment variables
	// and validate the configuration
	err := configManager.LoadConfiguration(ctx, *configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	return configManager, nil
}

// run contains the main application logic
// Responsibility: Initialization and launch of all system components
// Features: Sequentially initializes all necessary components
// and launches the server in the required mode
func run(ctx context.Context, configManager *configuration.Manager) error {

	log.SetLevel(configManager.GetLogConfig().Level)
	log.SetOutput(configManager.GetLogConfig().Output)

	// Create the app with the logger and initialized configuration manager
	app, err := agent.NewApp(log, configManager)
	if err != nil {
		return fmt.Errorf("failed to create agent app: %w", err)
	}

	// Initialize the app (creates the Agent and its dependencies)
	err = app.Initialize(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize agent app: %w", err)
	}

	// Start the agent
	err = app.Start(*daemonMode, ctx)
	if err != nil {
		return fmt.Errorf("failed to start agent: %w", err)
	}

	return nil
}

// Helper to output error in direct call mode and exit
func outputDirectCallErrorAndExit(errType string, err error) {
	result := agent.DirectCallResult{
		Success: false,
		Result:  map[string]any{"answer": ""},
		Meta:    agent.MetaInfo{},
		Error:   agent.DirectCallError{Type: errType, Message: err.Error()},
	}
	output, _ := json.Marshal(result)
	fmt.Println(string(output))
	if errType == "user" || errType == "config" {
		os.Exit(1)
	}
	os.Exit(2)
}
