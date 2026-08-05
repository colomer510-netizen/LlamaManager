from fastapi import FastAPI
from fastapi.staticfiles import StaticFiles
from fastapi.responses import FileResponse
from .api import routes
from backend.paths import get_app_dir
import os

app = FastAPI(title="Llama Admin Moderno")

app.include_router(routes.router, prefix="/api")

# Mount static files
app.mount("/static", StaticFiles(directory=os.path.join(get_app_dir(), "static")), name="static")

@app.get("/")
def serve_index():
    return FileResponse(os.path.join(get_app_dir(), "static", "index.html"))

if __name__ == "__main__":
    import uvicorn
    import webbrowser
    import threading
    
    def open_browser():
        webbrowser.open("http://127.0.0.1:8756")
        
    threading.Timer(1.5, open_browser).start()
    uvicorn.run(app, host="127.0.0.1", port=8756)
