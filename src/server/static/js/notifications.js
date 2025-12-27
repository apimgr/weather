/**
 * Notification Manager - WebUI Notification System
 * Handles toast, banner, and notification center
 * TEMPLATE.md Part 25 compliant
 */

class NotificationManager {
    constructor() {
        // WebSocket connection
        this.ws = null;
        this.wsReconnectTimer = null;
        this.wsReconnectAttempts = 0;
        this.wsMaxReconnectAttempts = 10;

        // Containers
        this.toastContainer = null;
        this.bannerContainer = null;
        this.centerDropdown = null;
        this.bellIcon = null;
        this.badge = null;

        // State
        this.preferences = null;
        this.unreadCount = 0;
        this.isUserContext = false;
        this.isAdminContext = false;

        // Audio
        this.notificationSound = null;

        // Initialize
        this.init();
    }

    /**
     * Initialize the notification manager
     */
    async init() {
        // Detect context (user or admin)
        this.detectContext();

        // Create DOM elements
        this.createContainers();

        // Load preferences
        await this.loadPreferences();

        // Load initial notifications
        await this.loadNotifications();

        // Connect WebSocket
        this.connectWebSocket();

        // Setup event listeners
        this.setupEventListeners();

        // Initialize notification sound
        this.initSound();

        console.log('[NotificationManager] Initialized');
    }

    /**
     * Detect if we're in user or admin context
     */
    detectContext() {
        // Check URL path
        const path = window.location.pathname;
        if (path.startsWith('/admin')) {
            this.isAdminContext = true;
            console.log('[NotificationManager] Admin context detected');
        } else {
            this.isUserContext = true;
            console.log('[NotificationManager] User context detected');
        }
    }

    /**
     * Create notification containers in DOM
     */
    createContainers() {
        // Toast container (top-right)
        this.toastContainer = document.createElement('div');
        this.toastContainer.id = 'notification-toast-container';
        this.toastContainer.className = 'notification-toast-container';
        document.body.appendChild(this.toastContainer);

        // Banner container (top of page, after nav)
        this.bannerContainer = document.createElement('div');
        this.bannerContainer.id = 'notification-banner-container';
        this.bannerContainer.className = 'notification-banner-container';

        // Insert after nav or at top of main content
        const nav = document.querySelector('nav');
        if (nav && nav.nextSibling) {
            nav.parentNode.insertBefore(this.bannerContainer, nav.nextSibling);
        } else {
            document.body.insertBefore(this.bannerContainer, document.body.firstChild);
        }

        // Find bell icon and badge in nav (should be added by template)
        this.bellIcon = document.getElementById('notification-bell');
        this.badge = document.getElementById('notification-badge');

        // Create notification center dropdown if bell icon exists
        if (this.bellIcon) {
            this.createNotificationCenter();
        }
    }

    /**
     * Create notification center dropdown
     */
    createNotificationCenter() {
        this.centerDropdown = document.createElement('div');
        this.centerDropdown.id = 'notification-center';
        this.centerDropdown.className = 'notification-center hidden';
        this.centerDropdown.innerHTML = `
            <div class="notification-center-header">
                <h3>Notifications</h3>
                <button class="notification-center-mark-all-read" title="Mark all as read">
                    <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
                        <path d="M13.78 4.22a.75.75 0 010 1.06l-7.25 7.25a.75.75 0 01-1.06 0L2.22 9.28a.75.75 0 011.06-1.06L6 10.94l6.72-6.72a.75.75 0 011.06 0z"/>
                    </svg>
                </button>
            </div>
            <div class="notification-center-list"></div>
            <div class="notification-center-footer">
                <a href="${this.isAdminContext ? '/admin' : ''}/notifications">View All</a>
            </div>
        `;

        // Insert dropdown after bell icon
        this.bellIcon.parentNode.insertBefore(this.centerDropdown, this.bellIcon.nextSibling);

        // Setup bell icon click handler
        this.bellIcon.addEventListener('click', (e) => {
            e.preventDefault();
            e.stopPropagation();
            this.toggleNotificationCenter();
        });

        // Setup mark all as read button
        const markAllBtn = this.centerDropdown.querySelector('.notification-center-mark-all-read');
        markAllBtn.addEventListener('click', () => this.markAllAsRead());

        // Close dropdown when clicking outside
        document.addEventListener('click', (e) => {
            if (!this.centerDropdown.contains(e.target) && !this.bellIcon.contains(e.target)) {
                this.hideNotificationCenter();
            }
        });
    }

