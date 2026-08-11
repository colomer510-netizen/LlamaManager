const $ = id => document.getElementById(id);

let state = {
    running: false,
    logsSince: 0,
    quantizeLogsSince: 0
};

async function fetchJSON(url, options = {}) {
    const res = await fetch(url, options);
    if (!res.ok) throw new Error(await res.text());
    return await res.json();
}

// --- Navigation Tabs ---
document.querySelectorAll('.nav-btn').forEach(btn => {
    btn.addEventListener('click', (e) => {
        // Remove active class from all
        document.querySelectorAll('.nav-btn').forEach(b => b.classList.remove('active'));
        document.querySelectorAll('.tab-pane').forEach(p => p.classList.remove('active'));
        
        // Add active class to clicked
        const targetId = e.target.getAttribute('data-tab');
        e.target.classList.add('active');
        $(targetId).classList.add('active');
    });
});

// --- Server Management ---
async function scanModels() {
    try {
        const customDirsRaw = $('customDirs').value;
        localStorage.setItem('llama_custom_dirs', customDirsRaw);
        
        const dirs = customDirsRaw.split('\n').map(d => d.trim()).filter(d => d);
        
        const res = await fetchJSON('/api/system/scan_models', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(dirs) 
        });
        
        const select = $('modelSelect');
        const qSelect = $('quantizeModelSelect');
        const oldVal = select.value; // Remember selection
        const qOldVal = qSelect ? qSelect.value : null;
        select.innerHTML = '';
        if (qSelect) qSelect.innerHTML = '';
        
        if (res.models.length === 0) {
            select.innerHTML = '<option value="">No se encontraron modelos (.gguf)</option>';
            if (qSelect) qSelect.innerHTML = '<option value="">No se encontraron modelos (.gguf)</option>';
            return;
        }
        
        res.models.forEach(m => {
            const opt = document.createElement('option');
            opt.value = m;
            opt.textContent = m.split('\\').pop().split('/').pop();
            select.appendChild(opt);
            
            if (qSelect) {
                const qOpt = document.createElement('option');
                qOpt.value = m;
                qOpt.textContent = m.split('\\').pop().split('/').pop();
                qSelect.appendChild(qOpt);
            }
        });
        
        if (oldVal) select.value = oldVal;
        if (qSelect && qOldVal) qSelect.value = qOldVal;
        
    } catch (e) {
        console.error(e);
        $('modelSelect').innerHTML = '<option value="">Error escaneando modelos</option>';
        if ($('quantizeModelSelect')) $('quantizeModelSelect').innerHTML = '<option value="">Error escaneando modelos</option>';
    }
}

async function startServer() {
    try {
        const model = $('modelSelect').value;
        if (!model) {
            alert('Selecciona un modelo primero.');
            return;
        }

        $('btnStart').disabled = true;
        
        const payload = {
            model: model,
            host: $('host').value,
            port: parseInt($('port').value),
            n_ctx: parseInt($('ctx').value),
            ngl: parseInt($('ngl').value),
            binary_strategy: $('binaryStrategy') ? $('binaryStrategy').value : 'auto'
        };

        await fetchJSON('/api/server/start', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload)
        });
        
        updateStatus();
    } catch (e) {
        alert("Error: " + e.message);
        $('btnStart').disabled = false;
    }
}

async function stopServer() {
    try {
        await fetchJSON('/api/server/stop', { method: 'POST' });
        updateStatus();
    } catch (e) {
        console.error(e);
    }
}

async function updateStatus() {
    try {
        const st = await fetchJSON('/api/server/status');
        state.running = st.running;
        
        const led = $('statusLed');
        const text = $('statusText');
        
        if (st.running) {
            led.className = 'led online';
            text.textContent = 'En línea (Uptime: ' + Math.floor(st.uptime) + 's)';
            $('btnStart').disabled = true;
            $('btnStop').disabled = false;
            $('btnWebUI').style.display = 'inline-block';
            
            // Poll logs
            const logs = await fetchJSON('/api/system/logs?since=' + state.logsSince);
            if (logs.lines) {
                const out = $('logsOutput');
                if (state.logsSince === 0) out.textContent = '';
                out.textContent += logs.lines;
                
                if ($('autoScroll').checked) {
                    $('consoleBody').scrollTop = $('consoleBody').scrollHeight;
                }
            }
            state.logsSince = logs.next;
            
        } else {
            led.className = 'led offline';
            text.textContent = 'Desconectado';
            $('btnStart').disabled = false;
            $('btnStop').disabled = true;
            $('btnWebUI').style.display = 'none';
            state.logsSince = 0;
        }
    } catch (e) {
        // Backend not ready
    }
}

