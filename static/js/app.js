"use strict";
const $ = id => document.getElementById(id);
const state = {
  logsSince: 0,
  status: null,
  models: [],
  scanning: false,
  streamAbort: null,
  lastPresetApplied: null,
};

// ---------------- idioma y tema ----------------
const I18N = {
  es: {
    brandSub: "Administrador de llama.cpp", statusChecking: "Comprobando…",
    tabServer: "⚙️ Servidor", tabChat: "💬 Chat", tabBench: "⚡ Benchmark", tabUtils: "🔧 Utilidades",
    themeDark: "🌙 Oscuro", themeLight: "☀️ Claro", themeAuto: "🖥 Auto",
    srvTitle: "Configuración del servidor", hintServer: "llama-server.exe",
    modelLabel: "Modelo (GGUF)", modelPh: "Elige o escribe la ruta…", scanBtn: "🔍 Buscar",
    presetLabel: "Presets guardados", presetEmpty: "— sin preset —", loadBtn: "Cargar",
    saveBtn: "Guardar", deleteBtn: "Borrar", hostLabel: "Host", portLabel: "Puerto",
    ctxLabel: "Contexto (tokens, 0 = del modelo)", nglLabel: "Capas en GPU (ngl)",
    threadsLabel: "Hilos CPU (vacío = auto)", slotsLabel: "Slots paralelos (vacío = auto)",
    tempLabel: "Temperatura", topPLabel: "Top-P", topKLabel: "Top-K",
    repeatLabel: "Penalización de repetición", seedLabel: "Semilla (-1 = aleatoria)",
    flashLabel: "Flash Attention", apiKeyLabel: "API key (opcional)",
    extraLabel: "Argumentos extra (avanzado)",
    dirsSummary: "Directorios de búsqueda de modelos", dirsPh: "Un directorio por línea…",
    saveScanBtn: "Guardar y escanear",
    btnStart: "▶ Iniciar servidor", btnAutoStart: "⚡ Inicio automático", btnStop: "⏹ Detener",
    consoleTitle: "Consola", hintConsole: "logs en vivo de llama-server",
    consoleEmpty: "Inicia el servidor para ver los logs…", autoscroll: "Autodesplazar", clearBtn: "Limpiar",
    benchTitle: "llama-bench — prueba de velocidad", hintBench: "mide tok/s de prompt y generación",
    npLabel: "Tokens de prompt (n_prompt)", ngLabel: "Tokens a generar (n_gen)",
    batchLabel: "Batch size", repsLabel: "Repeticiones por prueba", fmtLabel: "Formato de salida",
    optMd: "md (tabla)", btnRunBench: "▶ Ejecutar benchmark", btnBenchAuto: "⚡ Auto (según recursos)",
    btnStopTool: "⏹ Detener", benchHint: "El benchmark tarda varios minutos: carga el modelo y repite cada prueba. Los resultados aparecen en la consola al terminar.",
    benchConsoleTitle: "Consola de resultados", benchConsoleEmpty: "Ejecuta un benchmark para ver los resultados…",
    cliTitle: "llama-cli — chat por consola", hintCli: "generación única o terminal interactiva",
    promptLabel: "Prompt", promptPh: "Escribe aquí el prompt…", nLabel: "Tokens a generar",
    btnGenCli: "▶ Generar en consola", btnTermCli: "🖥 Abrir terminal interactiva", btnCliAuto: "⚡ Auto",
    cliHint: '"Generar en consola" responde una sola vez aquí abajo. La terminal interactiva abre una ventana de Windows donde puedes chatear de ida y vuelta con el modelo.',
    tokTitle: "llama-tokenize — contar tokens", hintTok: "revisa cuántos tokens consume tu texto",
    tokModelLabel: "Modelo (define el vocabulario)", tokTextLabel: "Texto", tokTextPh: "Pega aquí el texto…",
    btnTokenize: "🔢 Tokenizar",
    qTitle: "llama-quantize — convertir modelos", hintQ: "reduce el tamaño del GGUF",
    qInLabel: "Modelo origen", qInPh: "Elige un modelo .gguf…", qOutLabel: "Archivo de salida",
    qOutPh: "Se completa automáticamente…", qTypeLabel: "Tipo de cuantización",
    qQ4KM: "Q4_K_M — calidad/tamaño equilibrado (recomendado)", qQ40: "Q4_0 — rápido, menor calidad",
    qQ5KM: "Q5_K_M — mejor calidad, más grande", qQ6K: "Q6_K — alta calidad",
    qQ80: "Q8_0 — casi sin pérdida", qIQ3: "IQ3_M — ultra compacto", qIQ4: "IQ4_XS — ultra compacto",
    qF16: "F16 — sin pérdida (base)", qThreadsLabel: "Hilos (vacío = auto)",
    btnQuant: "▶ Cuantizar", btnStopToolU: "⏹ Detener",
    qHint: "Para mejores resultados conviene partir de un modelo F16/Q8_0 (cuantizar un modelo ya cuantizado pierde calidad). La conversión puede tardar varios minutos.",
    toolsConsoleTitle: "Consola de herramientas", hintTools: "salida de llama-cli y llama-quantize",
    toolsConsoleEmpty: "Aquí verás la salida de las herramientas…",
    chatInfoOff: "Inicia el servidor para chatear", maxTokLabel: "Máx. tokens",
    btnNewChat: "＋ Nuevo chat", chatOffline1: "El servidor no está en ejecución.<br>",
    chatOffline2: "Ve a la pestaña Servidor, elige un modelo y pulsa ▶ Iniciar servidor.",
    chatInputPh: "Escribe tu mensaje… (Enter para enviar, Shift+Enter para salto de línea)",
    btnSend: "Enviar", btnAbort: "⏹ Detener", footerLoading: "Cargando…",
    autoModalTitle: "⚡ Configuración automática", btnAutoGo: "▶ Iniciar con estos valores",
    btnAutoCancel: "Cancelar",
    optAuto: "auto", optAll: "all", optCpuOnly: "0 (solo CPU)",
    stOnline: "En línea", stLoading: "Cargando modelo…", stDead: "Proceso terminado",
    stStopped: "Servidor detenido", stOnlineChat: "Servidor en línea",
    errModelFirst: "Elige un modelo primero", errModel: "Elige un modelo",
    errPrompt: "Escribe un prompt", err: "Error",
    okAdjusted: "Campos ajustados: hilos", okAutoRun: "Hilos y GPU ajustados según tus recursos",
    updTitle: "Actualización de llama.cpp",
    updHint: "GitHub releases oficiales",
    updChecking: "Comprobando versión…", updCurrent: "Versión actual",
    updLatest: "Última versión en GitHub", updUpToDate: "Ya tienes la última versión",
    updAvailable: "¡Nueva versión disponible!", updDownload: "⬇ Descargar e instalar",
    updDownloading: "Descargando… esto puede tardar varios minutos",
    updInstalling: "Instalando…", updDone: "Actualizado correctamente",
    updError: "Error de actualización", updBackup: "Copia de seguridad guardada en",
    updNoUpdate: "No hay actualización disponible",
    updCheckingBtn: "🔍 Comprobar actualización",
    updRevert: "↩ Revertir", updNotInstalled: "No instalado", updRestored: "Copia restaurada",
  },
  en: {
    brandSub: "llama.cpp manager", statusChecking: "Checking…",
    tabServer: "⚙️ Server", tabChat: "💬 Chat", tabBench: "⚡ Benchmark", tabUtils: "🔧 Utilities",
    themeDark: "🌙 Dark", themeLight: "☀️ Light", themeAuto: "🖥 Auto",
    srvTitle: "Server settings", hintServer: "llama-server.exe",
    modelLabel: "Model (GGUF)", modelPh: "Choose or type a path…", scanBtn: "🔍 Search",
    presetLabel: "Saved presets", presetEmpty: "— no preset —", loadBtn: "Load",
    saveBtn: "Save", deleteBtn: "Delete", hostLabel: "Host", portLabel: "Port",
    ctxLabel: "Context (tokens, 0 = from model)", nglLabel: "GPU layers (ngl)",
    threadsLabel: "CPU threads (empty = auto)", slotsLabel: "Parallel slots (empty = auto)",
    tempLabel: "Temperature", topPLabel: "Top-P", topKLabel: "Top-K",
    repeatLabel: "Repeat penalty", seedLabel: "Seed (-1 = random)",
    flashLabel: "Flash Attention", apiKeyLabel: "API key (optional)",
    extraLabel: "Extra arguments (advanced)",
    dirsSummary: "Model search directories", dirsPh: "One directory per line…",
    saveScanBtn: "Save and scan",
    btnStart: "▶ Start server", btnAutoStart: "⚡ Auto start", btnStop: "⏹ Stop",
    consoleTitle: "Console", hintConsole: "live logs from llama-server",
    consoleEmpty: "Start the server to see logs…", autoscroll: "Autoscroll", clearBtn: "Clear",
    benchTitle: "llama-bench — speed test", hintBench: "measures prompt and generation tok/s",
    npLabel: "Prompt tokens (n_prompt)", ngLabel: "Tokens to generate (n_gen)",
    batchLabel: "Batch size", repsLabel: "Repeats per test", fmtLabel: "Output format",
    optMd: "md (table)", btnRunBench: "▶ Run benchmark", btnBenchAuto: "⚡ Auto (by resources)",
    btnStopTool: "⏹ Stop", benchHint: "The benchmark takes several minutes: it loads the model and repeats each test. Results appear in the console when done.",
    benchConsoleTitle: "Results console", benchConsoleEmpty: "Run a benchmark to see results…",
    cliTitle: "llama-cli — console chat", hintCli: "single generation or interactive terminal",
    promptLabel: "Prompt", promptPh: "Type your prompt here…", nLabel: "Tokens to generate",
    btnGenCli: "▶ Generate in console", btnTermCli: "🖥 Open interactive terminal", btnCliAuto: "⚡ Auto",
    cliHint: '"Generate in console" answers once below. The interactive terminal opens a Windows window where you can chat back and forth with the model.',
    tokTitle: "llama-tokenize — count tokens", hintTok: "check how many tokens your text uses",
    tokModelLabel: "Model (defines the vocabulary)", tokTextLabel: "Text", tokTextPh: "Paste your text here…",
    btnTokenize: "🔢 Tokenize",
    qTitle: "llama-quantize — convert models", hintQ: "reduces GGUF size",
    qInLabel: "Source model", qInPh: "Choose a .gguf model…", qOutLabel: "Output file",
    qOutPh: "Fills in automatically…", qTypeLabel: "Quantization type",
    qQ4KM: "Q4_K_M — balanced quality/size (recommended)", qQ40: "Q4_0 — fast, lower quality",
    qQ5KM: "Q5_K_M — better quality, larger", qQ6K: "Q6_K — high quality",
    qQ80: "Q8_0 — almost lossless", qIQ3: "IQ3_M — ultra compact", qIQ4: "IQ4_XS — ultra compact",
    qF16: "F16 — lossless (base)", qThreadsLabel: "Threads (empty = auto)",
    btnQuant: "▶ Quantize", btnStopToolU: "⏹ Stop",
    qHint: "For best results start from an F16/Q8_0 model (quantizing an already quantized model loses quality). Conversion can take several minutes.",
    toolsConsoleTitle: "Tools console", hintTools: "output of llama-cli and llama-quantize",
    toolsConsoleEmpty: "Tool output will appear here…",
    chatInfoOff: "Start the server to chat", maxTokLabel: "Max tokens",
    btnNewChat: "＋ New chat", chatOffline1: "The server is not running.<br>",
    chatOffline2: "Go to the Server tab, pick a model and press ▶ Start server.",
    chatInputPh: "Type your message… (Enter to send, Shift+Enter for new line)",
    btnSend: "Send", btnAbort: "⏹ Stop", footerLoading: "Loading…",
    autoModalTitle: "⚡ Automatic configuration", btnAutoGo: "▶ Start with these values",
    btnAutoCancel: "Cancel",
    optAuto: "auto", optAll: "all", optCpuOnly: "0 (CPU only)",
    stOnline: "Online", stLoading: "Loading model…", stDead: "Process ended",
    stStopped: "Server stopped", stOnlineChat: "Server online",
    errModelFirst: "Choose a model first", errModel: "Choose a model",
    errPrompt: "Type a prompt", err: "Error",
    okAdjusted: "Fields adjusted: threads", okAutoRun: "Threads and GPU adjusted to your resources",
    updTitle: "llama.cpp update",
    updHint: "official GitHub releases",
    updChecking: "Checking version…", updCurrent: "Current version",
    updLatest: "Latest on GitHub", updUpToDate: "You already have the latest version",
    updAvailable: "A new version is available!", updDownload: "⬇ Download and install",
    updDownloading: "Downloading… this can take several minutes",
    updInstalling: "Installing…", updDone: "Updated successfully",
    updError: "Update error", updBackup: "Backup saved to",
    updNoUpdate: "No update available",
    updCheckingBtn: "🔍 Check for updates",
    updRevert: "↩ Revert", updNotInstalled: "Not installed", updRestored: "Backup restored",
  },
};