    /**
     * Load user preferences from API
     */
    async loadPreferences() {
        try {
            const endpoint = this.isAdminContext
                ? '/api/v1/admin/notifications/preferences'
                : '/api/v1/user/notifications/preferences';

            const response = await fetch(endpoint, {
                credentials: 'include'
            });

            if (response.ok) {
                this.preferences = await response.json();
                console.log('[NotificationManager] Preferences loaded', this.preferences);
            } else {
                // Use defaults if preferences not found
                this.preferences = {
                    enable_toast: true,
                    enable_banner: true,
                    enable_center: true,
                    enable_sound: false,
                    toast_duration_success: 5,
                    toast_duration_info: 5,
                    toast_duration_warning: 10
                };
                console.log('[NotificationManager] Using default preferences');
            }
        } catch (error) {
            console.error('[NotificationManager] Failed to load preferences:', error);
            // Use defaults on error
            this.preferences = {
                enable_toast: true,
                enable_banner: true,
                enable_center: true,
                enable_sound: false,
                toast_duration_success: 5,
                toast_duration_info: 5,
                toast_duration_warning: 10
            };
        }
    }

    /**
     * Load initial notifications from API
     */
    async loadNotifications() {
        try {
            // Get unread count
            await this.updateUnreadCount();

            // Get unread notifications for notification center
            const endpoint = this.isAdminContext
                ? '/api/v1/admin/notifications/unread'
                : '/api/v1/user/notifications/unread';

            const response = await fetch(endpoint, {
                credentials: 'include'
            });

            if (response.ok) {
                const data = await response.json();
                this.updateNotificationCenter(data.notifications || []);
            }
        } catch (error) {
            console.error('[NotificationManager] Failed to load notifications:', error);
        }
    }

    /**
     * Update unread count badge
     */
    async updateUnreadCount() {
        try {
            const endpoint = this.isAdminContext
                ? '/api/v1/admin/notifications/count'
                : '/api/v1/user/notifications/count';

            const response = await fetch(endpoint, {
                credentials: 'include'
            });

            if (response.ok) {
                const data = await response.json();
                this.unreadCount = data.count || 0;
                this.updateBadge();
            }
        } catch (error) {
            console.error('[NotificationManager] Failed to update unread count:', error);
        }
    }

    /**
     * Update badge display
     */
    updateBadge() {
        if (!this.badge) return;

        if (this.unreadCount > 0) {
            this.badge.textContent = this.unreadCount > 99 ? '99+' : this.unreadCount;
            this.badge.classList.remove('hidden');
        } else {
            this.badge.classList.add('hidden');
        }
    }

