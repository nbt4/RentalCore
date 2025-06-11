// JobScanner Pro - JavaScript Application

// Theme management
function toggleTheme() {
    const html = document.documentElement;
    const currentTheme = html.getAttribute('data-theme');
    const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
    
    html.setAttribute('data-theme', newTheme);
    localStorage.setItem('theme', newTheme);
    
    // Update icon
    const icon = document.getElementById('theme-icon');
    if (icon) {
        icon.className = newTheme === 'dark' ? 'bi bi-sun-fill' : 'bi bi-moon-fill';
    }
}

// Initialize theme on page load
document.addEventListener('DOMContentLoaded', function() {
    const savedTheme = localStorage.getItem('theme') || 'dark';
    document.documentElement.setAttribute('data-theme', savedTheme);
    
    // Update icon
    const icon = document.getElementById('theme-icon');
    if (icon) {
        icon.className = savedTheme === 'dark' ? 'bi bi-sun-fill' : 'bi bi-moon-fill';
    }
});

// API utilities
class API {
    static async request(url, options = {}) {
        const defaultOptions = {
            headers: {
                'Content-Type': 'application/json',
            },
        };
        
        const config = { ...defaultOptions, ...options };
        
        try {
            const response = await fetch(url, config);
            const data = await response.json();
            
            if (!response.ok) {
                throw new Error(data.error || 'Request failed');
            }
            
            return data;
        } catch (error) {
            console.error('API request failed:', error);
            throw error;
        }
    }
    
    static async get(url) {
        return this.request(url);
    }
    
    static async post(url, data) {
        return this.request(url, {
            method: 'POST',
            body: JSON.stringify(data),
        });
    }
    
    static async put(url, data) {
        return this.request(url, {
            method: 'PUT',
            body: JSON.stringify(data),
        });
    }
    
    static async delete(url) {
        return this.request(url, {
            method: 'DELETE',
        });
    }
}

// Device scanner utilities
class DeviceScanner {
    constructor() {
        this.video = null;
        this.stream = null;
    }
    
    async startCamera() {
        try {
            const constraints = {
                video: {
                    facingMode: 'environment', // Use back camera if available
                    width: { ideal: 640 },
                    height: { ideal: 480 }
                }
            };
            
            this.stream = await navigator.mediaDevices.getUserMedia(constraints);
            this.video = document.getElementById('cameraVideo');
            
            if (this.video) {
                this.video.srcObject = this.stream;
                this.video.play();
                return true;
            }
            
            return false;
        } catch (error) {
            console.error('Error accessing camera:', error);
            alert('Could not access camera: ' + error.message);
            return false;
        }
    }
    
    stopCamera() {
        if (this.stream) {
            this.stream.getTracks().forEach(track => track.stop());
            this.stream = null;
        }
        
        if (this.video) {
            this.video.srcObject = null;
        }
    }
    
    // Placeholder for barcode detection - would require a barcode detection library
    async detectBarcode() {
        // This would integrate with a library like QuaggaJS or ZXing
        alert('Barcode detection not yet implemented. Please enter the device serial manually.');
    }
}

// Global scanner instance
const scanner = new DeviceScanner();

// Utility functions
function showAlert(message, type = 'info') {
    const alertHtml = `
        <div class="alert alert-${type} alert-dismissible fade show" role="alert">
            ${message}
            <button type="button" class="btn-close" data-bs-dismiss="alert"></button>
        </div>
    `;
    
    // Insert at the top of the main container
    const main = document.querySelector('main');
    if (main) {
        main.insertAdjacentHTML('afterbegin', alertHtml);
    }
}

function showLoading(container) {
    if (container) {
        container.innerHTML = '<div class="text-center"><div class="spinner"></div><p>Loading...</p></div>';
    }
}

function formatCurrency(amount) {
    return new Intl.NumberFormat('de-DE', {
        style: 'currency',
        currency: 'EUR'
    }).format(amount);
}

function formatDate(dateString) {
    const date = new Date(dateString);
    return date.toLocaleDateString('de-DE');
}

function formatDateTime(dateString) {
    const date = new Date(dateString);
    return date.toLocaleString('de-DE');
}

