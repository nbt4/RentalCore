/**
 * RentalCore UI Enhancements
 * Production-ready UI/UX utilities for enhanced user experience
 */

class UIEnhancements {
    constructor() {
        this.init();
    }

    init() {
        this.initButtonEnhancements();
        this.initFormValidation();
        this.initAccessibility();
        this.initKeyboardShortcuts();
        this.initLoadingStates();
        this.initToasts();
        this.initConfirmations();
    }

    /**
     * Enhanced Button Interactions
     */
    initButtonEnhancements() {
        // Add loading states to all forms
        document.addEventListener('submit', (e) => {
            if (e.target.tagName === 'FORM') {
                this.setFormLoading(e.target, true);
            }
        });

        // Enhanced button click feedback
        document.addEventListener('click', (e) => {
            if (e.target.matches('.btn') && !e.target.disabled) {
                this.addClickFeedback(e.target);
            }
        });

        // Auto-disable double-click submissions
        document.querySelectorAll('form').forEach(form => {
            const submitBtn = form.querySelector('button[type="submit"], input[type="submit"]');
            if (submitBtn) {
                form.addEventListener('submit', () => {
                    setTimeout(() => {
                        submitBtn.disabled = true;
                        this.setButtonLoading(submitBtn, true);
                    }, 10);
                });
            }
        });
    }

    /**
     * Enhanced Form Validation
     */
    initFormValidation() {
        // Real-time validation for common field types
        document.querySelectorAll('input[type="email"]').forEach(input => {
            input.addEventListener('blur', () => this.validateEmail(input));
            input.addEventListener('input', () => this.clearValidation(input));
        });

        document.querySelectorAll('input[data-type="iban"]').forEach(input => {
            input.addEventListener('blur', () => this.validateIBAN(input));
            input.addEventListener('input', () => this.formatIBAN(input));
        });

        document.querySelectorAll('input[type="tel"]').forEach(input => {
            input.addEventListener('blur', () => this.validatePhone(input));
            input.addEventListener('input', () => this.formatPhone(input));
        });

        document.querySelectorAll('input[required]').forEach(input => {
            input.addEventListener('blur', () => this.validateRequired(input));
        });

        // Form change detection
        document.querySelectorAll('form').forEach(form => {
            this.initChangeDetection(form);
        });
    }

    /**
     * Accessibility Enhancements
     */
    initAccessibility() {
        // Add ARIA labels to buttons without them
        document.querySelectorAll('.btn').forEach(btn => {
            if (!btn.getAttribute('aria-label') && !btn.getAttribute('aria-labelledby')) {
                const text = btn.textContent.trim() || btn.getAttribute('title') || 'Button';
                btn.setAttribute('aria-label', text);
            }
        });

        // Enhanced focus management for modals
        document.addEventListener('shown.bs.modal', (e) => {
            const modal = e.target;
            const focusableElements = modal.querySelectorAll(
                'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
            );
            if (focusableElements.length > 0) {
                focusableElements[0].focus();
            }
        });

        // Skip to main content link
        this.addSkipLink();

        // High contrast mode detection
        if (window.matchMedia('(prefers-contrast: high)').matches) {
            document.body.classList.add('high-contrast');
        }
    }

    /**
     * Keyboard Shortcuts
     */
    initKeyboardShortcuts() {
        document.addEventListener('keydown', (e) => {
            // Ctrl/Cmd + S - Save form
            if ((e.ctrlKey || e.metaKey) && e.key === 's') {
                e.preventDefault();
                const form = document.querySelector('form');
                if (form) {
                    const submitBtn = form.querySelector('button[type="submit"]');
                    if (submitBtn && !submitBtn.disabled) {
                        submitBtn.click();
                    }
                }
            }

            // Ctrl/Cmd + N - New record (if new button exists)
            if ((e.ctrlKey || e.metaKey) && e.key === 'n') {
                const newBtn = document.querySelector('a[href*="/new"], .btn-new, .btn[data-action="new"]');
                if (newBtn) {
                    e.preventDefault();
                    newBtn.click();
                }
            }

            // Escape - Close modals
            if (e.key === 'Escape') {
                const modal = document.querySelector('.modal.show');
                if (modal) {
                    const modalInstance = bootstrap.Modal.getInstance(modal);
                    if (modalInstance) {
                        modalInstance.hide();
                    }
                }
            }

            // Alt + S - Search
            if (e.altKey && e.key === 's') {
                e.preventDefault();
                const searchInput = document.querySelector('input[type="search"], #searchInput, .search-input');
                if (searchInput) {
                    searchInput.focus();
                }
            }
        });

        // Add keyboard shortcut indicators to buttons
        this.addShortcutIndicators();
    }

