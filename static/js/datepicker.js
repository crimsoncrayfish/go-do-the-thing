/**
 * SuperDatePicker Class
 * A customizable and accessible date picker component.
 * IMPORTANT: This library relies on Tailwind CSS being set up and processed
 *            in the consuming application for proper styling.
 *            It specifically leverages the following Tailwind color classes from your config:
 *            primary-dark, secondary-dark, accent-dark, border-dark
 *            primary-light, secondary-light, accent-light, border-light
 *            as well as generic Tailwind text-white, text-gray-900, text-gray-100, text-gray-500.
 *
 * This version REQUIRES initialization on an HTML <input> element.
 *
 * @param {HTMLInputElement} inputElement - The HTML <input> element to initialize the date picker on.
 * @param {object} [options={}] - Configuration options for the date picker.
 * @param {string} [options.defaultDate=null] - An initial date in 'YYYY-MM-DD' format.
 * @param {string} [options.dateFormat='YYYY-MM-DD'] - The format for displaying the date in the input. (Currently only 'YYYY-MM-DD' supported).
 * @param {string[]} [options.monthNames=defaultMonthNames] - Custom month names.
 * @param {string[]} [options.dayNamesShort=defaultDayNamesShort] - Custom short day names.
 */
class SuperDatePicker {
  static DEFAULT_MONTH_NAMES = [
    "January",
    "February",
    "March",
    "April",
    "May",
    "June",
    "July",
    "August",
    "September",
    "October",
    "November",
    "December",
  ];
  static DEFAULT_DAY_NAMES_SHORT = ["Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"];

  /**
   * Constructs a new SuperDatePicker instance.
   * @param {HTMLInputElement} inputElement - The <input> element to attach the date picker to.
   * @param {object} [options={}] - Options for the date picker.
   */
  constructor(inputElement, options = {}) {
    if (!(inputElement instanceof HTMLInputElement)) {
      console.error(
        "SuperDatePicker: Initialization failed. The SuperDatePicker must be initialized on an HTML <input> element.",
        inputElement,
      );
      throw new Error(
        "SuperDatePicker: Initialization failed. Invalid element type.",
      );
    }

    this.options = {
      defaultDate: null,
      dateFormat: "YYYY-MM-DD",
      monthNames: SuperDatePicker.DEFAULT_MONTH_NAMES,
      dayNamesShort: SuperDatePicker.DEFAULT_DAY_NAMES_SHORT,
      ...options,
    };

    this.input = inputElement;
    this.container = null;
    this.popup = null;

    this.selectedDate = null;
    this.currentDate = new Date();
    this.isOpen = false;

    this._initDOM();
    this._attachCoreEvents();

    const initialDate = this.options.defaultDate || this.input.value;
    if (initialDate) {
      this.setDate(initialDate);
    } else {
      this.renderCalendar();
    }
  }

  /**
   * Initializes the DOM structure for the date picker (Input-only mode).
   * @private
   */
  _initDOM() {
    if (
      !this.input.parentElement ||
      !this.input.parentElement.classList.contains("super-datepicker-wrapper")
    ) {
      const wrapper = document.createElement("div");
      wrapper.className = "super-datepicker-wrapper relative";
      this.input.parentElement.insertBefore(wrapper, this.input);
      wrapper.appendChild(this.input);
    }
    this.container = this.input.parentElement;

    this.input.classList.add(
      "datepicker-input",
      "w-full",
      "px-3",
      "py-2",
      "border",
      "rounded-lg",
      "focus:outline-none",
      "focus:ring",
      "cursor-pointer",
      "bg-primary-light",
      "text-text-on-light",
      "dark:bg-primary-dark",
      "dark:text-text-on-dark",
      "border-border-light",
      "dark:border-border-dark",
      "focus:ring-accent-light",
      "dark:focus:ring-accent-dark",
    );
    this.input.setAttribute("readonly", "readonly");
    this.input.setAttribute("type", "text");
    this.input.setAttribute("placeholder", this.options.dateFormat);
    this.input.setAttribute("aria-haspopup", "dialog");
    this.input.setAttribute("aria-expanded", "false");

    const popup = document.createElement("div");
    popup.classList.add(
      "datepicker-popup",
      "absolute",
      "w-72",
      "rounded-lg",
      "shadow-lg",
      "z-50",
      "hidden",
      "bg-primary-light",
      "border",
      "border-secondary-light",
      "dark:bg-primary-dark",
      "dark:border-secondary-dark",
    );

    document.body.appendChild(popup);

    this.popup = popup;

    this.popup.setAttribute("role", "dialog");
    this.popup.setAttribute("aria-modal", "true");
    this.popup.setAttribute("aria-label", `Calendar`);
  }

