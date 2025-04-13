// =============================================================================
// Utility Modules
// =============================================================================

/**
 * Basic error logging utility
 */
function logError(context, error) {
    const errorMsg = error instanceof Error ? error.message : error;
    console.error(`[${context}] ${errorMsg}`);
    return errorMsg;
}

/**
 * Initialize mobile menu functionality
 */
function initMobileMenu() {
    const hamburger = document.querySelector('.hamburger');
    const navLinks = document.querySelector('.nav-links');

    if (hamburger && navLinks) {
        hamburger.addEventListener('click', () => {
            navLinks.classList.toggle('active');
            hamburger.classList.toggle('active');
            hamburger.setAttribute('aria-expanded',
                hamburger.getAttribute('aria-expanded') === 'false' ? 'true' : 'false'
            );
        });
    }
}

/**
 * Initializes tabbed interfaces throughout the page
 */
function initTabs() {
    const tabBtns = document.querySelectorAll('.tab-btn');

    tabBtns.forEach(btn => {
        btn.addEventListener('click', () => {
            // Remove active class from all buttons
            tabBtns.forEach(b => b.classList.remove('active'));

            // Add active class to clicked button
            btn.classList.add('active');

            // Get the tab to show
            const tabToShow = btn.getAttribute('data-tab');

            // Get all tab content elements
            const allTabContent = document.querySelectorAll('.tab-content');

            // Hide all tabs
            allTabContent.forEach(tab => {
                tab.classList.remove('active');
            });

            // Show the selected tab config content
            const configTabId = `${tabToShow}-tab-config`;
            const configTab = document.getElementById(configTabId);
            if (configTab) {
                configTab.classList.add('active');
            }

            // Load the example file for this tab if it hasn't been loaded yet
            loadExampleFile(tabToShow, 'yaml');
        });
    });
}

/**
 * Set up copy buttons for code blocks
 */
function setupCodeCopyButtons() {
    const copyButtons = document.querySelectorAll('.copy-btn');
    copyButtons.forEach(button => {
        button.addEventListener('click', function() {
            // Find the code element - now the button is positioned before the code
            const codeBlock = this.closest('.code-block, .yaml-code').querySelector('code');
            if (!codeBlock) return;

            const textToCopy = codeBlock.textContent;

            navigator.clipboard.writeText(textToCopy).then(() => {
                // Visual feedback
                const originalIcon = this.innerHTML;
                this.innerHTML = '<i class="fas fa-check"></i>';
                setTimeout(() => {
                    this.innerHTML = originalIcon;
                }, 2000);
            }).catch(err => {
                console.error('Failed to copy text: ', err);
            });
        });
    });
}

/**
 * Initializes Mermaid diagrams with optimized settings
 */
function initializeMermaid() {
    if (typeof mermaid !== 'undefined') {
        mermaid.initialize({
            startOnLoad: true,
            theme: 'dark',
            securityLevel: 'loose',
            logLevel: 5,
            flowchart: {
                useMaxWidth: true,
                htmlLabels: true,
                curve: 'basis'
            }
        });
    }
}

/**
 * Initializes highlight.js for code syntax highlighting
 */
function initializeHighlightJS() {
    if (typeof hljs !== 'undefined') {
        // Apply highlighting to all code blocks
        document.querySelectorAll('pre code').forEach(block => {
            // If the code block has a language-* class, use that language
            const languageClass = Array.from(block.classList).find(cls => cls.startsWith('language-'));
            if (languageClass) {
                const language = languageClass.replace('language-', '');
                // Set the language explicitly
                block.setAttribute('data-language', language);
            }

            // Make sure any previous highlighting is cleared
            if (block.hasAttribute('data-highlighted')) {
                block.removeAttribute('data-highlighted');
            }

            // Apply highlighting
            hljs.highlightElement(block);
        });
    }
}

// Function to load example files
function loadExampleFile(exampleName, format) {
    // Only load YAML files since they're the only ones that exist
    if (format.toLowerCase() !== 'yaml') {
        return;
    }

    const lowerFormat = format.toLowerCase();
    console.log(`Loading example file: ${exampleName}.${lowerFormat}`);

    // Fetch the example file from site/examples directory
    fetch(`./examples/${exampleName}.${lowerFormat}`)
        .then(response => {
            if (!response.ok) {
                throw new Error(`Failed to load ${exampleName}.${lowerFormat} example file`);
            }
            return response.text();
        })
        .then(content => {
            // Find the code element within the tab-config div
            const tabConfigId = `${exampleName}-tab-config`;
            console.log(`Looking for code in tab config: ${tabConfigId}`);
            const tabConfigElement = document.getElementById(tabConfigId);

            if (tabConfigElement) {
                // Find the code element inside the pre element
                const codeElement = tabConfigElement.querySelector('code');
                if (codeElement) {
                    // Update the content of the code element
                    codeElement.textContent = content;
                    console.log(`Updated code content in ${tabConfigId}`);

                    // Remove any existing highlight.js classes
                    codeElement.classList.remove('hljs');
                    Array.from(codeElement.classList).forEach(cls => {
                        if (cls.startsWith('language-') || cls.startsWith('hljs-')) {
                            codeElement.classList.remove(cls);
                        }
                    });

                    // Add language class for YAML
                    codeElement.classList.add('language-yaml');

                    // Set the language explicitly
                    codeElement.setAttribute('data-language', 'yaml');

                    // Make sure any previous highlighting is cleared
                    if (codeElement.hasAttribute('data-highlighted')) {
                        codeElement.removeAttribute('data-highlighted');
                    }

                    // Apply syntax highlighting to the updated code
                    if (typeof hljs !== 'undefined') {
                        hljs.highlightElement(codeElement);
                    }
                } else {
                    console.error(`Code element not found in ${tabConfigId}`);
                }
            } else {
                console.error(`Tab config element with ID ${tabConfigId} not found`);
            }
        })
        .catch(error => {
            console.error(`Error loading example file: ${error.message}`);
        });
}

/**
 * Initialize agent selection in the Usage Guide
 */
function initAgentSelection() {
    const agentButtons = document.querySelectorAll('.tab-btn');

    // Make sure the first agent type is active by default
    if (agentButtons.length > 0 && !agentButtons[0].classList.contains('active')) {
        agentButtons[0].click();
    }
}

/**
 * Main initialization function
 */
function initializeApp() {
    try {
        // Initialize tabs
        initTabs();

        // Setup clipboard functionality
        setupCodeCopyButtons();

        // Initialize mermaid diagrams
        initializeMermaid();

        // Initialize syntax highlighting
        initializeHighlightJS();

        // Initialize mobile menu
        initMobileMenu();

        // Initialize agent selection
        initAgentSelection();

        // Load example files
        loadExampleFile('simple', 'yaml');
        loadExampleFile('ai-news', 'yaml');
        loadExampleFile('architect', 'yaml');
    } catch (error) {
        logError('App Initialization', error);
    }
}

// Initialize the application when DOM is fully loaded
document.addEventListener('DOMContentLoaded', initializeApp);
