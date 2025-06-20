// JobScanner Pro - JavaScript Application

// Enhanced Theme Management with Smooth Transitions
function toggleTheme() {
    const html = document.documentElement;
    const themeIcon = document.getElementById('theme-icon');
    const currentTheme = html.getAttribute('data-theme');
    
    // Add transition class for smooth theme switching
    document.body.classList.add('theme-transitioning');
    
    if (currentTheme === 'dark') {
        html.setAttribute('data-theme', 'light');
        if (themeIcon) themeIcon.className = 'bi bi-sun-fill';
        localStorage.setItem('theme', 'light');
    } else {
        html.setAttribute('data-theme', 'dark');
        if (themeIcon) themeIcon.className = 'bi bi-moon-fill';
        localStorage.setItem('theme', 'dark');
    }
    
    // Remove transition class after animation completes
    setTimeout(() => {
        document.body.classList.remove('theme-transitioning');
    }, 300);
}

// Initialize theme with proper loading sequence
document.addEventListener('DOMContentLoaded', function() {
    const savedTheme = localStorage.getItem('theme') || 'dark';
    const html = document.documentElement;
    const themeIcon = document.getElementById('theme-icon');
    
    html.setAttribute('data-theme', savedTheme);
    if (themeIcon) {
        themeIcon.className = savedTheme === 'light' ? 'bi bi-sun-fill' : 'bi bi-moon-fill';
    }
    
    // Initialize all interactive elements
    initializeInteractiveElements();
    initializeAnimations();
    initializeFormEnhancements();
    
    // Apply auto-enhancements after page load
    setTimeout(() => {
        autoEnhanceButtons();
        autoEnhanceBadges();
        autoEnhanceCards();
    }, 100);
});

// Interactive Elements Enhancement
function initializeInteractiveElements() {
    // Enhanced button interactions
    const buttons = document.querySelectorAll('.btn, button, input[type="submit"], input[type="button"]');
    buttons.forEach(button => {
        // Add ripple effect on click
        button.addEventListener('click', function(e) {
            const ripple = document.createElement('span');
            const rect = button.getBoundingClientRect();
            const size = Math.max(rect.width, rect.height);
            const x = e.clientX - rect.left - size / 2;
            const y = e.clientY - rect.top - size / 2;
            
            ripple.style.width = ripple.style.height = size + 'px';
            ripple.style.left = x + 'px';
            ripple.style.top = y + 'px';
            ripple.classList.add('ripple-effect');
            
            button.appendChild(ripple);
            
            setTimeout(() => {
                ripple.remove();
            }, 600);
        });
        
        // Automatically enhance button visibility
        enhanceButtonVisibility(button);
    });
    
    // Enhanced card hover effects
    const cards = document.querySelectorAll('.card');
    cards.forEach(card => {
        card.addEventListener('mouseenter', function() {
            this.style.transform = 'translateY(-2px)';
        });
        
        card.addEventListener('mouseleave', function() {
            this.style.transform = 'translateY(0)';
        });
    });
    
    // Enhanced dropdown animations
    const dropdowns = document.querySelectorAll('.dropdown-menu');
    dropdowns.forEach(dropdown => {
        dropdown.addEventListener('show.bs.dropdown', function() {
            this.classList.add('animate-fade-in');
        });
    });
}

// Animation System
function initializeAnimations() {
    // Intersection Observer for scroll animations
    const observerOptions = {
        threshold: 0.1,
        rootMargin: '0px 0px -50px 0px'
    };
    
    const observer = new IntersectionObserver((entries) => {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                entry.target.classList.add('animate-slide-up');
                observer.unobserve(entry.target);
            }
        });
    }, observerOptions);
    
    // Observe all cards and feature elements
    document.querySelectorAll('.card, .feature-card, .section-header').forEach(el => {
        observer.observe(el);
    });
    
    // Stagger animations for grid layouts
    const gridItems = document.querySelectorAll('.row .col-lg-3, .row .col-md-6, .row .col-lg-4');
    gridItems.forEach((item, index) => {
        item.style.animationDelay = `${index * 0.1}s`;
    });
}

