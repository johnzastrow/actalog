#!/usr/bin/env python3
"""Convert SVG placeholders in site/assets to PNGs using cairosvg.
Run with: uv run python site/scripts/convert_svgs.py
"""
import sys
from pathlib import Path

try:
    from cairosvg import svg2png
except Exception as e:
    print("Missing dependency 'cairosvg'. Install with: uv pip install cairosvg")
    sys.exit(2)

ASSET_DIR = Path(__file__).resolve().parent.parent / 'assets'
SVG_FILES = [
    'hero-screenshot.svg',
    'screenshot-1.svg',
    'screenshot-2.svg',
    'screenshot-3.svg',
]

for name in SVG_FILES:
    svg = ASSET_DIR / name
    if not svg.exists():
        print('skip, not found:', svg)
        continue
    out = svg.with_suffix('.png')
    try:
        svg2png(url=str(svg), write_to=str(out))
        print('written', out)
    except Exception as e:
        print('error converting', svg, e)

print('done')