  /**
   * Attaches global and input-specific event listeners.
   * @private
   */
  _attachCoreEvents() {
    this.popup.addEventListener("click", (e) => e.stopPropagation());

    this.input.addEventListener("click", this._handleInputClick.bind(this));

    document.addEventListener("click", this._handleDocumentClick.bind(this));

    document.addEventListener("keydown", this._handleKeyDown.bind(this));

    this.popup.addEventListener("click", this._handleCalendarClick.bind(this));

    window.addEventListener("resize", this._repositionPopup.bind(this));
    window.addEventListener("scroll", this._repositionPopup.bind(this), true);
  }

  /**
   * Repositions the popup absolutely based on the input's position,
   * performing viewport collision detection to keep it on screen.
   * @private
   */
  _repositionPopup() {
    if (!this.isOpen) return;

    const inputRect = this.input.getBoundingClientRect();
    const popupRect = this.popup.getBoundingClientRect();

    const viewportWidth =
      window.innerWidth || document.documentElement.clientWidth;
    const viewportHeight =
      window.innerHeight || document.documentElement.clientHeight;

    let finalTop =
      inputRect.bottom + (window.scrollY || window.pageYOffset) + 8;
    let finalLeft = inputRect.left + (window.scrollX || window.pageXOffset);

    if (
      finalLeft + popupRect.width >
      viewportWidth + (window.scrollX || window.pageXOffset)
    ) {
      finalLeft =
        inputRect.right +
        (window.scrollX || window.pageXOffset) -
        popupRect.width;

      if (finalLeft < (window.scrollX || window.pageXOffset)) {
        finalLeft = window.scrollX || window.pageXOffset;
      }
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
  _handleInputClick(e) {
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
      !this.container.contains(e.target) &&
      !this.popup.contains(e.target)
    ) {
      this.togglePopup(false);
      this.input.focus();
    }
  }

  /**
   * Handles click events delegated from the popup for navigation and date selection.
   * @private
   * @param {Event} e - The click event.
   */
  _handleCalendarClick(e) {
    if (e.target.classList.contains("datepicker-prev")) {
      this.currentDate.setMonth(this.currentDate.getMonth() - 1);
      this.renderCalendar();
    } else if (e.target.classList.contains("datepicker-next")) {
      this.currentDate.setMonth(this.currentDate.getMonth() + 1);
      this.renderCalendar();
    } else if (e.target.hasAttribute("data-date")) {
      const date = e.target.getAttribute("data-date");
      this.setDate(date);
      this.togglePopup(false);
      this.input.focus();
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
    const allDateButtons = Array.from(
      this.popup.querySelectorAll("button[data-date]"),
    );
    const currentDayIdx = allDateButtons.indexOf(currentFocusElement);

    if (currentDayIdx !== -1) {
      const nextDayIdx = currentDayIdx + direction;
      if (nextDayIdx >= 0 && nextDayIdx < allDateButtons.length) {
        allDateButtons[nextDayIdx].focus();
      } else {
        const newMonth = new Date(
          this.currentDate.getFullYear(),
          this.currentDate.getMonth() + direction,
          1,
        );
        this.currentDate = newMonth;
        this.renderCalendar();
        setTimeout(() => {
          const newButtons = this.popup.querySelectorAll("button[data-date]");
          if (newButtons.length > 0) {
            (direction === -1
              ? newButtons[newButtons.length - 1]
              : newButtons[0]
            )?.focus();
          } else {
            (direction === -1
              ? this.popup.querySelector(".datepicker-prev")
              : this.popup.querySelector(".datepicker-next")
            )?.focus();
          }
        }, 0);
      }
    } else {
      if (currentFocusElement.classList.contains("datepicker-prev")) {
        this.popup.querySelector(".datepicker-next")?.focus();
      } else if (currentFocusElement.classList.contains("datepicker-next")) {
        this.popup.querySelector(".datepicker-prev")?.focus();
      } else {
        if (allDateButtons.length > 0) {
          (direction === -1
            ? allDateButtons[allDateButtons.length - 1]
            : allDateButtons[0]
          )?.focus();
        } else {
          currentFocusElement.focus();
        }
      }
    }
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
    const allDateButtons = Array.from(
      this.popup.querySelectorAll("button[data-date]"),
    );
    const currentDayIdx = allDateButtons.indexOf(currentFocusElement);

    if (currentDayIdx !== -1) {
      const nextDayIdx = currentDayIdx + direction * 7;
      if (nextDayIdx >= 0 && nextDayIdx < allDateButtons.length) {
        allDateButtons[nextDayIdx].focus();
      } else {
        const newMonth = new Date(
          this.currentDate.getFullYear(),
          this.currentDate.getMonth() + direction,
          1,
        );
        this.currentDate = newMonth;
        this.renderCalendar();
        setTimeout(() => {
          const newButtons = this.popup.querySelectorAll("button[data-date]");
          if (newButtons.length > 0) {
            const targetColIdx = currentDayIdx % 7;
            let focusTarget = null;
            for (
              let i = direction === 1 ? 0 : newButtons.length - 1;
              direction === 1 ? i < newButtons.length : i >= 0;
              i += direction
            ) {
              if (
                newButtons[i].parentElement &&
                newButtons[i].parentElement.closest('[role="row"]') &&
                Array.from(
                  newButtons[i].parentElement.closest('[role="row"]').children,
                ).indexOf(newButtons[i].parentElement) === targetColIdx
              ) {
                focusTarget = newButtons[i];
                break;
              }
            }
            (
              focusTarget ||
              (direction === 1
                ? newButtons[0]
                : newButtons[newButtons.length - 1])
            )?.focus();
          } else {
            (direction === 1
              ? this.popup.querySelector(".datepicker-next")
              : this.popup.querySelector(".datepicker-prev")
            )?.focus();
          }
        }, 0);
      }
    } else {
      if (currentFocusElement.classList.contains("datepicker-prev")) {
        if (direction === 1) {
          this.popup.querySelector(".datepicker-next")?.focus();
        } else {
          currentFocusElement.focus();
        }
      } else if (currentFocusElement.classList.contains("datepicker-next")) {
        if (direction === -1) {
          this.popup.querySelector(".datepicker-prev")?.focus();
        } else {
          currentFocusElement.focus();
        }
      } else {
        if (allDateButtons.length > 0) {
          (direction === 1
            ? allDateButtons[0]
            : allDateButtons[allDateButtons.length - 1]
          )?.focus();
        } else {
          currentFocusElement.focus();
        }
      }
    }
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
    const currentFocusIdx = focusableElements.indexOf(currentFocusElement);

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
          currentFocusElement.hasAttribute("data-date")
        ) {
          currentFocusElement.click();
        } else if (
          currentFocusElement &&
          currentFocusElement.classList.contains("datepicker-prev")
        ) {
          this.currentDate.setMonth(this.currentDate.getMonth() - 1);
          this.renderCalendar();
          e.preventDefault();
        } else if (
          currentFocusElement &&
          currentFocusElement.classList.contains("datepicker-next")
        ) {
          this.currentDate.setMonth(this.currentDate.getMonth() + 1);
          this.renderCalendar();
          e.preventDefault();
        }
        break;

      case "Tab":
        if (focusableElements.length === 0) return;
        e.preventDefault();
        let nextIdx;
        if (e.shiftKey) {
          nextIdx =
            (currentFocusIdx - 1 + focusableElements.length) %
            focusableElements.length;
        } else {
          nextIdx = (currentFocusIdx + 1) % focusableElements.length;
        }
        focusableElements[nextIdx].focus();
        break;
    }
  }

