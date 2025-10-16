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
     * Close a modal
     */
    close: function(modalId) {
      const overlay = document.getElementById(modalId);
      if (overlay) {
        overlay.classList.remove('active');
        document.body.style.overflow = '';
        Utils.dispatchEvent('modal:closed', { modalId });
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
        size = 'md'
      } = options;

      const modalId = Utils.generateId();
      const modalHTML = `
        <div id="${modalId}" class="modal-overlay">
          <div class="modal modal-${size}">
            <div class="modal-header">
              <h3 class="modal-title">${title}</h3>
              <button class="modal-close" onclick="Modal.close('${modalId}')">&times;</button>
            </div>
            <div class="modal-body">${body}</div>
            ${footer ? `<div class="modal-footer">${footer}</div>` : ''}
          </div>
        </div>
      `;

      document.body.insertAdjacentHTML('beforeend', modalHTML);

      // Close on overlay click
      const overlay = document.getElementById(modalId);
      overlay.addEventListener('click', function(e) {
        if (e.target === overlay) {
          Modal.close(modalId);
          if (onClose) onClose();
        }
      });

      // Close on Escape key
      const escapeHandler = function(e) {
        if (e.key === 'Escape') {
          Modal.close(modalId);
          if (onClose) onClose();
          document.removeEventListener('keydown', escapeHandler);
        }
      };
      document.addEventListener('keydown', escapeHandler);

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
        body: `<p style="margin: 0; line-height: 1.6;">${message}</p>`,
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
        body: `<p style="margin: 0; line-height: 1.6;">${message}</p>`,
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

})();
