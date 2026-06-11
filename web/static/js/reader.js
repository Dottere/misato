let currentIndex = 0;
const imgElement = document.getElementById("manga-page");
const counterElement = document.getElementById("page-counter");

function updatePage() {
  if (window.CHAPTER_IMAGES.length === 0) {
    if (counterElement) counterElement.innerText = "No pages found.";
    return;
  }

  if (imgElement) imgElement.src = window.CHAPTER_IMAGES[currentIndex];

  if (counterElement)
    counterElement.innerText = `Page ${currentIndex + 1} / ${window.CHAPTER_IMAGES.length}`;

  window.scrollTo(0, 0);

  preloadNextPages();
}

function changePage(direction) {
  let newIndex = currentIndex + direction;

  if (newIndex >= 0 && newIndex < window.CHAPTER_IMAGES.length) {
    currentIndex = newIndex;
    updatePage();
  } else if (newIndex >= window.CHAPTER_IMAGES.length) {
    alert("You've reached the end of the chapter!");
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

document.addEventListener("keydown", function (event) {
  if (event.key === "ArrowRight") {
    changePage(1);
  } else if (event.key === "ArrowLeft") {
    changePage(-1);
  }
});

updatePage();
