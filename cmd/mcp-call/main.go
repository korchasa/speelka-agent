// mcp-call is a console application designed for end-to-end testing of mcp (model context protocol) servers. It should test the most of all their API methods and capabilities.
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// ANSI color codes for terminal output
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[1;31m"
	colorGreen  = "\033[1;32m"
	colorBlue   = "\033[1;34m"
	colorYellow = "\033[1;33m"
)

// Flags stores command-line flags for the application
type Flags struct {
	toolName     string
	paramsJSON   string
	initTimeout  int
	toolsTimeout int
	callTimeout  int
	setLogLevel  string
}

func main() {
	// Parse flags and arguments
	flags, args, params, err := prepare()
	if err != nil {
		logErrorf("Error: %v", err)
		logInfof("Usage: mcp-call --tool toolName [--params '{...}'] [--init-timeout seconds] [--tools-timeout seconds] [--call-timeout seconds] [--set-log-level level] command [args...]")
		os.Exit(1)
	}

	// Create a context with timeout
	ctx := context.Background()

	// === MCP LOGGING: notifications/message handler ===
	receivedNotification := make([]mcp.JSONRPCNotification, 0)

	// Initialize the MCP server
	command := args[0]
	commandArgs := args[1:]
	mcpClient, initResult, err := initStdioServer(ctx, command, commandArgs, time.Duration(flags.initTimeout)*time.Second, &receivedNotification)
	if err != nil {
		logErrorf("Error initializing server: %v", err)
		os.Exit(1)
	}

	// Output initialization result
	logSuccessf("Initialization result:\n%s", jsonify(initResult))
	defer func(mcpClient *client.Client) {
		err := mcpClient.Close()
		if err != nil {
			logErrorf("Error closing MCP client: %v", err)
			os.Exit(1)
		}
	}(mcpClient)

	// List tools and validate the requested tool exists
	err = listTools(ctx, mcpClient, time.Duration(flags.toolsTimeout)*time.Second)
	if err != nil {
		logErrorf("Error listing tools: %v", err)
		os.Exit(1)
	}

	// If the --set-log-level flag is set, send MCP logging/setLevel
	if flags.setLogLevel != "" {
		err := setLogLevel(ctx, mcpClient, flags.setLogLevel)
		if err != nil {
			logErrorf("Failed to set log level: %v", err)
			os.Exit(1)
		}
		logSuccessf("Log level set to '%s' on server", flags.setLogLevel)
	}

	// Call the tool and display results
	err = callTool(ctx, mcpClient, flags.toolName, params, time.Duration(flags.callTimeout)*time.Second)
	if err != nil {
		logErrorf("%v", err)
		os.Exit(1)
	}

	if len(receivedNotification) == 0 {
		logWarnf("Notifications received: 0")
	} else {
		logInfof("Notifications received: %d", len(receivedNotification))
	}

	os.Exit(0)
}

// prepare parses command line flags and arguments
// It validates inputs and returns the flags, command arguments, and parsed parameters
func prepare() (Flags, []string, map[string]interface{}, error) {
	// Define command-line flags
	flags := Flags{}
	flag.StringVar(&flags.toolName, "tool", "", "Name of the tool to call")
	flag.StringVar(&flags.paramsJSON, "params", "", "JSON string of parameters to pass to the tool")
	flag.IntVar(&flags.initTimeout, "init-timeout", 5, "Timeout in seconds for server initialization")
	flag.IntVar(&flags.toolsTimeout, "tools-timeout", 5, "Timeout in seconds for listing tools")
	flag.IntVar(&flags.callTimeout, "call-timeout", 10, "Timeout in seconds for tool call execution")
	flag.StringVar(&flags.setLogLevel, "set-log-level", "", "Set log level on the server (debug, info, warning, error, critical, alert, emergency)")

	// Parse flags but keep the remaining arguments for the command to execute
	flag.Parse()
	args := flag.Args()
	// Remove logging prefixes
	log.SetFlags(0)

	if len(args) == 0 {
		return flags, nil, nil, errors.New("no command specified")
	}

	// Parse the parameters JSON
	var params map[string]interface{}
	if flags.paramsJSON != "" {
		if err := json.Unmarshal([]byte(flags.paramsJSON), &params); err != nil {
			return flags, nil, nil, fmt.Errorf("error parsing parameters JSON: %v", err)
		}
	}

	return flags, args, params, nil
}