// Form Enhancements
function initializeFormEnhancements() {
    // Enhanced form validation
    const forms = document.querySelectorAll('form');
    forms.forEach(form => {
        form.addEventListener('submit', function(e) {
            if (!form.checkValidity()) {
                e.preventDefault();
                e.stopPropagation();
                
                // Highlight first invalid field
                const firstInvalid = form.querySelector(':invalid');
                if (firstInvalid) {
                    firstInvalid.focus();
                    firstInvalid.scrollIntoView({ behavior: 'smooth', block: 'center' });
                }
            }
            form.classList.add('was-validated');
        });
    });
    
    // Enhanced input focus effects
    const inputs = document.querySelectorAll('.form-control, .form-select');
    inputs.forEach(input => {
        input.addEventListener('focus', function() {
            this.parentElement.classList.add('input-focused');
        });
        
        input.addEventListener('blur', function() {
            this.parentElement.classList.remove('input-focused');
        });
    });
    
    // Auto-save functionality for long forms
    const autoSaveForms = document.querySelectorAll('[data-auto-save]');
    autoSaveForms.forEach(form => {
        const formId = form.id || 'auto-save-form';
        
        // Load saved data
        const savedData = localStorage.getItem(`form-${formId}`);
        if (savedData) {
            try {
                const data = JSON.parse(savedData);
                Object.keys(data).forEach(key => {
                    const field = form.querySelector(`[name="${key}"]`);
                    if (field && field.type !== 'password') {
                        field.value = data[key];
                    }
                });
            } catch (e) {
                console.warn('Failed to restore form data:', e);
            }
        }
        
        // Save data on change
        const debounce = (func, wait) => {
            let timeout;
            return function executedFunction(...args) {
                const later = () => {
                    clearTimeout(timeout);
                    func(...args);
                };
                clearTimeout(timeout);
                timeout = setTimeout(later, wait);
            };
        };
        
        const saveFormData = debounce(() => {
            const formData = new FormData(form);
            const data = {};
            for (let [key, value] of formData.entries()) {
                if (form.querySelector(`[name="${key}"]`).type !== 'password') {
                    data[key] = value;
                }
            }
            localStorage.setItem(`form-${formId}`, JSON.stringify(data));
        }, 1000);
        
        form.addEventListener('input', saveFormData);
        form.addEventListener('change', saveFormData);
        
        // Clear saved data on successful submit
        form.addEventListener('submit', function() {
            localStorage.removeItem(`form-${formId}`);
        });
    });
}

// Enhanced API utilities with better error handling and loading states
class API {
    static async request(url, options = {}) {
        const defaultOptions = {
            headers: {
                'Content-Type': 'application/json',
            },
        };
        
        const config = { ...defaultOptions, ...options };
        
        // Show loading indicator
        this.showLoading();
        
        try {
            const response = await fetch(url, config);
            let data;
            
            // Handle different content types
            const contentType = response.headers.get('content-type');
            if (contentType && contentType.includes('application/json')) {
                data = await response.json();
            } else {
                data = await response.text();
            }
            
            if (!response.ok) {
                throw new Error(data.error || data || `HTTP ${response.status}: ${response.statusText}`);
            }
            
            this.hideLoading();
            return data;
        } catch (error) {
            this.hideLoading();
            this.showError(error.message || 'Network error occurred');
            throw error;
        }
    }
    
    static showLoading() {
        let loader = document.getElementById('global-loader');
        if (!loader) {
            loader = document.createElement('div');
            loader.id = 'global-loader';
            loader.className = 'global-loader';
            loader.innerHTML = `
                <div class="spinner-professional spinner-lg"></div>
                <p class="mt-2 text-muted">Loading...</p>
            `;
            document.body.appendChild(loader);
        }
        loader.style.display = 'flex';
    }
    
    static hideLoading() {
        const loader = document.getElementById('global-loader');
        if (loader) {
            loader.style.display = 'none';
        }
    }
    
    static showError(message, type = 'danger') {
        this.showNotification(message, type, 5000);
    }
    
    static showSuccess(message) {
        this.showNotification(message, 'success', 3000);
    }
    