// Form validation
function validateForm(formId) {
    const form = document.getElementById(formId);
    if (!form) return false;
    
    let isValid = true;
    const inputs = form.querySelectorAll('input[required], select[required], textarea[required]');
    
    inputs.forEach(input => {
        if (!input.value.trim()) {
            input.classList.add('is-invalid');
            isValid = false;
        } else {
            input.classList.remove('is-invalid');
        }
    });
    
    return isValid;
}

// Search functionality
let searchTimeout;
function debounceSearch(callback, delay = 300) {
    clearTimeout(searchTimeout);
    searchTimeout = setTimeout(callback, delay);
}

// Initialize search on pages with search inputs
document.addEventListener('DOMContentLoaded', function() {
    const searchInputs = document.querySelectorAll('input[name="search"]');
    searchInputs.forEach(input => {
        input.addEventListener('input', function() {
            debounceSearch(() => {
                // Auto-submit form on search
                const form = input.closest('form');
                if (form) {
                    form.submit();
                }
            });
        });
    });
});

// Modal helpers
function openModal(modalId) {
    const modal = new bootstrap.Modal(document.getElementById(modalId));
    modal.show();
}

function closeModal(modalId) {
    const modal = bootstrap.Modal.getInstance(document.getElementById(modalId));
    if (modal) {
        modal.hide();
    }
}

// Keyboard shortcuts
document.addEventListener('keydown', function(e) {
    // Ctrl/Cmd + K for search
    if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
        e.preventDefault();
        const searchInput = document.querySelector('input[name="search"]');
        if (searchInput) {
            searchInput.focus();
        }
    }
    
    // Escape to close modals
    if (e.key === 'Escape') {
        const openModals = document.querySelectorAll('.modal.show');
        openModals.forEach(modal => {
            const bsModal = bootstrap.Modal.getInstance(modal);
            if (bsModal) {
                bsModal.hide();
            }
        });
    }
});

// Auto-save functionality for forms
function enableAutoSave(formId, endpoint, interval = 30000) {
    const form = document.getElementById(formId);
    if (!form) return;
    
    let autoSaveTimer;
    
    function autoSave() {
        const formData = new FormData(form);
        const data = Object.fromEntries(formData.entries());
        
        API.put(endpoint, data)
            .then(() => {
                showAlert('Auto-saved', 'success');
            })
            .catch(error => {
                console.error('Auto-save failed:', error);
            });
    }
    
    // Start auto-save timer
    function startAutoSave() {
        clearInterval(autoSaveTimer);
        autoSaveTimer = setInterval(autoSave, interval);
    }
    
    // Reset timer on user input
    form.addEventListener('input', startAutoSave);
    form.addEventListener('change', startAutoSave);
    
    // Initial timer
    startAutoSave();
    
    // Clean up on page unload
    window.addEventListener('beforeunload', () => {
        clearInterval(autoSaveTimer);
    });
}

// CSV export functionality
function exportToCSV(data, filename) {
    const csv = convertToCSV(data);
    const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' });
    const link = document.createElement('a');
    
    if (link.download !== undefined) {
        const url = URL.createObjectURL(blob);
        link.setAttribute('href', url);
        link.setAttribute('download', filename);
        link.style.visibility = 'hidden';
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
    }
}

function convertToCSV(data) {
    if (!data || !data.length) return '';
    
    const keys = Object.keys(data[0]);
    const csvContent = [
        keys.join(','),
        ...data.map(row => keys.map(key => {
            const value = row[key] || '';
            return typeof value === 'string' && value.includes(',') 
                ? `"${value.replace(/"/g, '""')}"` 
                : value;
        }).join(','))
    ].join('\n');
    
    return csvContent;
}

// Print functionality
function printPage() {
    window.print();
}

function printElement(elementId) {
    const element = document.getElementById(elementId);
    if (!element) return;
    
    const printWindow = window.open('', '_blank');
    printWindow.document.write(`
        <html>
            <head>
                <title>Print</title>
                <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.1.3/dist/css/bootstrap.min.css" rel="stylesheet">
                <link href="/static/css/app.css" rel="stylesheet">
            </head>
            <body>
                ${element.outerHTML}
            </body>
        </html>
    `);
    printWindow.document.close();
    printWindow.print();
    printWindow.close();
}

