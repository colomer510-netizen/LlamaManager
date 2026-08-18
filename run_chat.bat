@echo off
title LlamaManager Chat
"%~dp0bin\llama-cli.exe" -m "D:\OLLAMA AI\GGUF\gemma-4-E2B-it-Q5_K_M.gguf" -c 8192 -t 7 -ngl 0 --color -i
pause
