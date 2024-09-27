import Toastify from 'toastify-js'
import "toastify-js/src/toastify.css"

const logout = document.getElementById("logout");
logout.addEventListener("click", () => {
  window.logOut();
});

const selectFile = document.getElementById("select-file");
selectFile.addEventListener("click", () => {
  window.selectFile().then((response) => {
    Toastify({
        text: response,
        duration: 5000,
        gravity: "top",
        position: "right",
        stopOnFocus: true,
        style: {
          background: "linear-gradient(to right, #d90f0f, #860909)",
        },
    }).showToast();
    updateGallery();
  });
});

document
  .getElementById("capture-screenshot")
  .addEventListener("click", () => {
    window.fetchMonitors().then((monitors) => {
      const monitorList = document.getElementById("monitor-list");
      monitorList.innerHTML = "";

      if (monitors.length === 0) {
        Toastify({
            text: "No monitors found!",
            duration: 5000,
            gravity: "top",
            position: "right",
            stopOnFocus: true,
            style: {
              background: "linear-gradient(to right, #d90f0f, #860909)",
            },
        }).showToast();
        return;
      }

      monitors.forEach((monitor, index) => {
        const monitorDiv = document.createElement("div");
        monitorDiv.innerHTML = `
  <p>Monitor ${index}: ${monitor.width}x${monitor.height}</p>
  <button onclick="screenshot(${index})">Capture Screenshot</button>
`;
        monitorList.appendChild(monitorDiv);
      });
    });
  });

function screenshot(monitorIndex) {
  window.captureScreenshot(monitorIndex).then((response) => {
    Toastify({
        text: response,
        duration: 5000,
        gravity: "top",
        position: "right",
        stopOnFocus: true,
        style: {
          background: "linear-gradient(to right, #d90f0f, #860909)",
        },
    }).showToast();
    updateGallery();
  });
}

document.addEventListener("DOMContentLoaded", (event) => {
  window.fetchGallery().then((response) => {
    if (response.length > 0) {
      const imagesContainer = document.querySelector(".images");

      response.forEach((image) => {
        const img = document.createElement("img");
        img.src = image.url;

        img.setAttribute("data-url", image.url);

        img.addEventListener("click", (event) => {
          const clickedImageUrl = event.target.getAttribute("data-url");
          window.copyToClipboard(clickedImageUrl).then((response) => {
            Toastify({
                text: response,
                duration: 5000,
                gravity: "top",
                position: "right",
                stopOnFocus: true,
                style: {
                  background: "linear-gradient(to right, #d90f0f, #860909)",
                },
            }).showToast();
          });
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
        window.copyToClipboard(clickedImageUrl).then((response) => {
          Toastify({
            text: response,
            duration: 5000,
            gravity: "top",
            position: "right",
            stopOnFocus: true,
            style: {
              background: "linear-gradient(to right, #d90f0f, #860909)",
            },
        }).showToast();
        });
      });

      imagesContainer.appendChild(img);
    });
  });
}

const paste = document.getElementById("submit-paste");
paste.addEventListener("click", () => {
  const pasteTitleElement = document.getElementById("paste-title");
  const pasteContentElement = document.getElementById("paste-content");
  
  const pasteTitle = pasteTitleElement.value;
  const pasteText = pasteContentElement.value;
  
  window.pasteText(pasteTitle, pasteText).then((response) => {
    pasteTitleElement.value = "";
    pasteContentElement.value = "";
    Toastify({
        text: response,
        duration: 5000,
        gravity: "top",
        position: "right",
        stopOnFocus: true,
        style: {
          background: "linear-gradient(to right, #d90f0f, #860909)",
        },
    }).showToast();
  });
});