// QR Code generation helper
function generateQRCodeURL(data, size = 256) {
    return `/api/v1/qr?data=${encodeURIComponent(data)}&size=${size}`;
}

// Notification helpers
function showNotification(title, message, type = 'info') {
    if ('Notification' in window && Notification.permission === 'granted') {
        new Notification(title, {
            body: message,
            icon: '/static/images/icon.png'
        });
    } else {
        showAlert(`${title}: ${message}`, type);
    }
}

// Request notification permission on page load
document.addEventListener('DOMContentLoaded', function() {
    if ('Notification' in window && Notification.permission === 'default') {
        Notification.requestPermission();
    }
});

// Error handling
window.addEventListener('error', function(e) {
    console.error('JavaScript error:', e.error);
    showAlert('An error occurred. Please refresh the page if problems persist.', 'warning');
});

// PWA Management
class PWAManager {
    constructor() {
        this.deferredPrompt = null;
        this.isInstalled = false;
        this.pushSubscription = null;
        this.init();
    }
    
    async init() {
        // Check if already installed
        this.isInstalled = window.matchMedia('(display-mode: standalone)').matches || 
                           window.navigator.standalone === true;
        
        if (this.isInstalled) {
            console.log('PWA is installed');
            this.hidePWAPrompts();
        }
        
        // Register service worker
        await this.registerServiceWorker();
        
        // Setup install prompt
        this.setupInstallPrompt();
        
        // Setup push notifications
        this.setupPushNotifications();
        
        // Setup offline sync
        this.setupOfflineSync();
    }
    
    async registerServiceWorker() {
        if ('serviceWorker' in navigator) {
            try {
                const registration = await navigator.serviceWorker.register('/sw.js');
                console.log('ServiceWorker registered:', registration.scope);
                
                // Check for updates
                registration.addEventListener('updatefound', () => {
                    const newWorker = registration.installing;
                    newWorker.addEventListener('statechange', () => {
                        if (newWorker.state === 'installed' && navigator.serviceWorker.controller) {
                            this.showUpdateNotification();
                        }
                    });
                });
                
                return registration;
            } catch (error) {
                console.error('ServiceWorker registration failed:', error);
            }
        }
    }
    
    setupInstallPrompt() {
        window.addEventListener('beforeinstallprompt', (e) => {
            e.preventDefault();
            this.deferredPrompt = e;
            this.showInstallButton();
        });
        
        window.addEventListener('appinstalled', () => {
            console.log('PWA was installed');
            this.isInstalled = true;
            this.hidePWAPrompts();
            this.showNotification('App Installed', 'TS Equipment Manager has been installed successfully!');
        });
    }
    
    async installPWA() {
        if (!this.deferredPrompt) {
            this.showInstallInstructions();
            return;
        }
        
        this.deferredPrompt.prompt();
        const { outcome } = await this.deferredPrompt.userChoice;
        
        if (outcome === 'accepted') {
            console.log('User accepted PWA install prompt');
        } else {
            console.log('User dismissed PWA install prompt');
        }
        
        this.deferredPrompt = null;
    }
    
    showInstallButton() {
        // Create install button if it doesn't exist
        let installBtn = document.getElementById('pwa-install-btn');
        if (!installBtn && !this.isInstalled) {
            installBtn = document.createElement('button');
            installBtn.id = 'pwa-install-btn';
            installBtn.className = 'btn btn-primary mobile-fab';
            installBtn.innerHTML = '<i class="bi bi-download"></i>';
            installBtn.title = 'Install App';
            installBtn.onclick = () => this.installPWA();
            
            document.body.appendChild(installBtn);
        }
    }
    
    hidePWAPrompts() {
        const installBtn = document.getElementById('pwa-install-btn');
        if (installBtn) {
            installBtn.remove();
        }
    }
    
    showInstallInstructions() {
        const isIOS = /iPad|iPhone|iPod/.test(navigator.userAgent);
        const isAndroid = /Android/.test(navigator.userAgent);
        
        let instructions = 'To install this app:\n\n';
        
        if (isIOS) {
            instructions += '1. Tap the Share button\n2. Select "Add to Home Screen"\n3. Tap "Add"';
        } else if (isAndroid) {
            instructions += '1. Tap the menu button (⋮)\n2. Select "Add to Home Screen"\n3. Tap "Add"';
        } else {
            instructions += '1. Click the install button in your browser\'s address bar\n2. Or use your browser\'s menu to "Install" or "Add to Home Screen"';
        }
        
        alert(instructions);
    }
    
