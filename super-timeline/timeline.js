class Timeline {
  constructor(containerElement) {
    if (!containerElement) {
      throw new Error("Container element not found");
    }
    this.wrapper = containerElement;
    this.wrapper.className =
      "timeline-wrapper relative bg-gray-200 rounded-lg p-1";

    this.container = document.createElement("div");
    this.stateIcon = document.createElement("div");
    this.contextBar = document.createElement("div");
    this.wrapper.appendChild(this.container);
    this.wrapper.appendChild(this.stateIcon);
    this.wrapper.appendChild(this.contextBar);

    this.container.setAttribute("tabindex", "0");

    const dataset = containerElement.dataset;
    this.options = {
      tasks: JSON.parse(dataset.tasks || "[]"),
      updateUrl: dataset.updateUrl,
      createUrl: dataset.createUrl,
      mode: dataset.mode || "view",
    };

    this.state = {
      isResizing: false,
      resizeTarget: null,
      resizeHandle: null,
      startX: 0,
      startLeft: 0,
      startWidth: 0,

      resizeInitialStartMinutes: 0,
      resizeInitialEndMinutes: 0,
      view: { startMinutes: 0, endMinutes: 24 * 60 },
      isPanning: false,
      panStartX: 0,
      panInitialViewStart: 0,
      animationFrameId: null,
      panViewDuration: 0,

      isDraggingThumb: false,
      thumbStartX: 0,
      thumbInitialViewStart: 0,
      isZoomedOut: false,
    };

    this.sortedTasks = [];

    this._handleMouseDown = this._handleMouseDown.bind(this);
    this._handleMouseMove = this._handleMouseMove.bind(this);
    this._handleMouseUp = this._handleMouseUp.bind(this);
    this._handleWheel = this._handleWheel.bind(this);

    this._handlePanMove = this._handlePanMove.bind(this);
    this._handlePanEnd = this._handlePanEnd.bind(this);
    this._updatePanFrame = this._updatePanFrame.bind(this);

    this._handleContextClick = this._handleContextClick.bind(this);
    this._jumpToPosition = this._jumpToPosition.bind(this);
    this._startThumbDrag = this._startThumbDrag.bind(this);

    this._handleThumbDragMove = this._handleThumbDragMove.bind(this);
    this._handleThumbDragEnd = this._handleThumbDragEnd.bind(this);

    this.MINUTES_IN_DAY = 24 * 60;
  }

  init() {
    this.render();
    this._attachEventListeners();
  }

  setMode(newMode) {
    if (newMode === "scroll") {
      this.container.style.cursor = "grab";
    } else {
      this.container.style.cursor = "default";
    }
    if (!["view", "edit", "scroll"].includes(newMode)) return;
    this.options.mode = newMode;
    this.container.dataset.mode = newMode;
    this.render();
    this._attachEventListeners();
  }

  render() {
    this.container.innerHTML = "";
    this.contextBar.innerHTML = "";

    this.container.className = "timeline-container relative select-none";
    this.container.dataset.mode = this.options.mode;

    this._renderTimeMarkers();

    this.sortedTasks = [...this.options.tasks].sort(
      (a, b) => this._timeToMinutes(a.start) - this._timeToMinutes(b.start),
    );

    let lastEndMinute = 0;

    this.sortedTasks.forEach((task) => {
      const startMinute = this._timeToMinutes(task.start);
      const endMinute = this._timeToMinutes(task.end);

      if (startMinute > lastEndMinute) {
        this._renderGap(lastEndMinute, startMinute);
      }

      this._renderTask(task, startMinute, endMinute);
      lastEndMinute = endMinute;
    });

    if (lastEndMinute < this.MINUTES_IN_DAY) {
      this._renderGap(lastEndMinute, this.MINUTES_IN_DAY);
    }

    this._renderContextBar();
    this._renderStateIcon();
    if (this.state.isZoomedOut) {
      this.container.className += " h-12";
    } else {
      this.container.className += " h-10";
    }
  }
  _renderStateIcon() {
    this.stateIcon.className = "timeline-state-icon";

    const mode = this?.options?.mode;
    let svg = "";

    switch (mode) {
      case "edit":
        svg = `<svg xmlns="http://www.w3.org/2000/svg" height="16px" viewBox="0 -960 960 960" width="24px" fill="currentColor"><path d="M200-200h57l391-391-57-57-391 391v57Zm-80 80v-170l528-527q12-11 26.5-17t30.5-6q16 0 31 6t26 18l55 56q12 11 17.5 26t5.5 30q0 16-5.5 30.5T817-647L290-120H120Zm640-584-56-56 56 56Zm-141 85-28-29 57 57-29-28Z"/></svg>`;
        break;

      case "scroll":
        svg = `<svg xmlns="http://www.w3.org/2000/svg" height="16px" viewBox="0 -960 960 960" width="24px" fill="currentColor"><path d="M784-120 532-372q-30 24-69 38t-83 14q-109 0-184.5-75.5T120-580q0-109 75.5-184.5T380-840q109 0 184.5 75.5T640-580q0 44-14 83t-38 69l252 252-56 56ZM380-400q75 0 127.5-52.5T560-580q0-75-52.5-127.5T380-760q-75 0-127.5 52.5T200-580q0 75 52.5 127.5T380-400Zm-40-60v-80h-80v-80h80v-80h80v80h80v80h-80v80h-80Z"/></svg>`;
        break;

      case "view":
        svg = `<svg xmlns="http://www.w3.org/2000/svg" height="16px" viewBox="0 -960 960 960" width="24px" fill="currentColor"><path d="M480-320q75 0 127.5-52.5T660-500q0-75-52.5-127.5T480-680q-75 0-127.5 52.5T300-500q0 75 52.5 127.5T480-320Zm0-72q-45 0-76.5-31.5T372-500q0-45 31.5-76.5T480-608q45 0 76.5 31.5T588-500q0 45-31.5 76.5T480-392Zm0 192q-146 0-266-81.5T40-500q54-137 174-218.5T480-800q146 0 266 81.5T920-500q-54 137-174 218.5T480-200Zm0-300Zm0 220q113 0 207.5-59.5T832-500q-50-101-144.5-160.5T480-720q-113 0-207.5 59.5T128-500q50 101 144.5 160.5T480-280Z"/></svg>`;
        break;

      default:
        svg = "";
    }

    this.stateIcon.innerHTML = svg;
  }

  _renderContextBar() {
    this.contextBar.className =
      "timeline-context-bar relative bottom-0 left-0 w-full h-2 bg-gray-300 rounded-full overflow-hidden";

    this.state.isZoomedOut =
      this.state.view.endMinutes - this.state.view.startMinutes >=
      this.MINUTES_IN_DAY - 1;

    if (this.state.isZoomedOut) {
      this.contextBar.style.opacity = "0";
      this.contextBar.style.maxHeight = "0px";
      this.contextBar.style.marginTop = "0px";
      return;
    } else {
      this.contextBar.style.opacity = "1";
      this.contextBar.style.maxHeight = "0.5rem";
      this.contextBar.style.marginTop = "0.25rem";
    }

    this.sortedTasks.forEach((task) => {
      const taskEl = document.createElement("div");
      taskEl.className = "absolute h-full bg-blue-400";
      const startPercent =
        (this._timeToMinutes(task.start) / this.MINUTES_IN_DAY) * 100;
      const endPercent =
        (this._timeToMinutes(task.end) / this.MINUTES_IN_DAY) * 100;
      taskEl.style.left = `${startPercent}%`;
      taskEl.style.width = `${endPercent - startPercent}%`;
      this.contextBar.appendChild(taskEl);
    });

    const thumbEl = document.createElement("div");
    thumbEl.className =
      "absolute h-full bg-black/20 rounded-full border border-gray-700/50 cursor-grab";
    thumbEl.dataset.thumb = "true";

    const viewStartPercent =
      (this.state.view.startMinutes / this.MINUTES_IN_DAY) * 100;
    const viewEndPercent =
      (this.state.view.endMinutes / this.MINUTES_IN_DAY) * 100;

    thumbEl.style.left = `${viewStartPercent}%`;
    thumbEl.style.width = `${viewEndPercent - viewStartPercent}%`;

    this.contextBar.appendChild(thumbEl);
  }

  _renderTimeMarkers() {
    const viewDuration =
      this.state.view.endMinutes - this.state.view.startMinutes;
    let increment;

    if (viewDuration > 12 * 60) {
      increment = 180;
    } else if (viewDuration > 4 * 60) {
      increment = 60;
    } else if (viewDuration > 60) {
      increment = 30;
    } else {
      increment = 15;
    }

    const firstMarker =
      Math.ceil(this.state.view.startMinutes / increment) * increment;

    for (let m = firstMarker; m < this.state.view.endMinutes; m += increment) {
      const marker = document.createElement("div");
      marker.className = "absolute h-full w-px bg-gray-300 time-marker";
      marker.style.left = `${this._minutesToPercent(m)}%`;
      marker.dataset.time = this._minutesToTime(m);
      this.container.appendChild(marker);
    }
  }

  _renderTask(task, startMinute, endMinute) {
    if (
      endMinute < this.state.view.startMinutes ||
      startMinute > this.state.view.endMinutes
    ) {
      return;
    }

    const left = Math.max(
      0,
      Math.min(this._minutesToPercent(startMinute), 100),
    );
    const right = Math.max(0, Math.min(this._minutesToPercent(endMinute), 100));
    const width = right - left;
    const wrapperEl = document.createElement("div");
    wrapperEl.className = "group task-element absolute h-full";
    wrapperEl.dataset.taskId = task.id;

    wrapperEl.style.left = `${left}%`;
    wrapperEl.style.width = `${width}%`;

    const taskEl = document.createElement("div");
    taskEl.className =
      "relative h-full w-full bg-blue-500 border border-blue-700 rounded-md flex items-center justify-center text-white text-xs font-medium";

    const tooltipEl = document.createElement("div");
    tooltipEl.textContent = `${task.project}: ${task.name}`;
    tooltipEl.className =
      "absolute bottom-full left-1/2 mb-2 -translate-x-1/2 whitespace-nowrap rounded-md bg-gray-800 px-2 py-1 text-xs text-white opacity-0 transition-opacity group-hover:opacity-100 pointer-events-none";

    const nameEl = document.createElement("span");
    nameEl.className = "transition-opacity group-hover:opacity-50";
    taskEl.appendChild(nameEl);

    if (this.options.mode === "edit") {
      const leftHandle = document.createElement("div");
      leftHandle.className =
        "resize-handle absolute left-0 top-0 h-full w-4 flex items-center justify-start cursor-ew-resize opacity-0 group-hover:opacity-100 transition-opacity duration-200";
      leftHandle.dataset.handle = "left";
      const leftVisual = document.createElement("div");
      leftVisual.className = "h-1/2 w-1 bg-white/70 rounded-full";
      leftHandle.appendChild(leftVisual);
      taskEl.appendChild(leftHandle);

      const rightHandle = document.createElement("div");
      rightHandle.className =
        "resize-handle absolute right-0 top-0 h-full w-4 flex items-center justify-end cursor-ew-resize opacity-0 group-hover:opacity-100 transition-opacity duration-200";
      rightHandle.dataset.handle = "right";
      const rightVisual = document.createElement("div");
      rightVisual.className = "h-1/2 w-1 bg-white/70 rounded-full";
      rightHandle.appendChild(rightVisual);
      taskEl.appendChild(rightHandle);
    }

    wrapperEl.appendChild(tooltipEl);
    wrapperEl.appendChild(taskEl);
    this.container.appendChild(wrapperEl);
  }

  _renderGap(startMinute, endMinute) {
    const view = this.state.view;
    if (endMinute < view.startMinutes || startMinute > view.endMinutes) {
      return;
    }

    const left = this._minutesToPercent(startMinute);
    const right = this._minutesToPercent(endMinute);
    const width = right - left;

    const gapEl = document.createElement("div");
    gapEl.className = "absolute h-full group";
    gapEl.style.left = `${left}%`;
    gapEl.style.width = `${width}%`;

    if (this.options.mode === "edit") {
      gapEl.className +=
        " cursor-pointer hover:bg-gray-500/10 flex items-center justify-center text-gray-400 show-on-hover";
      gapEl.style.containerType = "inline-size";

      const svg = `
  <svg xmlns="http://www.w3.org/2000/svg"
       viewBox="0 -960 960 960"
       width="24" height="24"
       fill="currentColor" aria-hidden="true">
    <path d="M440-280h80v-160h160v-80H520v-160h-80v160H280v80h160v160Zm40 200q-83 0-156-31.5T197-197q-54-54-85.5-127T80-480q0-83 31.5-156T197-763q54-54 127-85.5T480-880q83 0 156 31.5T763-763q54 54 85.5 127T880-480q0 83-31.5 156T763-197q-54 54-127 85.5T480-80Zm0-80q134 0 227-93t93-227q0-134-93-227t-227-93q-134 0-227 93t-93 227q0 134 93 227t227 93Zm0-320Z"/>
  </svg>
`;
      gapEl.innerHTML = svg;
      gapEl.addEventListener("click", () => {
        const start = this._minutesToTime(
          Math.max(startMinute, view.startMinutes),
        );
        const end = this._minutesToTime(Math.min(endMinute, view.endMinutes));
        const message = `HTMX GET ${this.options.createUrl}?start=${start}&end=${end}`;
        this._showMockNotification(message);
      });
    }

    this.container.appendChild(gapEl);
  }

  _attachEventListeners() {
    this.container.removeEventListener("mousedown", this._handleMouseDown);
    this.container.removeEventListener("wheel", this._handleWheel);
    this.contextBar.removeEventListener("mousedown", this._handleContextClick);

    this.container.addEventListener("mousedown", this._handleMouseDown);
    this.container.addEventListener("wheel", this._handleWheel);
    this.contextBar.addEventListener("mousedown", this._handleContextClick);
  }

  _handleContextClick(e) {
    e.preventDefault();
    if (!e.target.dataset.thumb) {
      this._jumpToPosition(e);
    }
    this._startThumbDrag(e);
  }

  _jumpToPosition(e) {
    const rect = this.contextBar.getBoundingClientRect();
    const clickPercent = (e.clientX - rect.left) / rect.width;
    const targetCenterMinutes = this.MINUTES_IN_DAY * clickPercent;
    const viewDuration =
      this.state.view.endMinutes - this.state.view.startMinutes;
    let newStart = targetCenterMinutes - viewDuration / 2;
    if (newStart < 0) newStart = 0;
    if (newStart + viewDuration > this.MINUTES_IN_DAY)
      newStart = this.MINUTES_IN_DAY - viewDuration;
    this.state.view.startMinutes = newStart;
    this.state.view.endMinutes = newStart + viewDuration;
    this.render();
  }

  _startThumbDrag(e) {
    this.state.isDraggingThumb = true;
    this.state.thumbStartX = e.clientX;
    this.state.thumbInitialViewStart = this.state.view.startMinutes;

    const thumb = this.contextBar.querySelector("[data-thumb]");
    if (thumb) thumb.style.cursor = "grabbing";

    document.addEventListener("mousemove", this._handleThumbDragMove);
    document.addEventListener("mouseup", this._handleThumbDragEnd, {
      once: true,
    });
  }

  _handleThumbDragMove(e) {
    if (!this.state.isDraggingThumb) return;
    e.preventDefault();

    const { thumbStartX, thumbInitialViewStart, view } = this.state;
    const { offsetWidth } = this.contextBar;

    const duration = view.endMinutes - view.startMinutes;

    const deltaX = e.clientX - thumbStartX;
    const minutesPerPixel = this.MINUTES_IN_DAY / offsetWidth;
    const deltaMinutes = deltaX * minutesPerPixel;

    let newStart = thumbInitialViewStart + deltaMinutes;

    const minStart = 0;
    const maxStart = this.MINUTES_IN_DAY - duration;
    if (newStart < minStart) newStart = minStart;
    if (newStart > maxStart) newStart = maxStart;

    this.state.view.startMinutes = newStart;
    this.state.view.endMinutes = newStart + duration;

    this.state.panViewDuration = duration;

    if (!this.state.animationFrameId) {
      this.state.animationFrameId = requestAnimationFrame(this._updatePanFrame);
    }
  }

  _handleThumbDragEnd(_) {
    this.state.isDraggingThumb = false;

    const thumb = this.contextBar.querySelector("[data-thumb]");
    if (thumb) thumb.style.cursor = "grab";

    document.removeEventListener("mousemove", this._handleThumbDragMove);
    if (this.state.animationFrameId) {
      cancelAnimationFrame(this.state.animationFrameId);
      this.state.animationFrameId = null;
    }
  }

  _handleMouseDown(e) {
    if (this.options.mode === "edit") {
      if (e.button === 0) {
        const handleElement = e.target.closest("[data-handle]");
        if (handleElement) {
          this._startResize(e, handleElement);
          return;
        }
      }
    }

    if (this.options.mode === "scroll") {
      e.preventDefault();
      this.state.isPanning = true;
      this.state.panStartX = e.clientX;
      this.state.panInitialViewStart = this.state.view.startMinutes;

      this.state.panViewDuration =
        this.state.view.endMinutes - this.state.view.startMinutes;

      this.container.style.cursor = "grabbing";
      this.container.classList.add("is-panning");

      document.addEventListener("mousemove", this._handlePanMove);
      document.addEventListener("mouseup", this._handlePanEnd, { once: true });
    }
  }

  _startResize(e, handleElement) {
    e.preventDefault();
    this.state.isResizing = true;
    this.state.resizeHandle = handleElement.dataset.handle;
    this.state.resizeTarget = e.target.closest("[data-task-id]");

    const taskId = parseInt(this.state.resizeTarget.dataset.taskId, 10);
    const taskData = this.sortedTasks.find((t) => t.id === taskId);

    this.state.resizeInitialStartMinutes = this._timeToMinutes(taskData.start);
    this.state.resizeInitialEndMinutes = this._timeToMinutes(taskData.end);

    this.container.classList.add("is-resizing");
    this.container.style.cursor = "ew-resize";
    this.state.resizeTarget.classList.add("is-active");

    document.addEventListener("mousemove", this._handleMouseMove);
    document.addEventListener("mouseup", this._handleMouseUp, { once: true });
  }

  _startPan(e) {
    e.preventDefault();
    this.state.isPanning = true;
    this.state.panStartX = e.clientX;
    this.state.panInitialViewStart = this.state.view.startMinutes;
    this.state.panViewDuration =
      this.state.view.endMinutes - this.state.view.startMinutes;

    this.container.style.cursor = "grabbing";
    this.container.classList.add("is-panning");
    document.addEventListener("mousemove", this._handlePanMove);
    document.addEventListener("mouseup", this._handlePanEnd, { once: true });
  }

  _handlePanMove(e) {
    if (!this.state.isPanning) return;
    e.preventDefault();

    console.clear();

    const deltaX = e.clientX - this.state.panStartX;

    const viewDuration = this.state.panViewDuration;
    const minutesPerPixel = viewDuration / this.container.offsetWidth;
    const deltaMinutes = deltaX * minutesPerPixel;

    const newStart = this.state.panInitialViewStart - deltaMinutes;

    console.log(`LOG: Pan drag delta: ${deltaX.toFixed(2)}px`);
    console.log(
      `LOG: Minutes per pixel (constant): ${minutesPerPixel.toFixed(4)}`,
    );
    console.log(
      `LOG: Resulting time shift: ${deltaMinutes.toFixed(2)} minutes`,
    );

    this.state.view.startMinutes = newStart;

    if (!this.state.animationFrameId) {
      this.state.animationFrameId = requestAnimationFrame(this._updatePanFrame);
    }
  }

  _updatePanFrame() {
    let newStart = this.state.view.startMinutes;
    const viewDuration = this.state.panViewDuration;

    if (newStart < 0) newStart = 0;
    if (newStart + viewDuration > this.MINUTES_IN_DAY) {
      newStart = this.MINUTES_IN_DAY - viewDuration;
    }

    this.state.view.startMinutes = newStart;
    this.state.view.endMinutes = newStart + viewDuration;

    this.render();

    this.state.animationFrameId = null;
  }

  _handlePanEnd(_) {
    if (!this.state.isPanning) return;

    this.state.isPanning = false;
    this.container.style.cursor = "grab";
    this.container.classList.remove("is-panning");
    document.removeEventListener("mousemove", this._handlePanMove);

    if (this.state.animationFrameId) {
      cancelAnimationFrame(this.state.animationFrameId);
      this.state.animationFrameId = null;
    }
  }

  _handleWheel(e) {
    if (this.options.mode !== "scroll") return;
    e.preventDefault();

    const rect = this.container.getBoundingClientRect();
    const mouseX = e.clientX - rect.left;
    const mousePercent = mouseX / rect.width;

    const viewDuration =
      this.state.view.endMinutes - this.state.view.startMinutes;
    const mouseAtMinutes =
      this.state.view.startMinutes + viewDuration * mousePercent;

    const zoomIntensity = 0.1;
    const zoomDirection = e.deltaY < 0 ? 1 - zoomIntensity : 1 + zoomIntensity;

    const newDuration = viewDuration * zoomDirection;

    const clampedDuration = Math.max(
      15,
      Math.min(this.MINUTES_IN_DAY, newDuration),
    );

    let newStart = mouseAtMinutes - mousePercent * clampedDuration;
    let newEnd = newStart + clampedDuration;

    if (newStart < 0) {
      newStart = 0;
      newEnd = clampedDuration;
    }
    if (newEnd > this.MINUTES_IN_DAY) {
      newEnd = this.MINUTES_IN_DAY;
      newStart = newEnd - clampedDuration;
    }

    this.state.view.startMinutes = newStart;
    this.state.view.endMinutes = newEnd;

    this.render();
  }
  _handleMouseMove(e) {
    if (!this.state.isResizing) return;

    console.clear();

    const currentTaskId = parseInt(this.state.resizeTarget.dataset.taskId, 10);
    const currentIndex = this.sortedTasks.findIndex(
      (t) => t.id === currentTaskId,
    );
    const prevTask = this.sortedTasks[currentIndex - 1];
    const nextTask = this.sortedTasks[currentIndex + 1];
    const minBoundaryMinutes = prevTask ? this._timeToMinutes(prevTask.end) : 0;
    const maxBoundaryMinutes = nextTask
      ? this._timeToMinutes(nextTask.start)
      : this.MINUTES_IN_DAY;

    const rect = this.container.getBoundingClientRect();
    const mouseX = e.clientX - rect.left;
    const mouseMinute = this._pixelsToMinutes(mouseX);
    console.log(
      `LOG: Mouse is at ${mouseX.toFixed(2)}px which is minute ${mouseMinute.toFixed(2)}`,
    );

    let newStartMinutes, newEndMinutes;

    if (this.state.resizeHandle === "right") {
      newStartMinutes = this.state.resizeInitialStartMinutes;
      newEndMinutes = mouseMinute;
    } else {
      newStartMinutes = mouseMinute;
      newEndMinutes = this.state.resizeInitialEndMinutes;
    }
    console.log(
      `LOG: Raw new times -> start: ${newStartMinutes.toFixed(2)}, end: ${newEndMinutes.toFixed(2)}`,
    );

    newStartMinutes = Math.max(newStartMinutes, minBoundaryMinutes);
    newEndMinutes = Math.min(newEndMinutes, maxBoundaryMinutes);
    console.log(
      `LOG: Clamped to boundaries -> start: ${newStartMinutes.toFixed(2)}, end: ${newEndMinutes.toFixed(2)}`,
    );

    if (newEndMinutes - newStartMinutes < 5) {
      if (this.state.resizeHandle === "right") {
        newEndMinutes = newStartMinutes + 5;
      } else {
        newStartMinutes = newEndMinutes - 5;
      }
      console.log(
        `LOG: Adjusted for min duration -> start: ${newStartMinutes.toFixed(2)}, end: ${newEndMinutes.toFixed(2)}`,
      );
    }

    const newLeftPercent = this._minutesToPercent(newStartMinutes);
    const newRightPercent = this._minutesToPercent(newEndMinutes);
    const newWidthPercent = newRightPercent - newLeftPercent;
    console.log(
      `LOG: Final style -> left: ${newLeftPercent.toFixed(2)}%, width: ${newWidthPercent.toFixed(2)}%`,
    );

    this.state.resizeTarget.style.left = `${newLeftPercent}%`;
    this.state.resizeTarget.style.width = `${newWidthPercent}%`;
  }
  _handleMouseUp() {
    if (!this.state.isResizing) return;

    this.container.classList.remove("is-resizing");
    this.state.resizeTarget.classList.remove("is-active");

    document.removeEventListener("mousemove", this._handleMouseMove);

    const finalRect = this.state.resizeTarget.getBoundingClientRect();
    const finalLeftPx =
      finalRect.left - this.container.getBoundingClientRect().left;

    const startMinutes = this._pixelsToMinutes(finalLeftPx);
    const endMinutes = this._pixelsToMinutes(finalLeftPx + finalRect.width);

    const snappedStart = Math.round(startMinutes / 5) * 5;
    const snappedEnd = Math.round(endMinutes / 5) * 5;

    const newStart = this._minutesToTime(snappedStart);
    const newEnd = this._minutesToTime(snappedEnd);
    const taskId = parseInt(this.state.resizeTarget.dataset.taskId, 10);

    this.state.isResizing = false;

    this._updateTask(taskId, newStart, newEnd);
  }

  _minutesToPercent(minutes) {
    const viewDuration =
      this.state.view.endMinutes - this.state.view.startMinutes;
    if (viewDuration <= 0) return 0;

    const relativeMinutes = minutes - this.state.view.startMinutes;
    return (relativeMinutes / viewDuration) * 100;
  }

  _pixelsToMinutes(pixels) {
    const viewDuration =
      this.state.view.endMinutes - this.state.view.startMinutes;
    const percent = pixels / this.container.offsetWidth;
    return this.state.view.startMinutes + percent * viewDuration;
  }
  async _updateTask(taskId, newStart, newEnd) {
    const task = this.options.tasks.find((t) => t.id === taskId);
    if (task) {
      task.start = newStart;
      task.end = newEnd;
    }

    this.render();
    this._attachEventListeners();

    if (!this.options.updateUrl) return;

    //TODO: Whack the mock calls
    const payload = { id: taskId, start: newStart, end: newEnd };
    const message = `FETCH PUT ${
      this.options.updateUrl
    } with body: ${JSON.stringify(payload)}`;
    this._showMockNotification(message);

    //This is the real updates
    console.log(`Updating task ${taskId} to ${newStart} - ${newEnd}`);
    try {
      const response = await fetch(this.options.updateUrl, {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ id: taskId, start: newStart, end: newEnd }),
      });

      if (!response.ok) {
        throw new Error(`Server responded with status: ${response.status}`);
      }
      console.log("Task updated successfully.");
    } catch (error) {
      console.error("Failed to update task:", error);
    }
  }

  //TODO: Whack this method
  _showMockNotification(message) {
    const notification = document.createElement("div");
    notification.className =
      "fixed bottom-5 right-5 bg-gray-800 text-white text-sm py-2 px-4 rounded-lg shadow-lg transition-all duration-300 translate-y-16 opacity-0";
    notification.textContent = message;

    document.body.appendChild(notification);

    setTimeout(() => {
      notification.classList.remove("translate-y-16", "opacity-0");
    }, 10);

    setTimeout(() => {
      notification.classList.add("translate-y-16", "opacity-0");
      notification.addEventListener("transitionend", () =>
        notification.remove(),
      );
    }, 4000);
  }

  _timeToMinutes(timeStr) {
    const [hours, minutes] = timeStr.split(":").map(Number);
    return hours * 60 + minutes;
  }

  _minutesToTime(totalMinutes) {
    const hours = Math.floor(totalMinutes / 60)
      .toString()
      .padStart(2, "0");
    const minutes = (totalMinutes % 60).toString().padStart(2, "0");
    return `${hours}:${minutes}`;
  }
}