    /**
     * Loading States Management
     */
    initLoadingStates() {
        // Global AJAX loading indicator
        let loadingCount = 0;
        
        const showGlobalLoading = () => {
            loadingCount++;
            if (loadingCount === 1) {
                document.body.classList.add('loading');
                this.showLoadingBar();
            }
        };

        const hideGlobalLoading = () => {
            loadingCount = Math.max(0, loadingCount - 1);
            if (loadingCount === 0) {
                document.body.classList.remove('loading');
                this.hideLoadingBar();
            }
        };

        // Hook into fetch API
        const originalFetch = window.fetch;
        window.fetch = (...args) => {
            showGlobalLoading();
            return originalFetch(...args).finally(() => {
                hideGlobalLoading();
            });
        };

        // Add loading bar to page
        this.createLoadingBar();
    }

    /**
     * Toast Notifications
     */
    initToasts() {
        // Create toast container
        this.createToastContainer();

        // Auto-show toasts from server
        const urlParams = new URLSearchParams(window.location.search);
        const message = urlParams.get('message');
        const type = urlParams.get('type') || 'info';
        
        if (message) {
            this.showToast(decodeURIComponent(message), type);
            // Clean URL
            const newUrl = window.location.pathname;
            window.history.replaceState({}, document.title, newUrl);
        }
    }

    /**
     * Enhanced Confirmations
     */
    initConfirmations() {
        document.addEventListener('click', (e) => {
            if (e.target.matches('[data-confirm]')) {
                e.preventDefault();
                const message = e.target.getAttribute('data-confirm');
                const isDelete = e.target.classList.contains('btn-danger') || 
                               e.target.getAttribute('data-method') === 'DELETE';
                
                this.showConfirmation(message, isDelete).then((confirmed) => {
                    if (confirmed) {
                        if (e.target.tagName === 'A') {
                            window.location.href = e.target.href;
                        } else if (e.target.tagName === 'BUTTON') {
                            e.target.click();
                        }
                    }
                });
            }
        });
    }

    /**
     * Utility Methods
     */
    setButtonLoading(button, loading) {
        if (loading) {
            button.classList.add('btn-loading');
            button.disabled = true;
            if (!button.querySelector('.btn-text')) {
                button.innerHTML = `<span class="btn-text">${button.innerHTML}</span>`;
            }
        } else {
            button.classList.remove('btn-loading');
            button.disabled = false;
        }
    }

    setFormLoading(form, loading) {
        if (loading) {
            form.classList.add('form-loading');
        } else {
            form.classList.remove('form-loading');
        }
    }

    addClickFeedback(button) {
        button.classList.add('btn-success-feedback');
        setTimeout(() => {
            button.classList.remove('btn-success-feedback');
        }, 600);
    }

    validateEmail(input) {
        const email = input.value.trim();
        const isValid = /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
        
        if (email && !isValid) {
            this.setFieldError(input, 'Bitte geben Sie eine gültige E-Mail-Adresse ein.');
        } else if (isValid) {
            this.setFieldSuccess(input);
        }
        
        return isValid;
    }

    validateIBAN(input) {
        const iban = input.value.replace(/\s/g, '');
        const isValid = this.isValidIBAN(iban);
        
        if (iban && !isValid) {
            this.setFieldError(input, 'Ungültige IBAN. Bitte überprüfen Sie die Eingabe.');
        } else if (isValid) {
            this.setFieldSuccess(input);
        }
        
        return isValid;
    }

