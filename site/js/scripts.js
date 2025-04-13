// =============================================================================
// Utility Modules
// =============================================================================

/**
 * Centralized error handling utility
 */
const ErrorHandler = {
    /**
     * Log an error with consistent formatting
     * @param {string} context - Where the error occurred
     * @param {string|Error} error - The error message or object
     * @param {boolean} silent - Whether to show UI feedback
     * @returns {string} The formatted error message
     */
    logError: function(context, error, silent = true) {
        const errorMsg = error instanceof Error ? error.message : error;

        // In development, show detailed console errors
        if (typeof window !== 'undefined' && window.location.hostname === 'localhost') {
            console.error(`[${context}] Error:`, error);
        } else {
            // In production, log minimal information
            console.error(`[${context}] ${errorMsg}`);
        }

        // Show UI feedback if not silent
        if (!silent) {
            this.showErrorFeedback(errorMsg);
        }

        return errorMsg;
    },

    /**
     * Display an error message to the user
     * @param {string} message - The error message
     */
    showErrorFeedback: function(message) {
        // Find or create error container
        let errorContainer = document.querySelector('.error-feedback');

        if (!errorContainer) {
            errorContainer = document.createElement('div');
            errorContainer.className = 'error-feedback';
            errorContainer.setAttribute('role', 'alert');
            errorContainer.setAttribute('aria-live', 'assertive');
            document.body.appendChild(errorContainer);
        }

        // Set message and show
        errorContainer.textContent = message;
        errorContainer.classList.add('visible');

        // Hide after 5 seconds
        setTimeout(() => {
            if (errorContainer && errorContainer.parentNode) {
                errorContainer.classList.remove('visible');
                setTimeout(() => {
                    if (errorContainer && errorContainer.parentNode) {
                        errorContainer.parentNode.removeChild(errorContainer);
                    }
                }, 300); // Allow transition to complete
            }
        }, 5000);
    }
};

/**
 * Clipboard utility functions
 */
const ClipboardUtils = {
    /**
     * Copies text to clipboard with visual feedback
     * @param {string} text - The text to copy
     * @param {HTMLElement} button - Optional button element to show feedback on
     */
    copyToClipboard: function(text, button = null) {
        if (!text) {
            ErrorHandler.logError('Clipboard', 'No text to copy');
            return;
        }

        // Store original button content if button is provided
        let originalContent = '';
        if (button && button.innerHTML) {
            originalContent = button.innerHTML;
        }

        // Show loading state
        if (button) {
            button.innerHTML = '<i class="fas fa-spinner fa-spin"></i>';
        }

        if (navigator.clipboard) {
            navigator.clipboard.writeText(text)
                .then(() => {
                    // Show success state
                    if (button) {
                        button.innerHTML = '<i class="fas fa-check"></i>';
                        setTimeout(() => {
                            button.innerHTML = originalContent;
                        }, 2000);
                    } else {
                        this.showCopySuccess();
                    }
                })
                .catch(err => {
                    ErrorHandler.logError('Clipboard', 'Failed to copy text to clipboard', false);
                    // Show error state
                    if (button) {
                        button.innerHTML = '<i class="fas fa-times"></i>';
                        setTimeout(() => {
                            button.innerHTML = originalContent;
                        }, 2000);
                    }
                });
        } else {
            // Fallback for browsers that don't support clipboard API
            const textArea = document.createElement('textarea');
            textArea.value = text;
            textArea.style.position = 'fixed';
            textArea.style.left = '-9999px';
            textArea.style.top = '0';
            document.body.appendChild(textArea);

            try {
                textArea.select();
                const successful = document.execCommand('copy');

                if (successful) {
                    // Show success state
                    if (button) {
                        button.innerHTML = '<i class="fas fa-check"></i>';
                        setTimeout(() => {
                            button.innerHTML = originalContent;
                        }, 2000);
                    } else {
                        this.showCopySuccess();
                    }
                } else {
                    throw new Error('Copy command was unsuccessful');
                }
            } catch (err) {
                ErrorHandler.logError('Clipboard', 'Failed to copy text (fallback method)', false);
                // Show error state
                if (button) {
                    button.innerHTML = '<i class="fas fa-times"></i>';
                    setTimeout(() => {
                        button.innerHTML = originalContent;
                    }, 2000);
                }
            } finally {
                document.body.removeChild(textArea);
            }
        }
    },

    /**
     * Shows a success message when text is copied to clipboard
     */
    showCopySuccess: function() {
        // Check if message already exists
        let successMsg = document.querySelector('.copy-success');

        if (!successMsg) {
            // Create new message
            successMsg = document.createElement('div');
            successMsg.className = 'copy-success';
            successMsg.textContent = 'Copied to clipboard!';
            successMsg.setAttribute('role', 'status');
            successMsg.setAttribute('aria-live', 'polite');
            document.body.appendChild(successMsg);
        } else {
            // Reset animation by removing and re-adding the element
            successMsg.remove();
            successMsg = document.createElement('div');
            successMsg.className = 'copy-success';
            successMsg.textContent = 'Copied to clipboard!';
            successMsg.setAttribute('role', 'status');
            successMsg.setAttribute('aria-live', 'polite');
            document.body.appendChild(successMsg);
        }

        // Remove after animation completes
        setTimeout(() => {
            if (successMsg && successMsg.parentNode) {
                successMsg.parentNode.removeChild(successMsg);
            }
        }, 2000);
    }
};

