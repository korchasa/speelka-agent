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

    // Validate required fields
    const requiredFields = [
        { id: 'agentName', name: 'Agent Name' },
        { id: 'toolName', name: 'Tool Name' },
        { id: 'toolDescription', name: 'Tool Description' },
        { id: 'llmProvider', name: 'LLM Provider' },
        { id: 'llmApiKey', name: 'API Key' },
        { id: 'llmModel', name: 'Model' },
        { id: 'promptTemplate', name: 'Prompt Template' }
    ];

    let validationErrors = [];

    for (const field of requiredFields) {
        const element = document.getElementById(field.id);
        if (!element || !element.value.trim()) {
            validationErrors.push(`${field.name} is required`);
            if (element) {
                element.classList.add('error');

                // Show error message
                let errorMsg = element.parentNode.querySelector('.field-error');
                if (!errorMsg) {
                    errorMsg = document.createElement('span');
                    errorMsg.className = 'field-error';
                    element.parentNode.insertBefore(errorMsg, element.nextSibling);
                }
                errorMsg.textContent = `${field.name} is required`;
            }
        } else if (element) {
            element.classList.remove('error');

            // Remove error message if exists
            const errorMsg = element.parentNode.querySelector('.field-error');
            if (errorMsg) {
                errorMsg.remove();
            }
        }
    }

    // Special validation for prompt template
    const promptTemplate = getValue('promptTemplate', '');
    if (promptTemplate) {
        if (!promptTemplate.includes('{{input}}') || !promptTemplate.includes('{{tools}}')) {
            validationErrors.push('Prompt Template must include {{input}} and {{tools}} placeholders');
            const element = document.getElementById('promptTemplate');
            if (element) {
                element.classList.add('error');

                // Show error message
                let errorMsg = element.parentNode.querySelector('.field-error');
                if (!errorMsg) {
                    errorMsg = document.createElement('span');
                    errorMsg.className = 'field-error';
                    element.parentNode.insertBefore(errorMsg, element.nextSibling);
                }
                errorMsg.textContent = 'Prompt Template must include {{input}} and {{tools}} placeholders';
            }
        }
    }

    // If there are validation errors, show them and return null
    if (validationErrors.length > 0) {
        console.error('Validation errors:', validationErrors);
        return null;
    }

    // Form validation for numeric fields
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
    // Default values for comparison
    const defaults = {
        agent: {
            name: 'speelka-agent',
            version: '1.0.0',
            tool: {
                name: 'process',
                description: 'Process tool for handling user queries with LLM',
                argument_name: 'input',
                argument_description: 'User query to process'
            },
            llm: {
                provider: 'openai',
                model: 'gpt-4o',
                max_tokens: 0,
                temperature: 0.7,
                retry: {
                    max_retries: 3,
                    initial_backoff: 1.0,
                    max_backoff: 30.0,
                    backoff_multiplier: 2.0
                }
            },
            connections: {
                retry: {
                    max_retries: 3,
                    initial_backoff: 1.0,
                    max_backoff: 30.0,
                    backoff_multiplier: 2.0
                }
            }
        },
        runtime: {
            log: {
                level: 'info',
                output: 'stdout'
            },
            transports: {
                stdio: {
                    enabled: true,
                    buffer_size: 8192
                },
                http: {
                    enabled: false,
                    host: 'localhost',
                    port: 3000
                }
            }
        }
    };

    // Helper function to check if value is default
    const isDefault = (path, value) => {
        const parts = path.split('.');
        let defaultValue = defaults;

        for (const part of parts) {
            if (defaultValue === undefined || defaultValue[part] === undefined) {
                return false;
            }
            defaultValue = defaultValue[part];
        }

        return defaultValue === value;
    };

    // Generate environment variables from the configuration
    let envVars = [];

    // Agent settings
    envVars.push(`# Agent`);
    if (!isDefault('agent.name', config.agent.name)) {
        envVars.push(`export AGENT_NAME="${config.agent.name}"`);
    }
    if (!isDefault('agent.version', config.agent.version)) {
        envVars.push(`export AGENT_VERSION="${config.agent.version}"`);
    }

    // Tool settings
    envVars.push(`\n# Tool`);
    if (!isDefault('agent.tool.name', config.agent.tool.name)) {
        envVars.push(`export TOOL_NAME="${config.agent.tool.name}"`);
    }
    if (!isDefault('agent.tool.description', config.agent.tool.description)) {
        envVars.push(`export TOOL_DESCRIPTION="${config.agent.tool.description}"`);
    }
    if (!isDefault('agent.tool.argument_name', config.agent.tool.argument_name)) {
        envVars.push(`export TOOL_ARGUMENT_NAME="${config.agent.tool.argument_name}"`);
    }
    if (!isDefault('agent.tool.argument_description', config.agent.tool.argument_description)) {
        envVars.push(`export TOOL_ARGUMENT_DESCRIPTION="${config.agent.tool.argument_description}"`);
    }

    // LLM settings
    envVars.push(`\n# LLM`);
    if (!isDefault('agent.llm.provider', config.agent.llm.provider)) {
        envVars.push(`export LLM_PROVIDER="${config.agent.llm.provider}"`);
    }
    // API key is always required
    envVars.push(`export LLM_API_KEY="..."`);
    if (!isDefault('agent.llm.model', config.agent.llm.model)) {
        envVars.push(`export LLM_MODEL="${config.agent.llm.model}"`);
    }
    if (!isDefault('agent.llm.max_tokens', config.agent.llm.max_tokens)) {
        envVars.push(`export LLM_MAX_TOKENS=${config.agent.llm.max_tokens}`);
    }
    if (!isDefault('agent.llm.temperature', config.agent.llm.temperature)) {
        envVars.push(`export LLM_TEMPERATURE=${config.agent.llm.temperature}`);
    }

    // Only include prompt template if it's not the default
    const defaultPrompt = "You are a helpful AI assistant. Respond to the following request:\n\n{{input}}\n\nProvide a detailed and helpful response.\n\nAvailable tools:\n{{tools}}";
    if (config.agent.llm.prompt_template && config.agent.llm.prompt_template !== defaultPrompt) {
        const promptTemplate = config.agent.llm.prompt_template.replace(/"/g, '\\"');
        envVars.push(`export LLM_PROMPT_TEMPLATE="${promptTemplate}"`);
    }

    // LLM Retry settings
    let hasNonDefaultRetry = false;
    const retryConfig = [];

    if (!isDefault('agent.llm.retry.max_retries', config.agent.llm.retry.max_retries)) {
        hasNonDefaultRetry = true;
        retryConfig.push(`export LLM_RETRY_MAX_RETRIES=${config.agent.llm.retry.max_retries}`);
    }
    if (!isDefault('agent.llm.retry.initial_backoff', config.agent.llm.retry.initial_backoff)) {
        hasNonDefaultRetry = true;
        retryConfig.push(`export LLM_RETRY_INITIAL_BACKOFF=${config.agent.llm.retry.initial_backoff}`);
    }
    if (!isDefault('agent.llm.retry.max_backoff', config.agent.llm.retry.max_backoff)) {
        hasNonDefaultRetry = true;
        retryConfig.push(`export LLM_RETRY_MAX_BACKOFF=${config.agent.llm.retry.max_backoff}`);
    }
    if (!isDefault('agent.llm.retry.backoff_multiplier', config.agent.llm.retry.backoff_multiplier)) {
        hasNonDefaultRetry = true;
        retryConfig.push(`export LLM_RETRY_BACKOFF_MULTIPLIER=${config.agent.llm.retry.backoff_multiplier}`);
    }

    if (hasNonDefaultRetry) {
        envVars.push(`\n# LLM Retry`);
        envVars = envVars.concat(retryConfig);
    }

    // MCP Servers
    const mcpServers = config.agent.connections.mcpServers;
    if (Object.keys(mcpServers).length > 0) {
        envVars.push(`\n# MCP Servers`);
        let serverIndex = 0;
        for (const [serverId, serverConfig] of Object.entries(mcpServers)) {
            envVars.push(`export MCPS_${serverIndex}_ID="${serverId}"`);
            envVars.push(`export MCPS_${serverIndex}_COMMAND="${serverConfig.command}"`);

            if (serverConfig.args && serverConfig.args.length > 0) {
                envVars.push(`export MCPS_${serverIndex}_ARGS="${serverConfig.args.join(' ')}"`);
            }

            // Environment variables if any
            if (serverConfig.environment && Object.keys(serverConfig.environment).length > 0) {
                for (const [envKey, envValue] of Object.entries(serverConfig.environment)) {
                    envVars.push(`export MCPS_${serverIndex}_ENV_${envKey}="${envValue}"`);
                }
            }

            envVars.push(``);
            serverIndex++;
        }
    }

    // Connection retry settings
    let hasNonDefaultConnRetry = false;
    const connRetryConfig = [];

    if (!isDefault('agent.connections.retry.max_retries', config.agent.connections.retry.max_retries)) {
        hasNonDefaultConnRetry = true;
        connRetryConfig.push(`export MSPS_RETRY_MAX_RETRIES=${config.agent.connections.retry.max_retries}`);
    }
    if (!isDefault('agent.connections.retry.initial_backoff', config.agent.connections.retry.initial_backoff)) {
        hasNonDefaultConnRetry = true;
        connRetryConfig.push(`export MSPS_RETRY_INITIAL_BACKOFF=${config.agent.connections.retry.initial_backoff}`);
    }
    if (!isDefault('agent.connections.retry.max_backoff', config.agent.connections.retry.max_backoff)) {
        hasNonDefaultConnRetry = true;
        connRetryConfig.push(`export MSPS_RETRY_MAX_BACKOFF=${config.agent.connections.retry.max_backoff}`);
    }
    if (!isDefault('agent.connections.retry.backoff_multiplier', config.agent.connections.retry.backoff_multiplier)) {
        hasNonDefaultConnRetry = true;
        connRetryConfig.push(`export MSPS_RETRY_BACKOFF_MULTIPLIER=${config.agent.connections.retry.backoff_multiplier}`);
    }

    if (hasNonDefaultConnRetry) {
        envVars.push(`# MSPS Retry`);
        envVars = envVars.concat(connRetryConfig);
    }

    // Runtime settings
    envVars.push(`\n# Runtime`);
    if (!isDefault('runtime.log.level', config.runtime.log.level)) {
        envVars.push(`export RUNTIME_LOG_LEVEL="${config.runtime.log.level}"`);
    }
    if (!isDefault('runtime.log.output', config.runtime.log.output)) {
        envVars.push(`export RUNTIME_LOG_OUTPUT="${config.runtime.log.output}"`);
    }

    // Transport settings - only include if non-default
    let hasStdioSettings = false;
    const stdioSettings = [];

    if (!isDefault('runtime.transports.stdio.enabled', config.runtime.transports.stdio.enabled)) {
        hasStdioSettings = true;
        stdioSettings.push(`export RUNTIME_STDIO_ENABLED=${config.runtime.transports.stdio.enabled}`);
    }
    if (!isDefault('runtime.transports.stdio.buffer_size', config.runtime.transports.stdio.buffer_size)) {
        hasStdioSettings = true;
        stdioSettings.push(`export RUNTIME_STDIO_BUFFER_SIZE=${config.runtime.transports.stdio.buffer_size}`);
    }

    if (hasStdioSettings) {
        envVars.push(`\n# Transport - Stdio`);
        envVars = envVars.concat(stdioSettings);
    }

    // HTTP settings - only include if enabled or non-default
    if (config.runtime.transports.http && (
        config.runtime.transports.http.enabled ||
        !isDefault('runtime.transports.http.host', config.runtime.transports.http.host) ||
        !isDefault('runtime.transports.http.port', config.runtime.transports.http.port)
    )) {
        envVars.push(`\n# Transport - HTTP`);
        if (!isDefault('runtime.transports.http.enabled', config.runtime.transports.http.enabled)) {
            envVars.push(`export RUNTIME_HTTP_ENABLED=${config.runtime.transports.http.enabled}`);
        }
        if (!isDefault('runtime.transports.http.host', config.runtime.transports.http.host)) {
            envVars.push(`export RUNTIME_HTTP_HOST="${config.runtime.transports.http.host}"`);
        }
        if (!isDefault('runtime.transports.http.port', config.runtime.transports.http.port)) {
            envVars.push(`export RUNTIME_HTTP_PORT=${config.runtime.transports.http.port}`);
        }
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
        const binaryEnv = {};

        // Only include non-default values
        if (!isDefault('agent.name', config.agent.name)) binaryEnv.AGENT_NAME = config.agent.name;
        if (!isDefault('agent.version', config.agent.version)) binaryEnv.AGENT_VERSION = config.agent.version;
        if (!isDefault('agent.tool.name', config.agent.tool.name)) binaryEnv.TOOL_NAME = config.agent.tool.name;
        if (!isDefault('agent.tool.description', config.agent.tool.description)) binaryEnv.TOOL_DESCRIPTION = config.agent.tool.description;
        if (!isDefault('agent.tool.argument_name', config.agent.tool.argument_name)) binaryEnv.TOOL_ARGUMENT_NAME = config.agent.tool.argument_name;
        if (!isDefault('agent.tool.argument_description', config.agent.tool.argument_description)) binaryEnv.TOOL_ARGUMENT_DESCRIPTION = config.agent.tool.argument_description;

        // Always include API key placeholder
        binaryEnv.LLM_API_KEY = "...YOUR_LLM_API_KEY...";

        if (!isDefault('agent.llm.provider', config.agent.llm.provider)) binaryEnv.LLM_PROVIDER = config.agent.llm.provider;
        if (!isDefault('agent.llm.model', config.agent.llm.model)) binaryEnv.LLM_MODEL = config.agent.llm.model;
        if (!isDefault('agent.llm.max_tokens', config.agent.llm.max_tokens)) binaryEnv.LLM_MAX_TOKENS = config.agent.llm.max_tokens;
        if (!isDefault('agent.llm.temperature', config.agent.llm.temperature)) binaryEnv.LLM_TEMPERATURE = config.agent.llm.temperature;
        if (!isDefault('runtime.log.level', config.runtime.log.level)) binaryEnv.RUNTIME_LOG_LEVEL = config.runtime.log.level;

        let binaryJson = JSON.stringify({
            mcpServers: {
                "speelka-agent": {
                    command: "speelka-agent",
                    args: [],
                    environment: binaryEnv
                }
            }
        }, null, 4);
        binaryExample.textContent = binaryJson;
    }

    // Update docker example with environment variables
    const dockerExample = document.querySelector('.instructions pre.code-block:nth-of-type(3) code');
    if (dockerExample) {
        // Create environment arguments list for Docker
        const dockerEnvArgs = ["run", "-i", "--rm"];

        // Only add non-default env vars
        if (!isDefault('agent.name', config.agent.name)) dockerEnvArgs.push("-e", `AGENT_NAME=${config.agent.name}`);
        if (!isDefault('agent.version', config.agent.version)) dockerEnvArgs.push("-e", `AGENT_VERSION=${config.agent.version}`);
        if (!isDefault('agent.tool.name', config.agent.tool.name)) dockerEnvArgs.push("-e", `TOOL_NAME=${config.agent.tool.name}`);
        if (!isDefault('agent.tool.description', config.agent.tool.description)) dockerEnvArgs.push("-e", `TOOL_DESCRIPTION=${config.agent.tool.description}`);
        if (!isDefault('agent.tool.argument_name', config.agent.tool.argument_name)) dockerEnvArgs.push("-e", `TOOL_ARGUMENT_NAME=${config.agent.tool.argument_name}`);
        if (!isDefault('agent.tool.argument_description', config.agent.tool.argument_description)) dockerEnvArgs.push("-e", `TOOL_ARGUMENT_DESCRIPTION=${config.agent.tool.argument_description}`);

        // Always include LLM API key
        dockerEnvArgs.push("-e", "LLM_API_KEY=$LLM_API_KEY");

        if (!isDefault('agent.llm.provider', config.agent.llm.provider)) dockerEnvArgs.push("-e", `LLM_PROVIDER=${config.agent.llm.provider}`);
        if (!isDefault('agent.llm.model', config.agent.llm.model)) dockerEnvArgs.push("-e", `LLM_MODEL=${config.agent.llm.model}`);
        if (!isDefault('runtime.log.level', config.runtime.log.level)) dockerEnvArgs.push("-e", `RUNTIME_LOG_LEVEL=${config.runtime.log.level}`);

        // Add the image tag
        dockerEnvArgs.push("ghcr.io/korchasa/speelka-agent:latest");

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