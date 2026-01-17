#!/usr/bin/env python3
"""
Parse PDF files in `imports/` to extract WOD names (Hero, benchmark, daily,
or otherwise named workouts) and nearby description text, then append
normalized deduplicated entries to `seeds/more_wods.csv`.

Requirements (one of):
 - `pdftotext` available on PATH (from poppler)
 - or Python package `pdfminer.six` (install with `pip install pdfminer.six`)

Run from repo root:
    python scripts/parse_pdfs_to_wods.py

This script uses heuristics to find candidate WOD titles in PDF text:
- short lines that look like headings (ALL CAPS or Title Case)
- lines followed by a description that contains WOD keywords ("for time",
    "AMRAP", "rounds", numbers and reps)
- explicit patterns like "NAME - For time: ..." or "Workout: NAME"

Because PDFs vary widely, manual review of `seeds/more_wods.csv` is
recommended. The script will avoid adding duplicates already present in
`seeds/wods.csv` or the output file.
"""
import subprocess
import sys
import csv
import re
from pathlib import Path
from typing import List
import argparse

REPO = Path(__file__).resolve().parent.parent
IMPORTS = REPO / 'imports'
SEEDS = REPO / 'seeds'
OUT = SEEDS / 'more_wods.csv'

HERO_NAMES = [
    'murph','dt','the seven','grace','isabel','king kong','the chief','randy',
    'nick','jason','joshie','daniel','michael','bradley','nutts','garrett',
    'filthy fifty','fight gone bad','angie','barbara','chelsea','fran','cindy',
    'helen','annie','jackie','amanda','lynne'
]