// initStdioServer creates and initializes the MCP client
// It sets up stderr handling and connects to the MCP server
func initStdioServer(ctx context.Context, command string, commandArgs []string, timeout time.Duration, receivedNotification *[]mcp.JSONRPCNotification) (*client.Client, *mcp.InitializeResult, error) {
	// Create the MCP client
	logInfof("Run MCP-server : %s %v", command, commandArgs)
	mcpClient, err := client.NewStdioMCPClient(command, os.Environ(), commandArgs...)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating MCP client: %v", err)
	}

	// Set up a goroutine to print stderr from the subprocess (if available)
	if stderrReader, ok := client.GetStderr(mcpClient); ok {
		go func() {
			reader := bufio.NewReader(stderrReader)
			for {
				line, err := reader.ReadString('\n')
				if err != nil {
					if err == io.EOF {
						logErrorf("MCP-server has gone away (stderr closed)")
						os.Exit(1)
					}
					logErrorf("Error reading stderr: %v", err)
					return
				}
				logStderr(line)
			}
		}()
	}

	// === MCP LOGGING: notifications/message handler ===
	// MCP logs are now explicitly marked as [MCP-LOG] in the output for clarity.
	mcpClient.OnNotification(func(notification mcp.JSONRPCNotification) {
		*receivedNotification = append(*receivedNotification, notification)
		fmt.Printf("Received notification: %s\n", notification.Method)
		var logMsg mcp.LoggingMessageNotification
		params, err := json.Marshal(notification.Params)
		if err != nil {
			logErrorf("Failed to marshal notification params: %v", err)
			return
		}
		if err := json.Unmarshal(params, &logMsg.Params); err != nil {
			logErrorf("Failed to unmarshal LoggingMessageNotification.Params: %v", err)
			return
		}
		level := string(logMsg.Params.Level)
		color := colorBlue
		msg := ""
		if s, ok := logMsg.Params.Data.(string); ok {
			msg = s
		} else {
			b, _ := json.Marshal(logMsg.Params.Data)
			msg = string(b)
		}
		log.Printf(color+"[MCP-LOG %s] %s: %s"+colorReset, strings.ToUpper(level), logMsg.Params.Logger, msg)
	})

	// Initialize the client with a specific timeout
	fmt.Println("Initializing stdio client...")
	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "mcp-call Simple Client",
		Version: "1.0.0",
	}
	initRequest.Params.Capabilities = mcp.ClientCapabilities{}

	logInfof("Initializing MCP client with params: %s", jsonify(initRequest))
	initResult, err := mustRunWithTimeout(ctx, timeout, "Initialization", func(ctx context.Context) (*mcp.InitializeResult, error) {
		return mcpClient.Initialize(ctx, initRequest)
	})
	if err != nil {
		err := mcpClient.Close()
		if err != nil {
			return nil, nil, err
		}
		return nil, nil, err
	}

	logSuccessf("Connected to MCP server: %s", initResult.ServerInfo.Name)
	return mcpClient, initResult, nil
}

// listTools fetches the list of available tools from the MCP server
// It also validates if the specified tool exists
func listTools(ctx context.Context, client *client.Client, timeout time.Duration) error {
	// Get the list of available tools with timeout
	logInfof("Listing available tools...")
	toolsResult, err := mustRunWithTimeout(ctx, timeout, "List tools", func(ctx context.Context) (*mcp.ListToolsResult, error) {
		return client.ListTools(ctx, mcp.ListToolsRequest{})
	})
	if err != nil {
		return fmt.Errorf("error listing tools: %v", err)
	}

	// Display the list of available tools
	logSuccessf("Available tools:")
	for _, tool := range toolsResult.Tools {
		logInfof("  %s: %s", tool.Name, tool.Description)
	}
	logInfof("--------------------------------")

	return nil
}

