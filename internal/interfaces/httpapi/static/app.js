const overlay = document.querySelector("#overlay");
const cards = document.querySelector("#cards");
const wordHistory = document.querySelector("#wordHistory");
const scene = document.querySelector("#scene");
const cameraFeed = document.querySelector("#cameraFeed");
const captureCanvas = document.querySelector("#captureCanvas");
const statusEl = document.querySelector("#status");
const sceneHashEl = document.querySelector("#sceneHash");
const cacheStateEl = document.querySelector("#cacheState");
const cameraBtn = document.querySelector("#cameraBtn");
const recognizeBtn = document.querySelector("#recognizeBtn");
const autoBtn = document.querySelector("#autoBtn");
const cacheBtn = document.querySelector("#cacheBtn");
const clearHistoryBtn = document.querySelector("#clearHistoryBtn");

const localSceneKey = "glasses-english-ai:last-scene";
const learnedWordsKey = "glasses-english-ai:learned-words";
const deviceID = "demo_glasses";
const autoScanMs = 2500;

let lastSceneHash = "";
let cameraStream = null;
let autoTimer = null;

cameraBtn.addEventListener("click", toggleCamera);
recognizeBtn.addEventListener("click", () => recognize(false));
autoBtn.addEventListener("click", toggleAutoScan);
cacheBtn.addEventListener("click", () => recognize(true));
clearHistoryBtn.addEventListener("click", clearLearnedWords);
window.addEventListener("DOMContentLoaded", () => {
  renderLearnedWords();
  loadServerLearnedWords();
  restoreLocalScene();
  recognize(false);
});

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
  autoBtn.disabled = true;

  const frame = useCache ? "" : captureFrame();

  const payload = {
    device_id: deviceID,
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
    saveLocalScene(result);
    setStatus("Ready", "");
  } catch (error) {
    if (restoreLocalScene()) {
      setStatus("Offline cache", "is-offline");
    } else {
      setStatus("Error", "is-error");
    }
  } finally {
    recognizeBtn.disabled = false;
    cacheBtn.disabled = !lastSceneHash;
    cameraBtn.disabled = false;
    autoBtn.disabled = false;
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
  rememberLearnedWords(result.objects);
}

function saveLocalScene(result) {
  try {
    localStorage.setItem(localSceneKey, JSON.stringify({
      saved_at: new Date().toISOString(),
      result
    }));
  } catch (error) {
    // Local cache is best-effort; HUD should keep working without it.
  }
}

function restoreLocalScene() {
  try {
    const raw = localStorage.getItem(localSceneKey);
    if (!raw) {
      return false;
    }
    const cached = JSON.parse(raw);
    if (!cached.result?.objects?.length) {
      return false;
    }
    const result = {...cached.result, from_cache: true};
    lastSceneHash = result.scene_hash || lastSceneHash;
    renderResult(result);
    return true;
  } catch (error) {
    return false;
  }
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

  const actions = document.createElement("div");
  actions.className = "card-actions";

  const speakBtn = document.createElement("button");
  speakBtn.className = "speak-btn";
  speakBtn.type = "button";
  speakBtn.textContent = "朗读";
  speakBtn.addEventListener("click", () => speakObject(object));

  word.append(english, chinese);
  actions.append(speakBtn);
  content.append(word, sentence, actions);
  card.append(letter, content);
  return card;
}

