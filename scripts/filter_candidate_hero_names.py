#!/usr/bin/env python3
"""
Filter the `seeds/missing_hero_wods_from_pdfs.txt` to produce a stricter
candidate list of likely Hero WOD names (no digits, reasonable length).

Usage:
  python scripts/filter_candidate_hero_names.py --max 100

Writes `seeds/candidate_hero_wods.txt`.
"""
import argparse
import re
from pathlib import Path

REPO = Path(__file__).resolve().parent.parent
SEEDS = REPO / 'seeds'
INP = SEEDS / 'missing_hero_wods_from_pdfs.txt'
OUT = SEEDS / 'candidate_hero_wods.txt'

def is_candidate(s: str) -> bool:
    s = s.strip()
    if not s:
        return False
    # drop if contains digits (likely measurements or counts)
    if re.search(r'\d', s):
        return False
    # only letters, spaces, and a few punctuation
    if not re.match(r"^[A-Za-z \-\'\.]{3,60}$", s):
        return False
    # drop if too generic
    low = s.lower()
    if low in ('wod','workout','burpees','run','amrap','for time'):
        return False
    return True

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--max', type=int, default=100, help='Max candidates to write')
    args = parser.parse_args()

    if not INP.exists():
        print('Input file missing:', INP)
        return
    cand = []
    with open(INP, encoding='utf-8') as f:
        for line in f:
            s = line.strip()
            if is_candidate(s):
                cand.append(s)
            if len(cand) >= args.max:
                break
    with open(OUT, 'w', encoding='utf-8') as f:
        for c in cand:
            f.write(c + '\n')
    print('Wrote', len(cand), 'candidates to', OUT)

if __name__ == '__main__':
    main()