function currentLang() { return localStorage.getItem("lang") || "es"; }
function t(key) {
  const d = I18N[currentLang()];
  return (d && d[key]) || I18N.es[key] || key;
}
function applyI18n() {
  const lang = currentLang();
  document.querySelectorAll("[data-i18n]").forEach(el => { el.textContent = t(el.dataset.i18n); });
  document.querySelectorAll("[data-i18n-ph]").forEach(el => { el.placeholder = t(el.dataset.i18nPh); });
  document.documentElement.lang = lang === "en" ? "en" : "es";
}
function applyTheme() {
  const pref = localStorage.getItem("theme") || "dark";
  const light = pref === "light" || (pref === "auto" && window.matchMedia("(prefers-color-scheme: light)").matches);
  document.documentElement.dataset.theme = light ? "light" : "dark";
  $("themeSel").value = pref;
}
$("langSel").addEventListener("change", e => {
  localStorage.setItem("lang", e.target.value);
  applyI18n();
  updateChatHeader();
  setStatusPill(state.status);
});
$("themeSel").addEventListener("change", e => {
  localStorage.setItem("theme", e.target.value);
  applyTheme();
});
window.matchMedia("(prefers-color-scheme: light)").addEventListener("change", applyTheme);

// ---------------- utilidades ----------------
function toast(msg, type) {
  const t = document.createElement("div");
  t.className = "toast " + (type || "");
  t.textContent = msg;
  document.body.appendChild(t);
  setTimeout(() => t.remove(), 4000);
}
function esc(s) {
  return s.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;").replace(/"/g, "&quot;");
}
function fmtSize(bytes) {
  return (bytes / (1024 ** 3)).toFixed(2) + " GB";
}
function fmtTime(sec) {
  const h = Math.floor(sec / 3600), m = Math.floor((sec % 3600) / 60), s = sec % 60;
  return (h ? h + "h " : "") + (m ? m + "m " : "") + s + "s";
}