    static showNotification(message, type = 'info', duration = 4000) {
        const notification = document.createElement('div');
        notification.className = `alert alert-${type} notification-toast animate-slide-up`;
        notification.innerHTML = `
            <div class="d-flex align-items-center">
                <i class="bi bi-${this.getIconForType(type)} me-2"></i>
                <span>${message}</span>
                <button type="button" class="btn-close ms-auto" onclick="this.parentElement.parentElement.remove()"></button>
            </div>
        `;
        
        // Create notification container if it doesn't exist
        let container = document.getElementById('notification-container');
        if (!container) {
            container = document.createElement('div');
            container.id = 'notification-container';
            container.className = 'notification-container';
            document.body.appendChild(container);
        }
        
        container.appendChild(notification);
        
        // Auto-remove after duration
        setTimeout(() => {
            if (notification.parentNode) {
                notification.classList.add('animate-fade-out');
                setTimeout(() => notification.remove(), 300);
            }
        }, duration);
    }
    
    static getIconForType(type) {
        const icons = {
            'success': 'check-circle-fill',
            'danger': 'exclamation-triangle-fill',
            'warning': 'exclamation-triangle-fill',
            'info': 'info-circle-fill',
            'primary': 'info-circle-fill'
        };
        return icons[type] || 'info-circle-fill';
    }
}

// Enhanced Scanner Integration
class Scanner {
    static async initialize() {
        try {
            // Scanner interface is always available - skip API check for now
            this.setupScannerInterface();
        } catch (error) {
            console.warn('Scanner interface setup failed:', error);
        }
    }
    
    static setupScannerInterface() {
        // Add scanner shortcuts
        document.addEventListener('keydown', (e) => {
            // Ctrl+Shift+S to open scanner
            if (e.ctrlKey && e.shiftKey && e.key === 'S') {
                e.preventDefault();
                window.location.href = '/scan/select';
            }
        });
        
        // Add floating scanner button
        const scannerButton = document.createElement('button');
        scannerButton.className = 'btn btn-primary floating-scanner-btn';
        scannerButton.innerHTML = '<i class="bi bi-qr-code-scan"></i>';
        scannerButton.title = 'Quick Scanner (Ctrl+Shift+S)';
        scannerButton.onclick = () => window.location.href = '/scan/select';
        document.body.appendChild(scannerButton);
    }
}

// Enhanced Data Tables
class DataTable {
    static enhance(tableSelector) {
        const table = document.querySelector(tableSelector);
        if (!table) return;
        
        // Add sorting capability
        const headers = table.querySelectorAll('th[data-sortable]');
        headers.forEach(header => {
            header.style.cursor = 'pointer';
            header.innerHTML += ' <i class="bi bi-arrow-down-up ms-1 text-muted"></i>';
            
            header.addEventListener('click', () => {
                this.sortTable(table, header);
            });
        });
        
        // Add search functionality
        this.addTableSearch(table);
        
        // Add pagination if needed
        this.addPagination(table);
    }
    
    static sortTable(table, header) {
        const tbody = table.querySelector('tbody');
        const rows = Array.from(tbody.querySelectorAll('tr'));
        const columnIndex = Array.from(header.parentNode.children).indexOf(header);
        const isAscending = !header.classList.contains('sort-asc');
        
        // Remove existing sort classes
        header.parentNode.querySelectorAll('th').forEach(th => {
            th.classList.remove('sort-asc', 'sort-desc');
        });
        
        // Add new sort class
        header.classList.add(isAscending ? 'sort-asc' : 'sort-desc');
        
        // Sort rows
        rows.sort((a, b) => {
            const aVal = a.cells[columnIndex].textContent.trim();
            const bVal = b.cells[columnIndex].textContent.trim();
            
            // Try to parse as numbers
            const aNum = parseFloat(aVal);
            const bNum = parseFloat(bVal);
            
            if (!isNaN(aNum) && !isNaN(bNum)) {
                return isAscending ? aNum - bNum : bNum - aNum;
            }
            
            // String comparison
            return isAscending ? aVal.localeCompare(bVal) : bVal.localeCompare(aVal);
        });
        
        // Reorder rows in DOM
        rows.forEach(row => tbody.appendChild(row));
    }
    
