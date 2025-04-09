// Usage tabs functionality
document.addEventListener('DOMContentLoaded', function() {
    // Initialize usage tabs
    const navItems = document.querySelectorAll('.usage-nav-item');
    const usagePanels = document.querySelectorAll('.usage-panel');

    if (navItems.length > 0) {
        // Set first tab as active by default
        navItems[0].classList.add('active');
        usagePanels[0].classList.add('active');

        // Add click event listeners to all nav items
        navItems.forEach((item, index) => {
            item.addEventListener('click', function() {
                // Remove active class from all items
                navItems.forEach(item => item.classList.remove('active'));
                usagePanels.forEach(panel => panel.classList.remove('active'));

                // Add active class to clicked item and corresponding panel
                this.classList.add('active');
                usagePanels[index].classList.add('active');
            });
        });
    }

    // Initialize copy code buttons
    const copyButtons = document.querySelectorAll('.copy-code-btn');

    copyButtons.forEach(button => {
        button.addEventListener('click', function() {
            const codeBlock = this.closest('.code-block');
            if (!codeBlock) return;

            const codeText = codeBlock.querySelector('pre')?.innerText;
            if (!codeText) return;

            // Copy text to clipboard
            navigator.clipboard.writeText(codeText)
                .then(() => {
                    // Change button icon temporarily to show success
                    const originalIcon = this.innerHTML;
                    this.innerHTML = '<i class="fas fa-check"></i>';

                    // Reset button after 2 seconds
                    setTimeout(() => {
                        this.innerHTML = originalIcon;
                    }, 2000);
                })
                .catch(err => {
                    console.error('Failed to copy text: ', err);

                    // Show error icon
                    const originalIcon = this.innerHTML;
                    this.innerHTML = '<i class="fas fa-times"></i>';

                    // Reset button after 2 seconds
                    setTimeout(() => {
                        this.innerHTML = originalIcon;
                    }, 2000);
                });
        });
    });
});