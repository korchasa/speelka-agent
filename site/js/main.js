// Mobile Menu Toggle
document.addEventListener('DOMContentLoaded', function() {
    const hamburger = document.querySelector('.hamburger');
    const navLinks = document.querySelector('.nav-links');

    hamburger.addEventListener('click', function() {
        const isExpanded = this.getAttribute('aria-expanded') === 'true';

        navLinks.classList.toggle('active');
        hamburger.classList.toggle('active');

        // Update ARIA attributes
        this.setAttribute('aria-expanded', !isExpanded);
    });

    // Close menu when clicking on a link
    const links = document.querySelectorAll('.nav-links a');
    links.forEach(link => {
        link.addEventListener('click', function() {
            navLinks.classList.remove('active');
            hamburger.classList.remove('active');
            hamburger.setAttribute('aria-expanded', 'false');
        });
    });

    // Code block copy buttons
    const copyButtons = document.querySelectorAll('.copy-btn');
    copyButtons.forEach(button => {
        button.addEventListener('click', function() {
            const codeBlock = this.nextElementSibling;
            const textToCopy = codeBlock.textContent;

            copyToClipboard(textToCopy, function() {
                showCopySuccess();
            });
        });
    });

    // Auto-generate configuration when form inputs change
    const formInputs = document.querySelectorAll('.form-group input, .form-group select, .form-group textarea');
    formInputs.forEach(input => {
        input.addEventListener('change', function() {
            generateAndUpdateConfig();
        });

        // For text inputs and textareas, also listen for keyup events
        if (input.tagName === 'INPUT' && (input.type === 'text' || input.type === 'number') || input.tagName === 'TEXTAREA') {
            input.addEventListener('keyup', function() {
                generateAndUpdateConfig();
            });
        }
    });

    // HTTP settings toggle in configuration section
    const httpEnabledSelect = document.getElementById('httpEnabled');
    if (httpEnabledSelect) {
        httpEnabledSelect.addEventListener('change', function() {
            const httpSettings = document.getElementById('httpSettings');
            if (httpSettings) {
                httpSettings.style.display = this.value === 'true' ? 'block' : 'none';
            }
        });
    }

    // Initialize Mermaid for diagrams
    if (typeof mermaid !== 'undefined') {
        mermaid.initialize({
            startOnLoad: true,
            theme: 'dark',
            themeVariables: {
                primaryColor: '#8a5cf6',
                primaryTextColor: '#f9fafb',
                primaryBorderColor: '#6d28d9',
                lineColor: '#8a5cf6',
                secondaryColor: '#4c1d95',
                tertiaryColor: '#131525'
            },
            flowchart: {
                useMaxWidth: true,
                htmlLabels: true,
                curve: 'basis'
            },
            sequence: {
                useMaxWidth: true,
                mirrorActors: false,
                actorMargin: 150,
                messageMargin: 75
            }
        });
    }

    // Activate default tabs
    const firstTab = document.querySelector('.tab-container .tab');
    const firstTabContent = document.getElementById('agentTab');
    if (firstTab && firstTabContent) {
        firstTab.classList.add('active');
        firstTabContent.classList.add('active');
    }

    // Initialize the configuration tool
    initConfigTool();
});

// Smooth scrolling for anchor links
document.querySelectorAll('a[href^="#"]').forEach(anchor => {
    anchor.addEventListener('click', function(e) {
        e.preventDefault();

        const targetId = this.getAttribute('href');
        if (targetId === '#') return;

        const targetElement = document.querySelector(targetId);
        if (targetElement) {
            window.scrollTo({
                top: targetElement.offsetTop - 80,
                behavior: 'smooth'
            });
        }
    });
});

// Tabs functionality
function openTab(evt, tabName) {
    const tabContents = document.getElementsByClassName('tab-content');
    for (let i = 0; i < tabContents.length; i++) {
        tabContents[i].classList.remove('active');
    }

    const tabs = document.getElementsByClassName('tab');
    for (let i = 0; i < tabs.length; i++) {
        tabs[i].classList.remove('active');
    }

    document.getElementById(tabName).classList.add('active');
    evt.currentTarget.classList.add('active');

    // Update configuration when tabs are switched
    // This ensures any changes made in the newly activated tab are reflected
    setTimeout(generateAndUpdateConfig, 50);
}

