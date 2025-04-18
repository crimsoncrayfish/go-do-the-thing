htmx.config.useTemplateFragments = true;
htmx.config.allowNestedOobSwaps = true;
document.addEventListener("DOMContentLoaded", (_) => {
  document.body.addEventListener("htmx:beforeSwap", function (evt) {
    if (evt.detail.xhr.status === 422) {
      evt.detail.shouldSwap = true;
      evt.detail.isError = false;
    }
    if (evt.detail.xhr.status === 500) {
      evt.detail.shouldSwap = true;
      evt.detail.isError = false;
    }
  });
  document.body.addEventListener("htmx:configRequest", function (evt) {
    evt.detail.headers["accept"] = "text/html";
    evt.detail.headers["authorization"] = getAuthToken(); // add a new parameter into the mix
  });
  document.addEventListener("htmx:beforeRequest", function (_) {
    toggleLoader(false);
  });
  document.addEventListener("htmx:afterRequest", function (_) {
    toggleLoader(true);
  });
});

function getAuthToken() {
  return "bearer placeholder token";
}

function toggleClassForId(className, elementId) {
  const elem = document.getElementById(elementId);
  if (elem.classList.contains(className)) {
    elem.classList.remove(className);
    return;
  }
  elem.classList.add(className);
}
function toggleClassForIdExact(className, elementId, on) {
  const elem = document.getElementById(elementId);
  if (elem.classList.contains(className) && !on) {
    elem.classList.remove(className);
    return;
  }
  if (!elem.classList.contains(className) && on) {
    elem.classList.add(className);
  }
}

function swapClassForId(class1Name, class2Name, elementId) {
  const elem = document.getElementById(elementId);
  if (elem.classList.contains(class1Name)) {
    elem.classList.remove(class1Name);
    elem.classList.add(class2Name);
    return;
  }

  elem.classList.add(class1Name);
  elem.classList.remove(class2Name);
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