    /**
     * Connect to WebSocket for real-time notifications
     */
    connectWebSocket() {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/ws/notifications`;

        console.log('[NotificationManager] Connecting to WebSocket:', wsUrl);

        try {
            this.ws = new WebSocket(wsUrl);

            this.ws.onopen = () => {
                console.log('[NotificationManager] WebSocket connected');
                this.wsReconnectAttempts = 0;
                if (this.wsReconnectTimer) {
                    clearTimeout(this.wsReconnectTimer);
                    this.wsReconnectTimer = null;
                }
            };

            this.ws.onmessage = (event) => {
                this.handleWebSocketMessage(event);
            };

            this.ws.onerror = (error) => {
                console.error('[NotificationManager] WebSocket error:', error);
            };

            this.ws.onclose = () => {
                console.log('[NotificationManager] WebSocket closed');
                this.ws = null;
                this.scheduleReconnect();
            };
        } catch (error) {
            console.error('[NotificationManager] WebSocket connection failed:', error);
            this.scheduleReconnect();
        }
    }

    /**
     * Schedule WebSocket reconnection
     */
    scheduleReconnect() {
        if (this.wsReconnectAttempts >= this.wsMaxReconnectAttempts) {
            console.error('[NotificationManager] Max reconnect attempts reached');
            return;
        }

        this.wsReconnectAttempts++;
        const delay = Math.min(1000 * Math.pow(2, this.wsReconnectAttempts), 30000);

        console.log(`[NotificationManager] Reconnecting in ${delay}ms (attempt ${this.wsReconnectAttempts})`);

        this.wsReconnectTimer = setTimeout(() => {
            this.connectWebSocket();
        }, delay);
    }

    /**
     * Handle WebSocket message
     */
    handleWebSocketMessage(event) {
        try {
            const message = JSON.parse(event.data);

            if (message.type === 'notification') {
                this.handleNewNotification(message.data);
            } else if (message.type === 'ping') {
                // Send pong response
                if (this.ws && this.ws.readyState === WebSocket.OPEN) {
                    this.ws.send(JSON.stringify({ type: 'pong' }));
                }
            }
        } catch (error) {
            console.error('[NotificationManager] Failed to parse WebSocket message:', error);
        }
    }

    /**
     * Handle new notification from WebSocket
     */
    handleNewNotification(notification) {
        console.log('[NotificationManager] New notification received:', notification);

        // Update unread count
        this.unreadCount++;
        this.updateBadge();

        // Display based on display type and preferences
        if (notification.display === 'toast' && this.preferences.enable_toast) {
            this.showToast(notification);
        } else if (notification.display === 'banner' && this.preferences.enable_banner) {
            this.showBanner(notification);
        }

        // Always add to notification center if enabled
        if (this.preferences.enable_center) {
            this.addToNotificationCenter(notification);
        }

        // Play sound if enabled
        if (this.preferences.enable_sound) {
            this.playNotificationSound();
        }
    }

    /**
     * Show toast notification
     */
    showToast(notification) {
        const toast = document.createElement('div');
        toast.className = `notification-toast notification-${notification.type}`;
        toast.dataset.id = notification.id;

        toast.innerHTML = `
            <div class="notification-toast-icon">
                ${this.getNotificationIcon(notification.type)}
            </div>
            <div class="notification-toast-content">
                <div class="notification-toast-title">${this.escapeHtml(notification.title)}</div>
                <div class="notification-toast-message">${this.escapeHtml(notification.message)}</div>
                ${notification.action ? `
                    <a href="${this.escapeHtml(notification.action.url)}" class="notification-toast-action">
                        ${this.escapeHtml(notification.action.label)}
                    </a>
                ` : ''}
            </div>
            <button class="notification-toast-close" aria-label="Close">×</button>
        `;

        // Close button handler
        const closeBtn = toast.querySelector('.notification-toast-close');
        closeBtn.addEventListener('click', () => {
            this.dismissToast(toast, notification.id);
        });

        // Add to container
        this.toastContainer.appendChild(toast);

        // Trigger animation
        setTimeout(() => toast.classList.add('show'), 10);

        // Auto-dismiss based on type (except error)
        if (notification.type !== 'error') {
            const duration = this.getToastDuration(notification.type);
            setTimeout(() => {
                this.dismissToast(toast, notification.id);
            }, duration * 1000);
        }
    }

    /**
     * Get toast duration based on type
     */
    getToastDuration(type) {
        switch (type) {
            case 'success':
                return this.preferences.toast_duration_success || 5;
            case 'info':
                return this.preferences.toast_duration_info || 5;
            case 'warning':
                return this.preferences.toast_duration_warning || 10;
            default:
                return 5;
        }
    }

    /**
     * Dismiss toast notification
     */
    async dismissToast(toastElement, notificationId) {
        toastElement.classList.remove('show');

        // Mark as dismissed in API
        try {
            const endpoint = this.isAdminContext
                ? `/api/v1/admin/notifications/${notificationId}/dismiss`
                : `/api/v1/user/notifications/${notificationId}/dismiss`;

            await fetch(endpoint, {
                method: 'PATCH',
                credentials: 'include'
            });
        } catch (error) {
            console.error('[NotificationManager] Failed to dismiss notification:', error);
        }

        // Remove from DOM after animation
        setTimeout(() => {
            toastElement.remove();
        }, 300);
    }

    /**
     * Show banner notification
     */
    showBanner(notification) {
        const banner = document.createElement('div');
        banner.className = `notification-banner notification-${notification.type}`;
        banner.dataset.id = notification.id;

        banner.innerHTML = `
            <div class="notification-banner-icon">
                ${this.getNotificationIcon(notification.type)}
            </div>
            <div class="notification-banner-content">
                <div class="notification-banner-title">${this.escapeHtml(notification.title)}</div>
                <div class="notification-banner-message">${this.escapeHtml(notification.message)}</div>
                ${notification.action ? `
                    <a href="${this.escapeHtml(notification.action.url)}" class="notification-banner-action">
                        ${this.escapeHtml(notification.action.label)}
                    </a>
                ` : ''}
            </div>
            <button class="notification-banner-close" aria-label="Close">×</button>
        `;

        // Close button handler
        const closeBtn = banner.querySelector('.notification-banner-close');
        closeBtn.addEventListener('click', () => {
            this.dismissBanner(banner, notification.id);
        });

        // Add to container
        this.bannerContainer.appendChild(banner);

        // Trigger animation
        setTimeout(() => banner.classList.add('show'), 10);
    }

    /**
     * Dismiss banner notification
     */
    async dismissBanner(bannerElement, notificationId) {
        bannerElement.classList.remove('show');

        // Mark as dismissed in API
        try {
            const endpoint = this.isAdminContext
                ? `/api/v1/admin/notifications/${notificationId}/dismiss`
                : `/api/v1/user/notifications/${notificationId}/dismiss`;

            await fetch(endpoint, {
                method: 'PATCH',
                credentials: 'include'
            });
        } catch (error) {
            console.error('[NotificationManager] Failed to dismiss notification:', error);
        }

        // Remove from DOM after animation
        setTimeout(() => {
            bannerElement.remove();
        }, 300);
    }

    /**
     * Add notification to notification center
     */
    addToNotificationCenter(notification) {
        if (!this.centerDropdown) return;

        const list = this.centerDropdown.querySelector('.notification-center-list');
        const item = this.createNotificationCenterItem(notification);

        // Add to top of list
        list.insertBefore(item, list.firstChild);

        // Limit to 20 items in center
        const items = list.querySelectorAll('.notification-center-item');
        if (items.length > 20) {
            items[items.length - 1].remove();
        }
    }

    /**
     * Update notification center with notifications
     */
    updateNotificationCenter(notifications) {
        if (!this.centerDropdown) return;

        const list = this.centerDropdown.querySelector('.notification-center-list');
        list.innerHTML = '';

        if (notifications.length === 0) {
            list.innerHTML = '<div class="notification-center-empty">No new notifications</div>';
            return;
        }

        notifications.forEach(notification => {
            const item = this.createNotificationCenterItem(notification);
            list.appendChild(item);
        });
    }

    /**
     * Create notification center item element
     */
    createNotificationCenterItem(notification) {
        const item = document.createElement('div');
        item.className = `notification-center-item notification-${notification.type} ${notification.read ? 'read' : 'unread'}`;
        item.dataset.id = notification.id;

        item.innerHTML = `
            <div class="notification-center-item-icon">
                ${this.getNotificationIcon(notification.type)}
            </div>
            <div class="notification-center-item-content">
                <div class="notification-center-item-title">${this.escapeHtml(notification.title)}</div>
                <div class="notification-center-item-message">${this.escapeHtml(notification.message)}</div>
                <div class="notification-center-item-time">${this.formatTimeAgo(notification.created_at)}</div>
            </div>
            <div class="notification-center-item-actions">
                ${!notification.read ? '<button class="mark-read" title="Mark as read">✓</button>' : ''}
                <button class="delete" title="Delete">×</button>
            </div>
        `;

        // Mark as read handler
        const markReadBtn = item.querySelector('.mark-read');
        if (markReadBtn) {
            markReadBtn.addEventListener('click', async (e) => {
                e.stopPropagation();
                await this.markAsRead(notification.id);
                item.classList.add('read');
                item.classList.remove('unread');
                markReadBtn.remove();
                this.unreadCount--;
                this.updateBadge();
            });
        }

        // Delete handler
        const deleteBtn = item.querySelector('.delete');
        deleteBtn.addEventListener('click', async (e) => {
            e.stopPropagation();
            await this.deleteNotification(notification.id);
            item.remove();
            if (!notification.read) {
                this.unreadCount--;
                this.updateBadge();
            }
        });

        // Click to mark as read
        if (!notification.read) {
            item.addEventListener('click', async () => {
                await this.markAsRead(notification.id);
                item.classList.add('read');
                item.classList.remove('unread');
                if (markReadBtn) markReadBtn.remove();
                this.unreadCount--;
                this.updateBadge();
            });
        }

        return item;
    }

    /**
     * Toggle notification center visibility
     */
    toggleNotificationCenter() {
        if (this.centerDropdown.classList.contains('hidden')) {
            this.showNotificationCenter();
        } else {
            this.hideNotificationCenter();
        }
    }

    /**
     * Show notification center
     */
    showNotificationCenter() {
        if (!this.centerDropdown) return;
        this.centerDropdown.classList.remove('hidden');
        this.loadNotifications(); // Refresh notifications
    }

    /**
     * Hide notification center
     */
    hideNotificationCenter() {
        if (!this.centerDropdown) return;
        this.centerDropdown.classList.add('hidden');
    }

    /**
     * Mark notification as read
     */
    async markAsRead(notificationId) {
        try {
            const endpoint = this.isAdminContext
                ? `/api/v1/admin/notifications/${notificationId}/read`
                : `/api/v1/user/notifications/${notificationId}/read`;

            await fetch(endpoint, {
                method: 'PATCH',
                credentials: 'include'
            });
        } catch (error) {
            console.error('[NotificationManager] Failed to mark as read:', error);
        }
    }

    /**
     * Mark all notifications as read
     */
    async markAllAsRead() {
        try {
            const endpoint = this.isAdminContext
                ? '/api/v1/admin/notifications/read'
                : '/api/v1/user/notifications/read';

            const response = await fetch(endpoint, {
                method: 'PATCH',
                credentials: 'include'
            });

            if (response.ok) {
                // Update UI
                this.unreadCount = 0;
                this.updateBadge();

                // Update notification center items
                const items = this.centerDropdown.querySelectorAll('.notification-center-item');
                items.forEach(item => {
                    item.classList.add('read');
                    item.classList.remove('unread');
                    const markReadBtn = item.querySelector('.mark-read');
                    if (markReadBtn) markReadBtn.remove();
                });
            }
        } catch (error) {
            console.error('[NotificationManager] Failed to mark all as read:', error);
        }
    }

    /**
     * Delete notification
     */
    async deleteNotification(notificationId) {
        try {
            const endpoint = this.isAdminContext
                ? `/api/v1/admin/notifications/${notificationId}`
                : `/api/v1/user/notifications/${notificationId}`;

            await fetch(endpoint, {
                method: 'DELETE',
                credentials: 'include'
            });
        } catch (error) {
            console.error('[NotificationManager] Failed to delete notification:', error);
        }
    }

    /**
     * Get notification icon SVG
     */
    getNotificationIcon(type) {
        const icons = {
            success: '<svg width="20" height="20" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd"/></svg>',
            info: '<svg width="20" height="20" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clip-rule="evenodd"/></svg>',
            warning: '<svg width="20" height="20" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd"/></svg>',
            error: '<svg width="20" height="20" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd"/></svg>',
            security: '<svg width="20" height="20" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M2.166 4.999A11.954 11.954 0 0010 1.944 11.954 11.954 0 0017.834 5c.11.65.166 1.32.166 2.001 0 5.225-3.34 9.67-8 11.317C5.34 16.67 2 12.225 2 7c0-.682.057-1.35.166-2.001zm11.541 3.708a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd"/></svg>'
        };
        return icons[type] || icons.info;
    }

    /**
     * Format time ago
     */
    formatTimeAgo(timestamp) {
        const now = new Date();
        const then = new Date(timestamp);
        const seconds = Math.floor((now - then) / 1000);

        if (seconds < 60) return 'Just now';
        if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`;
        if (seconds < 86400) return `${Math.floor(seconds / 3600)}h ago`;
        if (seconds < 604800) return `${Math.floor(seconds / 86400)}d ago`;
        return then.toLocaleDateString();
    }

