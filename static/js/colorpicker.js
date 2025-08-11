if (window.SuperColorPickerInitialized) {
  if (typeof console !== "undefined" && console.warn) {
    console.warn(
      "SuperColorPicker already initialized, skipping duplicate execution.",
    );
  }
} else {
  const SPECIFIC_COLOR_CLASS_REGEX = /\b(dark:)?bg-\S+-(light|dark)\b/g;
  const colorPickerInstances = new Set();

  /**
   * SuperColorPicker Class
   * A customizable and accessible color picker component.
   * IMPORTANT: This library relies on Tailwind CSS being set up and processed
   *            in the consuming application for proper styling.
   *
   * This version REQUIRES initialization on an HTML <input> element.
   *
   * @param {HTMLInputElement} inputElement - The HTML <input> element to initialize the color picker on.
   * @param {object} [options={}] - Configuration options for the color picker.
   * @param {string[]} [options.colors={}] - The names of the colors used. Needs to match the tailwind config.
   * Requires both light and dark colors e.g. red-light & red-dark
   * @param {string[]} [options.rowWidth=5] - The number of colors in a row
   */
  class SuperColorPicker {
    /**
     * Constructs a new SuperColorPicker instance.
     * @param {HTMLInputElement} inputElement - The <input> element to attach the color picker to.
     * @param {object} [options={}] - Options for the color picker.
     */
    constructor(inputElement, options = {}) {
      if (!(inputElement instanceof HTMLInputElement)) {
        console.error(
          "SuperColorPicker: Initialization failed. The SuperColorPicker must be initialized on an HTML <input> element.",
          inputElement,
        );
        throw new Error(
          "SuperColorPicker: Initialization failed. Invalid element type.",
        );
      }

      this.input = inputElement;
      this.container = null;
      this.popup = null;
      this.button = null;
      this.swatchSpan = null;
      this.selectedColor = null;
      this.isOpen = false;

      this.options = {
        colors: [],
        rowWidth: 5,
        defaultColor: null,
        ...options,
      };

      if (
        !Array.isArray(this.options.colors) ||
        this.options.colors.length === 0
      ) {
        console.error(
          "SuperColorPicker: 'colors' option must be an array of color names and cannot be empty.",
        );
      }

      this._initDOM();
      this._attachCoreEvents();

      colorPickerInstances.add(this);

      const initialColor =
        this.input.value || this.options.defaultColor || this.options.colors[0];
      if (initialColor) {
        this.setColor(initialColor);
      } else {
        this.render();
      }
    }

    /**
     * Initializes the DOM structure for the color picker (Input-only mode).
     * @private
     */
    _initDOM() {
      if (
        !this.input.parentElement ||
        !this.input.parentElement.classList.contains(
          "super-colorpicker-wrapper",
        )
      ) {
        const wrapper = document.createElement("div");
        wrapper.className = "super-colorpicker-wrapper relative";
        this.input.parentElement.insertBefore(wrapper, this.input);
        wrapper.appendChild(this.input);
      }
      this.container = this.input.parentElement;

      this.input.setAttribute("type", "hidden");
      this.input.classList.add("color-input");
      this.input.setAttribute("aria-haspopup", "dialog");
      this.input.setAttribute("aria-expanded", "false");

      const button = document.createElement("button");
      button.setAttribute("type", "button");
      const inputId = this.input.id;
      button.setAttribute("id", `${inputId}-button`);
      button.classList.add(
        "flex",
        "h-6",
        "w-6",
        "items-center",
        "justify-center",
        "rounded-full",
        "transition-colors",
        "hover:bg-black/10",
        "dark:hover:bg-white/10",
      );
      button.setAttribute("aria-label", "Open color picker");
      button.setAttribute("aria-controls", `${inputId}-popup` || "");
      button.setAttribute("aria-haspopup", "dialog");
      button.setAttribute("aria-expanded", "false");

      const swatchSpan = document.createElement("span");
      swatchSpan.classList.add(
        "selected-color-swatch",
        "inline-block",
        "h-4",
        "w-4",
        "rounded-full",
        "border",
        "border-border-light",
        "dark:border-border-dark",
      );
      swatchSpan.setAttribute("aria-hidden", "true");
      button.appendChild(swatchSpan);
      this.input.after(button);
      this.button = button;
      this.swatchSpan = swatchSpan;

      const popup_temp = document.getElementById(`${inputId}-popup`);
      if (!popup_temp) {
        const popup = document.createElement("div");
        popup.classList.add(
          "colorpicker-popup",
          "absolute",
          "rounded-lg",
          "z-50",
          "hidden",
          "bg-secondary-light",
          "border-border-light",
          "dark:bg-secondary-dark",
          "dark:border-border-dark",
          "shadow-lg",
          "border",
          "p-1",
        );
        popup.setAttribute("id", `${inputId}-popup`);
        document.body.appendChild(popup);
        this.popup = popup;
      } else {
        this.popup = popup_temp;
      }

      this.popup.setAttribute("role", "dialog");
      this.popup.setAttribute("aria-modal", "true");
      this.popup.setAttribute("aria-label", `Color Picker`);
    }

    /**
     * Attaches global and input-specific event listeners.
     * @private
     */
    _attachCoreEvents() {
      this._boundButtonClick = this._handleButtonClick.bind(this);
      this._boundDocumentClick = this._handleDocumentClick.bind(this);
      this._boundKeyDown = this._handleKeyDown.bind(this);
      this._boundColorPickerClick = this._handleColorPickerClick.bind(this);
      this._boundRepositionPopup = this._repositionPopup.bind(this);

      this.popup.addEventListener("click", (e) => e.stopPropagation());
      this.button.addEventListener("click", this._boundButtonClick);
      document.addEventListener("click", this._boundDocumentClick);
      document.addEventListener("keydown", this._boundKeyDown);
      this.popup.addEventListener("click", this._boundColorPickerClick);
      window.addEventListener("resize", this._boundRepositionPopup);
      window.addEventListener("scroll", this._boundRepositionPopup, {
        capture: true,
        passive: true,
      });
    }

    /**
     * Repositions the popup absolutely based on the input's position,
     * performing viewport collision detection to keep it on screen.
     * @private
     */
    _repositionPopup() {
      if (!this.isOpen) return;

      const inputRect = this.button.getBoundingClientRect();
      const popupRect = this.popup.getBoundingClientRect();

      const viewportWidth =
        window.innerWidth || document.documentElement.clientWidth;
      const viewportHeight =
        window.innerHeight || document.documentElement.clientHeight;

      let finalTop =
        inputRect.bottom + (window.scrollY || window.pageYOffset) + 8;
      let finalLeft =
        inputRect.left +
        (window.scrollX || window.pageXOffset) +
        (inputRect.width - popupRect.width) / 2;

      if (
        finalLeft + popupRect.width >
        viewportWidth + (window.scrollX || window.pageXOffset)
      ) {
        finalLeft =
          inputRect.right +
          (window.scrollX || window.pageXOffset) -
          popupRect.width;
      }
      if (finalLeft < (window.scrollX || window.pageXOffset)) {
        finalLeft = window.scrollX || window.pageXOffset;
      }
      if (
        finalTop + popupRect.height >
        viewportHeight + (window.scrollY || window.pageYOffset)
      ) {
        finalTop =
          inputRect.top +
          (window.scrollY || window.pageYOffset) -
          popupRect.height -
          8;
        if (finalTop < (window.scrollY || window.pageYOffset)) {
          finalTop = window.scrollY || window.pageYOffset;
        }
      }

      this.popup.style.top = `${finalTop}px`;
      this.popup.style.left = `${finalLeft}px`;
      this.popup.style.position = "absolute";
    }

    /**
     * Handles click event on the input field.
     * @private
     * @param {Event} e - The click event.
     */
    _handleButtonClick(e) {
      e.stopPropagation();
      this.togglePopup(true);
    }

    /**
     * Handles click events on the document to close the popup.
     * @private
     * @param {Event} e - The click event.
     */
    _handleDocumentClick(e) {
      if (
        this.isOpen &&
        !this.button.contains(e.target) &&
        !this.input.contains(e.target) &&
        !this.popup.contains(e.target)
      ) {
        this.togglePopup(false);
        this.button.focus();
      }
    }

    /**
     * Handles click events delegated from the popup for navigation and color selection.
     * @private
     * @param {Event} e - The click event.
     */
    _handleColorPickerClick(e) {
      if (e.target.classList.contains("colorpicker-option")) {
        const color = e.target.getAttribute("data-color");
        this.setColor(color);
        this.togglePopup(false);
        this.button.focus();
      }
    }

    /**
     * Handles horizontal (left/right) keyboard navigation.
     * @private
     * @param {KeyboardEvent} e - The keyboard event.
     * @param {number} direction - -1 for left, 1 for right.
     */
    _handleHorizontalNavigation(e, direction) {
      e.preventDefault();
      const currentFocusElement = document.activeElement;
      const allColorButtons = Array.from(
        this.popup.querySelectorAll("button[data-color]"),
      );

      if (!allColorButtons.includes(currentFocusElement)) {
        if (allColorButtons.length > 0) {
          allColorButtons[0].focus();
        }
        return;
      }

      const currentColorIdx = allColorButtons.indexOf(currentFocusElement);
      const nextColorIdx = currentColorIdx + direction;
      if (nextColorIdx < 0) {
        nextColorIdx = allColorButtons.length - 1;
      } else if (nextColorIdx >= allColorButtons.length) {
        nextColorIdx = 0;
      }
      allColorButtons[nextColorIdx].focus();
    }

    /**
     * Handles vertical (up/down) keyboard navigation.
     * @private
     * @param {KeyboardEvent} e - The keyboard event.
     * @param {number} direction - -1 for up, 1 for down.
     */
    _handleVerticalNavigation(e, direction) {
      e.preventDefault();
      const currentFocusElement = document.activeElement;
      const allColorButtons = Array.from(
        this.popup.querySelectorAll("button[data-color]"),
      );

      if (!allColorButtons.includes(currentFocusElement)) {
        allColorButtons[0].focus();
        return;
      }
      const currentColorIdx = allColorButtons.indexOf(currentFocusElement);
      const nextColorIdx = currentColorIdx + direction * this.options.rowWidth;
      if (nextColorIdx < 0) {
        nextColorIdx += this.options.rowWidth;
      } else if (nextColorIdx >= allColorButtons.length) {
        nextColorIdx -= this.options.rowWidth;
      }
      allColorButtons[nextColorIdx].focus();
    }

    /**
     * Main keyboard event handler.
     * @private
     * @param {KeyboardEvent} e - The keyboard event.
     */
    _handleKeyDown(e) {
      if (!this.isOpen) return;

      const focusableElements = Array.from(
        this.popup.querySelectorAll(
          'button:not([disabled]), [tabindex]:not([tabindex="-1"])',
        ),
      ).filter(
        (el) =>
          window.getComputedStyle(el).display !== "none" &&
          window.getComputedStyle(el).visibility !== "hidden",
      );
      const currentFocusElement = document.activeElement;

      switch (e.key) {
        case "Escape":
          e.preventDefault();
          this.togglePopup(false);
          this.input.focus();
          break;

        case "ArrowLeft":
        case "Left":
          this._handleHorizontalNavigation(e, -1);
          break;

        case "ArrowRight":
        case "Right":
          this._handleHorizontalNavigation(e, 1);
          break;

        case "ArrowUp":
        case "Up":
          this._handleVerticalNavigation(e, -1);
          break;

        case "ArrowDown":
        case "Down":
          this._handleVerticalNavigation(e, 1);
          break;

        case "Enter":
        case " ":
          if (
            currentFocusElement &&
            currentFocusElement.hasAttribute("data-color")
          ) {
            currentFocusElement.click();
          }
          break;

        case "Tab":
          if (focusableElements.length === 0) return;
          e.preventDefault();
          const currentTabIdx = focusableElements.indexOf(currentFocusElement);
          let nextIdx;
          if (e.shiftKey) {
            nextIdx =
              (currentTabIdx - 1 + focusableElements.length) %
              focusableElements.length;
          } else {
            nextIdx = (currentTabIdx + 1) % focusableElements.length;
          }
          focusableElements[nextIdx].focus();
          break;
      }
    }

    /**
     * Renders the color grid
     */
    render() {
      let rowsHtml = "";
      for (const color of this.options.colors) {
        const isSelected = this.selectedColor === color;
        rowsHtml += `
            <button
              type="button"
              data-color="${color}"
              tabindex="-1" 
              class="colorpicker-option h-6 w-6 rounded-md bg-${color}-light dark:bg-${color}-dark ${isSelected ? "" : "border border-secondary-light dark:border-secondary-dark"}"
              aria-label="${color}"
              aria-selected="${isSelected ? "true" : "false"}"
              role="gridcell"
            />
          `;
      }

      this.popup.innerHTML = `
      <div class="grid grid-cols-${this.options.rowWidth} gap-1" role="grid">
        ${rowsHtml}
      </div>
    `;

      this.popup.setAttribute("aria-label", "Color Picker");

      const allRenderedColorButtons =
        this.popup.querySelectorAll("button[data-color]");
      let tabFocusedColorButton = null;

      if (this.selectedColor) {
        tabFocusedColorButton = this.popup.querySelector(
          `button[data-color="${this.selectedColor}"]`,
        );
        if (tabFocusedColorButton) {
          tabFocusedColorButton.tabIndex = 0;
        }
      }
      if (!tabFocusedColorButton && allRenderedColorButtons.length > 0) {
        allRenderedColorButtons[0].tabIndex = 0;
      }
    }

    /**
     * Toggles the visibility of the color picker popup.
     * @param {boolean} show - True to show, false to hide.
     */
    togglePopup(show) {
      this.isOpen = show;
      this.popup.classList.toggle("hidden", !show);
      this.input.setAttribute("aria-expanded", show.toString());
      this.button.setAttribute("aria-expanded", show.toString());

      if (show) {
        this.render();
        this._repositionPopup();

        setTimeout(() => {
          const focusBtn = this.popup.querySelector('button[tabindex="0"]');
          if (focusBtn) {
            focusBtn.focus();
          }
        }, 0);
      } else {
        this.popup
          .querySelectorAll("button[data-color]")
          .forEach((btn) => (btn.tabIndex = -1));
      }
    }

    /**
     * Sets the selected color for the color picker.
     * @param {string} colorString - The color to set as a string e.g. 'red'
     */
    setColor(colorString) {
      if (!this.options.colors.includes(colorString)) {
        console.warn(
          `SuperColorPicker: Invalid color string "${colorString}" provided to setColor. Skipping.`,
        );
        return;
      }
      this.selectedColor = colorString;
      this.input.value = colorString;
      if (this.swatchSpan) {
        const currentClasses = Array.from(this.swatchSpan.classList);
        currentClasses.forEach((cls) => {
          if (cls.match(SPECIFIC_COLOR_CLASS_REGEX)) {
            this.swatchSpan.classList.remove(cls);
          }
        });

        const lightClass = `bg-${colorString}-light`;
        const darkClass = `dark:bg-${colorString}-dark`;
        this.swatchSpan.classList.add(lightClass, darkClass);
      }
      this.render();
      this.input.dispatchEvent(new Event("change", { bubbles: true }));
    }

    /**
     * Gets the currently selected color.
     * @returns {string|null} The selected color as a string, or null if no color is selected.
     */
    getColor() {
      return this.selectedColor;
    }

    /**
     * Cleans up the color picker and removes all event listeners.
     */
    cleanup() {
      if (this._boundButtonClick) {
        this.button.removeEventListener("click", this._boundButtonClick);
      }
      if (this._boundDocumentClick) {
        document.removeEventListener("click", this._boundDocumentClick);
      }
      if (this._boundKeyDown) {
        document.removeEventListener("keydown", this._boundKeyDown);
      }
      if (this._boundColorPickerClick) {
        this.popup.removeEventListener("click", this._boundColorPickerClick);
      }
      if (this._boundRepositionPopup) {
        window.removeEventListener("resize", this._boundRepositionPopup);
        window.removeEventListener("scroll", this._boundRepositionPopup, {
          capture: true,
          passive: true,
        });
      }

      if (this.popup && this.popup.parentNode) {
        this.popup.parentNode.removeChild(this.popup);
      }

      if (this.button && this.button.parentNode) {
        this.button.parentNode.removeChild(this.button);
      }

      colorPickerInstances.delete(this);
      this.input = null;
      this.container = null;
      this.popup = null;
      this.button = null;
      this.swatchSpan = null;
      this._boundButtonClick = null;
      this._boundDocumentClick = null;
      this._boundKeyDown = null;
      this._boundColorPickerClick = null;
      this._boundRepositionPopup = null;
    }
  }

  function autoInitSuperColorPickers() {
    document
      .querySelectorAll("input[data-super-colorpicker]")
      .forEach((input) => {
        if (input._superColorPicker) {
          return;
        }
        const colorsData = input.getAttribute("data-colors");
        let colors = [];
        if (colorsData) {
          try {
            colors = JSON.parse(colorsData);
          } catch (e) {
            console.error(
              `SuperColorPicker: Failed to parse 'data-colors' attribute for input ID '${input.id || "N/A"}'. Make sure it's valid JSON array.`,
              e,
            );
          }
        }
        const rowWidthData = input.getAttribute("data-row-width");
        const rowWidth = rowWidthData ? parseInt(rowWidthData, 10) : undefined;
        const defaultColorData = input.getAttribute("data-default-color");
        input._superColorPicker = new SuperColorPicker(input, {
          colors: colors.length > 0 ? colors : undefined,
          rowWidth: rowWidth,
          defaultColor: defaultColorData,
        });
      });
  }

  function cleanupAllInstances() {
    const instancesToCleanup = Array.from(colorPickerInstances);
    instancesToCleanup.forEach((instance) => {
      instance.cleanup();
    });
  }

  document.addEventListener('htmx:beforeSwap', cleanupAllInstances);
  document.addEventListener('htmx:afterSwap', autoInitSuperColorPickers);

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', autoInitSuperColorPickers);
  } else {
    autoInitSuperColorPickers();
  }

  window.SuperColorPickerInitialized = true;
}
