## Investigate and Fix: text-extractor Timeout Not Respected

### Problem
- The timeout for `text-extractor` (set to 300s in `site/examples/ai-news.yaml`) is not respected in `internal/mcp_connector/mcp_connector.go`.
- Actual error: tool `extract-text` execution timed out after 30 seconds (should be 300).

### Findings
- The timeout is set correctly in YAML and parsed into the config struct.
- The `Apply` method merges the timeout field as expected.
- **BUG:** In `internal/configuration/manager.go`, `GetMCPConnectorConfig` did **not** copy the `Timeout` field from the loaded config to the connector config. As a result, the connector always used the fallback (30s).

### Solution
- Added a unit test to confirm the bug: the timeout from YAML was not respected in the connector config.
- Fixed `GetMCPConnectorConfig` to copy the `Timeout` field for each server.
- Removed duplicate test definitions and ensured the test runs with the correct build tag (`unit`).
- **Result:** The test now passes, confirming the fix. The timeout from YAML is now respected by the connector and tool execution.

---

# Whiteboard

## [DONE] Trim stderr lines before logging in MCP connector
- Location: `internal/mcp_connector/mcp_connector.go`, ConnectServer
- All stderr lines from MCP child processes are now trimmed of trailing newlines and whitespace before logging.
- Added a test (`Test_StderrLoggingTrimsNewlines`) in `internal/mcp_connector/mcp_connector_test.go` to verify trimming logic.
- All tests pass.

---

# Log Level Configuration Issue: Root Cause and Plan

## Problem
- The logger is always initialized with logrus.DebugLevel by default in internal/logger/logger.go.
- The configuration system correctly parses and stores the log level from the config file (e.g., warn in site/examples/text-extractor.yaml).
- However, after loading the configuration, the logger's log level is never set to match the configuration. The logger remains at DebugLevel unless changed via the MCP tool or programmatically.
- As a result, info/debug logs are always shown, even if the config requests warn or error.

## Root Cause
- The logger is created before configuration is loaded, and its level is not updated after config is loaded.
- There is no code that sets logger.SetLevel(configManager.GetLogConfig().Level) after loading config.

## Solution
- After configuration is loaded and applied (in loadConfiguration in cmd/server/main.go), update the logger to use the log level from configManager.GetLogConfig().Level.
- This ensures the logger respects the config file's log level.

## Plan
1. Update cmd/server/main.go to set the logger's level from the loaded configuration after config is loaded.
2. Add a test to ensure that the logger respects the config log level.
3. Document the change here and in implementation docs if needed.

---

## [DONE] Per-server MCP Timeout Support and Logger Config Fix

### Summary of Changes (2024-06)
- **Per-server MCP timeout:**
  - Each MCP server in config can now specify a `timeout` (seconds, float or int). If not set, defaults to 30s.
  - Timeout is now respected throughout: loaded from YAML/JSON, merged in `Apply`, copied in `GetMCPConnectorConfig`, and enforced in `MCPConnector.ExecuteTool`.
  - Manual timeout logic replaces context.WithTimeout for better control and logging.
  - Enhanced logging for tool execution, including timeout/cancellation details.
  - Comprehensive tests added for timeout propagation and enforcement.
- **Logger respects config log level:**
  - After config is loaded, logger's level is set to match config (`logger.SetLevel(configManager.GetLogConfig().Level)`).
  - Added test to ensure logger respects config log level.
- **File removals/cleanup:**
  - Deleted: `internal/app/direct_app_test.go`, `internal/app/direct_types.go`, `internal/app/util.go`, `site/examples/ai-news-subagent-extractor.yaml` (obsolete, replaced by `text-extractor.yaml`).
  - Tests and code referencing these files removed or updated.
- **Config/test updates:**
  - Example configs updated to use new timeout and server structure.
  - Tests for YAML/JSON loader extended to cover timeout field.

---

# Whiteboard

## [DONE] Per-server timeout and logger config fix
- All planned changes for timeout and logger config are complete and tested.
- Documentation updated in architecture, implementation, and file structure docs.

## Next Steps
- [ ] Review for any remaining references to deleted files or old config fields.
- [ ] Monitor for regressions in tool call timeout or logging behavior.
- [ ] Plan next feature or refactor as needed.

---

## [DONE] Logger JSON Output and Configurable Format
- Added `format` (RawFormat) to logging config (YAML/JSON/env).
- Default: `text`. Supports `json`.
- Main uses config to select formatter (`logrus.JSONFormatter` or custom).
- All config loaders and Apply logic updated.
- Tests for default, env, YAML, JSON config, and logger output.
- All tests pass.

---