def has_pdftotext() -> bool:
    try:
        subprocess.run(['pdftotext','-v'], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
        return True
    except Exception:
        return False

def extract_text_pdftotext(path: Path) -> str:
    # pdftotext -layout keeps relative layout, output to stdout with -
    try:
        p = subprocess.run(['pdftotext','-layout', str(path), '-'], stdout=subprocess.PIPE, stderr=subprocess.PIPE, check=True)
        return p.stdout.decode('utf-8', errors='replace')
    except subprocess.CalledProcessError:
        return ''

def extract_text_pdfminer(path: Path) -> str:
    try:
        from pdfminer.high_level import extract_text
    except Exception:
        print('pdfminer.six not installed (pip install pdfminer.six)', file=sys.stderr)
        return ''
    try:
        return extract_text(str(path))
    except Exception:
        return ''

def extract_text(path: Path) -> str:
    if has_pdftotext():
        return extract_text_pdftotext(path)
    else:
        return extract_text_pdfminer(path)

def find_matches(text: str, names: List[str]) -> List[tuple]:
    """
    Previously this matched a small whitelist. Now we scan for candidate
    titles and heading-like lines and also keep whitelist matches.
    Returns list of (normalized_name, snippet).
    """
    results = []
    lower = text.lower()
    # 1) whitelist exact matches (keep original behavior)
    for n in names:
        k = n.lower().strip()
        for m in re.finditer(r'\b' + re.escape(k) + r'\b', lower):
            start = max(0, m.start()-200)
            end = min(len(text), m.end()+400)
            snippet = text[start:end].strip()
            results.append((k, snippet))

    # 2) heading/title heuristics: look at lines
    lines = text.replace('\r','').split('\n')
    # join with indices for context
    for i, raw in enumerate(lines):
        line = raw.strip()
        if not line:
            continue
        # sanity filters
        if len(line) < 2 or len(line) > 120:
            continue
        # skip lines that look like page numbers or dates
        if re.match(r'^[0-9]{1,3}$', line):
            continue
        # candidate if line contains obvious markers
        if re.search(r'\b(workout|wod|benchmark|hero|rx)[:\-\s]', line, re.I):
            # try to extract name after marker
            m = re.search(r'(?:(?:workout|wod|benchmark|hero)[:\-\s]+)(.+)$', line, re.I)
            if m:
                cand = m.group(1).strip()
            else:
                cand = line
        else:
            # Title Case heuristic: many words start with uppercase (allow short words)
            words = [w for w in re.split(r'\s+', line) if w]
            if not words:
                continue
            alpha_chars = sum(ch.isalpha() for ch in line)
            upper_chars = sum(ch.isupper() for ch in line)
            # ALL CAPS line
            if alpha_chars > 0 and upper_chars / alpha_chars > 0.6:
                cand = line
            else:
                # Title case: at least half words start with uppercase letter
                title_like = sum(1 for w in words if w[0].isupper())
                if title_like >= max(1, len(words)//2) and len(words) <= 8:
                    cand = line
                else:
                    # check if next line looks like a WOD description -> use this line as name
                    nxt = lines[i+1].lower() if i+1 < len(lines) else ''
                    if re.search(r'\bfor time\b|\bamrap\b|\brounds\b|\breps\b|\brx\b|\bminutes?\b|\bseconds?\b', nxt, re.I):
                        cand = line
                    else:
                        continue

        # clean candidate
        cand = re.sub(r'^[\-\u2013\u2014\s:]+|[\-\u2013\u2014\s:]+$', '', cand).strip()
        if not cand:
            continue
        k = normalize_name(cand)
        # avoid tiny garbage
        if len(k) < 2 or len(k) > 60:
            continue
        # snippet: 3 lines before, up to 6 after
        start = max(0, i-3)
        end = min(len(lines), i+6)
        snippet = '\n'.join(l.strip() for l in lines[start:end] if l.strip())
        # de-duplicate by key
        results.append((k, snippet))

    # de-duplicate preserving order
    seen = set()
    deduped = []
    for k, s in results:
        if k in seen:
            continue
        seen.add(k)
        deduped.append((k, s))
    return deduped

def normalize_name(n: str) -> str:
    k = re.sub(r'[^0-9a-z\s]', ' ', n.lower()).strip()
    k = re.sub(r'\s+', ' ', k)
    return k

def load_existing_keys() -> set:
    keys = set()
    if OUT.exists():
        with open(OUT, newline='', encoding='utf-8') as f:
            r = csv.DictReader(f)
            for row in r:
                keys.add(normalize_name(row.get('name','')))
    # also include canonical seeds (avoid duplicates)
    wods = SEEDS / 'wods.csv'
    if wods.exists():
        with open(wods, newline='', encoding='utf-8') as f:
            r = csv.DictReader(f)
            for row in r:
                keys.add(normalize_name(row.get('name','')))
    return keys

def append_entries(entries: List[dict]):
    write_header = not OUT.exists()
    with open(OUT, 'a', newline='', encoding='utf-8') as f:
        writer = csv.DictWriter(f, fieldnames=['name','source','type','regime','score_type','description','url','notes','is_standard','created_by_email'])
        if write_header:
            writer.writeheader()
        for e in entries:
            writer.writerow(e)

def main():
    parser = argparse.ArgumentParser(description='Parse PDFs to extract WODs and append to seeds/more_wods.csv')
    parser.add_argument('--dry-run', action='store_true', help='Do not append; just show candidates')
    parser.add_argument('--append', action='store_true', help='Append found entries to output (otherwise dry-run-like)')
    parser.add_argument('--min-name-length', type=int, default=3, help='Minimum normalized name length to keep')
    parser.add_argument('--drop-patterns', type=str, default='home,search,result', help='Comma-separated substrings; drop names containing any')
    parser.add_argument('--max-garbage-run', type=int, default=10, help='Maximum allowed consecutive non-letter chars in description')
    parser.add_argument('--sample', type=int, default=20, help='Number of sample candidate names to print in dry-run')
    args = parser.parse_args()

    pdfs = list(IMPORTS.glob('*.pdf'))
    if not pdfs:
        print('No PDFs in imports/ to process.')
        return
    drop_patterns = [p.strip().lower() for p in args.drop_patterns.split(',') if p.strip()]
    existing = load_existing_keys()
    found = {}
    for p in pdfs:
        print('Processing', p.name)
        txt = extract_text(p)
        if not txt:
            print(' - failed to extract text from', p.name)
            continue
        matches = find_matches(txt, HERO_NAMES)
        for key, snippet in matches:
            if key in existing or key in found:
                continue
            # basic filters from args
            if len(key) < args.min_name_length:
                continue
            if any(pat in key for pat in drop_patterns):
                continue
            # remove descriptions with long runs of non-letter characters
            if re.search(r'[^A-Za-z\n\r]{' + str(args.max_garbage_run) + r',}', snippet):
                continue

            # heuristics to detect regime
            regime, score_type = ('','')
            if re.search(r'\bamrap\b', snippet, re.I):
                regime, score_type = ('AMRAP','Rounds+Reps')
            elif re.search(r'\bfor time\b|\bfor time:|\bfor time;|\bft:\b', snippet, re.I):
                regime, score_type = ('Fastest Time','Time (HH:MM:SS)')
            elif re.search(r'\bminutes?\b|\bseconds?\b|\breps?\b|\brounds?\b', snippet, re.I):
                regime, score_type = ('Timed/Rep-based','Rounds+Reps')
            else:
                regime, score_type = ('Unspecified','')
            name = key.title()
            entry = {
                'name': name,
                'source': p.name,
                'type': 'Imported',
                'regime': regime,
                'score_type': score_type,
                'description': snippet.replace('\n\n', '\n').strip(),
                'url': '',
                'notes': '',
                'is_standard': 'FALSE',
                'created_by_email': ''
            }
            found[key] = entry

    if not found:
        print('No new WOD candidates found in PDFs.')
        return

    entries = list(found.values())
    print(f'Found {len(entries)} candidate WODs from {len(pdfs)} PDFs.')
    # show a sample
    sample_n = min(args.sample, len(entries))
    print('\nSample candidate names:')
    for e in entries[:sample_n]:
        print(' -', e['name'], f"(source={e['source']})")

    if args.append:
        append_entries(entries)
        print(f'Appended {len(entries)} entries to {OUT}')
    else:
        print('\nRun with --append to append these cleaned candidates to the output file.')

if __name__ == '__main__':
    main()