/**
 * DOM utilities for common operations
 */
const DOMUtils = {
    /**
     * Cleans up event listeners to prevent memory leaks
     */
    cleanupEventListeners: function() {
        try {
            // Clean up form event listeners
            const formContainer = document.querySelector('#configForm');
            if (formContainer && formContainer._inputHandler) {
                formContainer.removeEventListener('input', formContainer._inputHandler);
                formContainer.removeEventListener('change', formContainer._changeHandler);
            }

            // Clean up server event listeners
            const serversContainer = document.getElementById('serversContainer');
            if (serversContainer && serversContainer._clickHandler) {
                serversContainer.removeEventListener('click', serversContainer._clickHandler);
            }

            // Clean up add server button
            const addServerBtn = document.querySelector('.add-server-btn');
            if (addServerBtn && addServerBtn._clickHandler) {
                addServerBtn.removeEventListener('click', addServerBtn._clickHandler);
            }

            // Clean up copy button
            const copyEnvBtn = document.getElementById('copyEnvBtn');
            if (copyEnvBtn && copyEnvBtn._clickHandler) {
                copyEnvBtn.removeEventListener('click', copyEnvBtn._clickHandler);
            }
        } catch (error) {
            ErrorHandler.logError('Cleanup', 'Failed to clean up event listeners');
        }
    },

    /**
     * Sets up lazy loading for images with data-src attribute
     */
    setupLazyLoading: function() {
        // Skip if no images with data-src are found
        const lazyImages = document.querySelectorAll('img[data-src]');
        if (lazyImages.length === 0) return;

        // Check for IntersectionObserver support
        if ('IntersectionObserver' in window) {
            const imgObserver = new IntersectionObserver((entries, observer) => {
                entries.forEach(entry => {
                    if (entry.isIntersecting) {
                        const img = entry.target;
                        const src = img.getAttribute('data-src');

                        if (src) {
                            // Add a load event listener to handle failures
                            img.addEventListener('load', () => {
                                img.removeAttribute('data-src');
                                // Add a class to trigger fade-in animation if wanted
                                img.classList.add('loaded');
                            });

                            img.addEventListener('error', () => {
                                ErrorHandler.logError('Image Loading', `Failed to load image: ${src}`);
                                // Remove observer even if loading failed
                                observer.unobserve(img);
                            });

                            img.src = src;
                        }

                        observer.unobserve(img);
                    }
                });
            }, {
                rootMargin: '50px 0px', // Load images slightly before they enter viewport
                threshold: 0.1 // Trigger when 10% of the image is visible
            });

            // Observe all lazy images
            lazyImages.forEach(img => {
                imgObserver.observe(img);
            });
        } else {
            // Fallback for browsers that don't support IntersectionObserver
            // Instead of loading all at once, use a more optimized approach
            const lazyLoad = () => {
                const viewportTop = window.pageYOffset;
                const viewportBottom = viewportTop + window.innerHeight;

                lazyImages.forEach(img => {
                    if (!img.dataset.src) return;

                    const rect = img.getBoundingClientRect();
                    const imageTop = rect.top + viewportTop;
                    const imageBottom = rect.bottom + viewportTop;

                    // Check if the image is in or near the viewport
                    if ((imageTop <= viewportBottom + 200) && (imageBottom >= viewportTop - 200)) {
                        img.src = img.dataset.src;
                        img.addEventListener('load', () => {
                            img.removeAttribute('data-src');
                            img.classList.add('loaded');
                        });
                    }
                });
            };

            // Initial load
            lazyLoad();

            // Add event listeners with throttling
            let throttleTimeout;
            const throttle = (callback, time) => {
                if (throttleTimeout) return;
                throttleTimeout = setTimeout(() => {
                    callback();
                    throttleTimeout = null;
                }, time);
            };

            window.addEventListener('scroll', () => {
                throttle(lazyLoad, 200);
            });

            window.addEventListener('resize', () => {
                throttle(lazyLoad, 200);
            });
        }
    }
};

