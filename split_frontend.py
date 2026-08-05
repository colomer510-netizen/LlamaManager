import os
import re

base_dir = r"c:\Users\Joaquin Obando\Desktop\administrador de llama.cpp\static"
index_path = os.path.join(base_dir, "index.html")
css_path = os.path.join(base_dir, "css", "style.css")
js_path = os.path.join(base_dir, "js", "app.js")

with open(index_path, "r", encoding="utf-8") as f:
    content = f.read()

# Extract CSS
style_match = re.search(r"<style>(.*?)</style>", content, re.DOTALL)
if style_match:
    css_content = style_match.group(1).strip()
    with open(css_path, "w", encoding="utf-8") as f:
        f.write(css_content)
    # Replace in index.html
    content = content.replace(style_match.group(0), '<link rel="stylesheet" href="/static/css/style.css">')

# Extract JS
script_match = re.search(r"<script>\n?\"use strict\";\n?(.*?)</script>", content, re.DOTALL)
if script_match:
    js_content = '"use strict";\n' + script_match.group(1).strip()
    with open(js_path, "w", encoding="utf-8") as f:
        f.write(js_content)
    # Replace in index.html
    # Additionally add marked.js for markdown parsing
    replacement = '<script src="https://cdn.jsdelivr.net/npm/marked/marked.min.js"></script>\n<script src="/static/js/app.js"></script>'
    content = content.replace(script_match.group(0), replacement)

with open(index_path, "w", encoding="utf-8") as f:
    f.write(content)

print("Split complete!")