  /**
   * Renders the calendar grid and month/year navigation.
   */
  renderCalendar() {
    const year = this.currentDate.getFullYear();
    const month = this.currentDate.getMonth();

    const firstDay = new Date(year, month, 1);
    const lastDay = new Date(year, month + 1, 0);
    const startDay = firstDay.getDay();
    const daysInMonth = lastDay.getDate();

    let weeksHtml = "";
    let dayCounter = 1;

    for (let i = 0; i < 6; i++) {
      let weekHtml = "";
      let hasDaysInWeek = false;

      for (let j = 0; j < 7; j++) {
        if ((i === 0 && j < startDay) || dayCounter > daysInMonth) {
          weekHtml += '<div class="w-full aspect-square"></div>';
        } else {
          const dateStr = `${year}-${String(month + 1).padStart(2, "0")}-${String(dayCounter).padStart(2, "0")}`;
          const isSelected = this.selectedDate === dateStr;
          const isToday = this._isToday(year, month, dayCounter);

          weekHtml += `
            <button
              type="button"
              data-date="${dateStr}"
              tabindex="-1" 
              class="w-full aspect-square rounded text-sm
                ${isSelected ? "bg-accent-light text-white dark:bg-accent-dark" : "text-primary-dark dark:text-primary-light hover:bg-secondary-light dark:hover:bg-secondary-dark"}
                ${isToday ? (isSelected ? "" : "font-bold") : ""}"
              aria-label="${this.options.monthNames[month]} ${dayCounter}, ${year}"
              aria-selected="${isSelected ? "true" : "false"}"
              role="gridcell"
            >
              ${dayCounter}
            </button>
          `;
          dayCounter++;
          hasDaysInWeek = true;
        }
      }
      if (hasDaysInWeek || dayCounter <= daysInMonth + 7) {
        weeksHtml += `<div class="grid grid-cols-7 gap-1" role="row">${weekHtml}</div>`;
      }
    }

    const currentMonthName = this.options.monthNames[month];
    this.popup.innerHTML = `
      <div class="flex justify-between items-center px-2 py-2">
        <button type="button" class="datepicker-prev p-1 rounded hover:bg-secondary-light dark:hover:bg-secondary-dark text-primary-dark dark:text-primary-light" aria-label="Previous month">${String.fromCharCode(8592)}</button>
        <span class="font-semibold text-primary-dark dark:text-primary-light" aria-live="polite">${currentMonthName} ${year}</span>
        <button type="button" class="datepicker-next p-1 rounded hover:bg-secondary-light dark:hover:bg-secondary-dark text-primary-dark dark:text-primary-light" aria-label="Next month">${String.fromCharCode(8594)}</button>
      </div>
      <div class="grid grid-cols-7 gap-1 px-2" role="rowgroup">
        ${this.options.dayNamesShort.map((d) => `<div class="text-xs text-center text-gray-500" role="columnheader">${d}</div>`).join("")}
      </div>
      <div class="px-2 pb-2" role="grid">
        ${weeksHtml}
      </div>
    `;

    this.popup.setAttribute(
      "aria-label",
      `Calendar for ${currentMonthName} ${year}`,
    );

    const allRenderedDateButtons =
      this.popup.querySelectorAll("button[data-date]");
    let tabFocusedDateButton = null;

    if (this.selectedDate) {
      tabFocusedDateButton = this.popup.querySelector(
        `button[data-date="${this.selectedDate}"]`,
      );
      if (tabFocusedDateButton) {
        tabFocusedDateButton.tabIndex = 0;
      }
    }

    if (!tabFocusedDateButton) {
      const today = new Date();
      const todayStr = `${today.getFullYear()}-${String(today.getMonth() + 1).padStart(2, "0")}-${String(today.getDate()).padStart(2, "0")}`;
      tabFocusedDateButton = this.popup.querySelector(
        `button[data-date="${todayStr}"]`,
      );

      if (tabFocusedDateButton) {
        tabFocusedDateButton.tabIndex = 0;
      }
    }

    if (!tabFocusedDateButton && allRenderedDateButtons.length > 0) {
      allRenderedDateButtons[0].tabIndex = 0;
    }
  }

