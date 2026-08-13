#!/bin/bash
# Copyright (C) 2026 colomer510-netizen
# SPDX-License-Identifier: GPL-3.0-or-later
# Licensed under the GNU General Public License v3.0; see the LICENSE file in the project root.

cd "$(dirname "$0")"

echo "[1/2] Checking virtual environment..."
if [ ! -d ".venv" ]; then
    echo "Creating virtual environment..."
    python3 -m venv .venv
    ./.venv/bin/pip install fastapi uvicorn pydantic
fi

echo "[2/2] Starting Llama Admin Pro on http://127.0.0.1:8756 ..."
if which xdg-open > /dev/null; then
  xdg-open http://127.0.0.1:8756
elif which open > /dev/null; then
  open http://127.0.0.1:8756
fi

./.venv/bin/uvicorn backend.main:app --host 127.0.0.1 --port 8756 --reload