    formatIBAN(input) {
        let value = input.value.replace(/\s/g, '').toUpperCase();
        value = value.replace(/(.{4})/g, '$1 ').trim();
        input.value = value;
    }

    validatePhone(input) {
        const phone = input.value.trim();
        const isValid = /^[\+]?[0-9\s\-\(\)]{7,}$/.test(phone);
        
        if (phone && !isValid) {
            this.setFieldError(input, 'Bitte geben Sie eine gültige Telefonnummer ein.');
        } else if (isValid) {
            this.setFieldSuccess(input);
        }
        
        return isValid;
    }

    formatPhone(input) {
        // Basic phone formatting for German numbers
        let value = input.value.replace(/\D/g, '');
        if (value.startsWith('49')) {
            value = '+49 ' + value.slice(2);
        } else if (value.startsWith('0')) {
            value = '+49 ' + value.slice(1);
        }
        input.value = value;
    }

    validateRequired(input) {
        const isEmpty = !input.value.trim();
        
        if (isEmpty) {
            this.setFieldError(input, 'Dieses Feld ist erforderlich.');
        } else {
            this.setFieldSuccess(input);
        }
        
        return !isEmpty;
    }

    setFieldError(input, message) {
        input.classList.remove('is-valid');
        input.classList.add('is-invalid');
        
        let feedback = input.parentNode.querySelector('.invalid-feedback');
        if (!feedback) {
            feedback = document.createElement('div');
            feedback.className = 'invalid-feedback';
            input.parentNode.appendChild(feedback);
        }
        feedback.textContent = message;
    }

    setFieldSuccess(input) {
        input.classList.remove('is-invalid');
        input.classList.add('is-valid');
        
        const feedback = input.parentNode.querySelector('.invalid-feedback');
        if (feedback) {
            feedback.textContent = '';
        }
    }

    clearValidation(input) {
        input.classList.remove('is-valid', 'is-invalid');
        const feedback = input.parentNode.querySelector('.invalid-feedback');
        if (feedback) {
            feedback.textContent = '';
        }
    }

    initChangeDetection(form) {
        const originalData = new FormData(form);
        let hasChanges = false;

        form.addEventListener('input', () => {
            hasChanges = true;
        });

        window.addEventListener('beforeunload', (e) => {
            if (hasChanges) {
                e.preventDefault();
                e.returnValue = 'Sie haben ungespeicherte Änderungen. Möchten Sie die Seite wirklich verlassen?';
            }
        });

        form.addEventListener('submit', () => {
            hasChanges = false;
        });
    }

    showToast(message, type = 'info', duration = 5000) {
        const toast = document.createElement('div');
        toast.className = `toast align-items-center text-white bg-${type} border-0`;
        toast.setAttribute('role', 'alert');
        toast.innerHTML = `
            <div class="d-flex">
                <div class="toast-body">${message}</div>
                <button type="button" class="btn-close btn-close-white me-2 m-auto" data-bs-dismiss="toast"></button>
            </div>
        `;

        const container = document.getElementById('toast-container');
        container.appendChild(toast);

        const bsToast = new bootstrap.Toast(toast, { delay: duration });
        bsToast.show();

        toast.addEventListener('hidden.bs.toast', () => {
            toast.remove();
        });
    }

    showConfirmation(message, isDangerous = false) {
        return new Promise((resolve) => {
            const modal = document.createElement('div');
            modal.className = 'modal fade';
            modal.innerHTML = `
                <div class="modal-dialog modal-dialog-centered">
                    <div class="modal-content">
                        <div class="modal-header">
                            <h5 class="modal-title">Bestätigung erforderlich</h5>
                            <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
                        </div>
                        <div class="modal-body">
                            <p>${message}</p>
                        </div>
                        <div class="modal-footer">
                            <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Abbrechen</button>
                            <button type="button" class="btn ${isDangerous ? 'btn-danger' : 'btn-primary'}" id="confirm-btn">
                                ${isDangerous ? 'Löschen' : 'Bestätigen'}
                            </button>
                        </div>
                    </div>
                </div>
            `;

            document.body.appendChild(modal);
            const bsModal = new bootstrap.Modal(modal);
            bsModal.show();

            const confirmBtn = modal.querySelector('#confirm-btn');
            confirmBtn.addEventListener('click', () => {
                resolve(true);
                bsModal.hide();
            });

            modal.addEventListener('hidden.bs.modal', () => {
                resolve(false);
                modal.remove();
            });
        });
    }