// =============================================================================
// Page Functionality
// =============================================================================

/**
 * Lazy load resources when they scroll into view
 * @param {HTMLElement[]} elements - Elements to lazy load
 * @param {Object} options - Intersection observer options
 */
function setupLazyLoading(elements, options = {}) {
    if (!elements || elements.length === 0) return;

    try {
        const lazyLoad = () => {
            if ('IntersectionObserver' in window) {
                // Use intersection observer
                const lazyObserver = new IntersectionObserver((entries) => {
                    entries.forEach(entry => {
                        if (entry.isIntersecting) {
                            const element = entry.target;
                            if (element.tagName === 'IMG' && element.dataset.src) {
                                element.src = element.dataset.src;
                                element.removeAttribute('data-src');
                            }
                            lazyObserver.unobserve(element);
                        }
                    });
                }, options);

                elements.forEach(el => lazyObserver.observe(el));
            } else {
                // Fallback for browsers that don't support intersection observer
                // Simple timeout to load all resources after a delay
                setTimeout(() => {
                    elements.forEach(element => {
                        if (element.tagName === 'IMG' && element.dataset.src) {
                            element.src = element.dataset.src;
                            element.removeAttribute('data-src');
                        }
                    });
                }, 500);
            }
        };

        // Call lazy load function
        lazyLoad();
    } catch (error) {
        ErrorHandler.logError('Lazy Loading', 'Failed to setup lazy loading', false);
    }
}

/**
 * Throttles a function to prevent it from being called too frequently
 * @param {Function} callback - The function to throttle
 * @param {number} time - The time in milliseconds to throttle by
 * @returns {Function} The throttled function
 */
function throttle(callback, time) {
    let throttlePause;

    return function() {
        if (throttlePause) return;

        throttlePause = true;
        setTimeout(() => {
            callback();
            throttlePause = false;
        }, time);
    };
}

// =============================================================================
// Application Initialization
// =============================================================================

/**
 * Initialize the application
 */
function initializeApp() {
    try {
        // Set up navigation
        setupNavigation();

        // Set up tabs if they exist
        initTabs();

        // Setup copy buttons
        setupCodeCopyButtons();

        // Set up UI components that need special handling
        setupConditionalUI();

        // Initialize Mermaid diagrams if available
        if (typeof mermaid !== 'undefined') {
            initializeMermaid();
        }
    } catch (error) {
        ErrorHandler.logError('App Initialization', 'Failed to initialize application', false);
    }
}

/**
 * Set up navigation functionality
 */
function setupNavigation() {
    try {
        // Setup mobile navigation toggle
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

        // Smooth scroll for anchor links
        document.querySelectorAll('a[href^="#"]').forEach(anchor => {
            anchor.addEventListener('click', function (e) {
                const targetId = this.getAttribute('href');
                if (targetId === '#') return; // Skip for # links that don't point to an element

                const targetElement = document.querySelector(targetId);
                if (targetElement) {
                    e.preventDefault();
                    targetElement.scrollIntoView({
                        behavior: 'smooth',
                        block: 'start'
                    });

                    // Close mobile menu if open
                    if (hamburger && navLinks && navLinks.classList.contains('active')) {
                        navLinks.classList.remove('active');
                        hamburger.classList.remove('active');
                        hamburger.setAttribute('aria-expanded', 'false');
                    }
                }
            });
        });
    } catch (error) {
        ErrorHandler.logError('Navigation', 'Failed to set up navigation', false);
    }
}