    /**
     * Escape HTML to prevent XSS
     */
    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    /**
     * Setup event listeners
     */
    setupEventListeners() {
        // Reload notifications when tab becomes visible
        document.addEventListener('visibilitychange', () => {
            if (!document.hidden) {
                this.updateUnreadCount();
            }
        });
    }

    /**
     * Initialize notification sound
     */
    initSound() {
        // Create audio element for notification sound
        // Using a subtle notification sound (data URI for embedded sound)
        this.notificationSound = new Audio('data:audio/wav;base64,UklGRnoGAABXQVZFZm10IBAAAAABAAEAQB8AAEAfAAABAAgAZGF0YQoGAACBhYqFbF1fdJivrJBhNjVgodDbq2EcBj+a2/LDciUFLIHO8tiJNwgZaLvt559NEAxQp+PwtmMcBjiR1/LMeSwFJHfH8N2QQAoUXrTp66hVFApGn+DyvmwhBSuBzvLZiTYIGGi98OScTgwOUKjj8bRiGwU7k9r0y3ksBS + Bzvvhi0QLFVy17PCDWBILS5/j8r1sIgUtgs7y2Ig2CBlouO3nm04MDlCo4/G0YhsFO5Pa9Mt5LAUwgc740YpDCxVctfD0g1gSC0uf4/O9bCAFLoHO8tiINggYaLjt55tODA5QqOPxtGIbBT2U2vTLeSw');
        this.notificationSound.volume = 0.3; // 30% volume
    }

    /**
     * Play notification sound
     */
    playNotificationSound() {
        if (this.notificationSound && this.preferences.enable_sound) {
            this.notificationSound.currentTime = 0;
            this.notificationSound.play().catch(err => {
                console.error('[NotificationManager] Failed to play sound:', err);
            });
        }
    }

    /**
     * Cleanup on destroy
     */
    destroy() {
        if (this.ws) {
            this.ws.close();
        }
        if (this.wsReconnectTimer) {
            clearTimeout(this.wsReconnectTimer);
        }
        if (this.toastContainer) {
            this.toastContainer.remove();
        }
        if (this.bannerContainer) {
            this.bannerContainer.remove();
        }
        if (this.centerDropdown) {
            this.centerDropdown.remove();
        }
    }
}

// Initialize notification manager when DOM is ready
let notificationManager;
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => {
        notificationManager = new NotificationManager();
    });
} else {
    notificationManager = new NotificationManager();
}

// Cleanup on page unload
window.addEventListener('beforeunload', () => {
    if (notificationManager) {
        notificationManager.destroy();
    }
});
