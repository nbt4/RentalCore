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
        
        // Ensure proper visibility and styling
        if (!button.classList.contains('btn-outline-') && 
            !button.classList.contains('btn-primary') && 
            !button.classList.contains('btn-secondary') &&
            !button.classList.contains('btn-success') &&
            !button.classList.contains('btn-danger') &&
            !button.classList.contains('btn-warning') &&
            !button.classList.contains('btn-info') &&
            !button.classList.contains('btn-light') &&
            !button.classList.contains('btn-dark')) {
            button.classList.add('btn-primary');
        }
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
            // Check if scanner is available
            const response = await API.request('/api/scanner/status');
            if (response.available) {
                this.setupScannerInterface();
            }
        } catch (error) {
            console.warn('Scanner not available:', error);
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

// Export utilities for use in other scripts
window.RentalCore = {
    API,
    Scanner,
    DataTable,
    Performance,
    toggleTheme
};