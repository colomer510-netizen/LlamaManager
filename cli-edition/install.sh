#!/bin/bash
# Instalador automático de Gestor Llama.cpp Super Agente
echo "Iniciando instalación del Gestor Llama..."

# 1. Crear carpeta especial del sistema (oculta para no ensuciar)
mkdir -p ~/.scripts
echo "[OK] Carpeta ~/.scripts creada/verificada."

# 2. Mover el binario ejecutable
cp gestor-llama ~/.scripts/gestor-llama
chmod +x ~/.scripts/gestor-llama
echo "[OK] Binario instalado en ~/.scripts/gestor-llama"

# 3. Crear el acceso directo en el menú de aplicaciones
DESKTOP_FILE=~/.local/share/applications/gestor-llama.desktop

cat << 'APP' > "$DESKTOP_FILE"
[Desktop Entry]
Version=1.0
Type=Application
Name=Gestor Llama.cpp Super Agente
Comment=Lanzador profesional de IA local y Super Agente Linux
Exec=/home/enoc-colomer/.scripts/gestor-llama
Icon=utilities-terminal
Terminal=true
Categories=System;Utility;
APP

# Corregir la ruta del nombre de usuario actual (por si no es enoc-colomer)
sed -i "s|/home/enoc-colomer|$HOME|g" "$DESKTOP_FILE"
chmod 755 "$DESKTOP_FILE"
echo "[OK] Acceso directo creado en el Menú de Aplicaciones."

# 4. Crear carpeta agrupada en GNOME (opcional, si usa GNOME)
if command -v gsettings &> /dev/null; then
    CURRENT_FOLDERS=$(gsettings get org.gnome.desktop.app-folders folder-children)
    if [[ "$CURRENT_FOLDERS" != *"SistemaScript"* ]]; then
        if [[ $CURRENT_FOLDERS == "@as []" ]]; then
            NEW_FOLDERS="['SistemaScript']"
        else
            NEW_FOLDERS=$(echo $CURRENT_FOLDERS | sed "s/\]/, 'SistemaScript'\]/")
        fi
        gsettings set org.gnome.desktop.app-folders folder-children "$NEW_FOLDERS"
        gsettings set org.gnome.desktop.app-folders.folder:/org/gnome/desktop/app-folders/folders/SistemaScript/ name 'Sistema de script'
    fi
    
    CURRENT_APPS=$(gsettings get org.gnome.desktop.app-folders.folder:/org/gnome/desktop/app-folders/folders/SistemaScript/ apps)
    if [[ "$CURRENT_APPS" != *"gestor-llama"* ]]; then
        if [[ $CURRENT_APPS == "@as []" || -z "$CURRENT_APPS" ]]; then
            NEW_APPS="['gestor-llama.desktop']"
        else
            NEW_APPS=$(echo $CURRENT_APPS | sed "s/\]/, 'gestor-llama.desktop'\]/")
        fi
        gsettings set org.gnome.desktop.app-folders.folder:/org/gnome/desktop/app-folders/folders/SistemaScript/ apps "$NEW_APPS"
    fi
    echo "[OK] Añadido a la carpeta 'Sistema de script' en GNOME."
fi

update-desktop-database ~/.local/share/applications/ 2>/dev/null || true

echo "================================================="
echo "✅ Instalación completada con éxito."
echo "Busca 'Gestor Llama.cpp Super Agente' en tu menú."
echo "================================================="
