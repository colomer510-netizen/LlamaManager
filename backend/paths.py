import sys
import os

def get_base_dir():
    """Devuelve el directorio externo donde está el .exe o el script.
    Ideal para config.json o la carpeta models/."""
    if getattr(sys, 'frozen', False):
        # Si es un ejecutable de PyInstaller
        return os.path.dirname(sys.executable)
    # Si es el código fuente
    return os.path.dirname(os.path.dirname(os.path.abspath(__file__)))

def get_app_dir():
    """Devuelve el directorio interno temporal (_MEIPASS) o el script.
    Ideal para assets empaquetados como static/ o bin/."""
    if getattr(sys, 'frozen', False):
        return sys._MEIPASS
    return os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
