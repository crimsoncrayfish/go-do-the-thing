if (window.SuperDialogInitialized) {
  if (typeof console !== "undefined" && console.warn) {
    console.warn(
      "SuperDialog already initialized, skipping duplicate execution.",
    );
  }
} else {
  /**
   * SuperDialog Class
   * A customizable and accessible dialog/popup component.
   * IMPORTANT: This library relies on Tailwind CSS being set up and processed
   *            in the consuming application for proper styling.
   *
   * @param {object} [options={}] - Configuration options for the dialog.
   * @param {string} [options.title=''] - The dialog title.
   * @param {string} [options.text=''] - The dialog text content.
   * @param {string} [options.confirmButtonText='OK'] - Text for the confirm button.
   * @param {string} [options.cancelButtonText='Cancel'] - Text for the cancel button.
   * @param {boolean} [options.showCancelButton=false] - Whether to show the cancel button.
   * @param {function} [options.onConfirm] - Callback when confirmed.
   * @param {function} [options.onCancel] - Callback when cancelled.
   */
  class SuperDialog {
    /**
     * Constructs a new SuperDialog instance.
     * @param {object} [options={}] - Options for the dialog.
     */
    constructor(options = {}) {
      this.options = {
        title: "",
        text: "",
        confirmButtonText: "OK",
        cancelButtonText: "Cancel",
        showCancelButton: false,
        onConfirm: null,
        onCancel: null,
        ...options,
      };

      this.backdrop = null;
      this.dialog = null;
      this.isOpen = false;
      this.resolvePromise = null;
      this.rejectPromise = null;

      this._initDOM();
      this._attachEvents();
    }

    /**
     * Initializes the DOM structure for the dialog.
     * @private
     */
    _initDOM() {
      this.backdrop = document.createElement("div");
      this.backdrop.classList.add(
        "super-dialog-backdrop",
        "fixed",
        "inset-0",
        "z-50",
        "flex",
        "items-center",
        "justify-center",
        "hidden",
        "transition-opacity",
        "duration-300",
        "ease-out",
      );

      this.backdrop.style.backgroundColor = "rgba(0, 0, 0, 0.4)";

      this.dialog = document.createElement("div");
      this.dialog.classList.add(
        "super-dialog",
        "bg-secondary-light",
        "dark:bg-secondary-dark",
        "rounded-lg",
        "shadow-xl",
        "p-4",
        "sm:p-6",
        "max-w-sm",
        "sm:max-w-md",
        "w-full",
        "mx-4",
        "transform",
        "scale-95",
        "opacity-0",
        "transition-all",
        "duration-300",
        "ease-out",
      );

      this.dialog.setAttribute("role", "dialog");
      this.dialog.setAttribute("aria-modal", "true");
      this.dialog.setAttribute("aria-labelledby", "super-dialog-title");
      this.dialog.setAttribute("aria-describedby", "super-dialog-text");

      this.backdrop.appendChild(this.dialog);
      document.body.appendChild(this.backdrop);

      this._renderContent();
    }

    /**
     * Renders the dialog content.
     * @private
     */
    _renderContent() {
      let titleHtml = "";
      if (this.options.title) {
        titleHtml = `
          <h2 id="super-dialog-title" class="text-2xl font-bold text-text-on-light dark:text-text-on-dark mb-4 text-center">
            ${this._escapeHtml(this.options.title)}
          </h2>
        `;
      }

      let textHtml = "";
      if (this.options.text) {
        textHtml = `
          <p id="super-dialog-text" class="text-text-on-light dark:text-text-on-dark mb-6">
            ${this._escapeHtml(this.options.text)}
          </p>
        `;
      }

      const cancelButtonHtml = this.options.showCancelButton
        ? `
        <button
          type="button"
          class="super-dialog-cancel bg-danger-light dark:bg-danger-dark text-text-on-light dark:text-text-on-dark font-bold py-2 px-4 rounded hover:opacity-80 transition-opacity"
        >
          ${this._escapeHtml(this.options.cancelButtonText)}
        </button>
      `
        : "";

      this.dialog.innerHTML = `
        ${titleHtml}
        ${textHtml}
        <div class="flex justify-end gap-4">
          ${cancelButtonHtml}
          <button
            type="button"
            class="super-dialog-confirm bg-success-light dark:bg-success-dark text-text-on-light dark:text-text-on-dark font-bold py-2 px-4 rounded hover:opacity-80 transition-opacity"
          >
            ${this._escapeHtml(this.options.confirmButtonText)}
          </button>
        </div>
      `;
    }

    /**
     * Attaches event listeners.
     * @private
     */
    _attachEvents() {
      this._boundKeyDownHandler = this._handleKeyDown.bind(this);

      this.dialog.addEventListener("click", (e) => e.stopPropagation());

      this.backdrop.addEventListener("click", () => {
        this._handleCancel();
      });

      this.dialog.addEventListener("click", (e) => {
        if (e.target.classList.contains("super-dialog-confirm")) {
          this._handleConfirm();
        } else if (e.target.classList.contains("super-dialog-cancel")) {
          this._handleCancel();
        }
      });

      document.addEventListener("keydown", this._boundKeyDownHandler);
    }

    /**
     * Handles keyboard events.
     * @private
     * @param {KeyboardEvent} e - The keyboard event.
     */
    _handleKeyDown(e) {
      if (!this.isOpen) return;

      switch (e.key) {
        case "Escape":
          e.preventDefault();
          this._handleCancel();
          break;
        case "Enter":
          e.preventDefault();
          this._handleConfirm();
          break;
        case "Tab":
          this._handleTabNavigation(e);
          break;
      }
    }

    /**
     * Handles tab navigation within the dialog.
     * @private
     * @param {KeyboardEvent} e - The keyboard event.
     */
    _handleTabNavigation(e) {
      const focusableElements = this.dialog.querySelectorAll(
        'button:not([disabled]), [tabindex]:not([tabindex="-1"])',
      );

      if (focusableElements.length === 0) return;

      const firstElement = focusableElements[0];
      const lastElement = focusableElements[focusableElements.length - 1];

      if (e.shiftKey) {
        if (document.activeElement === firstElement) {
          e.preventDefault();
          lastElement.focus();
        }
      } else {
        if (document.activeElement === lastElement) {
          e.preventDefault();
          firstElement.focus();
        }
      }
    }

    /**
     * Handles confirm action.
     * @private
     */
    _handleConfirm() {
      const result = { isConfirmed: true, isDismissed: false };

      if (this.options.onConfirm) {
        this.options.onConfirm(result);
      }

      if (this.resolvePromise) {
        this.resolvePromise(result);
      }

      this.close();
    }

    /**
     * Handles cancel action.
     * @private
     */
    _handleCancel() {
      const result = { isConfirmed: false, isDismissed: true };

      if (this.options.onCancel) {
        this.options.onCancel(result);
      }

      if (this.resolvePromise) {
        this.resolvePromise(result);
      }

      this.close();
    }

    /**
     * Shows the dialog with a bounce animation.
     * @returns {Promise} Promise that resolves with the user's choice.
     */
    show() {
      return new Promise((resolve, reject) => {
        this.resolvePromise = resolve;
        this.rejectPromise = reject;

        this.isOpen = true;
        this.backdrop.classList.remove("hidden");

        this.backdrop.offsetHeight;

        requestAnimationFrame(() => {
          this.dialog.classList.remove("scale-95", "opacity-0");
          this.dialog.classList.add("scale-100", "opacity-100");
          setTimeout(() => {
            this.dialog.classList.add("animate-bounce-in");
          }, 100);
        });

        setTimeout(() => {
          const firstButton = this.dialog.querySelector("button");
          if (firstButton) {
            firstButton.focus();
          }
        }, 150);
      });
    }

    /**
     * Closes the dialog.
     */
    close() {
      if (!this.isOpen) return;

      this.isOpen = false;

      this.dialog.classList.remove(
        "scale-100",
        "opacity-100",
        "animate-bounce-in",
      );
      this.dialog.classList.add("scale-95", "opacity-0");

      setTimeout(() => {
        this.backdrop.classList.add("hidden");
        this._cleanup();
      }, 300);
    }

    /**
     * Cleans up the dialog and removes it from DOM.
     * @private
     */
    _cleanup() {
      if (this._boundKeyDownHandler) {
        document.removeEventListener("keydown", this._boundKeyDownHandler);
        this._boundKeyDownHandler = null;
      }

      if (this.backdrop && this.backdrop.parentNode) {
        this.backdrop.parentNode.removeChild(this.backdrop);
      }

      this.backdrop = null;
      this.dialog = null;
      this.resolvePromise = null;
      this.rejectPromise = null;
    }

    /**
     * Escapes HTML to prevent XSS.
     * @private
     * @param {string} text - Text to escape.
     * @returns {string} Escaped text.
     */
    _escapeHtml(text) {
      const div = document.createElement("div");
      div.textContent = text;
      return div.innerHTML;
    }

    /**
     * Static method to create and show a dialog
     * @param {object} options - Dialog options.
     * @returns {Promise} Promise that resolves with the user's choice.
     */
    static fire(options = {}) {
      const dialog = new SuperDialog(options);
      return dialog.show();
    }
  }

  // Add custom CSS for tapering bounce animation
  const style = document.createElement("style");
  style.textContent = `
    @keyframes bounce-in {
      0% {
        transform: scale(0.8);
      }
      20% {
        transform: scale(1.1);
      }
      40% {
        transform: scale(0.92);
      }
      60% {
        transform: scale(1.03);
      }
      80% {
        transform: scale(0.99);
      }
      100% {
        transform: scale(1);
      }
    }
    
    .animate-bounce-in {
      animation: bounce-in 0.6s cubic-bezier(0.25, 0.46, 0.45, 0.94);
    }
  `;
  document.head.appendChild(style);

  window.SuperDialog = SuperDialog;
  window.SuperDialogInitialized = true;
}
