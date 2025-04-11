#!/bin/bash
#
# Script to convert Speelka Agent JSON configuration to environment variables
#
# Usage: json_to_env.sh <input_json_file> <output_env_file>
#
# Example: ./json_to_env.sh examples/architect.json examples/converted.env

set -e

if [ "$#" -ne 2 ]; then
    echo "Usage: $0 <input_json_file> <output_env_file>"
    exit 1
fi

INPUT_FILE="$1"
OUTPUT_FILE="$2"

if [ ! -f "$INPUT_FILE" ]; then
    echo "Error: Input file '$INPUT_FILE' not found."
    exit 1
fi

# Check if jq is installed
if ! command -v jq &> /dev/null; then
    echo "Error: jq is required but not installed. Please install jq."
    exit 1
fi

# Create or truncate output file
> "$OUTPUT_FILE"

echo "# Generated from $INPUT_FILE on $(date)" >> "$OUTPUT_FILE"
echo "" >> "$OUTPUT_FILE"

# Agent section
echo "# Agent" >> "$OUTPUT_FILE"
echo "export SPL_AGENT_NAME=\"$(jq -r '.agent.name // ""' "$INPUT_FILE")\"" >> "$OUTPUT_FILE"
echo "export SPL_AGENT_VERSION=\"$(jq -r '.agent.version // "1.0.0"' "$INPUT_FILE")\"" >> "$OUTPUT_FILE"
echo "" >> "$OUTPUT_FILE"

# Tool section
echo "# Tool" >> "$OUTPUT_FILE"
echo "export SPL_TOOL_NAME=\"$(jq -r '.agent.tool.name // ""' "$INPUT_FILE")\"" >> "$OUTPUT_FILE"
echo "export SPL_TOOL_DESCRIPTION=\"$(jq -r '.agent.tool.description // ""' "$INPUT_FILE")\"" >> "$OUTPUT_FILE"
echo "export SPL_TOOL_ARGUMENT_NAME=\"$(jq -r '.agent.tool.argument_name // "query"' "$INPUT_FILE")\"" >> "$OUTPUT_FILE"
echo "export SPL_TOOL_ARGUMENT_DESCRIPTION=\"$(jq -r '.agent.tool.argument_description // ""' "$INPUT_FILE")\"" >> "$OUTPUT_FILE"
echo "" >> "$OUTPUT_FILE"

# LLM section
echo "# LLM" >> "$OUTPUT_FILE"
echo "export SPL_LLM_PROVIDER=\"$(jq -r '.agent.llm.provider // ""' "$INPUT_FILE")\"" >> "$OUTPUT_FILE"
echo "export SPL_LLM_API_KEY=\"$(jq -r '.agent.llm.api_key // "no value"' "$INPUT_FILE")\"" >> "$OUTPUT_FILE"
echo "export SPL_LLM_MODEL=\"$(jq -r '.agent.llm.model // ""' "$INPUT_FILE")\"" >> "$OUTPUT_FILE"
echo "export SPL_LLM_MAX_TOKENS=$(jq -r '.agent.llm.max_tokens // 0' "$INPUT_FILE")" >> "$OUTPUT_FILE"
echo "export SPL_LLM_TEMPERATURE=$(jq -r '.agent.llm.temperature // 0.7' "$INPUT_FILE")" >> "$OUTPUT_FILE"

# Handle prompt template with potential newlines
PROMPT_TEMPLATE=$(jq -r '.agent.llm.prompt_template // ""' "$INPUT_FILE")
echo "export SPL_LLM_PROMPT_TEMPLATE=\"$PROMPT_TEMPLATE\"" >> "$OUTPUT_FILE"
echo "" >> "$OUTPUT_FILE"

# LLM Retry section
echo "# LLM Retry" >> "$OUTPUT_FILE"
echo "export SPL_LLM_RETRY_MAX_RETRIES=$(jq -r '.agent.llm.retry.max_retries // 3' "$INPUT_FILE")" >> "$OUTPUT_FILE"
echo "export SPL_LLM_RETRY_INITIAL_BACKOFF=$(jq -r '.agent.llm.retry.initial_backoff // 1.0' "$INPUT_FILE")" >> "$OUTPUT_FILE"
echo "export SPL_LLM_RETRY_MAX_BACKOFF=$(jq -r '.agent.llm.retry.max_backoff // 30.0' "$INPUT_FILE")" >> "$OUTPUT_FILE"
echo "export SPL_LLM_RETRY_BACKOFF_MULTIPLIER=$(jq -r '.agent.llm.retry.backoff_multiplier // 2.0' "$INPUT_FILE")" >> "$OUTPUT_FILE"
echo "" >> "$OUTPUT_FILE"