function renderMarkdown(txt) {
  if (typeof marked !== 'undefined') {
    return marked.parse(txt);
  }
  return esc(txt);
}

// ---------------- pestañas ----------------
document.querySelectorAll(".tab").forEach(btn => {
  btn.addEventListener("click", () => {
    document.querySelectorAll(".tab").forEach(b => b.classList.remove("active"));
    document.querySelectorAll(".tabpanel").forEach(p => p.classList.remove("active"));
    btn.classList.add("active");
    $("tab-" + btn.dataset.tab).classList.add("active");
    if (btn.dataset.tab === "chat") checkChatState();
  });
});

// ---------------- estado del servidor ----------------
async function fetchJSON(url, opts) {
  const r = await fetch(url, opts);
  if (!r.ok) throw new Error(r.status + " " + r.statusText);
  return r.json();
}

function setStatusPill(st) {
  const pill = $("statusPill"), txt = $("statusText");
  pill.className = "pill";
  if (st.running) {
    if (st.health === "ok") {
      pill.classList.add("online");
      txt.textContent = t("stOnline") + " · " + (st.model || "?") + " · :" + st.port + " · " + fmtTime(st.uptime);
    } else if (st.health === "loading") {
      pill.classList.add("loading");
      txt.textContent = t("stLoading") + " (" + fmtTime(st.uptime) + ")";
    } else {
      pill.classList.add("error");
      txt.textContent = t("stDead") + " (" + (st.model || "") + ")";
    }
  } else {
    pill.classList.add("error");
    txt.textContent = t("stStopped");
  }
}

async function pollStatus() {
  try {
    const st = await fetchJSON("/api/status");
    state.status = st;
    setStatusPill(st);
    $("btnStart").disabled = st.running;
    $("btnStop").disabled = !st.running;
    $("btnSend").disabled = !(st.running && st.health === "ok");
    updateChatHeader();
    if (st.running) {
      const logs = await fetchJSON("/api/logs?since=" + state.logsSince);
      appendLogs(logs.lines);
      state.logsSince = logs.next;
    } else {
      state.logsSince = 0;
    }
  } catch (e) { /* servidor aún arrancando */ }
}