function speakObject(object) {
  if (!window.speechSynthesis || !window.SpeechSynthesisUtterance) {
    setStatus("No TTS", "is-error");
    return;
  }

  window.speechSynthesis.cancel();
  const utterance = new SpeechSynthesisUtterance(object.speak_text || object.learning.example_sentence || object.english);
  utterance.lang = "en-US";
  utterance.rate = 0.9;
  window.speechSynthesis.speak(utterance);
  setStatus("Speaking", "is-busy");
  utterance.onend = () => setStatus("Ready", "");
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

function toggleAutoScan() {
  if (autoTimer) {
    clearInterval(autoTimer);
    autoTimer = null;
    autoBtn.textContent = "自动识别";
    autoBtn.classList.remove("is-active");
    setStatus("Ready", "");
    return;
  }

  autoTimer = setInterval(() => recognize(false), autoScanMs);
  autoBtn.textContent = "停止自动";
  autoBtn.classList.add("is-active");
  recognize(false);
}

function rememberLearnedWords(objects) {
  const learned = loadLearnedWords();
  for (const object of objects) {
    if (!object.english) {
      continue;
    }
    const key = object.english.toLowerCase();
    const previous = learned[key] || {
      english: object.english,
      chinese: object.chinese,
      count: 0,
      last_seen: ""
    };
    learned[key] = {
      ...previous,
      chinese: object.chinese || previous.chinese,
      count: previous.count + 1,
      last_seen: new Date().toISOString()
    };
  }
  saveLearnedWords(learned);
  renderLearnedWords(learned);
  syncLearnedWords(objects);
}

function loadLearnedWords() {
  try {
    return JSON.parse(localStorage.getItem(learnedWordsKey) || "{}");
  } catch (error) {
    return {};
  }
}

function saveLearnedWords(learned) {
  try {
    localStorage.setItem(learnedWordsKey, JSON.stringify(learned));
  } catch (error) {
    // Best-effort local learning history.
  }
}

function renderLearnedWords(learned = loadLearnedWords()) {
  const words = Object.values(learned)
    .sort((a, b) => b.count - a.count || a.english.localeCompare(b.english))
    .slice(0, 16);

  if (words.length === 0) {
    wordHistory.replaceChildren();
    wordHistory.textContent = "识别后会自动记录单词。";
    return;
  }

  wordHistory.replaceChildren(...words.map((word) => {
    const chip = document.createElement("div");
    chip.className = "word-chip";

    const english = document.createElement("strong");
    english.textContent = word.english;

    const chinese = document.createElement("span");
    chinese.textContent = word.chinese;

    const count = document.createElement("em");
    count.textContent = `x${word.count}`;

    chip.append(english, chinese, count);
    return chip;
  }));
}

function clearLearnedWords() {
  localStorage.removeItem(learnedWordsKey);
  renderLearnedWords({});
  fetch(`/api/learning/history?device_id=${encodeURIComponent(deviceID)}`, {method: "DELETE"})
    .catch(() => {});
}

async function syncLearnedWords(objects) {
  const words = objects
    .filter((object) => object.english)
    .map((object) => ({
      english: object.english,
      chinese: object.chinese
    }));
  if (words.length === 0) {
    return;
  }

  try {
    const response = await fetch("/api/learning/encounters", {
      method: "POST",
      headers: {"Content-Type": "application/json"},
      body: JSON.stringify({device_id: deviceID, words})
    });
    if (!response.ok) {
      return;
    }
    const history = await response.json();
    mergeServerLearnedWords(history.words || []);
  } catch (error) {
    // Local history still works when sync is unavailable.
  }
}

async function loadServerLearnedWords() {
  try {
    const response = await fetch(`/api/learning/history?device_id=${encodeURIComponent(deviceID)}`);
    if (!response.ok) {
      return;
    }
    const history = await response.json();
    mergeServerLearnedWords(history.words || []);
  } catch (error) {
    // Local history remains the offline source of truth.
  }
}

function mergeServerLearnedWords(words) {
  if (words.length === 0) {
    return;
  }

  const learned = loadLearnedWords();
  for (const word of words) {
    const key = word.english.toLowerCase();
    learned[key] = {
      english: word.english,
      chinese: word.chinese,
      count: Math.max(word.count, learned[key]?.count || 0),
      last_seen: word.last_seen || learned[key]?.last_seen || ""
    };
  }
  saveLearnedWords(learned);
  renderLearnedWords(learned);
}