# MCP Servers section
echo "# MCP Servers" >> "$OUTPUT_FILE"
index=0
jq -r '.agent.connections.mcpServers | keys[]' "$INPUT_FILE" | while read -r server_id; do
    echo "export SPL_MCPS_${index}_ID=\"$server_id\"" >> "$OUTPUT_FILE"

    # Command
    command=$(jq -r ".agent.connections.mcpServers[\"$server_id\"].command // \"\"" "$INPUT_FILE")
    echo "export SPL_MCPS_${index}_COMMAND=\"$command\"" >> "$OUTPUT_FILE"

    # Args - join with spaces
    args=$(jq -r ".agent.connections.mcpServers[\"$server_id\"].args | join(\" \") // \"\"" "$INPUT_FILE")
    echo "export SPL_MCPS_${index}_ARGS=\"$args\"" >> "$OUTPUT_FILE"

    # Environment variables
    jq -r ".agent.connections.mcpServers[\"$server_id\"].environment | keys[]" "$INPUT_FILE" 2>/dev/null | while read -r env_key; do
        env_value=$(jq -r ".agent.connections.mcpServers[\"$server_id\"].environment[\"$env_key\"]" "$INPUT_FILE")
        echo "export SPL_MCPS_${index}_ENV_${env_key}=\"$env_value\"" >> "$OUTPUT_FILE"
    done

    echo "" >> "$OUTPUT_FILE"
    index=$((index + 1))
done

# MSPS Retry section
echo "# MSPS Retry" >> "$OUTPUT_FILE"
echo "export SPL_MSPS_RETRY_MAX_RETRIES=$(jq -r '.agent.connections.retry.max_retries // 3' "$INPUT_FILE")" >> "$OUTPUT_FILE"
echo "export SPL_MSPS_RETRY_INITIAL_BACKOFF=$(jq -r '.agent.connections.retry.initial_backoff // 1.0' "$INPUT_FILE")" >> "$OUTPUT_FILE"
echo "export SPL_MSPS_RETRY_MAX_BACKOFF=$(jq -r '.agent.connections.retry.max_backoff // 30.0' "$INPUT_FILE")" >> "$OUTPUT_FILE"
echo "export SPL_MSPS_RETRY_BACKOFF_MULTIPLIER=$(jq -r '.agent.connections.retry.backoff_multiplier // 2.0' "$INPUT_FILE")" >> "$OUTPUT_FILE"
echo "" >> "$OUTPUT_FILE"

# Runtime section
echo "# Runtime" >> "$OUTPUT_FILE"
echo "export SPL_RUNTIME_LOG_LEVEL=\"$(jq -r '.runtime.log.level // "info"' "$INPUT_FILE")\"" >> "$OUTPUT_FILE"
echo "export SPL_RUNTIME_LOG_OUTPUT=\"$(jq -r '.runtime.log.output // "stdout"' "$INPUT_FILE")\"" >> "$OUTPUT_FILE"
echo "" >> "$OUTPUT_FILE"

# Transport - Stdio section
echo "# Transport - Stdio" >> "$OUTPUT_FILE"
echo "export SPL_RUNTIME_STDIO_ENABLED=$(jq -r '.runtime.transports.stdio.enabled // true' "$INPUT_FILE")" >> "$OUTPUT_FILE"
echo "export SPL_RUNTIME_STDIO_BUFFER_SIZE=$(jq -r '.runtime.transports.stdio.buffer_size // 8192' "$INPUT_FILE")" >> "$OUTPUT_FILE"
echo "" >> "$OUTPUT_FILE"

# Transport - HTTP section
if jq -e '.runtime.transports.http' "$INPUT_FILE" > /dev/null; then
    echo "# Transport - HTTP" >> "$OUTPUT_FILE"
    echo "export SPL_RUNTIME_HTTP_ENABLED=$(jq -r '.runtime.transports.http.enabled // false' "$INPUT_FILE")" >> "$OUTPUT_FILE"
    echo "export SPL_RUNTIME_HTTP_HOST=\"$(jq -r '.runtime.transports.http.host // "localhost"' "$INPUT_FILE")\"" >> "$OUTPUT_FILE"
    echo "export SPL_RUNTIME_HTTP_PORT=$(jq -r '.runtime.transports.http.port // 3000' "$INPUT_FILE")" >> "$OUTPUT_FILE"
fi

echo "Conversion complete. Environment variables written to $OUTPUT_FILE"
echo "To load these variables, run: source $OUTPUT_FILE"