function updateChatHeader() {
  const st = state.status;
  if (st && st.running) {
    const mi = st.model_info;
    $("chatModelLabel").textContent = st.model || "";
    if (mi && mi.meta) {
      const mm = mi.meta;
      const parts = [];
      if (mm.model_size) parts.push(fmtSize(mm.model_size));
      if (mm.context_length) parts.push("ctx " + mm.context_length);
      if (mm.quantization) parts.push(mm.quantization);
      $("chatInfoLabel").textContent = parts.join(" · ") || t("stOnlineChat");
    } else {
      $("chatInfoLabel").textContent = t("stOnlineChat");
    }
  } else {
    $("chatModelLabel").textContent = "—";
    $("chatInfoLabel").textContent = t("chatInfoOff");
  }
}

// ---------------- logs ----------------
function appendLogs(lines, con) {
  con = con || $("console");
  const wasEmpty = con.querySelector(".empty");
  if (wasEmpty) wasEmpty.remove();
  for (const l of lines) {
    const div = document.createElement("div");
    let cls = "";
    if (/error|fatal|failed|exception/i.test(l)) cls = "err";
    else if (/success|listo|OK|loaded|all good/i.test(l)) cls = "ok";
    else if (/warn/i.test(l)) cls = "warn";
    else if (/init|server|started|listening|n_ctx|model loaded/i.test(l)) cls = "info";
    div.className = cls;
    div.textContent = l;
    con.appendChild(div);
  }
  const auto = con.parentElement.querySelector("input[type=checkbox]");
  if (!auto || auto.checked) con.scrollTop = con.scrollHeight;
  if (con.children.length > 4000) {
    while (con.children.length > 3000) con.removeChild(con.firstChild);
  }
}
$("btnClearLogs").addEventListener("click", () => {
  $("console").innerHTML = '<span class="empty">Logs limpiados…</span>';
});
$("btnClearBench").addEventListener("click", () => {
  $("consoleBench").innerHTML = '<span class="empty">Logs limpiados…</span>';
});
$("btnClearUtils").addEventListener("click", () => {
  $("consoleUtils").innerHTML = '<span class="empty">Logs limpiados…</span>';
});

// ---------------- modelos ----------------
async function refreshModels(fresh) {
  const btn = $("btnScan");
  btn.disabled = true;
  btn.textContent = "Buscando…";
  try {
    const data = await fetchJSON("/api/models" + (fresh ? "?fresh=1" : ""));
    state.models = data.models;
    state.scanning = data.scanning;
    fillDatalist();
    const dirs = data.dirs || [];
    $("f_scan_dirs").value = dirs.join("\n");
    if (fresh || data.scanning) {
      pollScanUntilDone();
    }
  } catch (e) {
    toast("Error buscando modelos: " + e.message, "error");
  } finally {
    btn.disabled = false;
    btn.textContent = "🔍 Buscar";
  }
}

async function pollScanUntilDone() {
  while (state.scanning) {
    try {
      const data = await fetchJSON("/api/models");
      state.models = data.models;
      state.scanning = data.scanning;
      fillDatalist();
    } catch (e) { break; }
    await new Promise(r => setTimeout(r, 2000));
  }
  if (state.models.length === 0) {
    toast("No se encontraron modelos .gguf. Escribe la ruta manualmente.", "error");
  } else {
    toast("Encontrados " + state.models.length + " modelos");
  }
}

function fillDatalist() {
  const dl = $("modelsList");
  dl.innerHTML = "";
  for (const m of state.models) {
    const opt = document.createElement("option");
    opt.value = m.path;
    opt.label = m.name + " (" + m.size_gb + " GB)";
    dl.appendChild(opt);
  }
  if (state.models.length && !$("modelPath").value) {
    $("modelPath").value = state.models[0].path;
  }
}

$("btnScan").addEventListener("click", () => refreshModels(true));

// ---------------- config / presets ----------------
function readSettings() {
  return {
    host: $("f_host").value, port: $("f_port").value, ctx: $("f_ctx").value,
    ngl: $("f_ngl").value, threads: $("f_threads").value, slots: $("f_slots").value,
    temp: $("f_temp").value, top_p: $("f_top_p").value, top_k: $("f_top_k").value,
    repeat: $("f_repeat").value, seed: $("f_seed").value, flash: $("f_flash").value,
    api_key: $("f_api_key").value, extra_args: $("f_extra").value,
  };
}
function applySettings(s) {
  $("f_host").value = s.host ?? "127.0.0.1";
  $("f_port").value = s.port ?? "8080";
  $("f_ctx").value = s.ctx ?? "";
  $("f_ngl").value = s.ngl ?? "auto";
  $("f_threads").value = s.threads ?? "";
  $("f_slots").value = s.slots ?? "";
  $("f_temp").value = s.temp ?? "0.80";
  $("f_top_p").value = s.top_p ?? "0.95";
  $("f_top_k").value = s.top_k ?? "40";
  $("f_repeat").value = s.repeat ?? "1.00";
  $("f_seed").value = s.seed ?? "-1";
  $("f_flash").value = s.flash ?? "auto";
  $("f_api_key").value = s.api_key ?? "";
  $("f_extra").value = s.extra_args ?? "";
}

async function loadConfig() {
  try {
    const cfg = await fetchJSON("/api/config");
    applySettings(cfg.last_settings || {});
    fillPresets(cfg.presets || {});
  } catch (e) {
    toast("Error cargando configuración: " + e.message, "error");
  }
}

function fillPresets(presets) {
  const sel = $("presetSelect");
  sel.innerHTML = '<option value="">— sin preset —</option>';
  for (const name of Object.keys(presets)) {
    const opt = document.createElement("option");
    opt.value = name;
    opt.textContent = name;
    sel.appendChild(opt);
  }
}

async function saveLastSettings() {
  try {
    await fetch("/api/config", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ action: "save_last", settings: readSettings() }),
    });
  } catch (e) {}
}

