# Configuration Format Update

## Overview
The configuration format for MCP servers has been updated to match the standard MCP format. This change affects how server connections are defined in the configuration.

## Changes Made

### 1. From Array to Object
- **Before**: Servers were defined as an array under `connections.servers`
- **After**: Servers are now defined as an object map under `connections.mcpServers` where the key is the server ID

### 2. Removed Transport Field
- **Before**: Each server had a `transport` field to specify "http" or "stdio"
- **After**: Transport type is determined automatically based on the presence of `command` (stdio) or `URL` (http)

### 3. Renamed Arguments Field
- **Before**: Command line arguments were specified in an `arguments` array
- **After**: Command line arguments are now specified in an `args` array

### 4. Server ID as Key
- **Before**: Server ID was specified as a field inside each server object
- **After**: Server ID is now the key in the `mcpServers` object

## Example Before

```json
{
  "agent": {
    "connections": {
      "servers": [
        {
          "id": "time",
          "transport": "stdio",
          "command": "docker",
          "arguments": ["run", "-i", "--rm", "mcp/time"]
        }
      ]
    }
  }
}
```

## Example After

```json
{
  "agent": {
    "connections": {
      "mcpServers": {
        "time": {
          "command": "docker",
          "args": ["run", "-i", "--rm", "mcp/time"]
        }
      }
    }
  }
}
```

## Files Modified
- `internal/types/configuration_spec.go`: Updated the MCPConnectorConfig and MCPServerConnection structs
- `internal/configuration/manager.go`: Updated Configuration struct and loadFromJSON method
- `internal/mcp_connector/mcp_connector.go`: Updated to work with the new configuration structure
- `examples/architect.json` and `examples/simple.json`: Updated example configuration files
- `site/js/main.js` and `site/index.html`: Updated the website configuration generator

## Backward Compatibility
The system includes backward compatibility for loading old configuration formats, but all new configurations will be generated in the new format.