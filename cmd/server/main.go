// Package main is the entry point for the MCP server
// Responsibility: Initialization and launch of all system components
// Features: Supports two operating modes - daemon (HTTP server) and stdin/stdout
package main

import (
    "context"
    "encoding/json"
    "flag"
    "fmt"
    "github.com/korchasa/speelka-agent-go/internal/utils/log_formatter"
    "os"
    "os/signal"
    "runtime/debug"
    "syscall"

    "github.com/korchasa/speelka-agent-go/internal/application"

    "github.com/korchasa/speelka-agent-go/internal/configuration"
    "github.com/sirupsen/logrus"
)

// Command line parameters
// Responsibility: Determine the server operating mode
// Features: When true, the server runs as an HTTP daemon; otherwise, as a stdio server
var (
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

    startupLogger := logrus.New()
    startupLogger.SetLevel(logrus.WarnLevel)
    startupLogger.SetFormatter(log_formatter.NewCustomLogFormatter())
    conf, err := loadConfiguration(ctx)
    if err != nil {
        startupLogger.Fatalf("Failed to load configuration: %v", err)
    }
    // Get final log level and output from configuration
    logConfig, err := conf.BuildLogConfig()
    if err != nil {
        startupLogger.Fatalf("Invalid log config: %v", err)
    }

    // Global logger
    log := logrus.New()
    log.SetLevel(logConfig.Level)
    log.SetFormatter(logConfig.Formatter)

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

    app, err := application.NewMCPApp(log, conf)
    if err != nil {
        log.Fatalf("Failed to create application: %v", err)
    }
    err = app.Initialize(ctx)
    if err != nil {
        log.Fatalf("Failed to initialize application: %v", err)
    }
    if *callInput != "" {
        // Direct call mode
        log.Infof("Running in direct call mode with input: %s", *callInput)
        result, code, _ := app.ExecuteDirectCall(ctx, *callInput)
        _ = json.NewEncoder(os.Stdout).Encode(result)
        os.Exit(code)
        return
    } else {
        log.Infof("Running in MCP server mode")
        err = app.Start(ctx)
        if err != nil {
            log.Fatalf("Failed to start agent: %v", err)
        }
    }
}

func loadConfiguration(ctx context.Context) (*configuration.Configuration, error) {
    configManager := configuration.NewConfigurationManager()
    err := configManager.LoadConfiguration(ctx, *configFile)
    if err != nil {
        return nil, fmt.Errorf("failed to load configuration: %w", err)
    }
    if err := configManager.Validate(); err != nil {
        return nil, fmt.Errorf("configuration validation failed: %w", err)
    }
    conf := configManager.GetConfiguration()
    return conf, nil
}
