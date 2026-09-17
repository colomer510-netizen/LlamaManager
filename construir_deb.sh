#!/bin/bash

# --- CONFIGURACIÓN ---
APP_NAME="llamamanager"
APP_TITLE="LlamaManager"
VERSION="4.0.0"
ARCH="amd64"
MAINTAINER="Desarrollador LlamaManager <contacto@ejemplo.com>"
DESCRIPTION="Gestor Inteligente, Ligero y Persistente para Modelos GGUF locales."
WAILS_BIN="$HOME/go/bin/wails"

echo "==========================================="
echo "📦 Construyendo instalador .deb para Ubuntu"
echo "==========================================="

# 1. Compilar la aplicación con Wails
echo "🔨 1. Compilando la aplicación nativa (Wails)..."
cd cmd/desktop
$WAILS_BIN build -platform linux/amd64 -clean -tags webkit2_41
if [ $? -ne 0 ]; then
    echo "❌ Error al compilar con Wails."
    exit 1
fi
cd ../../

# 2. Preparar directorios del paquete
echo "📁 2. Creando estructura de carpetas DEB..."
DEB_DIR="build_deb/${APP_NAME}_${VERSION}_${ARCH}"
rm -rf build_deb
mkdir -p "$DEB_DIR/DEBIAN"
mkdir -p "$DEB_DIR/usr/bin"
mkdir -p "$DEB_DIR/usr/share/applications"
mkdir -p "$DEB_DIR/usr/share/icons/hicolor/512x512/apps"

# 3. Mover los archivos a la estructura
echo "🚚 3. Copiando binarios y recursos..."
# Wails genera el binario con el nombre de la carpeta (desktop) por defecto
cp cmd/desktop/build/bin/desktop "$DEB_DIR/usr/bin/$APP_NAME"
# Copiar el ícono de Wails
cp cmd/desktop/build/appicon.png "$DEB_DIR/usr/share/icons/hicolor/512x512/apps/$APP_NAME.png"

# 4. Crear el archivo de control (metadata del instalador)
echo "📝 4. Generando archivo de configuración (DEBIAN/control)..."
cat <<EOL > "$DEB_DIR/DEBIAN/control"
Package: $APP_NAME
Version: $VERSION
Section: utils
Priority: optional
Architecture: $ARCH
Maintainer: $MAINTAINER
Description: $DESCRIPTION
EOL

# 5. Crear el acceso directo (.desktop)
echo "🖥️ 5. Generando acceso directo del menú de aplicaciones..."
cat <<EOL > "$DEB_DIR/usr/share/applications/$APP_NAME.desktop"
[Desktop Entry]
Name=$APP_TITLE
Comment=$DESCRIPTION
Exec=/usr/bin/$APP_NAME
Icon=$APP_NAME
Terminal=false
Type=Application
Categories=Utility;Development;
EOL

# 6. Ajustar permisos
echo "🔐 6. Ajustando permisos de los archivos..."
chmod 755 "$DEB_DIR/DEBIAN"
chmod 755 "$DEB_DIR/usr/bin/$APP_NAME"
chmod 644 "$DEB_DIR/usr/share/applications/$APP_NAME.desktop"
chmod 644 "$DEB_DIR/usr/share/icons/hicolor/512x512/apps/$APP_NAME.png"

# 7. Empaquetar
echo "📦 7. Empaquetando archivo .deb..."
dpkg-deb --build "$DEB_DIR"

echo "==========================================="
echo "✅ ¡Éxito! Tu instalador está listo en:"
echo "👉 build_deb/${APP_NAME}_${VERSION}_${ARCH}.deb"
echo "==========================================="
