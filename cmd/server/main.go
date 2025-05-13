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

	"github.com/korchasa/speelka-agent-go/internal/app_direct"
	"github.com/korchasa/speelka-agent-go/internal/app_mcp"
	"github.com/korchasa/speelka-agent-go/internal/configuration"
	"github.com/korchasa/speelka-agent-go/internal/logger"
	"github.com/korchasa/speelka-agent-go/internal/types"
	"github.com/sirupsen/logrus"
)

// Global logger instance
// Responsibility: Providing access to the logger from anywhere in the program
// Features: Initialized at the start and used throughout the application
var log types.LoggerSpec

// Command line parameters
// Responsibility: Determine the server operating mode
// Features: When true, the server runs as an HTTP daemon; otherwise, as a stdio server
var (
	daemonMode = flag.Bool("daemon", false, "Run as a daemon with HTTP SSE MCP server (default: false, runs as stdio MCP server)")
	configFile = flag.String("config", "", "Path to configuration file (YAML or JSON format)")
	callInput  = flag.String("call", "", "Run in direct call mode with the given user query (bypasses MCP server)")
)

// initPanicHandler intercepts panics and logs them with a full call stack
// Responsibility: Providing panic information for debugging
// Features: Captures the panic, logs it, and then continues the panic
func initPanicHandler() {
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

	// Инициализация логгера с дефолтным конфигом (до загрузки конфига)
	logIface := app_mcp.NewLogger(types.LogConfig{
		Level:     logrus.ErrorLevel,
		RawOutput: "stderr",
		RawFormat: "text",
		Output:    nil,
	})
	log = logIface
	log.SetFormatter(logger.NewCustomLogFormatter())

	defer initPanicHandler()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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

	// Пересоздаём логгер с актуальным конфигом
	logConfig := configManager.GetLogConfig()
	logIface = app_mcp.NewLogger(logConfig)
	log = logIface
	// Set formatter based on configuration
	if logConfig.RawFormat == "json" {
		log.SetFormatter(&logrus.JSONFormatter{})
	} else {
		log.SetFormatter(logger.NewCustomLogFormatter())
	}
	log.SetLevel(logConfig.Level)
	log.Infof("Logger level set to %s and format set to %s from configuration", logConfig.Level.String(), logConfig.RawFormat)

	var application app_mcp.Application
	if *callInput != "" {
		// Direct call mode
		application = app_direct.NewDirectApp(log, configManager)
		// For CLI mode, execute and exit
		application.(*app_direct.DirectApp).Execute(ctx, *callInput)
		return
	} else {
		// MCP server mode
		var err error
		application, err = app_mcp.NewApp(log, configManager)
		if err != nil {
			log.Fatalf("Failed to create agent app: %v", err)
		}
		// Initialize the app (creates the Agent and its dependencies)
		err = application.Initialize(ctx)
		if err != nil {
			log.Fatalf("Failed to initialize agent app: %v", err)
		}
		// Start the agent
		err = application.Start(*daemonMode, ctx)
		if err != nil {
			log.Fatalf("Failed to start agent: %v", err)
		}
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