$("btnSavePreset").addEventListener("click", async () => {
  const name = prompt("Nombre del preset:");
  if (!name) return;
  const settings = readSettings();
  settings._model = $("modelPath").value;
  try {
    const r = await fetchJSON("/api/config", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ action: "save_preset", name, settings }),
    });
    fillPresets(r.presets);
    toast("Preset guardado: " + name, "ok");
  } catch (e) { toast("Error: " + e.message, "error"); }
});

$("btnLoadPreset").addEventListener("click", async () => {
  const name = $("presetSelect").value;
  if (!name) return;
  try {
    const cfg = await fetchJSON("/api/config");
    const s = cfg.presets[name];
    if (!s) return toast("Preset no encontrado", "error");
    applySettings(s);
    if (s._model) $("modelPath").value = s._model;
    toast("Preset cargado: " + name, "ok");
  } catch (e) { toast("Error: " + e.message, "error"); }
});

$("btnDeletePreset").addEventListener("click", async () => {
  const name = $("presetSelect").value;
  if (!name) return;
  if (!confirm("¿Borrar el preset \"" + name + "\"?")) return;
  try {
    const r = await fetchJSON("/api/config", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ action: "delete_preset", name }),
    });
    fillPresets(r.presets);
    toast("Preset borrado");
  } catch (e) { toast("Error: " + e.message, "error"); }
});

$("btnSaveDirs").addEventListener("click", async () => {
  const dirs = $("f_scan_dirs").value.split("\n").map(d => d.trim()).filter(Boolean);
  try {
    await fetchJSON("/api/config", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ action: "set_scan_dirs", dirs }),
    });
    toast("Directorio guardado, escaneando…", "ok");
    setTimeout(() => refreshModels(true), 400);
  } catch (e) { toast("Error: " + e.message, "error"); }
});

// ---------------- iniciar / detener ----------------
$("btnStart").addEventListener("click", async () => {
  const model = $("modelPath").value.trim();
  if (!model) return toast(t("errModelFirst"), "error");
  const settings = readSettings();
  $("btnStart").disabled = true;
  try {
    const r = await fetchJSON("/api/start", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ ...settings, model }),
    });
    if (r.ok) {
      toast(r.message, "ok");
      $("console").querySelector(".empty")?.remove();
      $("tab-chat").classList.remove("active");
      document.querySelector('[data-tab="servidor"]').click();
    } else {
      toast(r.error || "No se pudo iniciar", "error");
      appendLogs([(r.error || "").split("\n")]);
    }
  } catch (e) {
    toast("Error al iniciar: " + e.message, "error");
  } finally {
    $("btnStart").disabled = false;
  }
});

$("btnStop").addEventListener("click", async () => {
  try {
    const r = await fetchJSON("/api/stop", { method: "POST" });
    toast(r.message || "Detenido");
  } catch (e) { toast("Error al detener: " + e.message, "error"); }
});

// ---------------- inicio automático ----------------
let autoPendingModel = null;

$("btnAutoStart").addEventListener("click", async () => {
  const model = $("modelPath").value.trim();
  if (!model) return toast(t("errModelFirst"), "error");
  $("btnAutoStart").disabled = true;
  try {
    const r = await fetchJSON("/api/auto-config?model=" + encodeURIComponent(model) + "&lang=" + currentLang());
    if (!r.ok) return toast(r.error, "error");
    autoPendingModel = model;
    const html = r.explicacion.map(l =>
      l.startsWith("Capas en GPU: 0") && /insuficiente/.test(l)
        ? '<div class="warn">• ' + esc(l) + "</div>"
        : "<div>• " + esc(l) + "</div>"
    ).join("");
    $("autoExplain").innerHTML = html +
      '<div class="auto-args">' + esc(r.args.join(" ")) + "</div>";
    $("autoModal").style.display = "flex";
  } catch (e) {
    toast("Error: " + e.message, "error");
  } finally {
    $("btnAutoStart").disabled = false;
  }
});

$("btnAutoCancel").addEventListener("click", () => {
  $("autoModal").style.display = "none";
});

$("btnAutoGo").addEventListener("click", async () => {
  $("autoModal").style.display = "none";
  const settings = readSettings();
  $("btnStart").disabled = true;
  try {
    const r = await fetchJSON("/api/start", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        auto: true,
        model: autoPendingModel,
        lang: currentLang(),
        host: settings.host,
        port: settings.port,
        api_key: settings.api_key,
        extra_args: settings.extra_args,
      }),
    });
    if (r.ok) {
      toast(r.message, "ok");
      $("console").querySelector(".empty")?.remove();
      document.querySelector('[data-tab="servidor"]').click();
    } else {
      toast(r.error || "No se pudo iniciar", "error");
      appendLogs([(r.error || "").split("\n")]);
    }
  } catch (e) {
    toast("Error al iniciar: " + e.message, "error");
  } finally {
    $("btnStart").disabled = false;
  }
});

// ---------------- chat ----------------
const chatMessages = [];
function addMsg(role, content, meta) {
  const area = $("chat-area");
  area.querySelector(".chat-offline")?.remove();
  const div = document.createElement("div");
  div.className = "msg " + (role === "user" ? "user" : role === "error" ? "error" : "assistant");
  div.innerHTML = '<span class="role-tag">' + (role === "user" ? "Tú" : role === "error" ? "Error" : "Modelo") + "</span>" +
    '<span class="content">' + renderMarkdown(content || "") + "</span>" +
    (meta ? '<span class="meta">' + meta + "</span>" : "");
  area.appendChild(div);
  area.scrollTop = area.scrollHeight;
  return div;
}

function checkChatState() {
  const st = state.status;
  const area = $("chat-area");
  const off = area.querySelector(".chat-offline");
  if (!st || !st.running) {
    if (!off) {
      area.innerHTML = '<div class="chat-offline"><div class="big">💬</div>' +
        "El servidor no está en ejecución.<br>Ve a la pestaña <b>Servidor</b>, elige un modelo y pulsa <b>▶ Iniciar servidor</b>.</div>";
    }
  }
}