/**
 * Initializes tabbed interfaces throughout the page
 */
function initTabs() {
    // Set up any tabbed interfaces
    const tabContainers = document.querySelectorAll('.tabs-navigation');

        tabContainers.forEach(container => {
        const tabButtons = container.querySelectorAll('.tab-btn');

        tabButtons.forEach(button => {
            button.addEventListener('click', () => {
                // Remove active class from all buttons in this container
                container.querySelectorAll('.tab-btn').forEach(btn => {
                    btn.classList.remove('active');
                });

                // Add active class to clicked button
                button.classList.add('active');

                // Get the tab to show
                const tabId = button.getAttribute('data-tab');

                // Find the parent tabs container
                const tabsContainer = container.closest('section');

                // Hide all tab contents in this container
                tabsContainer.querySelectorAll('.tab-content').forEach(content => {
                    content.classList.remove('active');
                });

                // Show the selected tab content
                const tabContent = tabsContainer.querySelector(`#${tabId}-tab`);
                if (tabContent) {
                    tabContent.classList.add('active');
                }
            });
        });
    });

    // Set up format toggle buttons for code examples
    const toggleContainers = document.querySelectorAll('.code-toggle-buttons');

    toggleContainers.forEach(container => {
        const toggleButtons = container.querySelectorAll('.toggle-btn');

        toggleButtons.forEach(button => {
            button.addEventListener('click', () => {
                // Remove active class from all buttons in this container
                container.querySelectorAll('.toggle-btn').forEach(btn => {
                    btn.classList.remove('active');
                });

                // Add active class to clicked button
                button.classList.add('active');

                // Get the format to show
                const format = button.getAttribute('data-format');

                // Find the parent toggle container
                const codeToggleContainer = container.closest('.code-toggle');

                // Hide all code contents in this container
                codeToggleContainer.querySelectorAll('.code-toggle-content').forEach(content => {
                    content.classList.remove('active');
                });

                // Show the selected format content
                const formatContent = codeToggleContainer.querySelector(`#simple-${format}`);
                if (formatContent) {
                    formatContent.classList.add('active');
                    }
            });
        });
    });
}

/**
 * Set up copy buttons for code blocks
 */
function setupCodeCopyButtons() {
    try {
        const copyButtons = document.querySelectorAll('.copy-btn');
        copyButtons.forEach(button => {
            button.addEventListener('click', function() {
                const codeBlock = this.nextElementSibling;
                const textToCopy = codeBlock.textContent;

                ClipboardUtils.copyToClipboard(textToCopy, this);
            });
        });
    } catch (error) {
        ErrorHandler.logError('Copy Buttons', 'Failed to set up copy buttons', false);
    }
}

/**
 * Set up conditional UI elements
 */
function setupConditionalUI() {
    try {
        // Find all toggle buttons
        const toggleButtons = document.querySelectorAll('.advanced-toggle');

        toggleButtons.forEach(toggle => {
            toggle.addEventListener('click', function() {
                const targetId = this.getAttribute('data-target');
                if (targetId) {
                    const target = document.getElementById(targetId);
                    if (target) {
                        target.classList.toggle('open');
                        this.classList.toggle('open');
                    }
                }
            });
        });
    } catch (error) {
        ErrorHandler.logError('Conditional UI', 'Failed to set up conditional UI', false);
    }
}

/**
 * Initializes Mermaid diagrams with optimized settings
 */
function initializeMermaid() {
    try {
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
    } catch (error) {
        ErrorHandler.logError('Mermaid', 'Failed to initialize Mermaid diagrams', false);
    }
}

// Initialize the application when DOM is fully loaded
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initializeApp);
} else {
    initializeApp();
}