async function autoConfig() {
    try {
        $('btnAutoConfig').textContent = "Escaneando PC...";
        $('btnAutoConfig').disabled = true;
        
        const res = await fetchJSON('/api/system/auto-config');
        if (res.cpu_cores) {
            if ($('threads')) $('threads').value = res.recommended_threads;
            if ($('ngl')) $('ngl').value = res.recommended_ngl;
            
            const gpuText = res.gpu_name !== 'Ninguna' ? res.gpu_name : 'Gráficos Integrados';
            alert(`✅ Configuración Optimizada!\n\nProcesador: ${res.cpu_cores} núcleos (Usando ${res.recommended_threads} hilos)\nGráfica: ${gpuText}\nCapas GPU: ${res.recommended_ngl}`);
        }
    } catch (e) {
        alert("Error al intentar escanear el sistema.");
        console.error(e);
    } finally {
        $('btnAutoConfig').textContent = "✨ Auto-Configurar (Optimizado)";
        $('btnAutoConfig').disabled = false;
    }
}

// --- Quantize Management ---
async function startQuantize() {
    try {
        const inputModel = $('quantizeModelSelect').value;
        if (!inputModel) {
            alert('Selecciona un modelo de origen primero.');
            return;
        }

        const method = $('quantizeMethodSelect').value;
        let outputModel = $('quantizeOutputName').value.trim();
        
        if (!outputModel) {
            // Auto-generate name
            const baseName = inputModel.replace(/\.gguf$/i, '');
            outputModel = `${baseName}-${method}.gguf`;
        } else {
            // Ensure .gguf extension and absolute path if not provided
            if (!outputModel.toLowerCase().endsWith('.gguf')) {
                outputModel += '.gguf';
            }
            if (!outputModel.includes('/') && !outputModel.includes('\\')) {
                const lastSlash = Math.max(inputModel.lastIndexOf('/'), inputModel.lastIndexOf('\\'));
                if (lastSlash !== -1) {
                    const dir = inputModel.substring(0, lastSlash + 1);
                    outputModel = dir + outputModel;
                }
            }
        }

        $('btnStartQuantize').disabled = true;
        $('btnStopQuantize').disabled = false;
        
        state.quantizeLogsSince = 0;
        $('quantizeConsole').textContent = '';

        await fetchJSON('/api/tools/quantize/start', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                input_model: inputModel,
                output_model: outputModel,
                method: method,
                binary_strategy: $('binaryStrategy') ? $('binaryStrategy').value : 'auto'
            })
        });
        
    } catch (e) {
        alert("Error: " + e.message);
        $('btnStartQuantize').disabled = false;
        $('btnStopQuantize').disabled = true;
    }
}

async function stopQuantize() {
    try {
        await fetchJSON('/api/tools/quantize/stop', { method: 'POST' });
        $('btnStartQuantize').disabled = false;
        $('btnStopQuantize').disabled = true;
    } catch (e) {
        console.error(e);
    }
}

async function updateQuantizeStatus() {
    try {
        if ($('btnStopQuantize').disabled) return; 
        
        const logs = await fetchJSON('/api/tools/quantize/logs?since=' + state.quantizeLogsSince);
        if (logs.lines) {
            const out = $('quantizeConsole');
            out.textContent += logs.lines;
            out.scrollTop = out.scrollHeight;
        }
        state.quantizeLogsSince = logs.next;
        
        if (logs.lines.includes("Cuantización detenida") || logs.lines.includes("main: quantize time") || logs.lines.includes("error")) {
             $('btnStartQuantize').disabled = false;
             $('btnStopQuantize').disabled = true;
        }
    } catch (e) {
    }
}

// --- Installer Management ---
async function checkBinaries() {
    try {
        const strategy = $('binaryStrategy') ? $('binaryStrategy').value : 'auto';
        const res = await fetchJSON(`/api/system/check-binaries?strategy=${strategy}`);
        if (!res.installed) {
            if ($('installAlert')) $('installAlert').style.display = 'block';
            if ($('btnStart')) $('btnStart').disabled = true;
            if ($('btnStartQuantize')) $('btnStartQuantize').disabled = true;
        } else {
            if ($('installAlert')) $('installAlert').style.display = 'none';
            if ($('btnStart')) $('btnStart').disabled = state.running;
            if ($('btnStartQuantize')) $('btnStartQuantize').disabled = false;
        }
    } catch (e) {
        console.error("Error checking binaries", e);
    }
}