// Initialize the configuration tool
let serverCounter = 0;

function initConfigTool() {
    // Only initialize if we're on the page with the config tool
    if (!document.getElementById('serversContainer')) return;

    // Add a default server
    addServer();

    // Generate initial configuration
    // We use setTimeout to ensure DOM is fully ready after adding the server
    setTimeout(generateAndUpdateConfig, 100);
}

// Add a new server configuration
function addServer() {
    const serversContainer = document.getElementById('serversContainer');
    if (!serversContainer) return;

    const serverId = serverCounter++;

    const serverDiv = document.createElement('div');
    serverDiv.className = 'server-container';
    serverDiv.id = `server-${serverId}`;

    serverDiv.innerHTML = `
        <div class="form-group">
            <label for="serverId-${serverId}">Server ID:</label>
            <input type="text" id="serverId-${serverId}" value="server-${serverId}" />
        </div>

        <div class="form-group">
            <label for="serverCommand-${serverId}">Command:</label>
            <input type="text" id="serverCommand-${serverId}" value="docker" />
        </div>

        <div class="form-group">
            <label for="serverArgs-${serverId}">Arguments:</label>
            <input type="text" id="serverArgs-${serverId}" value="run, -i, --rm, mcp/time" placeholder="Comma-separated list" />
        </div>

        <div class="form-group">
            <label for="serverEnv-${serverId}">Environment:</label>
            <input type="text" id="serverEnv-${serverId}" value="NODE_ENV=production" placeholder="KEY=VALUE format, comma-separated" />
        </div>

        <button class="remove-server-btn" onclick="removeServer('server-${serverId}')">
            <i class="fas fa-trash"></i> Remove Server
        </button>
    `;

    serversContainer.appendChild(serverDiv);

    // Add event listeners to the new server's inputs
    const serverInputs = serverDiv.querySelectorAll('input, select');
    serverInputs.forEach(input => {
        input.addEventListener('change', function() {
            generateAndUpdateConfig();
        });

        if (input.tagName === 'INPUT' && (input.type === 'text' || input.type === 'number')) {
            input.addEventListener('keyup', function() {
                generateAndUpdateConfig();
            });
        }
    });

    // Generate updated configuration after adding server
    generateAndUpdateConfig();
}

// Remove a server configuration
function removeServer(serverId) {
    const serverDiv = document.getElementById(serverId);
    if (serverDiv) {
        serverDiv.remove();
        // Generate updated configuration after removing server
        generateAndUpdateConfig();
    }
}

// Automatically generate configuration and update examples
function generateAndUpdateConfig() {
    // Only proceed if we're on the page with the configuration tool
    if (!document.getElementById('generatedConfig')) return;

    // Generate the configuration
    const config = generateConfigObject();
    if (!config) return; // If validation failed

    // Update all code examples that need the configuration
    updateExampleConfigurations(config);
}