  /**
   * Toggles the visibility of the date picker popup.
   * @param {boolean} show - True to show, false to hide.
   */
  togglePopup(show) {
    this.isOpen = show;
    this.popup.classList.toggle("hidden", !show);
    this.input.setAttribute("aria-expanded", show.toString());

    if (show) {
      this.renderCalendar();
      this._repositionPopup();

      setTimeout(() => {
        const focusBtn =
          this.popup.querySelector('button[tabindex="0"]') ||
          this.popup.querySelector(".datepicker-prev");
        if (focusBtn) {
          focusBtn.focus();
        }
      }, 0);
    }
  }

  /**
   * Sets the selected date for the date picker.
   * @param {string} dateString - The date to set in 'YYYY-MM-DD' format.
   */
  setDate(dateString) {
    if (
      !/^\d{4}-\d{2}-\d{2}$/.test(dateString) ||
      isNaN(new Date(dateString).getTime())
    ) {
      console.warn(
        `SuperDatePicker: Invalid date string "${dateString}" provided to setDate. Skipping.`,
      );
      return;
    }

    this.selectedDate = dateString;
    this.input.value = dateString;

    const parts = dateString.split("-").map(Number);
    const dateObj = new Date(parts[0], parts[1] - 1, parts[2]);
    this.currentDate = dateObj;

    this.renderCalendar();
    this.input.dispatchEvent(new Event("change", { bubbles: true }));
  }

  /**
   * Gets the currently selected date.
   * @returns {string|null} The selected date in 'YYYY-MM-DD' format, or null if no date is selected.
   */
  getDate() {
    return this.selectedDate;
  }

  /**
   * Helper to check if a given day is today's date.
   * @private
   * @param {number} year - The year of the day to check.
   * @param {number} month - The 0-indexed month of the day to check.
   * @param {number} day - The day of the month to check.
   * @returns {boolean} True if the date is today, false otherwise.
   */
  _isToday(year, month, day) {
    const today = new Date();
    return (
      today.getFullYear() === year &&
      today.getMonth() === month &&
      today.getDate() === day
    );
  }
}

function autoInitSuperDatePickers() {
  document.querySelectorAll("input[data-super-datepicker]").forEach((input) => {
    if (!input._superDatePicker) {
      input._superDatePicker = new SuperDatePicker(input);
    }
  });
}

if (typeof window !== "undefined") {
  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", autoInitSuperDatePickers);
  } else {
    autoInitSuperDatePickers();
  }
}

if (typeof window !== "undefined" && window.htmx) {
  document.body.addEventListener("htmx:afterSwap", () => {
    autoInitSuperDatePickers();
  });
}
