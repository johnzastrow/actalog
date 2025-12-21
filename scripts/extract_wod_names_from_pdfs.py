#!/usr/bin/env python3
"""
Extract normalized WOD names from PDFs in `imports/` and list names
that are missing from `seeds/more_wods.csv`.

Usage:
  python scripts/extract_wod_names_from_pdfs.py

Writes `seeds/missing_wods_from_pdfs.txt` with one name per line.
"""
import re
import csv
from pathlib import Path
from parse_pdfs_to_wods import extract_text, find_matches, normalize_name

REPO = Path(__file__).resolve().parent.parent
IMPORTS = REPO / 'imports'
SEEDS = REPO / 'seeds'
OUT = SEEDS / 'missing_wods_from_pdfs.txt'

def load_existing_keys():
    keys = set()
    more = SEEDS / 'more_wods.csv'
    if more.exists():
        with open(more, newline='', encoding='utf-8') as f:
            r = csv.DictReader(f)
            for row in r:
                keys.add(normalize_name(row.get('name','')))
    wods = SEEDS / 'wods.csv'
    if wods.exists():
        with open(wods, newline='', encoding='utf-8') as f:
            r = csv.DictReader(f)
            for row in r:
                keys.add(normalize_name(row.get('name','')))
    return keys

def main():
    pdfs = list(IMPORTS.glob('*.pdf'))
    if not pdfs:
        print('No PDFs found in imports/')
        return
    existing = load_existing_keys()
    found = {}
    for p in pdfs:
        print('Scanning', p.name)
        txt = extract_text(p)
        if not txt:
            continue
        matches = find_matches(txt, [])
        for key, snippet in matches:
            # keep first snippet seen for each key
            if key not in found:
                found[key] = snippet

    missing = sorted(k for k in found.keys() if k not in existing)
    # hero-like if snippet contains 'hero' or 'memorial' or 'tribute'
    hero_like = [k for k in missing if re.search(r'\bhero\b|\bmemorial\b|\btribute\b', (found.get(k) or ''), re.I)]
    print(f'Found {len(found)} unique names in PDFs; {len(missing)} missing from seeds; {len(hero_like)} look like Hero WODs')
    with open(OUT, 'w', encoding='utf-8') as f:
        for m in missing:
            f.write(m + '\n')
    hero_out = SEEDS / 'missing_hero_wods_from_pdfs.txt'
    with open(hero_out, 'w', encoding='utf-8') as f:
        for m in hero_like:
            f.write(m + '\n')
    print('Wrote missing names to', OUT, 'and hero-like names to', hero_out)

if __name__ == '__main__':
    main()
