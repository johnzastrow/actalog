#!/usr/bin/env python3
"""
Clean `seeds/more_wods.csv` by removing noisy rows.

Filters applied by default:
- Drop rows where normalized `name` length < 3
- Drop rows where `name` contains any of the drop patterns (home, search, result, wod)
- Drop rows where `description` contains a run of non-letter characters longer than threshold

Backs up the original to `seeds/more_wods.csv.bak` before writing.

Usage:
  python scripts/clean_more_wods.py --min-name-length 3 --drop-patterns home,search,result,wod --max-garbage-run 8
"""
import csv
import re
import shutil
from pathlib import Path
import argparse

REPO = Path(__file__).resolve().parent.parent
SEEDS = REPO / 'seeds'
IN = SEEDS / 'more_wods.csv'

def normalize(n: str) -> str:
    k = re.sub(r'[^0-9a-z\s]', ' ', (n or '').lower()).strip()
    k = re.sub(r'\s+', ' ', k)
    return k

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--min-name-length', type=int, default=3)
    parser.add_argument('--drop-patterns', type=str, default='home,search,result,wod')
    parser.add_argument('--max-garbage-run', type=int, default=8)
    args = parser.parse_args()

    if not IN.exists():
        print('No', IN, 'found; nothing to do')
        return

    drop_patterns = [p.strip().lower() for p in args.drop_patterns.split(',') if p.strip()]
    bak = IN.with_suffix('.csv.bak')
    shutil.copy2(IN, bak)
    print('Backup written to', bak)

    rows = []
    removed = 0
    with open(IN, newline='', encoding='utf-8') as f:
        reader = csv.DictReader(f)
        fieldnames = reader.fieldnames
        for row in reader:
            name = row.get('name','')
            desc = row.get('description','') or ''
            key = normalize(name)
            if len(key) < args.min_name_length:
                removed += 1
                continue
            if any(pat in key for pat in drop_patterns):
                removed += 1
                continue
            if re.search(r'[^A-Za-z\n\r]{' + str(args.max_garbage_run) + r',}', desc):
                removed += 1
                continue
            rows.append(row)

    with open(IN, 'w', newline='', encoding='utf-8') as f:
        writer = csv.DictWriter(f, fieldnames=fieldnames)
        writer.writeheader()
        writer.writerows(rows)

    print('Wrote', len(rows), 'rows to', IN)
    print('Removed', removed, 'rows (backup at', bak, ')')

if __name__ == '__main__':
    main()
