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
    mcplogger "github.com/korchasa/speelka-agent-go/internal/logger"
    "github.com/sirupsen/logrus"
)

// Global logger instance
// Responsibility: Providing access to the logger from anywhere in the program
// Features: Initialized at the start and used throughout the application
var log *mcplogger.Logger

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
    // Parse command line parameters
    flag.Parse()

    // Create MCPLogger with internal configuration
    log = mcplogger.NewLogger()
    log.SetFormatter(mcplogger.NewCustomLogFormatter())
    log.Info("Logger initialized")

    // Set up panic handler
    defer panicHandler()

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
        _, _ = fmt.Fprintf(os.Stdout, "FATAL ERROR: %v\n", err)
        log.Fatalf("Main application failed: %v", err)
    }
}

// run contains the main application logic
// Responsibility: Initialization and launch of all system components
// Features: Sequentially initializes all necessary components
// and launches the server in the required mode
func run(ctx context.Context) error {
    // Create the app with the logger
    app, err := agent.NewApp(log)
    if err != nil {
        return fmt.Errorf("failed to create agent app: %w", err)
    }

    // Load configuration from environment variables
    err = app.LoadConfiguration(ctx)
    if err != nil {
        log.Fatalf("Failed to load configuration: %v", err)
    }

    // Initialize the app (creates the Agent and its dependencies)
    err = app.Initialize()
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
