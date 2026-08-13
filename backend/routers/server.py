# Copyright (C) 2026 colomer510-netizen
# SPDX-License-Identifier: GPL-3.0-or-later
# Licensed under the GNU General Public License v3.0; see the LICENSE file in the project root.

from fastapi import APIRouter, HTTPException
from backend.models.settings import ServerSettings
from backend.services.process_manager import server_process
from backend.services.binary_resolver import get_binary_path, is_binary_available
from backend.paths import get_data_dir
import os

router = APIRouter()

@router.post("/start")
async def start_server(settings: ServerSettings):
    if server_process.is_running():
        return {"status": "already_running"}

    if not is_binary_available("llama-server"):
        raise HTTPException(status_code=500, detail="llama-server binary not found")

    cmd = [get_binary_path("llama-server", settings.binary_strategy)]
    cmd.extend(["-m", settings.model])
    cmd.extend(["--host", settings.host])
    cmd.extend(["--port", str(settings.port)])
    cmd.extend(["-c", str(settings.n_ctx)])
    cmd.extend(["-ngl", str(settings.ngl)])

    if settings.threads:
        cmd.extend(["-t", str(settings.threads)])
    if settings.parallel:
        cmd.extend(["-np", str(settings.parallel)])

    # Extra config
    if settings.api_key:
        cmd.extend(["--api-key", settings.api_key])

    if settings.flash_attn == "on":
        cmd.append("-fa")

    if settings.extra_args:
        import shlex
        cmd.extend(shlex.split(settings.extra_args))

    log_path = os.path.join(get_data_dir(), "server.log")
    server_process.start(cmd, log_path)

    return {"status": "started", "port": settings.port}


@router.post("/stop")
async def stop_server():
    server_process.stop()
    return {"status": "stopped"}


@router.get("/status")
async def get_status():
    return {
        "running": server_process.is_running(),
        "uptime": server_process.get_uptime(),
    }
