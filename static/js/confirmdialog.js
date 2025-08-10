if (window.SuperConfirmDialogInitialized) {
  if (typeof console !== "undefined" && console.warn) {
    console.warn(
      "SuperConfirmDialog already initialized, skipping duplicate execution.",
    );
  }
} else {
  /**
   * SuperConfirmDialog Class
   * A lightweight, customizable confirmation dialog component.
   * IMPORTANT: This library relies on Tailwind CSS being set up and processed
   *            in the consuming application for proper styling.
   *
   * @param {object} [options={}] - Configuration options for the dialog.
   * @param {string} [options.title='Confirm'] - The dialog title.
   * @param {string} [options.message='Are you sure?'] - The dialog message.
   * @param {string} [options.confirmText='Confirm'] - Text for the confirm button.
   * @param {string} [options.cancelText='Cancel'] - Text for the cancel button.
   * @param {function} [options.onConfirm=null] - Callback function when confirmed.
   * @param {function} [options.onCancel=null] - Callback function when cancelled.
   */
  class SuperConfirmDialog {
    /**
     * Constructs a new SuperConfirmDialog instance.
     * @param {object} [options={}] - Options for the dialog.
     */
    constructor(options = {}) {
      this.options = {
        title: 'Confirm',
        message: 'Are you sure?',
        confirmText: 'Confirm',
        cancelText: 'Cancel',
        onConfirm: null,
        onCancel: null,
        ...options
      };

      this.overlay = null;
      this.dialog = null;
      this.isOpen = false;
      
      this.createDialog();
    }

    /**
     * Creates the dialog HTML structure
     */
    createDialog() {
      // Create overlay
      this.overlay = document.createElement('div');
      this.overlay.className = 'fixed inset-0 z-[9999] flex items-center justify-center p-4 opacity-0 transition-opacity duration-300 ease-out';
      this.overlay.style.backgroundColor = 'rgba(0, 0, 0, 0.4)';
      this.overlay.style.display = 'none';

      // Create dialog container
      this.dialog = document.createElement('div');
      this.dialog.className = 'bg-secondary-light dark:bg-secondary-dark rounded-xl shadow-2xl max-w-sm sm:max-w-md w-full mx-auto border border-border-light dark:border-border-dark';

      // Create dialog content
      this.dialog.innerHTML = `
        <div class="p-6 sm:p-8">
          <h3 class="text-lg sm:text-xl font-bold text-center text-text-on-light dark:text-text-on-dark mb-3 sm:mb-4" id="dialog-title">
            ${this.options.title}
          </h3>
          <p class="text-sm sm:text-base text-center text-text-on-light dark:text-text-on-dark opacity-80 mb-6 sm:mb-8 leading-relaxed" id="dialog-message">
            ${this.options.message}
          </p>
          <div class="flex space-x-4 justify-center">
            <button type="button" class="px-4 sm:px-6 py-2 sm:py-3 text-sm font-semibold text-white bg-danger-light dark:bg-danger-dark hover:opacity-90 rounded-lg focus:outline-none focus:ring-2 focus:ring-danger-dark transition-all duration-200 shadow-sm hover:shadow-md" id="dialog-cancel">
              ${this.options.cancelText}
            </button>
            <button type="button" class="px-4 sm:px-6 py-2 sm:py-3 text-sm font-semibold text-white bg-success-light dark:bg-success-dark hover:opacity-90 rounded-lg focus:outline-none focus:ring-2 focus:ring-success-dark transition-all duration-200 shadow-sm hover:shadow-md" id="dialog-confirm">
              ${this.options.confirmText}
            </button>
          </div>
        </div>
      `;

      this.overlay.appendChild(this.dialog);
      document.body.appendChild(this.overlay);

      // Bind events
      this.bindEvents();
    }

    /**
     * Binds event listeners
     */
    bindEvents() {
      const confirmBtn = this.dialog.querySelector('#dialog-confirm');
      const cancelBtn = this.dialog.querySelector('#dialog-cancel');

      confirmBtn.addEventListener('click', () => this.confirm());
      cancelBtn.addEventListener('click', () => this.cancel());
      
      // Close on overlay click
      this.overlay.addEventListener('click', (e) => {
        if (e.target === this.overlay) {
          this.cancel();
        }
      });

      // Close on Escape key
      document.addEventListener('keydown', (e) => {
        if (e.key === 'Escape' && this.isOpen) {
          this.cancel();
        }
      });
    }

    /**
     * Shows the dialog with SweetAlert2-style bounce animation
     */
    show() {
      if (this.isOpen) return;
      
      this.isOpen = true;
      this.overlay.style.display = 'flex';
      
      // Trigger reflow to ensure display change is applied
      this.overlay.offsetHeight;
      
      // Animate overlay fade in
      requestAnimationFrame(() => {
        this.overlay.classList.remove('opacity-0');
        this.overlay.classList.add('opacity-100');
      });
      
      // SweetAlert2-style bounce: big -> small -> slightly big -> settle
      this.dialog.style.transform = 'scale(1.1)';
      this.dialog.style.transition = 'transform 0.15s ease-out';
      
      setTimeout(() => {
        this.dialog.style.transform = 'scale(0.95)';
        this.dialog.style.transition = 'transform 0.12s ease-out';
      }, 150);
      
      setTimeout(() => {
        this.dialog.style.transform = 'scale(1.02)';
        this.dialog.style.transition = 'transform 0.1s ease-out';
      }, 262);
      
      setTimeout(() => {
        this.dialog.style.transform = 'scale(1)';
        this.dialog.style.transition = 'transform 0.15s ease-out';
      }, 362);

      // Focus the cancel button by default
      setTimeout(() => {
        this.dialog.querySelector('#dialog-cancel').focus();
      }, 500);
    }

    /**
     * Hides the dialog with animation
     */
    hide() {
      if (!this.isOpen) return;
      
      this.isOpen = false;
      
      // Animate out
      this.overlay.classList.remove('opacity-100');
      this.overlay.classList.add('opacity-0');
      this.dialog.style.transform = 'scale(0.8)';
      this.dialog.style.transition = 'transform 0.2s ease-in';
      
      // Hide after animation
      setTimeout(() => {
        this.overlay.style.display = 'none';
        this.destroy();
      }, 200);
    }

    /**
     * Handles confirm action
     */
    confirm() {
      if (this.options.onConfirm) {
        this.options.onConfirm();
      }
      this.hide();
    }

    /**
     * Handles cancel action
     */
    cancel() {
      if (this.options.onCancel) {
        this.options.onCancel();
      }
      this.hide();
    }

    /**
     * Destroys the dialog and removes it from DOM
     */
    destroy() {
      if (this.overlay && this.overlay.parentNode) {
        this.overlay.parentNode.removeChild(this.overlay);
      }
      this.overlay = null;
      this.dialog = null;
    }

    /**
     * Static method to show a confirmation dialog
     * @param {object} options - Dialog options
     * @returns {Promise} - Promise that resolves with true/false
     */
    static show(options = {}) {
      return new Promise((resolve) => {
        const dialog = new SuperConfirmDialog({
          ...options,
          onConfirm: () => {
            if (options.onConfirm) options.onConfirm();
            resolve(true);
          },
          onCancel: () => {
            if (options.onCancel) options.onCancel();
            resolve(false);
          }
        });
        dialog.show();
      });
    }
  }

  // Global function for easy access (similar to existing pattern)
  window.showConfirmationDialog = function(title, message, confirmText = 'Confirm', element = null) {
    return SuperConfirmDialog.show({
      title: title,
      message: message,
      confirmText: confirmText,
      onConfirm: () => {
        if (element) {
          htmx.trigger(element, 'confirmed');
        }
      }
    });
  };

  // Mark as initialized
  window.SuperConfirmDialogInitialized = true;
}