// callTool executes the specified tool with given parameters
// It handles the tool call and displays the results
func callTool(ctx context.Context, client *client.Client, toolName string, params map[string]interface{}, timeout time.Duration) error {
	// Create a CallToolRequest
	callToolRequest := mcp.CallToolRequest{}
	callToolRequest.Params.Name = toolName
	callToolRequest.Params.Arguments = params

	// Call the tool with timeout
	logInfof("Calling tool '%s' with parameters: %s", toolName, jsonify(params))
	result, err := mustRunWithTimeout(ctx, timeout, "Tool call", func(ctx context.Context) (*mcp.CallToolResult, error) {
		return client.CallTool(ctx, callToolRequest)
	})
	if err != nil {
		return fmt.Errorf("error calling tool: %v", err)
	}

	// Print the result
	if result.IsError {
		logErrorf("Tool call resulted in an error:")
	} else {
		logSuccessf("Tool call completed successfully:")
	}

	for _, content := range result.Content {
		switch c := content.(type) {
		case mcp.TextContent:
			logInfof("%s", c.Text)
		case *mcp.TextContent:
			logInfof("%s", c.Text)
		case mcp.ImageContent:
			logInfof("[Image: %s, %d bytes]", c.MIMEType, len(c.Data))
		case *mcp.ImageContent:
			logInfof("[Image: %s, %d bytes]", c.MIMEType, len(c.Data))
		case mcp.EmbeddedResource:
			logInfof("[Embedded Resource]")
		case *mcp.EmbeddedResource:
			logInfof("[Embedded Resource]")
		default:
			logInfof("[Unknown content type: %T]", content)
		}
	}

	return nil
}

// mustRunWithTimeout executes the given operation with a timeout
// operation should be a function that takes a context and returns a result and an error
// returns the result of the operation or an error if the operation times out or fails
func mustRunWithTimeout[T any](ctx context.Context, duration time.Duration, operationName string, operation func(context.Context) (T, error)) (T, error) {
	var zero T
	timeoutCtx, cancel := context.WithTimeout(ctx, duration)
	defer cancel()

	type opResult struct {
		result T
		err    error
	}

	resultChan := make(chan opResult)

	go func() {
		result, err := operation(timeoutCtx)
		if err != nil {
			// Check for subprocess termination signals
			if errors.Is(err, io.EOF) ||
				strings.Contains(err.Error(), "broken pipe") ||
				strings.Contains(err.Error(), "closed pipe") {
				logErrorf("Subprocess terminated unexpectedly during %s", operationName)
				resultChan <- opResult{result, fmt.Errorf("subprocess terminated: %w", err)}
				return
			}
		}
		resultChan <- opResult{result, err}
	}()

	select {
	case <-timeoutCtx.Done():
		if errors.Is(timeoutCtx.Err(), context.DeadlineExceeded) {
			return zero, fmt.Errorf("%s timed out after %v", operationName, duration)
		}
		return zero, fmt.Errorf("%s failed: %v", operationName, timeoutCtx.Err())
	case res := <-resultChan:
		return res.result, res.err
	}
}

// logInfof logs a message in blue color (for neutral information)
func logInfof(format string, args ...interface{}) {
	log.Printf(colorBlue+format+colorReset, args...)
}

// logSuccessf logs a message in green color (for successful operations)
func logSuccessf(format string, args ...interface{}) {
	log.Printf(colorGreen+format+colorReset, args...)
}

// logErrorf logs a message in red color (for errors)
func logWarnf(format string, args ...interface{}) {
	log.Printf(colorYellow+"mcp-call error: "+format+colorReset, args...)
}

// logErrorf logs a message in red color (for errors)
func logErrorf(format string, args ...interface{}) {
	log.Printf(colorRed+"mcp-call error: "+format+colorReset, args...)
}

// logStderr logs a message in magenta color (for stderr output)
// Stderr logs are now explicitly marked as [STDERR] in the output for clarity.
func logStderr(line string) {
	log.Printf("%s[STDERR]%s %s", colorYellow, colorReset, line)
}

func jsonify(v interface{}) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		logErrorf("Error marshalling JSON: %v", err)
		os.Exit(1)
	}
	return string(b)
}

// setLogLevel sends MCP logging/setLevel to the server
// Send as a tool call
func setLogLevel(ctx context.Context, mcpClient *client.Client, level string) error {
	err := mcpClient.SetLevel(ctx, mcp.SetLevelRequest{
		Params: struct {
			Level mcp.LoggingLevel `json:"level"`
		}{
			Level: mcp.LoggingLevel(level),
		},
	})
	return err
}
