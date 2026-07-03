const overlay = document.querySelector("#overlay");
const cards = document.querySelector("#cards");
const scene = document.querySelector("#scene");
const cameraFeed = document.querySelector("#cameraFeed");
const captureCanvas = document.querySelector("#captureCanvas");
const statusEl = document.querySelector("#status");
const sceneHashEl = document.querySelector("#sceneHash");
const cacheStateEl = document.querySelector("#cacheState");
const cameraBtn = document.querySelector("#cameraBtn");
const recognizeBtn = document.querySelector("#recognizeBtn");
const cacheBtn = document.querySelector("#cacheBtn");

let lastSceneHash = "";
let cameraStream = null;

cameraBtn.addEventListener("click", toggleCamera);
recognizeBtn.addEventListener("click", () => recognize(false));
cacheBtn.addEventListener("click", () => recognize(true));
window.addEventListener("DOMContentLoaded", () => recognize(false));

async function toggleCamera() {
  if (cameraStream) {
    stopCamera();
    return;
  }

  if (!navigator.mediaDevices?.getUserMedia) {
    setStatus("No camera", "is-error");
    return;
  }

  try {
    setStatus("Camera", "is-busy");
    cameraStream = await navigator.mediaDevices.getUserMedia({
      video: {
        facingMode: "environment",
        width: {ideal: 1280},
        height: {ideal: 720}
      },
      audio: false
    });
    cameraFeed.srcObject = cameraStream;
    scene.classList.add("is-camera");
    cameraBtn.textContent = "关闭摄像头";
    setStatus("Ready", "");
  } catch (error) {
    setStatus("Camera denied", "is-error");
  }
}

async function recognize(useCache) {
  setStatus("Scanning", "is-busy");
  recognizeBtn.disabled = true;
  cacheBtn.disabled = true;
  cameraBtn.disabled = true;

  const frame = useCache ? "" : captureFrame();

  const payload = {
    device_id: "demo_glasses",
    frame_id: `frame_${Date.now()}`,
    image_base64: frame,
    last_scene_hash: useCache ? lastSceneHash : "",
    offline_ok: true
  };

  try {
    const response = await fetch("/api/vision/recognize", {
      method: "POST",
      headers: {"Content-Type": "application/json"},
      body: JSON.stringify(payload)
    });
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`);
    }
    const result = await response.json();
    lastSceneHash = result.scene_hash;
    renderResult(result);
    setStatus("Ready", "");
  } catch (error) {
    setStatus("Error", "is-error");
  } finally {
    recognizeBtn.disabled = false;
    cacheBtn.disabled = !lastSceneHash;
    cameraBtn.disabled = false;
  }
}

function captureFrame() {
  if (!cameraStream || cameraFeed.readyState < HTMLMediaElement.HAVE_CURRENT_DATA) {
    return "desk demo";
  }

  const context = captureCanvas.getContext("2d");
  context.drawImage(cameraFeed, 0, 0, captureCanvas.width, captureCanvas.height);
  return captureCanvas.toDataURL("image/jpeg", 0.72);
}

function renderResult(result) {
  overlay.replaceChildren(...result.objects.map(renderTarget));
  cards.replaceChildren(...result.objects.map(renderCard));
  sceneHashEl.textContent = result.scene_hash || "-";
  cacheStateEl.textContent = String(result.from_cache);
}

function renderTarget(object) {
  const target = document.createElement("div");
  target.className = "target";
  target.style.left = `${(object.box.x / 800) * 100}%`;
  target.style.top = `${(object.box.y / 450) * 100}%`;
  target.style.width = `${(object.box.width / 800) * 100}%`;
  target.style.height = `${(object.box.height / 450) * 100}%`;

  const tag = document.createElement("div");
  tag.className = "tag";
  tag.textContent = object.display_text;
  target.append(tag);
  return target;
}

function renderCard(object) {
  const card = document.createElement("article");
  card.className = "card";

  const letter = document.createElement("div");
  letter.className = "letter";
  letter.textContent = object.letter;

  const content = document.createElement("div");

  const word = document.createElement("div");
  word.className = "word";

  const english = document.createElement("span");
  english.className = "english";
  english.textContent = object.english;

  const chinese = document.createElement("span");
  chinese.className = "chinese";
  chinese.textContent = object.chinese;

  const sentence = document.createElement("p");
  sentence.className = "sentence";
  sentence.textContent = `${object.phonetic}  ${object.learning.example_sentence} ${object.learning.example_meaning}`;

  word.append(english, chinese);
  content.append(word, sentence);
  card.append(letter, content);
  return card;
}

function setStatus(text, className) {
  statusEl.textContent = text;
  statusEl.className = `status ${className}`.trim();
}

function stopCamera() {
  cameraStream.getTracks().forEach((track) => track.stop());
  cameraStream = null;
  cameraFeed.srcObject = null;
  scene.classList.remove("is-camera");
  cameraBtn.textContent = "打开摄像头";
}
