#!/usr/bin/env python3
"""Render `claude-statusline --demo` to docs/demo.png via headless Chromium.

Needs chromium, ImageMagick and MesloLGS Nerd Font Mono installed locally.
"""
import html
import re
import subprocess
import sys
import tempfile
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
FONT = "/usr/share/fonts/TTF/MesloLGSNerdFontMono-Regular.ttf"
BG = "#1a1b26"
FG = "#c0caf5"


def ansi_to_html(text):
    out, fg, bg = [], None, None
    for part in re.split(r"(\x1b\[[0-9;]*m)", text):
        if not part.startswith("\x1b["):
            if part:
                style = (f"color:{fg};" if fg else "") + (f"background:{bg};" if bg else "")
                out.append(f'<span style="{style}">{html.escape(part)}</span>')
            continue
        codes = [int(c) for c in part[2:-1].split(";") if c]
        i = 0
        while i < len(codes):
            c = codes[i]
            if c == 0 or not codes:
                fg = bg = None
            elif c in (38, 48) and codes[i + 1] == 2:
                color = "#%02x%02x%02x" % tuple(codes[i + 2 : i + 5])
                fg, bg = (color, bg) if c == 38 else (fg, color)
                i += 4
            i += 1
        if not codes:
            fg = bg = None
    return "".join(out)


def main():
    demo = subprocess.run([str(ROOT / "claude-statusline"), "--demo"], cwd=ROOT, capture_output=True, text=True, check=True).stdout
    page = f"""<!doctype html><meta charset="utf-8"><style>
@font-face {{ font-family: Meslo; src: url(file://{FONT}); }}
body {{ margin: 0; background: {BG}; }}
pre {{ font: 15px/1.9 Meslo, monospace; color: {FG}; padding: 18px 22px; margin: 0; display: inline-block; }}
</style><pre>{ansi_to_html(demo)}</pre>"""
    with tempfile.TemporaryDirectory() as tmp:
        page_path = Path(tmp) / "demo.html"
        raw = Path(tmp) / "raw.png"
        page_path.write_text(page)
        subprocess.run(
            ["chromium", "--headless", "--disable-gpu", "--hide-scrollbars", "--force-device-scale-factor=2",
             "--window-size=1700,260", f"--screenshot={raw}", page_path.as_uri()],
            check=True, capture_output=True,
        )
        out = ROOT / "docs" / "demo.png"
        out.parent.mkdir(exist_ok=True)
        # Trim the empty viewport, then add back an even margin.
        subprocess.run(["magick", raw, "-trim", "-bordercolor", BG, "-border", "36", out], check=True)
    print(out, file=sys.stderr)


if __name__ == "__main__":
    main()