    static addTableSearch(table) {
        const searchContainer = document.createElement('div');
        searchContainer.className = 'table-search mb-3';
        searchContainer.innerHTML = `
            <div class="input-group">
                <span class="input-group-text">
                    <i class="bi bi-search"></i>
                </span>
                <input type="text" class="form-control" placeholder="Search table...">
            </div>
        `;
        
        table.parentNode.insertBefore(searchContainer, table);
        
        const searchInput = searchContainer.querySelector('input');
        searchInput.addEventListener('input', (e) => {
            this.filterTable(table, e.target.value);
        });
    }
    
    static filterTable(table, searchTerm) {
        const tbody = table.querySelector('tbody');
        const rows = tbody.querySelectorAll('tr');
        
        rows.forEach(row => {
            const text = row.textContent.toLowerCase();
            const matches = text.includes(searchTerm.toLowerCase());
            row.style.display = matches ? '' : 'none';
        });
    }
    
    static addPagination(table) {
        const tbody = table.querySelector('tbody');
        const rows = Array.from(tbody.querySelectorAll('tr'));
        const rowsPerPage = 10;
        
        if (rows.length <= rowsPerPage) return;
        
        const totalPages = Math.ceil(rows.length / rowsPerPage);
        let currentPage = 1;
        
        const paginationContainer = document.createElement('nav');
        paginationContainer.className = 'mt-3';
        paginationContainer.innerHTML = `
            <ul class="pagination justify-content-center">
                <li class="page-item">
                    <a class="page-link" href="#" data-page="prev">Previous</a>
                </li>
                ${Array.from({ length: totalPages }, (_, i) => 
                    `<li class="page-item ${i === 0 ? 'active' : ''}">
                        <a class="page-link" href="#" data-page="${i + 1}">${i + 1}</a>
                    </li>`
                ).join('')}
                <li class="page-item">
                    <a class="page-link" href="#" data-page="next">Next</a>
                </li>
            </ul>
        `;
        
        table.parentNode.appendChild(paginationContainer);
        
        const showPage = (page) => {
            const start = (page - 1) * rowsPerPage;
            const end = start + rowsPerPage;
            
            rows.forEach((row, index) => {
                row.style.display = (index >= start && index < end) ? '' : 'none';
            });
            
            // Update pagination active state
            paginationContainer.querySelectorAll('.page-item').forEach(item => {
                item.classList.remove('active');
            });
            paginationContainer.querySelector(`[data-page="${page}"]`).parentNode.classList.add('active');
        };
        
        paginationContainer.addEventListener('click', (e) => {
            e.preventDefault();
            const pageData = e.target.dataset.page;
            
            if (pageData === 'prev' && currentPage > 1) {
                currentPage--;
            } else if (pageData === 'next' && currentPage < totalPages) {
                currentPage++;
            } else if (!isNaN(pageData)) {
                currentPage = parseInt(pageData);
            }
            
            showPage(currentPage);
        });
        
        showPage(1);
    }
}

// Performance monitoring
class Performance {
    static monitor() {
        // Monitor page load performance
        window.addEventListener('load', () => {
            const perfData = performance.getEntriesByType('navigation')[0];
            if (perfData) {
                console.log('Page Load Performance:', {
                    loadTime: perfData.loadEventEnd - perfData.loadEventStart,
                    domReady: perfData.domContentLoadedEventEnd - perfData.domContentLoadedEventStart,
                    totalTime: perfData.loadEventEnd - perfData.navigationStart
                });
            }
        });
        
        // Monitor API response times
        const originalFetch = window.fetch;
        window.fetch = function(...args) {
            const startTime = performance.now();
            return originalFetch.apply(this, args).then(response => {
                const endTime = performance.now();
                console.log(`API Request to ${args[0]} took ${endTime - startTime}ms`);
                return response;
            });
        };
    }
}

// Initialize all enhancements
document.addEventListener('DOMContentLoaded', function() {
    // Initialize performance monitoring
    Performance.monitor();
    
    // Initialize scanner
    Scanner.initialize();
    
    // Initialize device categories
    DeviceCategories.initialize();
    
    // Enhance existing tables
    document.querySelectorAll('table.table').forEach(table => {
        DataTable.enhance(`#${table.id || 'table'}`);
    });
    
    // Initialize tooltips and popovers
    const tooltipTriggerList = [].slice.call(document.querySelectorAll('[data-bs-toggle="tooltip"]'));
    tooltipTriggerList.map(function (tooltipTriggerEl) {
        return new bootstrap.Tooltip(tooltipTriggerEl);
    });
    
    const popoverTriggerList = [].slice.call(document.querySelectorAll('[data-bs-toggle="popover"]'));
    popoverTriggerList.map(function (popoverTriggerEl) {
        return new bootstrap.Popover(popoverTriggerEl);
    });
});

