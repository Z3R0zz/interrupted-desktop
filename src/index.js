import Toastify from 'toastify-js';
import "toastify-js/src/toastify.css";
import './views/assets/styles.scss';

function showToast(text) {
  Toastify({
    text: text,
    duration: 5000,
    gravity: "top",
    position: "right",
    stopOnFocus: true,
    style: {
      background: "linear-gradient(to right, #d90f0f, #860909)",
    },
  }).showToast();
}

function attachEventListener(elementId, event, callback) {
  const element = document.getElementById(elementId);
  if (element) {
    element.addEventListener(event, callback);
  }
}

attachEventListener("logout", "click", () => {
  window.logOut();
});

attachEventListener("select-file", "click", () => {
  window.selectFile().then((response) => {
    showToast(response);
    updateGallery();
  });
});

attachEventListener("capture-screenshot", "click", () => {
  window.fetchMonitors().then((monitors) => {
    const monitorList = document.getElementById("monitor-list");
    monitorList.innerHTML = "";

    if (monitors.length === 0) {
      showToast("No monitors found!");
      return;
    }

    monitors.forEach((monitor, index) => {
      const monitorDiv = document.createElement("div");
      monitorDiv.innerHTML = `
        <p class="text-gray-300 plus-jakarta-sans">Monitor ${index}: ${monitor.width}x${monitor.height}</p>
        <button class="bg-red-500 hover:bg-red-400 text-white font-semibold py-2 px-4 rounded-md ring-2 ring-red-400 transition-all duration-300" onclick="screenshot(${index})">Capture Screenshot</button>
      `;
      monitorList.appendChild(monitorDiv);
    });
  });
});

window.screenshot = function(monitorIndex) {
  window.captureScreenshot(monitorIndex).then((response) => {
    showToast(response);
    updateGallery();
  });
};

document.addEventListener("DOMContentLoaded", () => {
  window.fetchGallery().then((response) => {
    if (response.length > 0) {
      const imagesContainer = document.querySelector(".images");
      imagesContainer.innerHTML = "";

      response.forEach((image) => {
        const img = document.createElement("img");
        img.src = image.url;
        img.setAttribute("data-url", image.url);

        img.addEventListener("click", (event) => {
          const clickedImageUrl = event.target.getAttribute("data-url");
          window.copyToClipboard(clickedImageUrl).then(showToast);
        });

        imagesContainer.appendChild(img);
      });
    }
  });
});

function updateGallery() {
  window.fetchGallery().then((response) => {
    const imagesContainer = document.querySelector(".images");
    imagesContainer.innerHTML = "";

    response.forEach((image) => {
      const img = document.createElement("img");
      img.src = image.url;
      img.setAttribute("data-url", image.url);

      img.addEventListener("click", (event) => {
        const clickedImageUrl = event.target.getAttribute("data-url");
        window.copyToClipboard(clickedImageUrl).then(showToast);
      });

      imagesContainer.appendChild(img);
    });
  });
}

attachEventListener("submit-paste", "click", () => {
  const pasteTitleElement = document.getElementById("paste-title");
  const pasteContentElement = document.getElementById("paste-content");

  const pasteTitle = pasteTitleElement.value;
  const pasteText = pasteContentElement.value;

  window.pasteText(pasteTitle, pasteText).then((response) => {
    pasteTitleElement.value = "";
    pasteContentElement.value = "";
    showToast(response);
  });
});

attachEventListener("shorten", "click", () => {
  const url = document.getElementById("shorten-url");
  const urlValue = url.value;

  window.shortenUrl(urlValue).then((response) => {
    url.value = "";
    showToast(response);
  });
});