// Generate configuration object from form inputs
function generateConfigObject() {
    // Form validation
    const numericFields = [
        { id: 'llmMaxTokens', min: 0, name: 'Max Tokens' },
        { id: 'llmTemperature', min: 0, max: 1, name: 'Temperature' },
        { id: 'maxRetries', min: 1, name: 'Max Retries' },
        { id: 'initialBackoff', min: 0.1, name: 'Initial Backoff' },
        { id: 'maxBackoff', min: 1, name: 'Max Backoff' },
        { id: 'backoffMultiplier', min: 1, name: 'Backoff Multiplier' },
        { id: 'connMaxRetries', min: 1, name: 'Connection Max Retries' },
        { id: 'connInitialBackoff', min: 0.1, name: 'Connection Initial Backoff' },
        { id: 'connMaxBackoff', min: 1, name: 'Connection Max Backoff' },
        { id: 'connBackoffMultiplier', min: 1, name: 'Connection Backoff Multiplier' },
        { id: 'stdioBufferSize', min: 1024, name: 'STDIO Buffer Size' },
        { id: 'httpPort', min: 1, max: 65535, name: 'HTTP Port' }
    ];

    // Validate all numeric fields
    for (const field of numericFields) {
        const element = document.getElementById(field.id);
        if (!element) continue;

        const value = parseFloat(element.value);

        if (isNaN(value)) {
            console.error(`${field.name} must be a valid number.`);
            return null;
        }

        if (field.min !== undefined && value < field.min) {
            console.error(`${field.name} must be at least ${field.min}.`);
            return null;
        }

        if (field.max !== undefined && value > field.max) {
            console.error(`${field.name} must be at most ${field.max}.`);
            return null;
        }
    }

    // Agent section
    const agentName = document.getElementById('agentName').value;
    const agentVersion = document.getElementById('agentVersion').value;
    const toolName = document.getElementById('toolName').value;
    const toolDescription = document.getElementById('toolDescription').value;
    const toolArgumentName = document.getElementById('toolArgumentName').value;
    const toolArgumentDescription = document.getElementById('toolArgumentDescription').value;

    // LLM section
    const llmProvider = document.getElementById('llmProvider').value;
    const llmAPIKey = document.getElementById('llmAPIKey').value;
    const llmModel = document.getElementById('llmModel').value;
    const llmMaxTokens = parseInt(document.getElementById('llmMaxTokens').value);
    const llmTemperature = parseFloat(document.getElementById('llmTemperature').value);
    const llmPromptTemplate = document.getElementById('llmPromptTemplate').value;
    const llmRetryMaxRetries = parseInt(document.getElementById('llmRetryMaxRetries').value);
    const llmRetryInitialBackoff = parseFloat(document.getElementById('llmRetryInitialBackoff').value);
    const llmRetryMaxBackoff = parseFloat(document.getElementById('llmRetryMaxBackoff').value);
    const llmRetryBackoffMultiplier = parseFloat(document.getElementById('llmRetryBackoffMultiplier').value);

    // Connections section
    const mcpServers = {};
    const serverDivs = document.querySelectorAll('.server-container');
    serverDivs.forEach(div => {
        const id = div.id;
        const serverId = document.getElementById(`serverId-${id.split('-')[1]}`).value;
        const command = document.getElementById(`serverCommand-${id.split('-')[1]}`).value;
        const argsStr = document.getElementById(`serverArgs-${id.split('-')[1]}`).value;
        const envStr = document.getElementById(`serverEnv-${id.split('-')[1]}`).value;

        // Parse arguments
        const args = argsStr.split(',').map(arg => arg.trim());

        // Parse environment
        const envPairs = envStr.split(',').map(pair => pair.trim());
        const env = {};
        envPairs.forEach(pair => {
            if (pair.includes('=')) {
                const [key, value] = pair.split('=');
                env[key.trim()] = value.trim();
            }
        });

        mcpServers[serverId] = {
            command,
            args,
            environment: env
        };
    });

    const connRetryMaxRetries = parseInt(document.getElementById('connRetryMaxRetries').value);
    const connRetryInitialBackoff = parseFloat(document.getElementById('connRetryInitialBackoff').value);
    const connRetryMaxBackoff = parseFloat(document.getElementById('connRetryMaxBackoff').value);
    const connRetryBackoffMultiplier = parseFloat(document.getElementById('connRetryBackoffMultiplier').value);

    // Runtime section
    const logLevel = document.getElementById('logLevel').value;
    const logOutput = document.getElementById('logOutput').value;
    const stdioEnabled = document.getElementById('stdioEnabled').value === 'true';
    const stdioBufferSize = parseInt(document.getElementById('stdioBufferSize').value);
    const httpEnabled = document.getElementById('httpEnabled').value === 'true';
    const httpHost = document.getElementById('httpHost').value;
    const httpPort = parseInt(document.getElementById('httpPort').value);

    // Create the configuration object
    return {
        agent: {
            name: agentName,
            version: agentVersion,
            tool: {
                name: toolName,
                description: toolDescription,
                argument_name: toolArgumentName,
                argument_description: toolArgumentDescription
            },
            llm: {
                provider: llmProvider,
                api_key: llmAPIKey || "YOUR_API_KEY_HERE",
                model: llmModel,
                max_tokens: llmMaxTokens,
                temperature: llmTemperature,
                prompt_template: llmPromptTemplate,
                retry: {
                    max_retries: llmRetryMaxRetries,
                    initial_backoff: llmRetryInitialBackoff,
                    max_backoff: llmRetryMaxBackoff,
                    backoff_multiplier: llmRetryBackoffMultiplier
                }
            },
            connections: {
                mcpServers: mcpServers,
                retry: {
                    max_retries: connRetryMaxRetries,
                    initial_backoff: connRetryInitialBackoff,
                    max_backoff: connRetryMaxBackoff,
                    backoff_multiplier: connRetryBackoffMultiplier
                }
            }
        },
        runtime: {
            log: {
                level: logLevel,
                output: logOutput
            },
            transports: {
                stdio: {
                    enabled: stdioEnabled,
                    buffer_size: stdioBufferSize
                },
                http: {
                    enabled: httpEnabled,
                    host: httpHost,
                    port: httpPort
                }
            }
        }
    };
}

