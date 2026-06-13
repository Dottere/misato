let currentIndex = 0;
let imgElement;
let pageInput;
let pageTotal;

function showToast(message) {
  let toast = document.getElementById("system-toast");
  if (!toast) {
    toast = document.createElement("div");
    toast.id = "system-toast";
    toast.className = "toast";
    document.body.appendChild(toast);
  }
  toast.innerText = message;
  toast.classList.add("show");
  if (toast.hideTimeout) clearTimeout(toast.hideTimeout);
  toast.hideTimeout = setTimeout(() => toast.classList.remove("show"), 3000);
}

function updatePage() {
  if (!window.CHAPTER_IMAGES || window.CHAPTER_IMAGES.length === 0) {
    if (pageTotal) pageTotal.innerText = "No pages found.";
    return;
  }

  const nextImageUrl = window.CHAPTER_IMAGES[currentIndex];

  if (pageInput && pageTotal) {
    pageInput.value = currentIndex + 1;
    pageInput.max = window.CHAPTER_IMAGES.length;
    pageTotal.innerText = ` / ${window.CHAPTER_IMAGES.length}`;
  }

  const tempImg = new Image();
  tempImg.onload = () => {
    if (imgElement) {
      imgElement.src = nextImageUrl;
    }
  };
  tempImg.src = nextImageUrl;

  preloadNextPages();
}

function changePage(direction) {
  let newIndex = currentIndex + direction;
  if (newIndex >= 0 && newIndex < window.CHAPTER_IMAGES.length) {
    currentIndex = newIndex;
    updatePage();
  } else if (newIndex >= window.CHAPTER_IMAGES.length) {
    showToast("You've reached the end of the chapter.");
  } else if (newIndex < 0) {
    showToast("You are on the first page.");
  }
}

function preloadNextPages() {
  if (currentIndex + 1 < window.CHAPTER_IMAGES.length) {
    new Image().src = window.CHAPTER_IMAGES[currentIndex + 1];
  }
  if (currentIndex + 2 < window.CHAPTER_IMAGES.length) {
    new Image().src = window.CHAPTER_IMAGES[currentIndex + 2];
  }
}

document.addEventListener("DOMContentLoaded", () => {
  imgElement = document.getElementById("manga-page");
  pageInput = document.getElementById("page-input");
  pageTotal = document.getElementById("page-total");

  document.addEventListener("keydown", function (event) {
    if (document.activeElement === pageInput) return;

    if (event.key === "ArrowRight") {
      changePage(1);
    } else if (event.key === "ArrowLeft") {
      changePage(-1);
    }
  });

  if (pageInput) {
    pageInput.addEventListener("change", (e) => {
      let requestedPage = parseInt(e.target.value, 10);

      if (isNaN(requestedPage)) {
        pageInput.value = currentIndex + 1;
        return;
      }

      let newIndex = requestedPage - 1;

      if (newIndex < 0) newIndex = 0;
      if (newIndex >= window.CHAPTER_IMAGES.length)
        newIndex = window.CHAPTER_IMAGES.length - 1;

      currentIndex = newIndex;
      updatePage();
      pageInput.blur();
    });
  }

  updatePage();
});