$("btnNewChat").addEventListener("click", () => {
  chatMessages.length = 0;
  $("chat-area").innerHTML = "";
});

$("chatInput").addEventListener("keydown", e => {
  if (e.key === "Enter" && !e.shiftKey) {
    e.preventDefault();
    $("btnSend").click();
  }
});

async function sendChat() {
  const input = $("chatInput");
  const text = input.value.trim();
  if (!text || !state.status || !state.status.running) return;
  chatMessages.push({ role: "user", content: text });
  addMsg("user", text);
  input.value = "";
  $("btnSend").disabled = true;
  $("btnAbort").style.display = "";
  const msgDiv = addMsg("assistant", "");
  const contentDiv = msgDiv.querySelector(".content");

  const maxTokens = parseInt($("f_chat_max").value, 10) || 2048;
  const payload = {
    model: state.status.model || "local",
    messages: chatMessages,
    temperature: parseFloat($("f_chat_temp").value || "0.8"),
    top_p: parseFloat($("f_top_p").value || "0.95"),
    max_tokens: maxTokens,
    stream: true,
  };

  const ac = new AbortController();
  state.streamAbort = ac;
  let full = "", thinking = "", usage = null, finish = "", timings = null;
  const render = () => {
    let html = "";
    if (thinking) {
      html += '<span class="thinking">' + esc(thinking) + "</span>";
    }
    html += renderMarkdown(full);
    contentDiv.innerHTML = html;
    $("chat-area").scrollTop = $("chat-area").scrollHeight;
  };
  try {
    const res = await fetch("/api/chat", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
      signal: ac.signal,
    });
    if (!res.ok) {
      let msg = "Error del servidor (" + res.status + ")";
      try { const j = await res.json(); if (j.error) msg = j.error; } catch (e) {}
      throw new Error(msg);
    }
    const reader = res.body.getReader();
    const decoder = new TextDecoder();
    let buf = "";
    while (true) {
      const { done, value } = await reader.read();
      if (done) break;
      buf += decoder.decode(value, { stream: true });
      const parts = buf.split("\n\n");
      buf = parts.pop();
      for (const part of parts) {
        for (const line of part.split("\n")) {
          if (!line.startsWith("data:")) continue;
          const data = line.slice(5).trim();
          if (data === "[DONE]") continue;
          try {
            const obj = JSON.parse(data);
            const delta = obj.choices?.[0]?.delta;
            if (delta) {
              if (typeof delta.reasoning_content === "string") {
                thinking += delta.reasoning_content;
              }
              if (typeof delta.content === "string") {
                full += delta.content;
              }
            }
            if (obj.choices?.[0]?.finish_reason) finish = obj.choices[0].finish_reason;
            if (obj.usage) usage = obj.usage;
            if (obj.timings) timings = obj.timings;
            render();
          } catch (e) {}
        }
      }
    }
  } catch (e) {
    if (e.name === "AbortError") {
      full += "\n\n_[generación detenida por el usuario]_";
      render();
    } else {
      addMsg("error", String(e.message || e));
      chatMessages.pop();
    }
  } finally {
    state.streamAbort = null;
    $("btnSend").disabled = !(state.status && state.status.running);
    $("btnAbort").style.display = "none";
    if (full.trim() || thinking.trim()) {
      chatMessages.push({ role: "assistant", content: full });
      const st = state.status;
      const metaBits = [];
      if (finish === "length") metaBits.push("límite de tokens alcanzado");
      if (usage) {
        metaBits.push("tokens: " + usage.total_tokens +
          " (prompt " + usage.prompt_tokens + " + generados " + usage.completion_tokens + ")");
      }
      if (timings && timings.predicted_per_second) {
        metaBits.push(Math.round(timings.predicted_per_second) + " tok/s");
      }
      if (st && st.uptime) metaBits.push("servidor activo: " + fmtTime(st.uptime));
      if (metaBits.length) {
        const meta = document.createElement("span");
        meta.className = "meta";
        meta.textContent = metaBits.join(" · ");
        msgDiv.appendChild(meta);
      }
    }
  }
}
function objTimings(usage) {
  return null;
}

$("btnSend").addEventListener("click", sendChat);
$("btnAbort").addEventListener("click", () => state.streamAbort?.abort());

// ---------------- footer ----------------
async function loadTools() {
  try {
    const t = await fetchJSON("/api/tools");
    const names = t.tools.map(x => x.name).join(", ");
    $("footerInfo").textContent = "Herramientas detectadas en bin/ (" + t.tools.length + "): " + names;
  } catch (e) {
    $("footerInfo").textContent = "bin/ no encontrado";
  }
}

// ---------------- herramientas (bench, cli, quantize) ----------------
const toolState = { since: 0, running: false };

async function runTool(tool, args) {
  try {
    const r = await fetchJSON("/api/tool/run", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ tool, args }),
    });
    if (r.ok) {
      toolState.since = 0;
      $("consoleBench").innerHTML = "";
      $("consoleUtils").innerHTML = "";
      toast(r.message, "ok");
    } else {
      toast(r.error || "No se pudo iniciar", "error");
      appendLogs([(r.error || "").split("\n")], $("consoleUtils"));
    }
  } catch (e) {
    toast("Error: " + e.message, "error");
  }
}

async function stopTool() {
  try {
    const r = await fetchJSON("/api/tool/stop", { method: "POST" });
    toast(r.message || "Detenido");
  } catch (e) { toast("Error: " + e.message, "error"); }
}

function updateToolButtons(st) {
  toolState.running = st.running;
  for (const id of ["btnRunBench", "btnRunCli", "btnQuant"]) {
    $(id).disabled = st.running;
  }
  $("btnStopTool").disabled = !st.running;
  $("btnStopToolU").disabled = !st.running;
}

