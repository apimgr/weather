/**
 * Weather Service - Application JavaScript
 * Provides utilities for modals, alerts, toasts, and interactive components
 */

(function() {
  'use strict';

  // ============================================
  // UTILITY FUNCTIONS
  // ============================================

  const Utils = {
    /**
     * Create and dispatch a custom event
     */
    dispatchEvent: function(eventName, detail = {}) {
      const event = new CustomEvent(eventName, { detail, bubbles: true });
      document.dispatchEvent(event);
    },

    /**
     * Debounce function calls
     */
    debounce: function(func, wait) {
      let timeout;
      return function executedFunction(...args) {
        const later = () => {
          clearTimeout(timeout);
          func(...args);
        };
        clearTimeout(timeout);
        timeout = setTimeout(later, wait);
      };
    },

    /**
     * Generate unique ID
     */
    generateId: function() {
      return 'id_' + Math.random().toString(36).substr(2, 9);
    }
  };

  // ============================================
  // MODAL SYSTEM
  // ============================================

  const Modal = {
    /**
     * Open a modal
     */
    open: function(modalId) {
      const overlay = document.getElementById(modalId);
      if (overlay) {
        overlay.classList.add('active');
        document.body.style.overflow = 'hidden';
        Utils.dispatchEvent('modal:opened', { modalId });
      }
    },

    /**
     * Close a modal with animation
     */
    close: function(modalId) {
      const overlay = document.getElementById(modalId);
      if (overlay) {
        // Clear auto-close interval if exists
        const intervalId = overlay.getAttribute('data-auto-close-interval');
        if (intervalId) {
          clearInterval(parseInt(intervalId));
        }

        // Add closing animation
        overlay.classList.add('closing');
        overlay.classList.remove('active');

        // Remove from DOM after animation
        setTimeout(() => {
          overlay.remove();
          document.body.style.overflow = '';
          Utils.dispatchEvent('modal:closed', { modalId });
        }, 200);
      }
    },

    /**
     * Create and show a modal programmatically
     */
    create: function(options) {
      const {
        title = 'Modal',
        body = '',
        footer = '',
        onClose = null,
        size = 'md',
        autoClose = 0,  // Auto-close after N seconds (0 = no auto-close)
        closeable = true
      } = options;

      const modalId = Utils.generateId();
      const autoCloseHtml = autoClose > 0
        ? `<div class="modal-auto-close">Auto-closing in <span id="${modalId}-countdown">${autoClose}</span>s</div>`
        : '';

      const modalHTML = `
        <div id="${modalId}" class="modal-overlay">
          <div class="modal modal-${size}">
            <div class="modal-header">
              <h3 class="modal-title">${title}</h3>
              ${closeable ? `<button class="modal-close" onclick="Modal.close('${modalId}')">&times;</button>` : ''}
            </div>
            <div class="modal-body">
              ${body}
              ${autoCloseHtml}
            </div>
            ${footer ? `<div class="modal-footer">${footer}</div>` : ''}
          </div>
        </div>
      `;

      document.body.insertAdjacentHTML('beforeend', modalHTML);

      // Close on overlay click (if closeable)
      const overlay = document.getElementById(modalId);

      // Prevent clicks on modal content from closing the modal
      const modalContent = overlay.querySelector('.modal');
      if (modalContent) {
        modalContent.addEventListener('click', function(e) {
          e.stopPropagation();
        });
      }

      if (closeable) {
        overlay.addEventListener('click', function(e) {
          if (e.target === overlay) {
            Modal.close(modalId);
            if (onClose) onClose();
          }
        });
      }

      // Close on Escape key (if closeable)
      if (closeable) {
        const escapeHandler = function(e) {
          if (e.key === 'Escape') {
            Modal.close(modalId);
            if (onClose) onClose();
            document.removeEventListener('keydown', escapeHandler);
          }
        };
        document.addEventListener('keydown', escapeHandler);
      }

      // Auto-close countdown
      if (autoClose > 0) {
        let remaining = autoClose;
        const countdownEl = document.getElementById(`${modalId}-countdown`);

        const interval = setInterval(() => {
          remaining--;
          if (countdownEl) {
            countdownEl.textContent = remaining;
          }

          if (remaining <= 0) {
            clearInterval(interval);
            Modal.close(modalId);
            if (onClose) onClose();
          }
        }, 1000);

        // Store interval ID to clear if manually closed
        overlay.setAttribute('data-auto-close-interval', interval);
      }

      Modal.open(modalId);
      return modalId;
    }
  };

  // ============================================
  // TOAST NOTIFICATIONS
  // ============================================

  const Toast = {
    container: null,

    /**
     * Initialize toast container
     */
    init: function() {
      if (!this.container) {
        this.container = document.createElement('div');
        this.container.className = 'toast-container';
        document.body.appendChild(this.container);
      }
    },

    /**
     * Show a toast notification
     */
    show: function(message, options = {}) {
      this.init();

      const {
        type = 'info',  // 'success', 'error', 'warning', 'info'
        title = '',
        duration = 5000,
        dismissible = true
      } = options;

      const icons = {
        success: '✓',
        error: '✗',
        warning: '⚠',
        info: 'ℹ'
      };

      const toastId = Utils.generateId();
      const toastHTML = `
        <div id="${toastId}" class="toast toast-${type}">
          <div class="toast-icon">${icons[type]}</div>
          <div class="toast-content">
            ${title ? `<div class="toast-title">${title}</div>` : ''}
            <div class="toast-message">${message}</div>
          </div>
          ${dismissible ? '<button class="toast-close" onclick="Toast.dismiss(\'' + toastId + '\')">&times;</button>' : ''}
        </div>
      `;

      this.container.insertAdjacentHTML('beforeend', toastHTML);

      if (duration > 0) {
        setTimeout(() => this.dismiss(toastId), duration);
      }

      Utils.dispatchEvent('toast:shown', { toastId, type, message });
      return toastId;
    },

    /**
     * Dismiss a toast
     */
    dismiss: function(toastId) {
      const toast = document.getElementById(toastId);
      if (toast) {
        toast.style.opacity = '0';
        toast.style.transform = 'translateX(100%)';
        setTimeout(() => toast.remove(), 300);
        Utils.dispatchEvent('toast:dismissed', { toastId });
      }
    },

    /**
     * Convenience methods
     */
    success: function(message, options = {}) {
      return this.show(message, { ...options, type: 'success' });
    },

    error: function(message, options = {}) {
      return this.show(message, { ...options, type: 'error' });
    },

    warning: function(message, options = {}) {
      return this.show(message, { ...options, type: 'warning' });
    },

    info: function(message, options = {}) {
      return this.show(message, { ...options, type: 'info' });
    }
  };

  // ============================================
  // ALERT SYSTEM
  // ============================================

  const Alert = {
    /**
     * Create an alert element
     */
    create: function(message, options = {}) {
      const {
        type = 'info',  // 'success', 'error', 'warning', 'info'
        title = '',
        dismissible = true,
        container = null
      } = options;

      const icons = {
        success: '✓',
        error: '✗',
        warning: '⚠',
        info: 'ℹ'
      };

      const alertId = Utils.generateId();
      const alertHTML = `
        <div id="${alertId}" class="alert alert-${type} ${dismissible ? 'alert-dismissible' : ''}">
          <div class="alert-icon">${icons[type]}</div>
          <div class="alert-content">
            ${title ? `<div class="alert-title">${title}</div>` : ''}
            <div class="alert-message">${message}</div>
          </div>
          ${dismissible ? '<button class="alert-close" onclick="Alert.dismiss(\'' + alertId + '\')">&times;</button>' : ''}
        </div>
      `;

      if (container) {
        const containerEl = typeof container === 'string'
          ? document.getElementById(container) || document.querySelector(container)
          : container;

        if (containerEl) {
          containerEl.insertAdjacentHTML('beforeend', alertHTML);
        }
      }

      return alertId;
    },

    /**
     * Dismiss an alert
     */
    dismiss: function(alertId) {
      const alert = document.getElementById(alertId);
      if (alert) {
        alert.style.opacity = '0';
        setTimeout(() => alert.remove(), 300);
      }
    }
  };

  // ============================================
  // DROPDOWN SYSTEM
  // ============================================

  const Dropdown = {
    activeDropdown: null,

    /**
     * Toggle dropdown visibility
     */
    toggle: function(dropdownId, triggerId = null) {
      const dropdown = document.getElementById(dropdownId);
      if (!dropdown) return;

      const isActive = dropdown.classList.contains('active');

      // Close any other open dropdowns
      this.closeAll();

      if (!isActive) {
        dropdown.classList.add('active');
        this.activeDropdown = dropdownId;

        // Position dropdown if trigger is provided
        if (triggerId) {
          const trigger = document.getElementById(triggerId);
          if (trigger) {
            const rect = trigger.getBoundingClientRect();
            dropdown.style.top = `${rect.bottom + 8}px`;
            dropdown.style.right = `${window.innerWidth - rect.right}px`;
          }
        }

        Utils.dispatchEvent('dropdown:opened', { dropdownId });
      }
    },

    /**
     * Close a specific dropdown
     */
    close: function(dropdownId) {
      const dropdown = document.getElementById(dropdownId);
      if (dropdown) {
        dropdown.classList.remove('active');
        if (this.activeDropdown === dropdownId) {
          this.activeDropdown = null;
        }
        Utils.dispatchEvent('dropdown:closed', { dropdownId });
      }
    },

    /**
     * Close all dropdowns
     */
    closeAll: function() {
      document.querySelectorAll('.profile-dropdown.active, .dropdown.active').forEach(dropdown => {
        dropdown.classList.remove('active');
      });
      this.activeDropdown = null;
    }
  };

  // ============================================
  // MOBILE MENU
  // ============================================

  const MobileMenu = {
    /**
     * Toggle mobile menu
     */
    toggle: function() {
      const menu = document.querySelector('.navbar-menu-mobile');
      if (menu) {
        menu.classList.toggle('active');
        Utils.dispatchEvent('mobile-menu:toggled', {
          isActive: menu.classList.contains('active')
        });
      }
    },

    /**
     * Close mobile menu
     */
    close: function() {
      const menu = document.querySelector('.navbar-menu-mobile');
      if (menu) {
        menu.classList.remove('active');
      }
    }
  };

  // ============================================
  // NOTIFICATION SYSTEM
  // ============================================

  const Notifications = {
    unreadCount: 0,

    /**
     * Update notification badge count
     */
    updateBadge: function(count) {
      this.unreadCount = count;
      const badge = document.querySelector('.notification-badge');

      if (badge) {
        if (count > 0) {
          badge.textContent = count > 99 ? '99+' : count;
          badge.style.display = 'flex';
        } else {
          badge.style.display = 'none';
        }
      }

      Utils.dispatchEvent('notifications:updated', { count });
    },

    /**
     * Fetch and update notification count
     */
    fetch: async function() {
      try {
        const response = await fetch('/api/v1/notifications/unread', {
          credentials: 'same-origin'
        });

        // Stop polling if unauthorized (user logged out)
        if (response.status === 401) {
          this.stopPolling();
          return;
        }

        if (response.ok) {
          const data = await response.json();
          this.updateBadge(data.unread_count || 0);
        }
      } catch (error) {
        console.error('Failed to fetch notifications:', error);
      }
    },

    /**
     * Start polling for notifications
     */
    startPolling: function(interval = 30000) {
      this.fetch(); // Initial fetch
      this.pollInterval = setInterval(() => this.fetch(), interval);
    },

    /**
     * Stop polling
     */
    stopPolling: function() {
      if (this.pollInterval) {
        clearInterval(this.pollInterval);
        this.pollInterval = null;
      }
    }
  };

  // ============================================
  // FORM UTILITIES
  // ============================================

  const Forms = {
    /**
     * Serialize form data to JSON
     */
    serializeToJSON: function(formElement) {
      const formData = new FormData(formElement);
      const json = {};

      for (const [key, value] of formData.entries()) {
        if (json[key]) {
          if (!Array.isArray(json[key])) {
            json[key] = [json[key]];
          }
          json[key].push(value);
        } else {
          json[key] = value;
        }
      }

      return json;
    },

    /**
     * Validate form
     */
    validate: function(formElement) {
      if (!formElement.checkValidity()) {
        formElement.reportValidity();
        return false;
      }
      return true;
    },

    /**
     * Show form errors
     */
    showErrors: function(formElement, errors) {
      // Clear existing errors
      formElement.querySelectorAll('.form-error').forEach(el => el.remove());

      // Add new errors
      Object.entries(errors).forEach(([field, message]) => {
        const input = formElement.querySelector(`[name="${field}"]`);
        if (input) {
          const error = document.createElement('div');
          error.className = 'form-error';
          error.textContent = message;
          input.parentNode.appendChild(error);
        }
      });
    }
  };

  // ============================================
  // LOADING STATES
  // ============================================

  const Loading = {
    /**
     * Show loading spinner
     */
    show: function(element, size = 'md') {
      const spinner = document.createElement('div');
      spinner.className = `spinner spinner-${size}`;
      spinner.setAttribute('data-loading', 'true');

      if (typeof element === 'string') {
        element = document.querySelector(element);
      }

      if (element) {
        element.appendChild(spinner);
      }
    },

    /**
     * Hide loading spinner
     */
    hide: function(element) {
      if (typeof element === 'string') {
        element = document.querySelector(element);
      }

      if (element) {
        const spinner = element.querySelector('[data-loading="true"]');
        if (spinner) {
          spinner.remove();
        }
      }
    },

    /**
     * Set button loading state
     */
    setButtonLoading: function(button, isLoading) {
      if (typeof button === 'string') {
        button = document.querySelector(button);
      }

      if (!button) return;

      if (isLoading) {
        button.disabled = true;
        button.setAttribute('data-original-text', button.textContent);
        button.innerHTML = '<span class="loading-dots"><span class="loading-dot"></span><span class="loading-dot"></span><span class="loading-dot"></span></span>';
      } else {
        button.disabled = false;
        const originalText = button.getAttribute('data-original-text');
        if (originalText) {
          button.textContent = originalText;
          button.removeAttribute('data-original-text');
        }
      }
    }
  };

  // ============================================
  // GLOBAL EVENT HANDLERS
  // ============================================

  document.addEventListener('DOMContentLoaded', function() {
    // Close dropdowns when clicking outside
    document.addEventListener('click', function(e) {
      if (!e.target.closest('.profile-avatar') && !e.target.closest('.dropdown')) {
        Dropdown.closeAll();
      }
    });

    // Close mobile menu when clicking links
    document.querySelectorAll('.navbar-menu-mobile .navbar-link').forEach(link => {
      link.addEventListener('click', () => MobileMenu.close());
    });

    // Handle modal close on overlay click
    document.querySelectorAll('.modal-overlay').forEach(overlay => {
      overlay.addEventListener('click', function(e) {
        if (e.target === overlay) {
          Modal.close(overlay.id);
        }
      });
    });

    // Handle ESC key to close modals
    document.addEventListener('keydown', function(e) {
      if (e.key === 'Escape') {
        // Close active modal
        const activeModal = document.querySelector('.modal-overlay.active');
        if (activeModal) {
          Modal.close(activeModal.id);
        }

        // Close dropdowns
        Dropdown.closeAll();

        // Close mobile menu
        MobileMenu.close();
      }
    });
  });

  // ============================================
  // MODERN ALERT & CONFIRM REPLACEMENTS
  // ============================================

  /**
   * Modern alert replacement using modals
   */
  window.showAlert = function(message, title = 'Alert') {
    return new Promise((resolve) => {
      const modalId = Modal.create({
        title: title,
        body: `<p class="modal-body-text">${message}</p>`,
        footer: `
          <button class="btn btn-primary" onclick="Modal.close('${modalId}'); window._alertResolve();">
            OK
          </button>
        `,
        size: 'sm',
        onClose: () => resolve()
      });
      window._alertResolve = resolve;
    });
  };

  /**
   * Modern confirm replacement using modals
   */
  window.showConfirm = function(message, title = 'Confirm') {
    return new Promise((resolve) => {
      const modalId = Modal.create({
        title: title,
        body: `<p class="modal-body-text">${message}</p>`,
        footer: `
          <button class="btn btn-secondary" onclick="Modal.close('${modalId}'); window._confirmResolve(false);">
            Cancel
          </button>
          <button class="btn btn-primary" onclick="Modal.close('${modalId}'); window._confirmResolve(true);">
            OK
          </button>
        `,
        size: 'sm',
        onClose: () => resolve(false)
      });
      window._confirmResolve = resolve;
    });
  };

  /**
   * Modern prompt replacement using modals
   */
  window.showPrompt = function(message, defaultValue = '', title = 'Input') {
    return new Promise((resolve) => {
      const inputId = Utils.generateId();
      const modalId = Modal.create({
        title: title,
        body: `
          <p class="modal-body-text-spacing">${message}</p>
          <input type="text" id="${inputId}" class="modal-input-full" value="${defaultValue}"
                 placeholder="Enter value...">
        `,
        footer: `
          <button class="btn btn-secondary" onclick="Modal.close('${modalId}'); window._promptResolve(null);">
            Cancel
          </button>
          <button class="btn btn-primary" onclick="
            const value = document.getElementById('${inputId}').value;
            Modal.close('${modalId}');
            window._promptResolve(value);
          ">
            OK
          </button>
        `,
        size: 'sm',
        onClose: () => resolve(null)
      });

      window._promptResolve = resolve;

      // Focus input and allow Enter to submit
      setTimeout(() => {
        const input = document.getElementById(inputId);
        if (input) {
          input.focus();
          input.addEventListener('keypress', (e) => {
            if (e.key === 'Enter') {
              const value = input.value;
              Modal.close(modalId);
              resolve(value);
            }
          });
        }
      }, 100);
    });
  };

  // ============================================
  // EXPOSE TO GLOBAL SCOPE
  // ============================================

  window.Modal = Modal;
  window.Toast = Toast;
  window.Alert = Alert;
  window.Dropdown = Dropdown;
  window.MobileMenu = MobileMenu;
  window.Notifications = Notifications;
  window.Forms = Forms;
  window.Loading = Loading;
  window.Utils = Utils;

  // Auto-start notification polling if user is authenticated
  // Only poll if notification bell exists (user is logged in)
  if (document.querySelector('.notification-bell')) {
    Notifications.startPolling();
  }

  // ============================================
  // ADMIN PANEL (AI.md PART 18)
  // ============================================

  const AdminPanel = {
    /**
     * Initialize admin panel
     */
    init: function() {
      this.initializeSearch();
      this.initializeKeyboardShortcuts();
      this.initializeMobileSidebar();
    },

    /**
     * Search functionality
     */
    initializeSearch: function() {
      const searchInput = document.getElementById('admin-search');
      if (!searchInput) return;

      searchInput.addEventListener('input', Utils.debounce(function(e) {
        const query = e.target.value.toLowerCase();
        if (query.length < 2) return;

        // Search through sidebar items
        const navItems = document.querySelectorAll('.nav-item, .nav-subitem');
        navItems.forEach(item => {
          const text = item.textContent.toLowerCase();
          const match = text.includes(query);
          item.style.display = match ? '' : 'none';
        });
      }, 300));
    },

    /**
     * Keyboard shortcuts per AI.md PART 18
     */
    initializeKeyboardShortcuts: function() {
      document.addEventListener('keydown', function(e) {
        // Ctrl/Cmd + K: Focus search
        if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
          e.preventDefault();
          document.getElementById('admin-search')?.focus();
        }

        // Ctrl/Cmd + B: Toggle sidebar
        if ((e.ctrlKey || e.metaKey) && e.key === 'b') {
          e.preventDefault();
          AdminPanel.toggleSidebar();
        }
      });
    },

    /**
     * Mobile sidebar
     */
    initializeMobileSidebar: function() {
      const sidebar = document.getElementById('adminSidebar');
      const toggle = document.querySelector('.sidebar-toggle');

      if (!sidebar || !toggle) return;

      toggle.addEventListener('click', function() {
        if (window.innerWidth <= 768) {
          sidebar.classList.toggle('mobile-open');
        }
      });

      // Close sidebar when clicking outside on mobile
      document.addEventListener('click', function(e) {
        if (window.innerWidth <= 768) {
          if (!sidebar.contains(e.target) && !toggle.contains(e.target)) {
            sidebar.classList.remove('mobile-open');
          }
        }
      });
    },

    /**
     * Toggle sidebar
     */
    toggleSidebar: function() {
      const sidebar = document.getElementById('adminSidebar');
      if (sidebar) {
        sidebar.classList.toggle('collapsed');
      }
    }
  };

  /**
   * Confirmation dialog (replaces JS alerts per AI.md PART 18)
   */
  window.confirmAction = function(message, callback) {
    Modal.create({
      title: 'Confirm Action',
      body: `<p class="modal-body-text">${message}</p>`,
      footer: `
        <button class="btn btn-secondary" onclick="Modal.close(this.closest('.modal-overlay').id)">Cancel</button>
        <button class="btn btn-danger" id="confirmActionBtn">Confirm</button>
      `,
      size: 'sm',
      onClose: function() {}
    });

    // Set up confirm button handler after modal is created
    setTimeout(function() {
      const confirmBtn = document.getElementById('confirmActionBtn');
      if (confirmBtn) {
        confirmBtn.addEventListener('click', function() {
          const overlay = this.closest('.modal-overlay');
          if (overlay) Modal.close(overlay.id);
          if (callback) callback();
        });
      }
    }, 50);
  };

  // Expose AdminPanel
  window.AdminPanel = AdminPanel;

  // Initialize admin panel if on admin page
  if (document.querySelector('#adminSidebar') || window.location.pathname.startsWith('/admin')) {
    AdminPanel.init();
  }

  // ============================================
  // THEME SYSTEM (AI.md PART 16)
  // Dark theme is DEFAULT per AI.md spec
  // ============================================

  const Theme = {
    // Available themes: dark (default), light, auto
    THEMES: ['dark', 'light', 'auto'],

    /**
     * Get current theme from localStorage
     */
    get: function() {
      return localStorage.getItem('theme') || 'dark';
    },

    /**
     * Set theme and persist to localStorage
     */
    set: function(theme) {
      if (!this.THEMES.includes(theme)) {
        theme = 'dark';
      }
      localStorage.setItem('theme', theme);
      this.apply(theme);
      Utils.dispatchEvent('theme:changed', { theme });
    },

    /**
     * Apply theme to document
     */
    apply: function(theme) {
      var effectiveTheme = theme;
      if (theme === 'auto') {
        effectiveTheme = window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
      }
      document.documentElement.setAttribute('data-theme', effectiveTheme);
      // Update theme-color meta tag
      var themeColor = effectiveTheme === 'dark' ? '#282a36' : '#ffffff';
      var metaThemeColor = document.querySelector('meta[name="theme-color"]');
      if (metaThemeColor) {
        metaThemeColor.setAttribute('content', themeColor);
      }
    },

    /**
     * Toggle between dark and light themes
     */
    toggle: function() {
      var current = this.get();
      var next = current === 'dark' ? 'light' : 'dark';
      this.set(next);
      return next;
    },

    /**
     * Cycle through all themes: dark -> light -> auto -> dark
     */
    cycle: function() {
      var current = this.get();
      var index = this.THEMES.indexOf(current);
      var next = this.THEMES[(index + 1) % this.THEMES.length];
      this.set(next);
      return next;
    },

    /**
     * Initialize theme system
     */
    init: function() {
      // Apply saved theme
      this.apply(this.get());

      // Listen for system preference changes when using auto theme
      window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', function(e) {
        if (Theme.get() === 'auto') {
          Theme.apply('auto');
        }
      });

      // Set up theme toggle buttons
      document.querySelectorAll('[data-theme-toggle]').forEach(function(btn) {
        btn.addEventListener('click', function() {
          Theme.toggle();
        });
      });

      document.querySelectorAll('[data-theme-cycle]').forEach(function(btn) {
        btn.addEventListener('click', function() {
          Theme.cycle();
        });
      });
    }
  };

  // Expose Theme globally
  window.Theme = Theme;

  // Initialize theme system
  Theme.init();

})();