// Device Category Management
class DeviceCategories {
    static initialize() {
        // Add enhanced collapse behavior for device categories
        this.initializeCategoryCollapse();
        this.addExpandCollapseAllButtons();
        this.enhanceCollapseButtons();
    }
    
    static initializeCategoryCollapse() {
        // Add event listeners for category collapse events
        document.querySelectorAll('[data-bs-toggle="collapse"]').forEach(button => {
            const target = document.querySelector(button.getAttribute('data-bs-target'));
            if (target) {
                target.addEventListener('show.bs.collapse', () => {
                    const icon = button.querySelector('i');
                    if (icon) {
                        icon.classList.remove('bi-chevron-down');
                        icon.classList.add('bi-chevron-up');
                    }
                });
                
                target.addEventListener('hide.bs.collapse', () => {
                    const icon = button.querySelector('i');
                    if (icon) {
                        icon.classList.remove('bi-chevron-up');
                        icon.classList.add('bi-chevron-down');
                    }
                });
            }
        });
    }
    
    static addExpandCollapseAllButtons() {
        // Check if we're on the categorized devices page
        const categorizedContent = document.querySelector('.card-header.bg-primary');
        if (!categorizedContent) return;
        
        // Create expand/collapse all buttons
        const buttonsContainer = document.createElement('div');
        buttonsContainer.className = 'mb-3 d-flex gap-2';
        buttonsContainer.innerHTML = `
            <button type="button" class="btn btn-sm btn-outline-primary" id="expand-all-categories">
                <i class="bi bi-arrows-expand"></i> Expand All
            </button>
            <button type="button" class="btn btn-sm btn-outline-secondary" id="collapse-all-categories">
                <i class="bi bi-arrows-collapse"></i> Collapse All
            </button>
        `;
        
        // Insert before the first category card
        const firstCard = document.querySelector('.card.mb-3');
        if (firstCard) {
            firstCard.parentNode.insertBefore(buttonsContainer, firstCard);
        }
        
        // Add event listeners
        document.getElementById('expand-all-categories')?.addEventListener('click', () => {
            this.expandAllCategories();
        });
        
        document.getElementById('collapse-all-categories')?.addEventListener('click', () => {
            this.collapseAllCategories();
        });
    }
    
    static expandAllCategories() {
        // Expand all category and subcategory collapses
        document.querySelectorAll('.collapse').forEach(collapse => {
            const bsCollapse = new bootstrap.Collapse(collapse, { show: true });
        });
    }
    
    static collapseAllCategories() {
        // Collapse all category and subcategory collapses
        document.querySelectorAll('.collapse.show').forEach(collapse => {
            const bsCollapse = new bootstrap.Collapse(collapse, { hide: true });
        });
    }
    
    static enhanceCollapseButtons() {
        // Add smooth animations and better visual feedback
        document.querySelectorAll('[data-bs-toggle="collapse"]').forEach(button => {
            button.addEventListener('click', function() {
                // Add click animation
                this.style.transform = 'scale(0.95)';
                setTimeout(() => {
                    this.style.transform = 'scale(1)';
                }, 150);
            });
        });
    }
    
    static addDeviceCountAnimations() {
        // Animate device count badges when categories are expanded
        const countBadges = document.querySelectorAll('.badge');
        countBadges.forEach(badge => {
            if (badge.textContent.includes('devices')) {
                badge.style.transition = 'all 0.3s ease';
                badge.addEventListener('mouseenter', function() {
                    this.style.transform = 'scale(1.1)';
                });
                badge.addEventListener('mouseleave', function() {
                    this.style.transform = 'scale(1)';
                });
            }
        });
    }
}