async function pollToolLogs() {
  try {
    const data = await fetchJSON("/api/tool/logs?since=" + toolState.since);
    if (data.lines.length) {
      appendLogs(data.lines, $("consoleBench"));
      appendLogs(data.lines, $("consoleUtils"));
    }
    toolState.since = data.next;
  } catch (e) { /* admin aún arrancando */ }
}

async function pollToolStatus() {
  try {
    const st = await fetchJSON("/api/tool/status");
    updateToolButtons(st);
  } catch (e) {}
}

$("btnRunBench").addEventListener("click", () => {
  const model = $("bm_model").value.trim();
  if (!model) return toast(t("errModel"), "error");
  const args = ["-m", model, "-p", $("bm_np").value || "512", "-n", $("bm_ng").value || "128",
    "-b", $("bm_batch").value || "2048", "-r", $("bm_reps").value || "5",
    "--progress", "-o", $("bm_fmt").value || "md"];
  const threads = $("bm_threads").value.trim();
  if (threads) args.push("-t", threads);
  const ngl = $("bm_ngl").value;
  if (ngl && ngl !== "-1") args.push("-ngl", ngl);
  const flash = $("bm_flash").value;
  if (flash !== "auto") args.push("-fa", flash);
  const extra = $("bm_extra").value.trim();
  if (extra) args.push(...extra.split(/\s+/).filter(Boolean));
  runTool("bench", args);
});

$("btnBenchAuto").addEventListener("click", async () => {
  const model = $("bm_model").value.trim();
  if (!model) return toast(t("errModel"), "error");
  try {
    const r = await fetchJSON("/api/auto-config?model=" + encodeURIComponent(model) + "&lang=" + currentLang());
    if (!r.ok) return toast(r.error, "error");
    const getVal = f => { const i = r.args.indexOf(f); return i >= 0 ? r.args[i + 1] : null; };
    const t = getVal("-t"), ngl = getVal("-ngl");
    $("bm_threads").value = t || "";
    $("bm_ngl").value = (ngl && ngl !== "-1") ? ngl : "-1";
    $("bm_batch").value = "2048";
    toast(t("okAdjusted") + " " + (t || "auto") + " · GPU " + (ngl || "auto") +
      " (según tus recursos)", "ok");
  } catch (e) { toast("Error: " + e.message, "error"); }
});

$("btnRunCli").addEventListener("click", () => {  const model = $("cli_model").value.trim();
  const prompt = $("cli_prompt").value.trim();
  if (!model) return toast(t("errModel"), "error");
  if (!prompt) return toast(t("errPrompt"), "error");
  const args = ["-m", model, "-p", prompt, "-n", $("cli_n").value || "256",
    "--temp", $("cli_temp").value || "0.80", "--no-conversation", "--no-display-prompt"];
  runTool("cli", args);
});

$("btnCliAuto").addEventListener("click", async () => {
  const model = $("cli_model").value.trim();
  const prompt = $("cli_prompt").value.trim();
  if (!model) return toast(t("errModel"), "error");
  if (!prompt) return toast(t("errPrompt"), "error");
  try {
    const r = await fetchJSON("/api/auto-config?model=" + encodeURIComponent(model) + "&lang=" + currentLang());
    if (!r.ok) return toast(r.error, "error");
    const getVal = f => { const i = r.args.indexOf(f); return i >= 0 ? r.args[i + 1] : null; };
    const args = ["-m", model, "-p", prompt, "-n", $("cli_n").value || "256",
      "--temp", $("cli_temp").value || "0.80", "--no-conversation", "--no-display-prompt",
      "-t", getVal("-t") || "", "-ngl", getVal("-ngl") || "0"];
    runTool("cli", args);
    toast(t("okAutoRun"), "ok");
  } catch (e) { toast("Error: " + e.message, "error"); }
});

$("btnTermCli").addEventListener("click", async () => {
  const model = $("cli_model").value.trim();
  if (!model) return toast(t("errModel"), "error");
  const args = ["-m", model, "-i", "-c", "4096"];
  try {
    const r = await fetchJSON("/api/tool/terminal", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ tool: "cli", args }),
    });
    toast(r.message, r.ok ? "ok" : "error");
  } catch (e) { toast("Error: " + e.message, "error"); }
});

$("btnQuant").addEventListener("click", () => {
  const src = $("q_in").value.trim();
  const out = $("q_out").value.trim();
  if (!src) return toast("Elige el modelo origen", "error");
  if (!out) return toast("Indica el archivo de salida", "error");
  if (src.toLowerCase() === out.toLowerCase()) return toast("El archivo de salida no puede ser el mismo", "error");
  const args = [src, out, $("q_type").value];
  const threads = $("q_threads").value.trim();
  if (threads) args.push(threads);
  runTool("quantize", args);
});

$("q_in").addEventListener("change", () => {
  if (!$("q_out").value.trim()) {
    const src = $("q_in").value.trim();
    if (src) {
      const dir = src.replace(/[\\/][^\\/]+$/, "");
      const name = src.replace(/^.*[\\/]/, "").replace(/\.gguf$/i, "");
      $("q_out").value = dir + "\\" + name + "-" + $("q_type").value + ".gguf";
    }
  }
});

$("q_type").addEventListener("change", () => {
  const out = $("q_out").value.trim();
  const src = $("q_in").value.trim();
  if (out && src && /-Q[0-9A-Z_]+\.gguf$/i.test(out)) {
    const name = src.replace(/^.*[\\/]/, "").replace(/\.gguf$/i, "");
    $("q_out").value = out.replace(/[\\/][^\\/]+$/, "") + "\\" + name + "-" + $("q_type").value + ".gguf";
  }
});

$("btnStopTool").addEventListener("click", stopTool);
$("btnStopToolU").addEventListener("click", stopTool);

