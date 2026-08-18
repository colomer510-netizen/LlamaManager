@echo off
title LlamaManager Server
"%~dp0bin\llama-server.exe" -m "D:\OLLAMA AI\GGUF\gemma-4-E2B-it-Q5_K_M.gguf" -c 8192 -t 7 -ngl 0 --port 8080
pause