// Export utilities for use in other scripts
// Auto-Enhancement Functions
function enhanceButtonVisibility(button) {
    // Skip if already enhanced
    if (button.classList.contains('btn-enhanced-visibility') || 
        button.classList.contains('btn-outline-primary-enhanced') ||
        button.classList.contains('btn-outline-secondary-enhanced') ||
        button.classList.contains('btn-outline-success-enhanced') ||
        button.classList.contains('btn-outline-danger-enhanced') ||
        button.classList.contains('btn-outline-warning-enhanced') ||
        button.classList.contains('btn-outline-info-enhanced')) {
        return;
    }
    
    // Add enhanced visibility class
    button.classList.add('btn-enhanced-visibility');
    
    // Enhance outline buttons specifically
    if (button.classList.contains('btn-outline-primary')) {
        button.classList.remove('btn-outline-primary');
        button.classList.add('btn-outline-primary-enhanced');
    } else if (button.classList.contains('btn-outline-secondary')) {
        button.classList.remove('btn-outline-secondary');
        button.classList.add('btn-outline-secondary-enhanced');
    } else if (button.classList.contains('btn-outline-success')) {
        button.classList.remove('btn-outline-success');
        button.classList.add('btn-outline-success-enhanced');
    } else if (button.classList.contains('btn-outline-danger')) {
        button.classList.remove('btn-outline-danger');
        button.classList.add('btn-outline-danger-enhanced');
    } else if (button.classList.contains('btn-outline-warning')) {
        button.classList.remove('btn-outline-warning');
        button.classList.add('btn-outline-warning-enhanced');
    } else if (button.classList.contains('btn-outline-info')) {
        button.classList.remove('btn-outline-info');
        button.classList.add('btn-outline-info-enhanced');
    }
    
    // Ensure proper base styling for non-outlined buttons
    if (!button.classList.toString().includes('btn-outline-') && 
        !button.classList.contains('btn-primary') && 
        !button.classList.contains('btn-secondary') &&
        !button.classList.contains('btn-success') &&
        !button.classList.contains('btn-danger') &&
        !button.classList.contains('btn-warning') &&
        !button.classList.contains('btn-info') &&
        !button.classList.contains('btn-light') &&
        !button.classList.contains('btn-dark') &&
        button.classList.contains('btn')) {
        button.classList.add('btn-primary');
    }
}

function autoEnhanceButtons() {
    // Find all buttons that need enhancement
    const buttons = document.querySelectorAll('.btn, button, input[type="submit"], input[type="button"]');
    buttons.forEach(button => {
        enhanceButtonVisibility(button);
    });
}

function autoEnhanceBadges() {
    // Enhance all badges for better visibility
    const badges = document.querySelectorAll('.badge');
    badges.forEach(badge => {
        if (!badge.classList.contains('status-badge')) {
            // Add status badge class for enhanced styling
            if (badge.textContent.toLowerCase().includes('free') || 
                badge.textContent.toLowerCase().includes('available')) {
                badge.classList.add('status-free');
            } else if (badge.textContent.toLowerCase().includes('rented') ||
                      badge.textContent.toLowerCase().includes('assigned')) {
                badge.classList.add('status-rented');
            } else if (badge.textContent.toLowerCase().includes('maintenance')) {
                badge.classList.add('status-maintenance');
            }
        }
    });
}

function autoEnhanceCards() {
    // Enhance all cards for better interactivity
    const cards = document.querySelectorAll('.card');
    cards.forEach(card => {
        if (!card.classList.contains('feature-card')) {
            // Add hover effects
            card.addEventListener('mouseenter', function() {
                this.style.transform = 'translateY(-2px)';
                this.style.transition = 'all 0.3s ease';
            });
            
            card.addEventListener('mouseleave', function() {
                this.style.transform = 'translateY(0)';
            });
        }
    });
}

// Enhanced search functionality
function enhanceSearchInputs() {
    const searchInputs = document.querySelectorAll('input[type="search"], input[placeholder*="search"], input[placeholder*="Search"]');
    searchInputs.forEach(input => {
        if (!input.parentElement.classList.contains('search-enhanced')) {
            input.parentElement.classList.add('search-enhanced');
        }
    });
}

