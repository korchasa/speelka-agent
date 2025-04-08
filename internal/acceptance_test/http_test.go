package acceptance_test

import (
	"testing"
)

// TestHTTPRequest tests that the server can be started and that it responds correctly to HTTP requests
func TestHTTPRequest(t *testing.T) {
	t.Skip("Skipping HTTP test, because it's not implemented in mcp-go yet")
	// Generate a random port number in the dynamic/private port range
	// rand.Seed(time.Now().UnixNano())
	// httpPort := 49152 + rand.Intn(16383) // Random port between 49152 and 65535

	// // Constants for the test
	// const httpHost = "localhost"
	// httpURL := fmt.Sprintf("http://localhost:%d", httpPort)
	// const timeoutSeconds = 30
	// const maxRetries = 5

	// t.Logf("Using random port %d for HTTP server", httpPort)

	// // Create a configuration object directly in the test
	// config := map[string]interface{}{
	// 	"server": map[string]interface{}{
	// 		"name":    "simple-speelka-agent",
	// 		"version": "1.0.0",
	// 		"tool": map[string]interface{}{
	// 			"name":                 "process",
	// 			"description":          "Process tool for handling user queries with LLM",
	// 			"argument_name":        "input",
	// 			"argument_description": "User query to process",
	// 		},
	// 		"http": map[string]interface{}{
	// 			"enabled": true,
	// 			"host":    httpHost,
	// 			"port":    httpPort,
	// 		},
	// 		"stdio": map[string]interface{}{
	// 			"enabled":     false,
	// 			"buffer_size": 8192,
	// 			"auto_detect": false,
	// 		},
	// 		"debug": false,
	// 	},
	// 	"mcp_connector": map[string]interface{}{
	// 		"servers": []map[string]interface{}{
	// 			{
	// 				"id":        "time",
	// 				"transport": "stdio",
	// 				"command":   "docker",
	// 				"arguments": []string{"run", "-i", "--rm", "mcp/time"},
	// 			},
	// 			{
	// 				"id":        "mcp-filesystem-server",
	// 				"transport": "stdio",
	// 				"command":   "mcp-filesystem-server",
	// 				"arguments": []string{"/Users/korchasa/www/speelka/speelka-agent"},
	// 			},
	// 		},
	// 		"retry": map[string]interface{}{
	// 			"max_retries":        3,
	// 			"initial_backoff":    1.0,
	// 			"max_backoff":        30.0,
	// 			"backoff_multiplier": 2.0,
	// 		},
	// 	},
	// 	"llm": map[string]interface{}{
	// 		"provider":        "openai",
	// 		"api_key":         "will be replaced",
	// 		"model":           "gpt-4o",
	// 		"max_tokens":      0,
	// 		"temperature":     nil,
	// 		"prompt_template": "You are a helpful AI assistant. Respond to the following request:\n\n{{input}}\n\nProvide a detailed and helpful response.\n\nAvailable tools:\n{{tools}}",
	// 		"retry": map[string]interface{}{
	// 			"max_retries":        3,
	// 			"initial_backoff":    1.0,
	// 			"max_backoff":        30.0,
	// 			"backoff_multiplier": 2.0,
	// 		},
	// 	},
	// 	"log": map[string]interface{}{
	// 		"level":  "info",
	// 		"output": "./app.log",
	// 	},
	// }

	// // Read LLM_API_KEY from environment or .env file
	// apiKey := os.Getenv("LLM_API_KEY")
	// if apiKey == "" {
	// 	// Try to find the .env file
	// 	envPaths := []string{
	// 		".env",                                   // Current directory
	// 		"../../.env",                             // Project root
	// 		"../../../.env",                          // One level up from project root
	// 		filepath.Join(os.Getenv("HOME"), ".env"), // Home directory
	// 	}

	// 	for _, envPath := range envPaths {
	// 		if envData, err := os.ReadFile(envPath); err == nil {
	// 			for _, line := range strings.Split(string(envData), "\n") {
	// 				if strings.HasPrefix(line, "LLM_API_KEY=") {
	// 					apiKey = strings.TrimPrefix(line, "LLM_API_KEY=")
	// 					t.Logf("Found LLM_API_KEY in %s", envPath)
	// 					break
	// 				}
	// 			}
	// 			if apiKey != "" {
	// 				break
	// 			}
	// 		}
	// 	}
	// }
	// require.NotEmpty(t, apiKey, "LLM_API_KEY not found in environment or .env files")

	// // Update LLM config with the API key
	// llmConfig := config["llm"].(map[string]interface{})
	// llmConfig["api_key"] = apiKey

	// // Convert the configuration to JSON
	// configJSON, err := json.Marshal(config)
	// require.NoError(t, err, "Failed to marshal configuration")

	// // Print the configuration for debugging (excluding API key)
	// debugConfig := make(map[string]interface{})
	// json.Unmarshal(configJSON, &debugConfig)
	// if llm, ok := debugConfig["llm"].(map[string]interface{}); ok {
	// 	llm["api_key"] = "[REDACTED]"
	// }
	// t.Logf("Using configuration: %v", debugConfig)

	// // Setup logger
	// logger := logrus.New()
	// logger.SetLevel(logrus.InfoLevel)

	// // Create a context with timeout for controlling the test
	// ctx, cancel := context.WithTimeout(context.Background(), timeoutSeconds*time.Second)
	// defer cancel()

	// // Create a configuration manager
	// os.Setenv("CONFIG_JSON", string(configJSON))
	// configManager := configuration.NewConfigurationManager(logger)
	// err = configManager.LoadConfiguration(ctx)
	// require.NoError(t, err, "Failed to load configuration")

	// // Create and start the agent
	// agentApp, err := agent.NewAgent(configManager, logger)
	// require.NoError(t, err, "Failed to create agent")

	// // Start the agent in a goroutine
	// t.Log("Starting agent with HTTP server...")
	// go func() {
	// 	err := agentApp.Start(true, ctx) // true for daemon mode (HTTP server)
	// 	if err != nil && err.Error() != "context canceled" {
	// 		t.Errorf("Agent failed: %v", err)
	// 	}
	// }()

	// // Make sure to clean up the agent when the test finishes
	// defer func() {
	// 	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	// 	defer shutdownCancel()

	// 	if err := agentApp.Stop(shutdownCtx); err != nil {
	// 		t.Logf("Error stopping agent: %v", err)
	// 	}
	// }()

	// // Wait for the server to start
	// t.Log("Waiting for server to start...")
	// time.Sleep(5 * time.Second)

	// // Test the HTTP connection with retries
	// var response *http.Response
	// var responseBody []byte
	// var responseError error

	// url := fmt.Sprintf("%s/message", httpURL)
	// t.Logf("Testing HTTP connection to %s", url)

	// // Prepare a simple MCP request to the process tool
	// requestBody := map[string]interface{}{
	// 	"method": "tools/call",
	// 	"params": map[string]interface{}{
	// 		"name": "process",
	// 		"arguments": map[string]interface{}{
	// 			"input": "What time is it now?",
	// 		},
	// 	},
	// }
	// requestJSON, err := json.Marshal(requestBody)
	// require.NoError(t, err, "Failed to marshal request body")

	// // Try to connect with retries
	// for retry := 0; retry < maxRetries; retry++ {
	// 	response, responseError = http.Post(
	// 		url,
	// 		"application/json",
	// 		strings.NewReader(string(requestJSON)),
	// 	)

	// 	if responseError == nil && response.StatusCode == http.StatusOK {
	// 		// Connection successful
	// 		defer response.Body.Close()
	// 		responseBody, err = io.ReadAll(response.Body)
	// 		require.NoError(t, err, "Failed to read response body")
	// 		break
	// 	}

	// 	if responseError != nil {
	// 		t.Logf("Connection attempt %d failed: %v. Retrying in 2 seconds...", retry+1, responseError)
	// 	} else {
	// 		t.Logf("Connection attempt %d failed with status code %d. Retrying in 2 seconds...",
	// 			retry+1, response.StatusCode)
	// 		response.Body.Close()
	// 	}

	// 	time.Sleep(2 * time.Second)
	// }

	// // Assert that we got a successful response
	// require.NoError(t, responseError, "Failed to connect to server after multiple attempts")
	// require.Equal(t, http.StatusOK, response.StatusCode, "HTTP status code should be 200 OK")
	// require.NotEmpty(t, responseBody, "Response body should not be empty")

	// t.Logf("Received response: %s", string(responseBody))

	// // Verify the response structure (basic checks only)
	// var responseJSON map[string]interface{}
	// err = json.Unmarshal(responseBody, &responseJSON)
	// require.NoError(t, err, "Failed to parse response JSON")

	// // Check that we got a response and not an error
	// assert.False(t, responseJSON["isError"].(bool), "Response should not be an error")
}
