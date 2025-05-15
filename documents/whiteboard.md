// 2024-06-15
// Changes:
// - Extended LogConfig: support for output (stdout, stderr, file, MCP), format (custom, json, text, unknown), dynamic level
// - Added constants LogOutputStdout, LogOutputStderr, LogOutputMCP
// - Updated interfaces: LoggerSpec, MCPServerNotifier
// - Added and extended tests: BuildLogConfig, GetAgentConfig, GetLLMConfig, GetMCPServerConfig, GetMCPConnectorConfig
// - Golden serialization and property-based overlay tests for configuration
