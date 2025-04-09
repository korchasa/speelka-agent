# Speelka Agent Configuration Tool

This directory contains a static HTML configuration tool for the Speelka Agent.

## Overview

The `config.html` file provides a user-friendly interface for configuring the Speelka Agent. It allows users to:

- Customize all aspects of the agent configuration
- Dynamically add/remove MCP server connections
- Generate a properly formatted JSON configuration
- Copy the configuration for use with the agent
- View instructions for running the agent with the generated configuration

## Usage

1. Open `config.html` in any modern web browser
2. Configure the agent settings through the tabbed interface
3. Click "Generate Configuration" to create the JSON configuration
4. Copy the configuration using the "Copy Configuration" button
5. Follow the provided instructions to run the agent with your configuration

## Features

- **Tabbed Interface**: Organizes settings into logical groups for easier navigation
- **Dynamic Server Management**: Add or remove MCP server connections as needed
- **JSON Generation**: Automatically formats your settings into valid JSON
- **Copy to Clipboard**: Easily copy the generated configuration
- **Usage Instructions**: Clear instructions for different deployment methods

## Technical Details

This is a purely client-side HTML page with JavaScript for functionality. No server-side processing is required, making it easy to deploy and use anywhere.

The page is designed to match the configuration structure described in the Speelka Agent documentation, ensuring that all generated configurations are valid and ready to use.

## Maintenance

When updating the Speelka Agent's configuration structure, make sure to update this configuration tool as well to maintain compatibility.