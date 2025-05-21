// Package mcp_connector: logging routing logic for MCP clients
package mcp_connector

import (
	"bufio"
	"encoding/json"
	"github.com/korchasa/speelka-agent-go/internal/configuration"
	"strings"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// setupLoggingRoute configures logging routing for a server depending on capabilities and transport.
func (mc *MCPConnector) setupLoggingRoute(serverID string, mcpClient client.MCPClient, capabilities mcp.ServerCapabilities, serverConfig configuration.MCPServerConnection) {
	if capabilities.Logging != nil {
		// MCP logging via notifications/message
		if stdioClient, ok := mcpClient.(*client.Client); ok {
			stdioClient.OnNotification(func(notification mcp.JSONRPCNotification) {
				mc.log.Debugf("[MCP-LOG] notification: %s", notification.Method)
				if notification.Method != "notifications/message" {
					return
				}
				var logMsg mcp.LoggingMessageNotification
				params, err := json.Marshal(notification.Params)
				if err != nil {
					return
				}
				if err := json.Unmarshal(params, &logMsg.Params); err != nil {
					return
				}
				level := string(logMsg.Params.Level)
				msg := ""
				if s, ok := logMsg.Params.Data.(string); ok {
					msg = s
				} else {
					b, _ := json.Marshal(logMsg.Params.Data)
					msg = string(b)
				}
				switch level {
				case "debug":
					mc.log.Debugf("[MCP %s] %s", level, msg)
				case "info", "notice":
					mc.log.Infof("[MCP %s] %s", level, msg)
				case "warning":
					mc.log.Warnf("[MCP %s] %s", level, msg)
				case "error", "critical", "alert", "emergency":
					mc.log.Errorf("[MCP %s] %s", level, msg)
				default:
					mc.log.Infof("[MCP %s] %s", level, msg)
				}
			})
		}
	} else if serverConfig.Command != "" {
		// Fallback: read stderr of child process (only for stdio)
		if stdioClient, ok := mcpClient.(*client.Client); ok {
			if stderr, ok := client.GetStderr(stdioClient); ok && stderr != nil {
				go func() {
					scanner := bufio.NewScanner(stderr)
					for scanner.Scan() {
						line := scanner.Text()
						trimmed := strings.TrimRight(line, "\r\n \t")
						if trimmed != "" {
							mc.log.Infof("`%s` stderr: %s", serverID, trimmed)
						}
					}
				}()
			}
		}
	} else if serverConfig.URL != "" {
		// For HTTP fallback is not possible
		mc.log.Infof("[MCP-CONNECT] Fallback to stderr is not possible for HTTP transport (server '%s')", serverID)
	}
}
