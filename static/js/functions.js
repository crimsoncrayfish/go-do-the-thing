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
  document.addEventListener("htmx:afterSwap", function (_) {
    if (typeof initFlowbite === "function") {
      initFlowbite();
    } else if (
      typeof Flowbite !== "undefined" &&
      typeof Flowbite.init === "function"
    ) {
      Flowbite.init();
    }
  });
});



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