// ---------------- tokenize ----------------
$("btnTokenize").addEventListener("click", async () => {
  const model = $("tok_model").value.trim();
  const text = $("tok_text").value;
  if (!model) return toast(t("errModel"), "error");
  if (!text.trim()) return toast("Escribe algún texto", "error");
  $("btnTokenize").disabled = true;
  $("btnTokenize").textContent = "Tokenizando…";
  try {
    const r = await fetchJSON("/api/tokenize", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ model, text }),
    });
    if (r.ok) {
      $("tokResult").style.display = "";
      $("tokCount").textContent = r.count + " tokens";
      const list = $("tokList");
      list.innerHTML = "";
      for (const t of r.tokens) {
        const chip = document.createElement("span");
        chip.className = "tok-chip";
        chip.textContent = t.token;
        const id = document.createElement("span");
        id.className = "id";
        id.textContent = " " + t.id;
        chip.appendChild(id);
        list.appendChild(chip);
      }
    } else {
      toast(r.error || "Error tokenizando", "error");
    }
  } catch (e) {
    toast("Error: " + e.message, "error");
  } finally {
    $("btnTokenize").disabled = false;
    $("btnTokenize").textContent = "🔢 Tokenizar";
  }
});

// ---------------- actualización de llama.cpp ----------------
let updPolling = false;

async function checkUpdate() {
  const status = $("updStatus");
  $("btnUpdCheck").disabled = true;
  status.textContent = t("updChecking");
  try {
    const r = await fetchJSON("/api/update/check");
    if (!r.ok) return toast(r.error, "error");
    const cur = $("updCurrent");
    cur.querySelector(".dot").style.background = r.up_to_date ? "var(--accent2)" : "var(--warn)";
    $("updCurrentText").textContent = t("updCurrent") + ": " + (r.current ? r.current : t("updNotInstalled"));
    const lines = [t("updLatest") + ": " + (r.latest_tag || "?")];
    if (r.asset) lines.push(r.asset.name + " (" + r.asset.size_mb + " MB)");
    if (r.up_to_date) {
      lines.push("✓ " + t("updUpToDate"));
      $("btnUpdRun").style.display = "none";
    } else {
      lines.push("▲ " + t("updAvailable"));
      $("btnUpdRun").style.display = "";
      $("btnUpdRun").dataset.asset = r.asset ? JSON.stringify(r.asset) : "";
    }
    if (r.backups && r.backups.length) $("btnUpdRevert").style.display = "";
    status.textContent = lines.join("\n");
  } catch (e) {
    status.textContent = t("updError") + ": " + e.message;
    toast(t("updError") + ": " + e.message, "error");
  } finally {
    $("btnUpdCheck").disabled = false;
  }
}

$("btnUpdCheck").addEventListener("click", checkUpdate);

$("btnUpdRun").addEventListener("click", async () => {
  const status = $("updStatus");
  try {
    const r = await fetchJSON("/api/update/run", {
      method: "POST", headers: { "Content-Type": "application/json" },
      body: "{}",
    });
    if (!r.ok) return toast(r.error, "error");
    status.textContent = t("updDownloading");
    $("btnUpdRun").disabled = true;
    updPolling = true;
    toast(r.message, "ok");
  } catch (e) { toast(t("updError") + ": " + e.message, "error"); }
});

$("btnUpdRevert").addEventListener("click", async () => {
  try {
    const r = await fetchJSON("/api/update/revert", {
      method: "POST", headers: { "Content-Type": "application/json" },
      body: "{}",
    });
    toast(r.message, r.ok ? "ok" : "error");
    if (r.ok) $("btnUpdRevert").style.display = "none";
    checkUpdate();
  } catch (e) { toast(t("updError") + ": " + e.message, "error"); }
});

async function pollUpdateStatus() {
  if (!updPolling) return;
  try {
    const r = await fetchJSON("/api/update/status");
    const status = $("updStatus");
    if (r.phase === "done") {
      updPolling = false;
      $("btnUpdRun").disabled = false;
      $("btnUpdRun").style.display = "none";
      status.textContent = "✓ " + t("updDone") +
        (r.backup ? "\n" + t("updBackup") + ": " + r.backup : "");
      toast(t("updDone"), "ok");
    } else if (r.phase === "error") {
      updPolling = false;
      $("btnUpdRun").disabled = false;
      status.textContent = t("updError") + ": " + (r.error || "");
      toast(t("updError") + ": " + (r.error || ""), "error");
    } else if (r.phase !== "idle") {
      status.textContent = r.detail || r.phase;
    }
  } catch (e) { /* admin reiniciando */ }
}

// ---------------- inicio ----------------
(async function init() {
  $("langSel").value = currentLang();
  applyI18n();
  applyTheme();
  await Promise.allSettled([loadConfig(), refreshModels(false), loadTools(), pollStatus(),
    pollToolStatus(), pollToolLogs()]);
  setInterval(pollStatus, 3000);
  setInterval(pollToolStatus, 2000);
  setInterval(pollToolLogs, 1500);
  setInterval(pollUpdateStatus, 1500);
})();

// --- WEBSOCKETS LOGS ---
let wsLogs = null;
function connectWebSocket() {
    if (wsLogs) wsLogs.close();
    wsLogs = new WebSocket('ws://' + location.host + '/api/ws/logs');
    wsLogs.onmessage = (e) => {
        const data = JSON.parse(e.data);
        if (data.lines && data.lines.length > 0) {
            const consoleDiv = document.getElementById('console');
            const wasAtBottom = Math.abs(consoleDiv.scrollHeight - consoleDiv.scrollTop - consoleDiv.clientHeight) < 10;
            
            data.lines.forEach(line => {
                const el = document.createElement('div');
                el.textContent = line;
                consoleDiv.appendChild(el);
            });
            
            if (document.getElementById('chkAutoscroll').checked && wasAtBottom) {
                consoleDiv.scrollTop = consoleDiv.scrollHeight;
            }
        }
    };
    wsLogs.onclose = () => setTimeout(connectWebSocket, 3000);
}
connectWebSocket();