    async setupPushNotifications() {
        if (!('serviceWorker' in navigator) || !('PushManager' in window)) {
            console.log('Push messaging is not supported');
            return;
        }
        
        const registration = await navigator.serviceWorker.getRegistration();
        if (!registration) return;
        
        // Check if already subscribed
        this.pushSubscription = await registration.pushManager.getSubscription();
        
        if (this.pushSubscription) {
            console.log('Already subscribed to push notifications');
            await this.sendSubscriptionToServer(this.pushSubscription);
        }
    }
    
    async requestPushPermission() {
        const permission = await Notification.requestPermission();
        
        if (permission === 'granted') {
            await this.subscribeToPush();
            return true;
        } else {
            console.log('Push notification permission denied');
            return false;
        }
    }
    
    async subscribeToPush() {
        try {
            const registration = await navigator.serviceWorker.getRegistration();
            if (!registration) throw new Error('No service worker registration');
            
            // You would need VAPID keys for production
            const subscription = await registration.pushManager.subscribe({
                userVisibleOnly: true,
                applicationServerKey: null // Add your VAPID public key here
            });
            
            this.pushSubscription = subscription;
            await this.sendSubscriptionToServer(subscription);
            
            this.showNotification('Notifications Enabled', 'You will now receive important updates');
            return subscription;
        } catch (error) {
            console.error('Push subscription failed:', error);
            throw error;
        }
    }
    
    async sendSubscriptionToServer(subscription) {
        try {
            await API.post('/pwa/subscribe', subscription.toJSON());
            console.log('Push subscription sent to server');
        } catch (error) {
            console.error('Failed to send subscription to server:', error);
        }
    }
    
    setupOfflineSync() {
        // Handle online/offline events
        window.addEventListener('online', () => {
            console.log('App is online');
            this.showConnectionStatus('online');
            this.syncOfflineData();
        });
        
        window.addEventListener('offline', () => {
            console.log('App is offline');
            this.showConnectionStatus('offline');
        });
        
        // Sync on page load if online
        if (navigator.onLine) {
            this.syncOfflineData();
        }
    }
    
    async syncOfflineData() {
        try {
            const offlineActions = this.getOfflineActions();
            if (offlineActions.length === 0) return;
            
            console.log(`Syncing ${offlineActions.length} offline actions`);
            
            const response = await API.post('/pwa/sync', { actions: offlineActions });
            
            if (response.results) {
                this.processSyncResults(response.results);
            }
            
            this.clearOfflineActions();
            console.log('Offline data synced successfully');
        } catch (error) {
            console.error('Offline sync failed:', error);
        }
    }
    
    getOfflineActions() {
        const stored = localStorage.getItem('offlineActions');
        return stored ? JSON.parse(stored) : [];
    }
    
    addOfflineAction(action) {
        const actions = this.getOfflineActions();
        actions.push({
            id: Date.now(),
            ...action,
            timestamp: new Date().toISOString()
        });
        localStorage.setItem('offlineActions', JSON.stringify(actions));
    }
    
    clearOfflineActions() {
        localStorage.removeItem('offlineActions');
    }
    
    processSyncResults(results) {
        const failed = results.filter(r => r.status === 'error');
        const success = results.filter(r => r.status === 'success');
        
        if (success.length > 0) {
            this.showNotification('Sync Complete', `${success.length} actions synchronized`);
        }
        
        if (failed.length > 0) {
            console.error('Sync failures:', failed);
            this.showNotification('Sync Warning', `${failed.length} actions failed to sync`);
        }
    }
    
    showConnectionStatus(status) {
        const statusEl = document.getElementById('connection-status');
        if (statusEl) {
            statusEl.textContent = status === 'online' ? 'Connected' : 'Offline';
            statusEl.className = status === 'online' ? 'text-success' : 'text-warning';
        }
        
        // Show toast notification
        const message = status === 'online' ? 'Connection restored' : 'Working offline';
        this.showToast(message, status === 'online' ? 'success' : 'warning');
    }
    
