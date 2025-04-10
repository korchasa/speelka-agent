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

// Toggle advanced sections
function toggleAdvanced(sectionId) {
    const section = document.getElementById(sectionId);
    const toggle = section.previousElementSibling;

    if (section.classList.contains('open')) {
        section.classList.remove('open');
        toggle.classList.remove('open');
    } else {
        section.classList.add('open');
        toggle.classList.add('open');
    }
}

// Initialize the configuration tool
let serverCounter = 0;

function initConfigTool() {
    // Only initialize if we're on the page with the config tool
    if (!document.getElementById('serversContainer')) return;

    // Add a default server
    addServer();

    // Initialize all advanced sections to be closed by default
    const advancedSections = document.querySelectorAll('.advanced-section');
    advancedSections.forEach(section => {
        section.classList.remove('open');
        const toggle = section.previousElementSibling;
        if (toggle && toggle.classList.contains('advanced-toggle')) {
            toggle.classList.remove('open');
        }
    });

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
            <span class="field-description">Unique identifier for this MCP server connection</span>
        </div>

        <div class="form-group">
            <label for="serverCommand-${serverId}">Command:</label>
            <input type="text" id="serverCommand-${serverId}" value="docker" />
            <span class="field-description">Command to execute for this MCP server</span>
        </div>

        <div class="advanced-toggle" onclick="toggleAdvanced('serverAdvanced-${serverId}')">
            <i class="fas fa-caret-right"></i> Advanced Server Settings
        </div>
        
        <div id="serverAdvanced-${serverId}" class="advanced-section">
            <div class="form-group">
                <label for="serverArgs-${serverId}">Arguments:</label>
                <input type="text" id="serverArgs-${serverId}" value="run, -i, --rm, mcp/time" placeholder="Comma-separated list" />
                <span class="field-description">Command arguments as a comma-separated list</span>
            </div>

            <div class="form-group">
                <label for="serverEnv-${serverId}">Environment:</label>
                <input type="text" id="serverEnv-${serverId}" value="NODE_ENV=production" placeholder="KEY=VALUE format, comma-separated" />
                <span class="field-description">Environment variables in KEY=VALUE format, comma-separated</span>
            </div>
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

// Create a configuration object from form inputs
function createConfigObjectFromForm() {
    // Helper function to safely get value from an element
    const getValue = (id, defaultValue) => {
        const element = document.getElementById(id);
        return element ? element.value : defaultValue;
    };

    // Helper function to safely get numeric value from an element
    const getNumericValue = (id, defaultValue, parser = parseInt) => {
        const element = document.getElementById(id);
        if (!element) return defaultValue;
        const value = parser(element.value);
        return isNaN(value) ? defaultValue : value;
    };

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
    const agentName = getValue('agentName', 'speelka-agent');
    const agentVersion = getValue('agentVersion', '1.0.0');
    const toolName = getValue('toolName', 'process');
    const toolDescription = getValue('toolDescription', 'Process tool for handling user queries with LLM');
    const toolArgumentName = getValue('argumentName', 'input');
    const toolArgumentDescription = getValue('argumentDescription', 'User query to process');

    // LLM section
    const llmProvider = getValue('llmProvider', 'openai');
    const llmApiKey = getValue('llmApiKey', '');
    const llmModel = getValue('llmModel', 'gpt-4o');
    const llmMaxTokens = getNumericValue('llmMaxTokens', 0);
    const llmTemperature = getNumericValue('llmTemperature', 0.7, parseFloat);
    const llmPromptTemplate = getValue('promptTemplate', 'You are a helpful AI assistant...');
    const llmRetryMaxRetries = getNumericValue('maxRetries', 3);
    const llmRetryInitialBackoff = getNumericValue('initialBackoff', 1.0, parseFloat);
    const llmRetryMaxBackoff = getNumericValue('maxBackoff', 30.0, parseFloat);
    const llmRetryBackoffMultiplier = getNumericValue('backoffMultiplier', 2.0, parseFloat);

    // Connections section
    const mcpServers = {};
    const serverDivs = document.querySelectorAll('.server-container');
    serverDivs.forEach(div => {
        const id = div.id;
        const serverId = getValue(`serverId-${id.split('-')[1]}`, `server-${id.split('-')[1]}`);
        const command = getValue(`serverCommand-${id.split('-')[1]}`, 'docker');
        const argsStr = getValue(`serverArgs-${id.split('-')[1]}`, '');
        const envStr = getValue(`serverEnv-${id.split('-')[1]}`, '');

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

    const connMaxRetries = getNumericValue('connMaxRetries', 3);
    const connInitialBackoff = getNumericValue('connInitialBackoff', 1.0, parseFloat);
    const connMaxBackoff = getNumericValue('connMaxBackoff', 30.0, parseFloat);
    const connBackoffMultiplier = getNumericValue('connBackoffMultiplier', 2.0, parseFloat);

    // Runtime section
    const logLevel = getValue('logLevel', 'info');
    const logOutput = getValue('logOutput', 'stdout');
    const stdioEnabled = getValue('stdioEnabled', 'true') === 'true';
    const stdioBufferSize = getNumericValue('stdioBufferSize', 8192);
    const httpEnabled = getValue('httpEnabled', 'false') === 'true';
    const httpHost = getValue('httpHost', 'localhost');
    const httpPort = getNumericValue('httpPort', 3000);

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
                api_key: llmApiKey || "YOUR_API_KEY_HERE",
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
                    max_retries: connMaxRetries,
                    initial_backoff: connInitialBackoff,
                    max_backoff: connMaxBackoff,
                    backoff_multiplier: connBackoffMultiplier
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

// Generate and update configuration in the output area
function generateAndUpdateConfig() {
    const config = createConfigObjectFromForm();

    // Format the JSON with indentation for better readability
    const formattedJson = JSON.stringify(config, null, 2);

    // Update the displayed configuration
    const configOutput = document.getElementById('generatedConfig');
    if (configOutput) {
        configOutput.textContent = formattedJson;
    }

    // Also update the examples with the current configuration
    updateExamplesWithConfig(config);

    return formattedJson;
}

// Update examples with the current configuration
function updateExamplesWithConfig(config) {
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