htmx.config.useTemplateFragments = true;
htmx.config.allowNestedOobSwaps = true;
// NOTE: If debugging set: htmx.logAll();
document.addEventListener("DOMContentLoaded", (_) => {
  document.body.addEventListener("htmx:beforeSwap", function (evt) {
    if (evt.detail.xhr.status >= 400 && evt.detail.xhr.status < 600) {
      evt.detail.shouldSwap = true;
    }
  });
  document.body.addEventListener("htmx:configRequest", function (evt) {
    evt.detail.headers["accept"] = "text/html";
    // Authentication is handled via cookies, no need for authorization header
  });
  document.body.addEventListener("htmx:afterSwap", () => {
    autoInitSuperDatePickers();
  });
  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", autoInitSuperDatePickers);
  } else {
    autoInitSuperDatePickers();
  }
});

function addClassForId(className, elementId) {
  const elem = document.getElementById(elementId);
  if (!elem.classList.contains(className)) {
    elem.classList.add(className);
  }
}

function toggleClassForId(className, elementId) {
  const elem = document.getElementById(elementId);
  if (!elem) {
    console.error(
      `could not toggle class ${className} for element ${elementId}`,
    );
    return;
  }
  if (elem.classList.contains(className)) {
    elem.classList.remove(className);
    return;
  }
  elem.classList.add(className);
}

function swapClassesForId(class1List, class2List, elementId) {
  for (let i = 0; i < class1List.length; i++) {
    const elem = document.getElementById(elementId);
    if (elem.classList.contains(class1List[i])) {
      elem.classList.remove(class1List[i]);
      elem.classList.add(class2List[i]);
    } else {
      elem.classList.remove(class2List[i]);
      elem.classList.add(class1List[i]);
    }
  }
}

if ("dark-mode" in localStorage) {
  if (localStorage.getItem("dark-mode") === "true") {
    document.querySelector("html").classList.add("dark");
  } else {
    document.querySelector("html").classList.remove("dark");
  }
} else if (window.matchMedia("(prefers-color-scheme: dark)").matches) {
  document.querySelector("html").classList.add("dark");
} else {
  document.querySelector("html").classList.remove("dark");
}

if (!window.showAndHideToast) {
  let showToastTimer = null;
  let hideToastTimer = null;

  function hideToast(toastId) {
    const toast = document.getElementById(toastId);
    if (toast) {
      clearTimeout(showToastTimer);
      clearTimeout(hideToastTimer);

      toast.classList.remove("opacity-100");
      toast.classList.add("opacity-0");

      setTimeout(() => {
        toast.classList.remove("flex");
        toast.classList.add("hidden");
      }, 500);
    }
  }

  function showAndHideToast(toastId) {
    const toast = document.getElementById(toastId);
    if (toast) {
      clearTimeout(showToastTimer);
      clearTimeout(hideToastTimer);

      showToastTimer = setTimeout(() => {
        toast.classList.remove("hidden");
        toast.classList.add("flex");
        setTimeout(() => {
          toast.classList.remove("opacity-0");
          toast.classList.add("opacity-100");
        }, 10);
      }, 10);

      hideToastTimer = setTimeout(() => {
        hideToast(toastId);
      }, 5000);
    }
  }
}