// Navigation active state management
function updateNavigationActiveState() {
    const currentPath = window.location.pathname;
    const navLinks = document.querySelectorAll('.navbar-nav .nav-link');
    
    navLinks.forEach(link => {
        link.classList.remove('active');
        const href = link.getAttribute('href');
        
        // Exact match or path starts with href (except root)
        if (href === currentPath || 
            (currentPath.startsWith(href) && href !== '/' && href.length > 1)) {
            link.classList.add('active');
        }
        
        // Special handling for home page
        if (currentPath === '/' && href === '/') {
            link.classList.add('active');
        }
    });
}

// Light theme optimizations
function optimizeLightTheme() {
    const theme = document.documentElement.getAttribute('data-theme');
    if (theme === 'light') {
        // Apply light theme specific enhancements
        document.body.style.setProperty('--text-primary-override', '#2d3748');
        document.body.style.setProperty('--text-secondary-override', '#4a5568');
        document.body.style.setProperty('--bg-primary-override', '#ffffff');
        document.body.style.setProperty('--bg-secondary-override', '#f7fafc');
    } else {
        // Reset overrides for dark theme
        document.body.style.removeProperty('--text-primary-override');
        document.body.style.removeProperty('--text-secondary-override');
        document.body.style.removeProperty('--bg-primary-override');
        document.body.style.removeProperty('--bg-secondary-override');
    }
}

// Performance optimizations
function performanceOptimizations() {
    // Lazy load images
    const images = document.querySelectorAll('img[data-src]');
    const imageObserver = new IntersectionObserver((entries, observer) => {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                const img = entry.target;
                img.src = img.dataset.src;
                img.removeAttribute('data-src');
                imageObserver.unobserve(img);
            }
        });
    });
    
    images.forEach(img => imageObserver.observe(img));
    
    // Optimize table rendering for large datasets
    const largeTables = document.querySelectorAll('table tbody tr:nth-child(n+50)');
    if (largeTables.length > 0) {
        // Implement virtual scrolling for large tables
        console.log('Large table detected, consider implementing virtual scrolling');
    }
}

// Force Theme Consistency across all pages
function forceThemeConsistency() {
    // Ensure data-theme attribute exists
    if (!document.documentElement.hasAttribute('data-theme')) {
        const savedTheme = localStorage.getItem('theme') || 'dark';
        document.documentElement.setAttribute('data-theme', savedTheme);
    }
    
    // Update theme icon if it exists
    const themeIcon = document.getElementById('theme-icon');
    if (themeIcon) {
        const currentTheme = document.documentElement.getAttribute('data-theme');
        themeIcon.className = currentTheme === 'light' ? 'bi bi-sun-fill' : 'bi bi-moon-fill';
    }
    
    // Force enhance all elements immediately
    setTimeout(() => {
        autoEnhanceButtons();
        autoEnhanceBadges();
        autoEnhanceCards();
        updateNavigationActiveState();
        enhanceSearchInputs();
    }, 50);
}

// Page Load Handler - Multiple attempts to ensure consistency
document.addEventListener('DOMContentLoaded', function() {
    forceThemeConsistency();
    optimizeLightTheme();
    performanceOptimizations();
    
    // Re-apply enhancements multiple times to catch dynamic content
    setTimeout(forceThemeConsistency, 100);
    setTimeout(forceThemeConsistency, 300);
    setTimeout(forceThemeConsistency, 500);
    
    // Watch for theme changes
    const observer = new MutationObserver(mutations => {
        mutations.forEach(mutation => {
            if (mutation.type === 'attributes' && mutation.attributeName === 'data-theme') {
                optimizeLightTheme();
                forceThemeConsistency();
            }
        });
    });
    
    observer.observe(document.documentElement, {
        attributes: true,
        attributeFilter: ['data-theme']
    });
});

// Page visibility change handler
document.addEventListener('visibilitychange', function() {
    if (!document.hidden) {
        setTimeout(forceThemeConsistency, 100);
    }
});

// Window focus handler
window.addEventListener('focus', function() {
    setTimeout(forceThemeConsistency, 100);
});

window.RentalCore = {
    API,
    Scanner,
    DataTable,
    Performance,
    toggleTheme,
    enhanceButtonVisibility,
    autoEnhanceButtons,
    autoEnhanceBadges,
    autoEnhanceCards,
    forceThemeConsistency,
    updateNavigationActiveState
};