async function installLocal(event) {
    const btn = event && event.target ? event.target : $('btnInstallLocal');
    const originalText = btn.textContent;
    try {
        btn.disabled = true;
        btn.textContent = "Descargando... (Puede tardar)";
        await fetchJSON('/api/system/install/local', { method: 'POST' });
        alert("Instalación/Actualización local completada con éxito.");
        checkBinaries();
    } catch (e) {
        alert("Error al instalar/actualizar local: " + e.message);
    } finally {
        btn.disabled = false;
        btn.textContent = originalText;
    }
}

async function installGlobal(event) {
    const btn = event && event.target ? event.target : $('btnInstallGlobal');
    const originalText = btn.textContent;
    try {
        btn.disabled = true;
        btn.textContent = "Ejecutando PowerShell...";
        const res = await fetchJSON('/api/system/install/global', { method: 'POST' });
        alert("Comando ejecutado.\n\nResultado:\n" + (res.output || "Hecho."));
        checkBinaries();
    } catch (e) {
        alert("Error al instalar/actualizar global: " + e.message);
    } finally {
        btn.disabled = false;
        btn.textContent = originalText;
    }
}

async function checkUpdates() {
    const btn = $('btnCheckUpdates');
    const msg = $('updateStatusMsg');
    const strategy = $('binaryStrategy') ? $('binaryStrategy').value : 'auto';
    
    try {
        btn.disabled = true;
        btn.innerHTML = "🔎 Buscando...";
        msg.textContent = "";
        msg.style.color = "var(--text-secondary)";
        
        const res = await fetchJSON(`/api/system/check-updates?strategy=${strategy}`);
        
        if (!res.installed) {
            msg.textContent = "No instalado. Última versión: b" + res.latest_version;
            msg.style.color = "var(--danger-color)";
        } else if (res.has_update) {
            msg.textContent = `¡Actualización disponible! (b${res.local_version} -> b${res.latest_version})`;
            msg.style.color = "var(--success-color)";
        } else {
            msg.textContent = `Estás al día (Versión b${res.local_version}).`;
            msg.style.color = "var(--success-color)";
        }
    } catch (e) {
        msg.textContent = "Error al buscar actualizaciones.";
        msg.style.color = "var(--danger-color)";
    } finally {
        btn.disabled = false;
        btn.innerHTML = "🔎 Buscar Actualizaciones";
    }
}

// Events
$('btnScan').addEventListener('click', scanModels);
$('btnStart').addEventListener('click', startServer);
$('btnStop').addEventListener('click', stopServer);

$('btnWebUI').addEventListener('click', () => {
    const host = $('host').value || '127.0.0.1';
    const port = $('port').value || '8081';
    window.open(`http://${host}:${port}`, '_blank');
});

if ($('btnAutoConfig')) $('btnAutoConfig').addEventListener('click', autoConfig);
if ($('btnStartQuantize')) $('btnStartQuantize').addEventListener('click', startQuantize);
if ($('btnStopQuantize')) $('btnStopQuantize').addEventListener('click', stopQuantize);

// --- CONVERTER LOGIC ---
let convertInterval = null;
let convertLogCursor = 0;

async function fetchConvertLogs() {
    try {
        const res = await fetch(`/api/tools/convert/logs?since=${convertLogCursor}`);
        const data = await res.json();
        if (data.lines) {
            const consoleEl = $('convertConsole');
            consoleEl.textContent += data.lines;
            consoleEl.scrollTop = consoleEl.scrollHeight;
            convertLogCursor = data.next;
        }
    } catch (e) {
        console.error("Error fetching convert logs", e);
    }
}

async function startConvert() {
    const inputDir = $('convertInputDir').value.trim();
    const outtype = $('convertOutType').value;
    const output = $('convertOutput').value.trim();

    if (!inputDir) {
        alert("Debes ingresar la ruta de la carpeta original del modelo.");
        return;
    }

    try {
        const res = await fetch('/api/tools/convert/start', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({
                model_dir: inputDir,
                outtype: outtype,
                output_path: output
            })
        });
        
        if (!res.ok) {
            const data = await res.json();
            throw new Error(data.detail || "Error al iniciar conversión");
        }
        
        $('btnStartConvert').disabled = true;
        $('btnStopConvert').disabled = false;
        $('convertConsole').textContent = "Iniciando Conversión...\nDependiendo de tu conexión a Internet, si es la primera vez que se usa el convertidor se descargarán ~2GB de librerías de Python de IA (PyTorch).\nPor favor, ten paciencia.\n\n";
        convertLogCursor = 0;
        
        convertInterval = setInterval(fetchConvertLogs, 1500);
    } catch (e) {
        alert(e.message);
    }
}

