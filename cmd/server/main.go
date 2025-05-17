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

	"github.com/korchasa/speelka-agent-go/internal/app_direct"
	"github.com/korchasa/speelka-agent-go/internal/app_mcp"
	"github.com/korchasa/speelka-agent-go/internal/configuration"
	"github.com/korchasa/speelka-agent-go/internal/logger"
	"github.com/korchasa/speelka-agent-go/internal/types"
	"github.com/sirupsen/logrus"
)

// Command line parameters
// Responsibility: Determine the server operating mode
// Features: When true, the server runs as an HTTP daemon; otherwise, as a stdio server
var (
	daemonMode = flag.Bool("daemon", false, "Run as a daemon with HTTP SSE MCP server (default: false, runs as stdio MCP server)")
	configFile = flag.String("config", "", "Path to configuration file (YAML or JSON format)")
	callInput  = flag.String("call", "", "Run in direct call mode with the given user query (bypasses MCP server)")
)

// main - application entry point
// Responsibility: Starting the server and handling termination
// Features: Sets up signal handling for graceful shutdown
func main() {
	flag.Parse()

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

	startupLogger := logger.NewLogger(types.LogConfig{
		DefaultLevel: "warn",
		Format:       "text",
		Level:        logrus.WarnLevel,
		DisableMCP:   true,
	})
	startupLogger.SetFormatter(logger.NewCustomLogFormatter())
	configManager, err := loadConfiguration(ctx, startupLogger)
	if err != nil {
		startupLogger.Fatalf("Failed to load configuration: %v", err)
	}

	// Get final log level and output from configuration
	conf := configManager.GetConfiguration()
	logConfig, err := types.BuildLogConfig(conf.Runtime.Log)
	if err != nil {
		startupLogger.Fatalf("Invalid log config: %v", err)
	}
	level := logConfig.Level
	// Global logger
	log := logger.NewLogger(logConfig)
	log.SetLevel(level)
	log.SetFormatter(logger.NewCustomLogFormatter())

	// panic handler now uses the global logger
	defer func() {
		if r := recover(); r != nil {
			stackTrace := debug.Stack()
			log.WithFields(logrus.Fields{
				"panic": r,
				"stack": string(stackTrace),
			}).Error("PANIC OCCURRED")
			panic(r)
		}
	}()

	var application app_mcp.Application
	if *callInput != "" {
		// Direct call mode
		application := app_direct.NewDirectApp(log, configManager)
		result, code, _ := application.Execute(ctx, *callInput)
		_ = json.NewEncoder(os.Stdout).Encode(result)
		os.Exit(code)
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

func loadConfiguration(ctx context.Context, log types.LoggerSpec) (*configuration.Manager, error) {
	configManager := configuration.NewConfigurationManager(log)
	err := configManager.LoadConfiguration(ctx, *configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}
	return configManager, nil
}