// Update all code examples in the instructions section
function updateExampleConfigurations(config) {
    // Create compact configuration JSON without indentation or line breaks
    const compactConfigJson = JSON.stringify(config);
    const escapedCompactConfigJson = compactConfigJson.replace(/'/g, "\\'");

    // For display in the generated config section, keep the formatted version
    const prettyConfigJson = JSON.stringify(config, null, 2);
    document.getElementById('generatedConfig').textContent = prettyConfigJson;

    // Generate environment variables from the configuration
    let envVars = [];

    // Agent settings
    envVars.push(`# Agent`);
    envVars.push(`export AGENT_NAME="${config.agent.name}"`);
    envVars.push(`export AGENT_VERSION="${config.agent.version}"`);

    // Tool settings
    envVars.push(`\n# Tool`);
    envVars.push(`export TOOL_NAME="${config.agent.tool.name}"`);
    envVars.push(`export TOOL_DESCRIPTION="${config.agent.tool.description}"`);
    envVars.push(`export TOOL_ARGUMENT_NAME="${config.agent.tool.argument_name}"`);
    envVars.push(`export TOOL_ARGUMENT_DESCRIPTION="${config.agent.tool.argument_description}"`);

    // LLM settings
    envVars.push(`\n# LLM`);
    envVars.push(`export LLM_PROVIDER="${config.agent.llm.provider}"`);
    envVars.push(`export LLM_API_KEY="..."`);
    envVars.push(`export LLM_MODEL="${config.agent.llm.model}"`);
    envVars.push(`export LLM_MAX_TOKENS=${config.agent.llm.max_tokens}`);
    envVars.push(`export LLM_TEMPERATURE=${config.agent.llm.temperature}`);
    const promptTemplate = config.agent.llm.prompt_template.replace(/"/g, '\\"');
    envVars.push(`export LLM_PROMPT_TEMPLATE="${promptTemplate}"`);

    // LLM Retry settings
    envVars.push(`\n# LLM Retry`);
    envVars.push(`export LLM_RETRY_MAX_RETRIES=${config.agent.llm.retry.max_retries}`);
    envVars.push(`export LLM_RETRY_INITIAL_BACKOFF=${config.agent.llm.retry.initial_backoff}`);
    envVars.push(`export LLM_RETRY_MAX_BACKOFF=${config.agent.llm.retry.max_backoff}`);
    envVars.push(`export LLM_RETRY_BACKOFF_MULTIPLIER=${config.agent.llm.retry.backoff_multiplier}`);

    // MCP Servers
    envVars.push(`\n# MCP Servers`);
    let serverIndex = 0;
    for (const [serverId, serverConfig] of Object.entries(config.agent.connections.mcpServers)) {
        envVars.push(`export MCPS_${serverIndex}_ID="${serverId}"`);
        envVars.push(`export MCPS_${serverIndex}_COMMAND="${serverConfig.command}"`);
        envVars.push(`export MCPS_${serverIndex}_ARGS="${serverConfig.args.join(' ')}"`);

        // Environment variables if any
        if (serverConfig.environment && Object.keys(serverConfig.environment).length > 0) {
            for (const [envKey, envValue] of Object.entries(serverConfig.environment)) {
                envVars.push(`export MCPS_${serverIndex}_ENV_${envKey}="${envValue}"`);
            }
        }

        envVars.push(``);
        serverIndex++;
    }

    // Connection retry settings
    envVars.push(`# MSPS Retry`);
    envVars.push(`export MSPS_RETRY_MAX_RETRIES=${config.agent.connections.retry.max_retries}`);
    envVars.push(`export MSPS_RETRY_INITIAL_BACKOFF=${config.agent.connections.retry.initial_backoff}`);
    envVars.push(`export MSPS_RETRY_MAX_BACKOFF=${config.agent.connections.retry.max_backoff}`);
    envVars.push(`export MSPS_RETRY_BACKOFF_MULTIPLIER=${config.agent.connections.retry.backoff_multiplier}`);

    // Runtime settings
    envVars.push(`\n# Runtime`);
    envVars.push(`export RUNTIME_LOG_LEVEL="${config.runtime.log.level}"`);
    envVars.push(`export RUNTIME_LOG_OUTPUT="${config.runtime.log.output}"`);

    // Transport settings
    envVars.push(`\n# Transport - Stdio`);
    envVars.push(`export RUNTIME_STDIO_ENABLED=${config.runtime.transports.stdio.enabled}`);
    envVars.push(`export RUNTIME_STDIO_BUFFER_SIZE=${config.runtime.transports.stdio.buffer_size}`);

    if (config.runtime.transports.http) {
        envVars.push(`\n# Transport - HTTP`);
        envVars.push(`export RUNTIME_HTTP_ENABLED=${config.runtime.transports.http.enabled}`);
        envVars.push(`export RUNTIME_HTTP_HOST="${config.runtime.transports.http.host}"`);
        envVars.push(`export RUNTIME_HTTP_PORT=${config.runtime.transports.http.port}`);
    }

    // Join all environment variables
    const envVarsText = envVars.join('\n');

    // Update environment variables example
    const envExample = document.querySelector('.instructions pre.code-block:nth-of-type(1) code');
    if (envExample) {
        envExample.textContent = envVarsText;
    }

    // Update binary example with environment variables
    const binaryExample = document.querySelector('.instructions pre.code-block:nth-of-type(2) code');
    if (binaryExample) {
        let binaryJson = JSON.stringify({
            mcpServers: {
                "speelka-agent": {
                    command: "speelka-agent",
                    args: [],
                    environment: {
                        AGENT_NAME: config.agent.name,
                        AGENT_VERSION: config.agent.version,
                        TOOL_NAME: config.agent.tool.name,
                        TOOL_DESCRIPTION: config.agent.tool.description,
                        TOOL_ARGUMENT_NAME: config.agent.tool.argument_name,
                        TOOL_ARGUMENT_DESCRIPTION: config.agent.tool.argument_description,
                        LLM_PROVIDER: config.agent.llm.provider,
                        LLM_API_KEY: "...YOUR_LLM_API_KEY...",
                        LLM_MODEL: config.agent.llm.model,
                        LLM_MAX_TOKENS: config.agent.llm.max_tokens,
                        LLM_TEMPERATURE: config.agent.llm.temperature,
                        RUNTIME_LOG_LEVEL: config.runtime.log.level
                    }
                }
            }
        }, null, 4);
        binaryExample.textContent = binaryJson;
    }

    // Update docker example with environment variables
    const dockerExample = document.querySelector('.instructions pre.code-block:nth-of-type(3) code');
    if (dockerExample) {
        // Create environment arguments list for Docker
        const dockerEnvArgs = [
            "run", "-i", "--rm",
            "-e", "AGENT_NAME=" + config.agent.name,
            "-e", "AGENT_VERSION=" + config.agent.version,
            "-e", "TOOL_NAME=" + config.agent.tool.name,
            "-e", "TOOL_DESCRIPTION=" + config.agent.tool.description,
            "-e", "TOOL_ARGUMENT_NAME=" + config.agent.tool.argument_name,
            "-e", "TOOL_ARGUMENT_DESCRIPTION=" + config.agent.tool.argument_description,
            "-e", "LLM_PROVIDER=" + config.agent.llm.provider,
            "-e", "LLM_API_KEY=$LLM_API_KEY",
            "-e", "LLM_MODEL=" + config.agent.llm.model,
            "-e", "RUNTIME_LOG_LEVEL=" + config.runtime.log.level,
            "ghcr.io/korchasa/speelka-agent:latest"
        ];

        let dockerJson = JSON.stringify({
            mcpServers: {
                "speelka-agent": {
                    command: "docker",
                    args: dockerEnvArgs,
                    environment: {
                        LLM_API_KEY: "...YOUR_LLM_API_KEY..."
                    }
                }
            }
        }, null, 4);
        dockerExample.textContent = dockerJson;
    }
}

// Generate the configuration JSON for manual button click
function generateConfig() {
    generateAndUpdateConfig();

    // Scroll to the generated configuration
    document.getElementById('generatedConfig').scrollIntoView({ behavior: 'smooth', block: 'center' });
}

// Copy configuration to clipboard
function copyConfig() {
    const configText = document.getElementById('generatedConfig').textContent;

    // Use the shared copy function
    copyToClipboard(configText, function() {
        showCopySuccess();
    });
}

// Copy text to clipboard and execute callback on success
function copyToClipboard(text, callback) {
    if (navigator.clipboard) {
        navigator.clipboard.writeText(text)
            .then(() => {
                if (callback) callback();
            })
            .catch(err => {
                console.error('Failed to copy text: ', err);
            });
    } else {
        // Fallback for browsers that don't support clipboard API
        const textArea = document.createElement('textarea');
        textArea.value = text;
        document.body.appendChild(textArea);
        textArea.select();

        try {
            const successful = document.execCommand('copy');
            if (successful && callback) callback();
        } catch (err) {
            console.error('Failed to copy text: ', err);
        }

        document.body.removeChild(textArea);
    }
}

// Show copy success message
function showCopySuccess() {
    // Check if message already exists
    let successMsg = document.querySelector('.copy-success');

    if (!successMsg) {
        // Create new message
        successMsg = document.createElement('div');
        successMsg.className = 'copy-success';
        successMsg.textContent = 'Copied to clipboard!';
        document.body.appendChild(successMsg);
    } else {
        // Reset animation by removing and re-adding the element
        successMsg.remove();
        successMsg = document.createElement('div');
        successMsg.className = 'copy-success';
        successMsg.textContent = 'Copied to clipboard!';
        document.body.appendChild(successMsg);
    }

    // Remove after animation completes
    setTimeout(() => {
        if (successMsg.parentNode) {
            successMsg.parentNode.removeChild(successMsg);
        }
    }, 2000);
}

// Download configuration as JSON file
function downloadConfig() {
    const config = generateConfigObject();
    if (!config) {
        showMessage('Error: Could not generate configuration', 'error');
        return;
    }

    const configJson = JSON.stringify(config, null, 2);
    const blob = new Blob([configJson], { type: 'application/json' });
    const url = URL.createObjectURL(blob);

    const a = document.createElement('a');
    a.href = url;
    a.download = 'speelka-agent-config.json';
    document.body.appendChild(a);
    a.click();

    setTimeout(() => {
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
        showMessage('Configuration downloaded successfully', 'success');
    }, 100);
}

// Upload configuration from JSON file
function uploadConfig(input) {
    if (!input.files || input.files.length === 0) return;

    const file = input.files[0];
    const reader = new FileReader();

    reader.onload = function(e) {
        try {
            const config = JSON.parse(e.target.result);

            // Validate the uploaded configuration
            if (!validateConfig(config)) {
                showMessage('Invalid configuration format', 'error');
                return;
            }

            // Apply the configuration to the form
            applyConfigToForm(config);

            // Update the examples
            updateExampleConfigurations(config);

            // Display success message
            showMessage('Configuration loaded successfully', 'success');
        } catch (error) {
            console.error('Error parsing JSON:', error);
            showMessage('Error parsing JSON file', 'error');
        }
    };

    reader.onerror = function() {
        showMessage('Error reading file', 'error');
    };

    reader.readAsText(file);

    // Reset the input to allow uploading the same file again
    input.value = '';
}

// Validate uploaded configuration
function validateConfig(config) {
    // Check basic structure
    if (!config || typeof config !== 'object') return false;
    if (!config.agent || typeof config.agent !== 'object') return false;
    if (!config.runtime || typeof config.runtime !== 'object') return false;

    return true;
}

// Apply configuration to form inputs
function applyConfigToForm(config) {
    try {
        // Agent settings
        if (config.agent) {
            setValue('agentName', config.agent.name);
            setValue('agentVersion', config.agent.version);

            // Tool settings
            if (config.agent.tool) {
                setValue('toolName', config.agent.tool.name);
                setValue('toolDescription', config.agent.tool.description);
                setValue('toolArgumentName', config.agent.tool.argument_name);
                setValue('toolArgumentDescription', config.agent.tool.argument_description);
            }

            // LLM settings
            if (config.agent.llm) {
                setValue('llmProvider', config.agent.llm.provider);
                setValue('llmAPIKey', config.agent.llm.api_key);
                setValue('llmModel', config.agent.llm.model);
                setValue('llmMaxTokens', config.agent.llm.max_tokens);
                setValue('llmTemperature', config.agent.llm.temperature);
                setValue('llmPromptTemplate', config.agent.llm.prompt_template);

                // Retry settings
                if (config.agent.llm.retry) {
                    setValue('llmRetryMaxRetries', config.agent.llm.retry.max_retries);
                    setValue('llmRetryInitialBackoff', config.agent.llm.retry.initial_backoff);
                    setValue('llmRetryMaxBackoff', config.agent.llm.retry.max_backoff);
                    setValue('llmRetryBackoffMultiplier', config.agent.llm.retry.backoff_multiplier);
                }
            }

            // Connection settings
            if (config.agent.connections) {
                // Clear existing servers
                const serversContainer = document.getElementById('serversContainer');
                if (serversContainer) {
                    serversContainer.innerHTML = '';
                }

                // Add servers from config
                if (config.agent.connections.mcpServers && typeof config.agent.connections.mcpServers === 'object') {
                    serverCounter = 0; // Reset counter

                    // Handle new mcpServers object format - ensure all servers are properly processed
                    const mcpServersEntries = Object.keys(config.agent.connections.mcpServers).map(key => [
                        key,
                        config.agent.connections.mcpServers[key]
                    ]);

                    // Process each server explicitly using array
                    for (let i = 0; i < mcpServersEntries.length; i++) {
                        const [id, server] = mcpServersEntries[i];
                        addServerFromConfig(id, server);
                    }
                } else if (config.agent.connections.servers && Array.isArray(config.agent.connections.servers)) {
                    // Handle legacy servers array format for backward compatibility
                    serverCounter = 0; // Reset counter

                    for (let i = 0; i < config.agent.connections.servers.length; i++) {
                        const server = config.agent.connections.servers[i];
                        const id = server.id || `server-${serverCounter}`;
                        const serverObj = {
                            command: server.command,
                            args: server.arguments || [],
                            environment: server.environment || {}
                        };
                        addServerFromConfig(id, serverObj);
                    }
                } else {
                    // Add a default server if none in config
                    addServer();
                }

                // Connection retry settings
                if (config.agent.connections.retry) {
                    setValue('connRetryMaxRetries', config.agent.connections.retry.max_retries);
                    setValue('connRetryInitialBackoff', config.agent.connections.retry.initial_backoff);
                    setValue('connRetryMaxBackoff', config.agent.connections.retry.max_backoff);
                    setValue('connRetryBackoffMultiplier', config.agent.connections.retry.backoff_multiplier);
                }
            } else {
                console.log("No connections found in config");
            }
        }

        // Runtime settings
        if (config.runtime) {
            // Log settings
            if (config.runtime.log) {
                setValue('logLevel', config.runtime.log.level);
                setValue('logOutput', config.runtime.log.output);
            }

            // Transport settings
            if (config.runtime.transports) {
                // STDIO settings
                if (config.runtime.transports.stdio) {
                    setValue('stdioEnabled', config.runtime.transports.stdio.enabled.toString());
                    setValue('stdioBufferSize', config.runtime.transports.stdio.buffer_size);
                }

                // HTTP settings
                if (config.runtime.transports.http) {
                    setValue('httpEnabled', config.runtime.transports.http.enabled.toString());
                    setValue('httpHost', config.runtime.transports.http.host);
                    setValue('httpPort', config.runtime.transports.http.port);

                    // Show/hide HTTP settings based on enabled state
                    const httpSettings = document.getElementById('httpSettings');
                    if (httpSettings) {
                        httpSettings.style.display = config.runtime.transports.http.enabled ? 'block' : 'none';
                    }
                }
            }
        }
    } catch (error) {
        console.error('Error applying configuration to form:', error);
        showMessage('Error applying configuration to form', 'error');
    }
}

// Add a server from configuration
function addServerFromConfig(id, serverConfig) {
    const serversContainer = document.getElementById('serversContainer');
    if (!serversContainer) return;

    const serverId = serverCounter++;

    const serverDiv = document.createElement('div');
    serverDiv.className = 'server-container';
    serverDiv.id = `server-${serverId}`;

    // Handle both new args and legacy arguments
    const argsArray = serverConfig.args || serverConfig.arguments || [];
    const argsStr = Array.isArray(argsArray) ? argsArray.join(', ') : '';

    // Build environment string
    let envStr = '';
    if (serverConfig.environment && typeof serverConfig.environment === 'object') {
        if (Array.isArray(serverConfig.environment)) {
            // Handle array of environment variables format
            envStr = serverConfig.environment.join(', ');
        } else {
            // Handle object format
            envStr = Object.entries(serverConfig.environment)
                .map(([key, value]) => `${key}=${value}`)
                .join(', ');
        }
    }

    // Escape special characters in id to avoid issues with user-provided values
    const safeId = id.replace(/[^a-zA-Z0-9-_]/g, '_');

    serverDiv.innerHTML = `
        <div class="form-group">
            <label for="serverId-${serverId}">Server ID:</label>
            <input type="text" id="serverId-${serverId}" value="${safeId}" />
        </div>

        <div class="form-group">
            <label for="serverCommand-${serverId}">Command:</label>
            <input type="text" id="serverCommand-${serverId}" value="${serverConfig.command || 'docker'}" />
        </div>

        <div class="form-group">
            <label for="serverArgs-${serverId}">Arguments:</label>
            <input type="text" id="serverArgs-${serverId}" value="${argsStr}" placeholder="Comma-separated list" />
        </div>

        <div class="form-group">
            <label for="serverEnv-${serverId}">Environment:</label>
            <input type="text" id="serverEnv-${serverId}" value="${envStr}" placeholder="KEY=VALUE format, comma-separated" />
        </div>

        <button class="remove-server-btn" onclick="removeServer('server-${serverId}')">
            <i class="fas fa-trash"></i> Remove Server
        </button>
    `;

    serversContainer.appendChild(serverDiv);

    // Add event listeners to the new server's inputs
    const serverInputs = serverDiv.querySelectorAll('input, select');
    serverInputs.forEach(input => {
        input.addEventListener('change', function() {
            generateAndUpdateConfig();
        });

        if (input.tagName === 'INPUT' && (input.type === 'text' || input.type === 'number')) {
            input.addEventListener('keyup', function() {
                generateAndUpdateConfig();
            });
        }
    });
}

// Helper function to set form field values
function setValue(id, value) {
    const element = document.getElementById(id);
    if (!element) return;

    if (element.tagName === 'SELECT' || element.tagName === 'INPUT' || element.tagName === 'TEXTAREA') {
        element.value = value !== undefined && value !== null ? value : '';

        // Trigger change event for select elements to ensure proper state
        if (element.tagName === 'SELECT') {
            const changeEvent = new Event('change', { bubbles: true });
            element.dispatchEvent(changeEvent);
        }
    }
}

// Show message function that works with success and error messages
function showMessage(message, type = 'success') {
    // Check if message already exists
    let msgElement = document.querySelector(`.message-${type}`);

    if (!msgElement) {
        // Create new message
        msgElement = document.createElement('div');
        msgElement.className = `message-${type}`;
        document.body.appendChild(msgElement);
    } else {
        // Reset animation by removing and re-adding the element
        msgElement.remove();
        msgElement = document.createElement('div');
        msgElement.className = `message-${type}`;
        document.body.appendChild(msgElement);
    }

    msgElement.textContent = message;

    // Remove after animation completes
    setTimeout(() => {
        if (msgElement.parentNode) {
            msgElement.parentNode.removeChild(msgElement);
        }
    }, 3000);
}