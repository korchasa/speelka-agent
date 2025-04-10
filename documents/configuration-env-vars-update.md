# Environment Variables Configuration Update

## Overview

The configuration format for Speelka Agent has been updated to use direct environment variables instead of the previous JSON-based approach. This change makes the configuration more transparent, easier to manage, and integrates better with container environments.

## Changes Made

### 1. From Single JSON to Individual Variables
- **Before**: Configuration was loaded from a single `CONFIG_JSON` environment variable
- **After**: Configuration is loaded from multiple individual environment variables

### 2. Variable Naming Conventions
- Agent settings: `AGENT_NAME`, `AGENT_VERSION`
- Tool settings: `TOOL_NAME`, `TOOL_DESCRIPTION`, etc.
- LLM settings: `LLM_PROVIDER`, `LLM_MODEL`, etc.
- MCP Server settings: Indexed variables like `MCPS_0_ID`, `MCPS_0_COMMAND`, etc.
- Runtime settings: `RUNTIME_LOG_LEVEL`, `RUNTIME_STDIO_ENABLED`, etc.

### 3. Array Handling with Indexed Variables
- MCP Servers are now configured using indexed environment variables
- Each server has a set of variables prefixed with `MCPS_<index>_`

### 4. Better Validation and Error Reporting
- Required variables are explicitly validated
- Clear error messages for missing required variables
- Type conversion is handled automatically with sensible defaults

## Example Before

```bash
CONFIG_JSON='{"agent":{"name":"speelka-agent","version":"1.0.0","tool":{"name":"process","description":"Process tool for handling user queries with LLM","argument_name":"input","argument_description":"User query to process"},"llm":{"provider":"openai","api_key":"your_api_key_here","model":"gpt-4o","max_tokens":0,"temperature":0.7,"prompt_template":"You are a helpful AI assistant...","retry":{"max_retries":3,"initial_backoff":1.0,"max_backoff":30.0,"backoff_multiplier":2.0}},"connections":{"mcpServers":{"time":{"command":"docker","args":["run","-i","--rm","mcp/time"]}},"retry":{"max_retries":3,"initial_backoff":1.0,"max_backoff":30.0,"backoff_multiplier":2.0}}},"runtime":{"log":{"level":"info","output":"stdout"},"transports":{"stdio":{"enabled":true,"buffer_size":8192},"http":{"enabled":true,"host":"localhost","port":3000}}}}'
```

## Example After

```bash
# Agent
export AGENT_NAME="architect-speelka-agent"
export AGENT_VERSION="1.0.0"

# Tool
export TOOL_NAME="architect"
export TOOL_DESCRIPTION="Architecture design and assessment tool for software systems"
export TOOL_ARGUMENT_NAME="query"
export TOOL_ARGUMENT_DESCRIPTION="Architecture query or task to process"

# LLM
export LLM_PROVIDER="openai"
export LLM_API_KEY="your_api_key_here"
export LLM_MODEL="gpt-4o"
export LLM_MAX_TOKENS=0
export LLM_TEMPERATURE=0.2
export LLM_PROMPT_TEMPLATE="# ROLE\nYou are a Senior Software Architect..."

# LLM Retry Config
export LLM_RETRY_MAX_RETRIES=3
export LLM_RETRY_INITIAL_BACKOFF=1.0
export LLM_RETRY_MAX_BACKOFF=30.0
export LLM_RETRY_BACKOFF_MULTIPLIER=2.0

# MCP Servers
export MCPS_0_ID="time"
export MCPS_0_COMMAND="docker"
export MCPS_0_ARGS="run -i --rm mcp/time"

export MCPS_1_ID="mcp-filesystem-server"
export MCPS_1_COMMAND="mcp-filesystem-server"
export MCPS_1_ARGS="."

# MSPS Retry Configuration
export MSPS_RETRY_MAX_RETRIES=3
export MSPS_RETRY_INITIAL_BACKOFF=1.0
export MSPS_RETRY_MAX_BACKOFF=30.0
export MSPS_RETRY_BACKOFF_MULTIPLIER=2.0

# Runtime Configuration
export RUNTIME_LOG_LEVEL="debug"
export RUNTIME_LOG_OUTPUT="./architect.log"
export RUNTIME_STDIO_ENABLED=true
export RUNTIME_STDIO_BUFFER_SIZE=8192
export RUNTIME_HTTP_ENABLED=false
export RUNTIME_HTTP_HOST="localhost"
export RUNTIME_HTTP_PORT=3000
```

## Files Modified

- `internal/configuration/manager.go`: Updated to load configuration from environment variables
- `internal/configuration/manager_test.go`: Added tests for environment variable configuration
- `documents/implementation.md`: Updated configuration examples
- `documents/configuration-env-vars-update.md`: Added new documentation for this change
- `examples/architect.env` and `examples/simple.env`: Added new example configuration files
- `README.md`: Updated with the new configuration approach
- `Dockerfile`: Removed CONFIG_JSON environment variable

## Conversion Tool

A conversion script is provided to help migrate existing JSON configurations to the new environment variables format:

```bash
# Convert a JSON configuration file to environment variables
./scripts/json_to_env.sh examples/architect.json examples/converted.env

# Load the converted environment variables
source examples/converted.env
```

The script requires `jq` to be installed on your system. It will extract all configuration values from the JSON file and generate the equivalent environment variables.

## Backward Compatibility

The system still includes the ability to load the old configuration format for backward compatibility, but all new deployments should use the environment variables approach. The JSON-based configuration method is now considered deprecated and will be removed in a future version.

## Required Environment Variables

The following environment variables are required for proper operation:

- `AGENT_NAME`: The name of the agent
- `TOOL_NAME`: The name of the tool provided by the agent
- `TOOL_DESCRIPTION`: Description of the tool functionality
- `LLM_PROVIDER`: The provider of the LLM service (e.g., "openai")
- `LLM_MODEL`: The specific model to use (e.g., "gpt-4o")
- `LLM_PROMPT_TEMPLATE`: The template to use for system prompts (must include `{{query}}` and `{{tools}}` placeholders)

## MCP Servers Configuration

MCP servers are configured using indexed environment variables in the format:

```
MCPS_<index>_ID="server-id"
MCPS_<index>_COMMAND="command"
MCPS_<index>_ARGS="arg1 arg2 arg3"
```

The `<index>` should start from 0 and increment for each server. The `ID` field is used as the key in the map of MCP servers.