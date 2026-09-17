#!/bin/bash

# Configuración básica
APP_NAME="llamamanager"
VERSION="4.0.0"
ARCH="amd64"
DEB_DIR="build_deb/${APP_NAME}_${VERSION}_${ARCH}"

echo "🔨 Creando estructura del paquete .deb..."
rm -rf build_deb
mkdir -p "$DEB_DIR/DEBIAN"
mkdir -p "$DEB_DIR/opt/llamamanager/public"
mkdir -p "$DEB_DIR/usr/bin"
mkdir -p "$DEB_DIR/usr/share/applications"

echo "🚚 Copiando binario y archivos web..."
cp LlamaManager-Server "$DEB_DIR/opt/llamamanager/"
cp -r public/* "$DEB_DIR/opt/llamamanager/public/"
cp iniciar_agente.sh "$DEB_DIR/opt/llamamanager/"

echo "⚙️ Creando ejecutable global..."
cat <<SCRIPT > "$DEB_DIR/usr/bin/llamamanager"
#!/bin/bash
# Forzar directorio de trabajo para que encuentre "public/"
cd /opt/llamamanager
./LlamaManager-Server
SCRIPT
chmod +x "$DEB_DIR/usr/bin/llamamanager"

echo "📝 Generando archivo DEBIAN/control..."
cat <<CONTROL > "$DEB_DIR/DEBIAN/control"
Package: llamamanager
Version: $VERSION
Section: utils
Priority: optional
Architecture: $ARCH
Maintainer: Enoc
Description: Gestor Web Avanzado para modelos locales llama.cpp
CONTROL

echo "🖥️ Creando acceso directo (.desktop)..."
cat <<DESKTOP > "$DEB_DIR/usr/share/applications/llamamanager.desktop"
[Desktop Entry]
Name=LlamaManager Web
Comment=Arranca el servidor local de IA
Exec=gnome-terminal -- llamamanager
Icon=utilities-terminal
Terminal=false
Type=Application
Categories=Utility;Development;
DESKTOP

echo "🔐 Ajustando permisos..."
chmod 755 "$DEB_DIR/DEBIAN"
chmod -R 755 "$DEB_DIR/opt/llamamanager"
chmod 644 "$DEB_DIR/usr/share/applications/llamamanager.desktop"

echo "📦 Empaquetando..."
dpkg-deb --build "$DEB_DIR"

echo "✅ ¡Debian package creado exitosamente en build_deb/ !"
