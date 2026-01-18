/**
 * Card flip animations for task and project cards.
 * 
 * Cards use a 3D rotateX flip animation with front/back faces.
 * The back face shows a status message during HTMX requests.
 * 
 * Expected HTML structure:
 *   #[prefix]-card-[id]          - Outer wrapper with hx-on handlers
 *   #[prefix]-card-inner-[id]    - Inner element that rotates
 *   #[prefix]-card-back-*-[id]   - Back face elements (hidden by default)
 */

const CardFlip = (function () {
  /**
   * Extract the numeric ID from a card element's id attribute.
   * @param {HTMLElement} cardEl - Element with id like "task-card-123"
   * @param {string} prefix - The prefix (e.g., "task-card")
   * @returns {string|null} The extracted ID or null
   */
  function getIdFromCard(cardEl, prefix) {
    if (!cardEl || !cardEl.id) return null;
    return cardEl.id.replace(prefix + "-", "");
  }

  /**
   * Flip a card to show its back face.
   * @param {string} prefix - Card prefix ("task-card" or "project-card")
   * @param {string} id - The card's numeric ID
   * @param {string} backType - The back face type to show (e.g., "update", "delete", "time-start")
   */
  function flipToBack(prefix, id, backType) {
    const backId = prefix + "-back-" + backType + "-" + id;
    toggleClassForId("hidden", backId);

    const inner = document.getElementById(prefix + "-inner-" + id);
    if (!inner) return;
    inner.style.transform = "rotateX(180deg)";
  }

  /**
   * Flip a card back to show its front face.
   * @param {string} prefix - Card prefix ("task-card" or "project-card")
   * @param {string} id - The card's numeric ID
   * @param {string[]} backTypes - All back face types to hide after flip
   */
  function flipToFront(prefix, id, backTypes) {
    const inner = document.getElementById(prefix + "-inner-" + id);
    if (!inner) return;

    inner.style.transform = "rotateX(0deg)";
    inner.addEventListener(
      "transitionend",
      function hideBack() {
        backTypes.forEach(function (type) {
          addClassForId("hidden", prefix + "-back-" + type + "-" + id);
        });
        inner.removeEventListener("transitionend", hideBack);
      },
      { once: true },
    );
  }

  // ============================================================
  // Task Card Flip Handlers
  // ============================================================

  const TASK_PREFIX = "task-card";
  const TASK_BACK_TYPES = ["update", "delete", "time-start", "time-stop"];

  /**
   * Determine which back face to show based on the HTMX request.
   * @param {string} method - HTTP method (POST, DELETE, etc.)
   * @param {string|null} action - The data-action attribute value
   * @returns {string|null} The back type to show, or null to skip
   */
  function getTaskBackType(method, action) {
    if (method === "DELETE") return "delete";
    if (action === "time-start") return "time-start";
    if (action === "time-stop") return "time-stop";
    if (method === "POST") return "update";
    return null;
  }

  /**
   * Handle flipping task card to back before HTMX request.
   * Called from hx-on--before-request on the card wrapper.
   * @param {CustomEvent} event - HTMX before-request event
   * @param {number|string} id - The task ID
   */
  function flipTaskCardToBack(event, id) {
    const method = event.detail.requestConfig.verb.toUpperCase();
    if (method === "GET") return;

    const triggerEl = event.detail.requestConfig.elt;
    const action = triggerEl ? triggerEl.getAttribute("data-action") : null;
    const backType = getTaskBackType(method, action);

    if (!backType) return;

    // If starting a new task, also flip any currently active task
    if (action === "time-start") {
      flipActiveTaskToBack(id);
    }

    flipToBack(TASK_PREFIX, id, backType);
  }

  /**
   * Find and flip any currently active (in-progress) task to back.
   * This is called when starting a new task to animate the old one stopping.
   * @param {number|string} excludeId - Task ID to exclude (the one being started)
   */
  function flipActiveTaskToBack(excludeId) {
    const activeStopBtn = document.querySelector(
      'button[data-action="time-stop"]',
    );
    if (!activeStopBtn) return;

    // Extract the task ID from any parent element's ID (e.g., task-card-front-3 → 3)
    const activeCard = activeStopBtn.closest('[id^="task-card-"]');
    if (!activeCard) return;

    // The ID is always the last number in the element ID
    const match = activeCard.id.match(/(\d+)$/);
    if (!match) return;

    const activeId = match[1];
    if (activeId === String(excludeId)) return;

    flipToBack(TASK_PREFIX, activeId, "time-stop");
  }

  /**
   * Handle flipping task card to front after HTMX swap.
   * Called from hx-on--after-swap on the card wrapper.
   * @param {CustomEvent} event - HTMX after-swap event
   * @param {number|string} id - The task ID
   */
  function flipTaskCardToFront(event, id) {
    const method = event.detail.requestConfig.verb.toUpperCase();
    if (method === "GET") return;

    flipToFront(TASK_PREFIX, id, TASK_BACK_TYPES);
  }

  // ============================================================
  // Project Card Flip Handlers
  // ============================================================

  const PROJECT_PREFIX = "project-card";
  const PROJECT_BACK_TYPES = ["delete"];

  /**
   * Handle flipping project card to back before HTMX request.
   * Called from hx-on--before-request on the card wrapper.
   * @param {CustomEvent} event - HTMX before-request event
   * @param {number|string} id - The project ID
   */
  function flipProjectCardToBack(event, id) {
    const method = event.detail.requestConfig.verb.toUpperCase();
    if (method === "GET") return;

    if (method === "DELETE") {
      flipToBack(PROJECT_PREFIX, id, "delete");
    }
    // POST not implemented for projects yet
  }

  /**
   * Handle flipping project card to front after HTMX swap.
   * Called from hx-on--after-swap on the card wrapper.
   * @param {CustomEvent} event - HTMX after-swap event
   * @param {number|string} id - The project ID
   */
  function flipProjectCardToFront(event, id) {
    const method = event.detail.requestConfig.verb.toUpperCase();
    if (method === "GET") return;

    flipToFront(PROJECT_PREFIX, id, PROJECT_BACK_TYPES);
  }

  // ============================================================
  // Public API
  // ============================================================

  return {
    // Task cards
    flipTaskCardToBack: flipTaskCardToBack,
    flipTaskCardToFront: flipTaskCardToFront,
    // Project cards
    flipProjectCardToBack: flipProjectCardToBack,
    flipProjectCardToFront: flipProjectCardToFront,
  };
})();

// Global aliases for use in hx-on attributes
function flipTaskCardUpdateToBack(event, id) {
  CardFlip.flipTaskCardToBack(event, id);
}
function flipTaskCardUpdateToFront(event, id) {
  CardFlip.flipTaskCardToFront(event, id);
}
function flipProjectCardUpdateToBack(event, id) {
  CardFlip.flipProjectCardToBack(event, id);
}
function flipProjectCardUpdateToFront(event, id) {
  CardFlip.flipProjectCardToFront(event, id);
}
