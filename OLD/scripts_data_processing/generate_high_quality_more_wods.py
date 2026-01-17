#!/usr/bin/env python3
"""
Generate a high-quality `seeds/more_wods.csv` by combining curated entries
and canonical seed entries, deduplicating by normalized name, and backing up
the previous `more_wods.csv`.

Usage:
  python scripts/generate_high_quality_more_wods.py

This script reads these inputs if present:
- `seeds/more_wods_curated.csv` (created from reliable web sources)
- `seeds/wods.csv` (existing canonical seeds)

It writes `seeds/more_wods.csv` and backs up the previous file to
`seeds/more_wods.csv.bak`.
"""
import csv
import shutil
import re
from pathlib import Path

REPO = Path(__file__).resolve().parent.parent
SEEDS = REPO / 'seeds'
CURATED = SEEDS / 'more_wods_curated.csv'
CANONICAL = SEEDS / 'wods.csv'
OUT = SEEDS / 'more_wods.csv'

FIELDNAMES = ['name','source','type','regime','score_type','description','url','notes','is_standard','created_by_email']

def normalize(n: str) -> str:
    if not n:
        return ''
    k = re.sub(r'[^0-9a-z\s]',' ', n.lower()).strip()
    k = re.sub(r'\s+',' ', k)
    return k

def read_csv(path: Path):
    rows = []
    if not path.exists():
        return rows
    with open(path, newline='', encoding='utf-8') as f:
        r = csv.DictReader(f)
        for row in r:
            rows.append({k: (row.get(k,'') or '') for k in FIELDNAMES})
    return rows

def main():
    curated = read_csv(CURATED)
    canonical = read_csv(CANONICAL)

    print(f'Read {len(curated)} curated entries, {len(canonical)} canonical entries')

    # preference order: curated -> canonical
    by_key = {}

    def add_row(row, source_priority=0):
        key = normalize(row.get('name',''))
        if not key:
            return
        if key in by_key:
            return
        by_key[key] = row

    for r in canonical:
        add_row(r, source_priority=1)
    for r in curated:
        add_row(r, source_priority=0)

    entries = list(by_key.values())

    # sort by name for stable output
    entries.sort(key=lambda r: normalize(r.get('name','')))

    # backup
    if OUT.exists():
        bak = OUT.with_suffix('.csv.bak')
        shutil.copy2(OUT, bak)
        print('Backup of existing more_wods.csv written to', bak)

    with open(OUT, 'w', newline='', encoding='utf-8') as f:
        w = csv.DictWriter(f, fieldnames=FIELDNAMES)
        w.writeheader()
        w.writerows(entries)

    print(f'Wrote {len(entries)} high-quality entries to {OUT}')

if __name__ == '__main__':
    main()