    showUpdateNotification() {
        const updateBtn = document.createElement('button');
        updateBtn.className = 'btn btn-success btn-sm';
        updateBtn.innerHTML = '<i class="bi bi-arrow-clockwise"></i> Update Available';
        updateBtn.onclick = () => window.location.reload();
        
        const navbar = document.querySelector('.navbar .container');
        if (navbar) {
            navbar.appendChild(updateBtn);
        }
    }
    
    showNotification(title, message, type = 'info') {
        if ('Notification' in window && Notification.permission === 'granted') {
            new Notification(title, {
                body: message,
                icon: '/static/images/icon-192.png',
                badge: '/static/images/icon-192.png'
            });
        } else {
            this.showToast(`${title}: ${message}`, type);
        }
    }
    
    showToast(message, type = 'info') {
        const toast = document.createElement('div');
        toast.className = `mobile-toast mobile-toast-${type}`;
        toast.innerHTML = `
            <div class="d-flex align-items-center">
                <i class="bi bi-${this.getIconForType(type)} me-2"></i>
                <span>${message}</span>
                <button class="btn-close ms-auto" onclick="this.parentElement.parentElement.remove()"></button>
            </div>
        `;
        
        document.body.appendChild(toast);
        
        setTimeout(() => toast.classList.add('show'), 100);
        setTimeout(() => {
            toast.classList.remove('show');
            setTimeout(() => toast.remove(), 300);
        }, 3000);
    }
    
    getIconForType(type) {
        const icons = {
            success: 'check-circle',
            warning: 'exclamation-triangle',
            error: 'x-circle',
            info: 'info-circle'
        };
        return icons[type] || 'info-circle';
    }
    
    // Public API methods
    async enableNotifications() {
        return await this.requestPushPermission();
    }
    
    isOnline() {
        return navigator.onLine;
    }
    
    isAppInstalled() {
        return this.isInstalled;
    }
}

// Initialize PWA Manager
const pwaManager = new PWAManager();

// Enhanced Service worker registration (for PWA features and performance)
if ('serviceWorker' in navigator) {
    // Listen for service worker messages
    navigator.serviceWorker.addEventListener('message', event => {
        if (event.data && event.data.type === 'CACHE_UPDATED') {
            console.log('Cache updated:', event.data.url);
        }
    });
}

// Add PWA install button to navbar if supported
document.addEventListener('DOMContentLoaded', function() {
    // Add install button to navbar
    const navbarActions = document.querySelector('.navbar .header-actions, .navbar .navbar-nav');
    if (navbarActions && !pwaManager.isInstalled) {
        const installBtn = document.createElement('button');
        installBtn.className = 'btn btn-outline-light btn-sm me-2';
        installBtn.innerHTML = '<i class="bi bi-download"></i> Install';
        installBtn.title = 'Install TS Equipment Manager';
        installBtn.onclick = () => pwaManager.installPWA();
        installBtn.id = 'navbar-install-btn';
        
        navbarActions.insertBefore(installBtn, navbarActions.firstChild);
    }
    
    // Add notification permission button
    if ('Notification' in window && Notification.permission === 'default') {
        const notificationBtn = document.createElement('button');
        notificationBtn.className = 'btn btn-outline-warning btn-sm me-2';
        notificationBtn.innerHTML = '<i class="bi bi-bell"></i> Notifications';
        notificationBtn.onclick = async () => {
            const enabled = await pwaManager.enableNotifications();
            if (enabled) {
                notificationBtn.remove();
            }
        };
        
        if (navbarActions) {
            navbarActions.insertBefore(notificationBtn, navbarActions.firstChild);
        }
    }
});

// Mobile-specific enhancements
if (window.innerWidth <= 768) {
    // Add mobile fab for quick scanner access
    document.addEventListener('DOMContentLoaded', function() {
        if (window.location.pathname !== '/mobile/scanner') {
            const scanFab = document.createElement('button');
            scanFab.className = 'mobile-fab';
            scanFab.innerHTML = '<i class="bi bi-camera"></i>';
            scanFab.title = 'Quick Scan';
            scanFab.onclick = () => {
                window.location.href = '/scan/select';
            };
            
            document.body.appendChild(scanFab);
        }
    });
}