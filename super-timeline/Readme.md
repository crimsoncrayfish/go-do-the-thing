# Timeline Component

A standalone, interactive timeline component for displaying and editing daily tasks.

## Features

-   **Visual Timeline:** Displays tasks over a 24-hour period.
-   **Interactive Editing:** Supports resizing tasks and creating new tasks in gaps.
-   **Multiple Modes:**
    -   `view`: Read-only mode.
    -   `edit`: Allows all interactions.
    -   `scroll`: Allows panning and zooming.
-   **HTMX Integration:** Designed to work seamlessly with an HTMX-powered backend. All data mutations (create, update) are sent to server endpoints.

## Setup

1.  **Include Dependencies:** Add Tailwind CSS, HTMX, and the `timeline.js` script to your HTML file.

    ```html
    <head>
      <script src="https://cdn.tailwindcss.com"></script>
      <script src="https://unpkg.com/htmx.org@1.9.10"></script>
      <link rel="stylesheet" href="style.css" />
    </head>
    <body>
      <!-- Your component container -->
      <script src="timeline.js"></script>
    </body>
    ```

2.  **HTML Container:** Create a `div` element to host the timeline. Configuration is passed via `data-` attributes.

    ```html
    <div
      id="daily-timeline"
      hx-swap="outerHTML"
      data-tasks='[{"id": 1, "name": "Initial Planning", "project": "Project A", "start": "09:00", "end": "10:30"}]'
      data-update-url="/tasks/update"
      data-create-url="/tasks/create-form"
    ></div>
    ```

3.  **Initialization:** Initialize the component in a script tag.

    ```html
    <script>
      document.addEventListener('DOMContentLoaded', () => {
        const timelineEl = document.getElementById('daily-timeline');
        new Timeline(timelineEl).init();
      });
    </script>
    ```

## Configuration Options

Options are passed via `data-*` attributes on the container element.

-   `data-tasks` (required): A JSON string array of task objects. Each object needs `id`, `name`, `project`, `start` (HH:mm), and `end` (HH:mm).
-   `data-update-url` (optional): The URL to send `PUT` requests to when a task is resized. The request body will be JSON: `{ id, start, end }`.
-   `data-create-url` (optional): The URL to fetch an HTMX form/modal from when a gap is clicked. The request will include `?start=HH:mm&end=HH:mm` as query parameters.
-   `data-mode` (optional): The initial mode. Can be `view`, `edit`, or `scroll`. Defaults to `view`.