    createToastContainer() {
        if (!document.getElementById('toast-container')) {
            const container = document.createElement('div');
            container.id = 'toast-container';
            container.className = 'toast-container position-fixed bottom-0 end-0 p-3';
            container.style.zIndex = '1055';
            document.body.appendChild(container);
        }
    }

    createLoadingBar() {
        const loadingBar = document.createElement('div');
        loadingBar.id = 'global-loading-bar';
        loadingBar.style.cssText = `
            position: fixed;
            top: 0;
            left: 0;
            width: 0%;
            height: 3px;
            background: linear-gradient(90deg, #3b82f6, #1d4ed8);
            z-index: 9999;
            transition: width 0.3s ease;
            opacity: 0;
        `;
        document.body.appendChild(loadingBar);
    }

    showLoadingBar() {
        const bar = document.getElementById('global-loading-bar');
        bar.style.opacity = '1';
        bar.style.width = '70%';
    }

    hideLoadingBar() {
        const bar = document.getElementById('global-loading-bar');
        bar.style.width = '100%';
        setTimeout(() => {
            bar.style.opacity = '0';
            bar.style.width = '0%';
        }, 200);
    }

    addSkipLink() {
        if (!document.getElementById('skip-link')) {
            const skipLink = document.createElement('a');
            skipLink.id = 'skip-link';
            skipLink.href = '#main-content';
            skipLink.textContent = 'Zum Hauptinhalt springen';
            skipLink.style.cssText = `
                position: absolute;
                top: -40px;
                left: 6px;
                background: var(--accent-color);
                color: white;
                padding: 8px;
                text-decoration: none;
                border-radius: 4px;
                z-index: 10000;
                transition: top 0.3s;
            `;
            skipLink.addEventListener('focus', () => {
                skipLink.style.top = '6px';
            });
            skipLink.addEventListener('blur', () => {
                skipLink.style.top = '-40px';
            });
            document.body.insertBefore(skipLink, document.body.firstChild);
        }
    }

    addShortcutIndicators() {
        // Add keyboard shortcut hints to common buttons
        const shortcuts = {
            'button[type="submit"]': 'Ctrl+S',
            '.btn-new, a[href*="/new"]': 'Ctrl+N',
            '.search-input, input[type="search"]': 'Alt+S'
        };

        Object.entries(shortcuts).forEach(([selector, shortcut]) => {
            document.querySelectorAll(selector).forEach(element => {
                if (!element.getAttribute('title')) {
                    const currentTitle = element.getAttribute('title') || element.textContent.trim();
                    element.setAttribute('title', `${currentTitle} (${shortcut})`);
                }
            });
        });
    }

    isValidIBAN(iban) {
        if (iban.length < 15 || iban.length > 34) return false;
        
        const rearranged = iban.slice(4) + iban.slice(0, 4);
        const numeric = rearranged.replace(/[A-Z]/g, char => char.charCodeAt(0) - 55);
        
        let remainder = '';
        for (let i = 0; i < numeric.length; i++) {
            remainder = (remainder + numeric[i]) % 97;
        }
        
        return remainder === 1;
    }
}

// Initialize when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    window.uiEnhancements = new UIEnhancements();
});

// Global utility functions
window.showToast = (message, type, duration) => {
    if (window.uiEnhancements) {
        window.uiEnhancements.showToast(message, type, duration);
    }
};

window.showConfirmation = (message, isDangerous) => {
    if (window.uiEnhancements) {
        return window.uiEnhancements.showConfirmation(message, isDangerous);
    }
    return Promise.resolve(confirm(message));
};

window.setButtonLoading = (button, loading) => {
    if (window.uiEnhancements) {
        window.uiEnhancements.setButtonLoading(button, loading);
    }
};