async function stopConvert() {
    try {
        await fetch('/api/tools/convert/stop', { method: 'POST' });
        if (convertInterval) clearInterval(convertInterval);
        $('btnStartConvert').disabled = false;
        $('btnStopConvert').disabled = true;
    } catch (e) {
        alert("Error al detener la conversión: " + e.message);
    }
}

if ($('btnStartConvert')) $('btnStartConvert').addEventListener('click', startConvert);
if ($('btnStopConvert')) $('btnStopConvert').addEventListener('click', stopConvert);

// --- UPDATE LOGIC ---
if ($('binaryStrategy')) $('binaryStrategy').addEventListener('change', checkBinaries);
if ($('btnInstallLocal')) $('btnInstallLocal').addEventListener('click', installLocal);
if ($('btnInstallGlobal')) $('btnInstallGlobal').addEventListener('click', installGlobal);
if ($('btnUpdateLocal')) $('btnUpdateLocal').addEventListener('click', installLocal);
if ($('btnUpdateGlobal')) $('btnUpdateGlobal').addEventListener('click', installGlobal);
if ($('btnCheckUpdates')) $('btnCheckUpdates').addEventListener('click', checkUpdates);

// --- PROFILES LOGIC ---
const PROFILES_KEY = 'llama_profiles';

function getProfiles() {
    try {
        return JSON.parse(localStorage.getItem(PROFILES_KEY)) || {};
    } catch {
        return {};
    }
}

function saveProfiles(profiles) {
    localStorage.setItem(PROFILES_KEY, JSON.stringify(profiles));
}

function updateProfileSelect() {
    const select = $('profileSelect');
    if (!select) return;
    select.innerHTML = '<option value="">-- Perfiles Guardados --</option>';
    const profiles = getProfiles();
    for (const name in profiles) {
        const opt = document.createElement('option');
        opt.value = name;
        opt.textContent = name;
        select.appendChild(opt);
    }
}

function saveCurrentProfile() {
    const name = prompt("Nombre del perfil (ej. Llama-3 Ligero):");
    if (!name) return;
    
    const config = {
        model: $('modelSelect').value,
        customDirs: $('customDirs').value,
        binaryStrategy: $('binaryStrategy').value,
        host: $('host').value,
        port: $('port').value,
        ctx: $('ctx').value,
        ngl: $('ngl').value,
        threads: $('threads').value
    };
    
    const profiles = getProfiles();
    profiles[name] = config;
    saveProfiles(profiles);
    updateProfileSelect();
    $('profileSelect').value = name;
}

function loadProfile() {
    const name = $('profileSelect').value;
    if (!name) return;
    
    const profiles = getProfiles();
    const config = profiles[name];
    if (config) {
        if (config.model && $('modelSelect').querySelector(`option[value="${config.model}"]`)) {
            $('modelSelect').value = config.model;
        }
        if (config.customDirs !== undefined) $('customDirs').value = config.customDirs;
        if (config.binaryStrategy) $('binaryStrategy').value = config.binaryStrategy;
        if (config.host) $('host').value = config.host;
        if (config.port) $('port').value = config.port;
        if (config.ctx) $('ctx').value = config.ctx;
        if (config.ngl !== undefined) $('ngl').value = config.ngl;
        if (config.threads) $('threads').value = config.threads;
    }
}

function deleteProfile() {
    const name = $('profileSelect').value;
    if (!name) return;
    
    if (confirm(`¿Eliminar el perfil '${name}'?`)) {
        const profiles = getProfiles();
        delete profiles[name];
        saveProfiles(profiles);
        updateProfileSelect();
    }
}

if ($('btnSaveProfile')) $('btnSaveProfile').addEventListener('click', saveCurrentProfile);
if ($('btnDeleteProfile')) $('btnDeleteProfile').addEventListener('click', deleteProfile);
if ($('profileSelect')) $('profileSelect').addEventListener('change', loadProfile);

// Init
updateProfileSelect();

if (localStorage.getItem('llama_custom_dirs')) {
    $('customDirs').value = localStorage.getItem('llama_custom_dirs');
}

scanModels();
checkBinaries();

setInterval(() => {
    updateStatus();
    updateQuantizeStatus();
}, 1000);

// PWA Service Worker Registration
if ('serviceWorker' in navigator) {
    window.addEventListener('load', () => {
        navigator.serviceWorker.register('sw.js').then(registration => {
            console.log('ServiceWorker registrado con éxito.');
        }).catch(err => {
            console.log('Error al registrar ServiceWorker:', err);
        });
    });
}
