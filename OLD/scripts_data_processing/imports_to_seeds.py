#!/usr/bin/env python3
import csv
import glob
import html
import re
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parent.parent
IMPORTS_DIR = REPO_ROOT / 'imports'
SEEDS_DIR = REPO_ROOT / 'seeds'
SEEDS_DIR.mkdir(exist_ok=True)

# Load existing names to avoid duplicates
existing_wods = set()
existing_movements = set()

with open(SEEDS_DIR / 'wods.csv', newline='', encoding='utf-8') as f:
    reader = csv.DictReader(f)
    for r in reader:
        existing_wods.add(r['name'].strip())

with open(SEEDS_DIR / 'movements.csv', newline='', encoding='utf-8') as f:
    reader = csv.DictReader(f)
    for r in reader:
        existing_movements.add(r['name'].strip())


# Output files
more_wods_path = SEEDS_DIR / 'more_wods.csv'
more_mov_path = SEEDS_DIR / 'more_movements.csv'

wod_fields = ['name','source','type','regime','score_type','description','url','notes','is_standard','created_by_email']
mov_fields = ['name','type','description','is_standard','created_by_email']

wods = []
movements = []

# Helpers

def detect_regime_and_score(text):
    t = text.lower()
    regime = ''
    score_type = ''
    if 'amrap' in t or 'amrap:' in t:
        regime = 'AMRAP'
        score_type = 'Rounds+Reps'
    elif 'for time' in t or 'for time:' in t or re.search(r'\b\d+ for time', t):
        regime = 'Fastest Time'
        score_type = 'Time (HH:MM:SS)'
    elif 'emom' in t or 'every minute' in t:
        regime = 'EMOM'
        score_type = 'Rounds+Reps'
    elif 'reps for time' in t or 'rft' in t:
        regime = 'Fastest Time'
        score_type = 'Time (HH:MM:SS)'
    elif 'tabata' in t:
        regime = 'Tabata'
        score_type = 'Rounds+Reps'
    return regime, score_type


def clean_description(s):
    if s is None:
        return ''
    # Unescape html entities, strip leading/trailing whitespace
    s = html.unescape(s)
    s = s.strip()
    # Replace CRLF with LF, ensure markdown lists preserved
    s = s.replace('\r\n', '\n')
    # If description contains HTML tags (like <p>), keep them but convert simple tags to markdown
    s = re.sub(r'<\/?p>', '\n\n', s)
    s = re.sub(r'<br\s*/?>', '\n', s)
    # Strip any remaining HTML tags except keep anchor (<a>) tags so links remain
    # This preserves raw <a href="...">links</a> in the description. Other tags
    # are removed to keep markdown/plaintext style descriptions.
    s = re.sub(r'</?(?!a\b)[^>]+>', '', s)
    return s


def normalize_name(n: str) -> str:
    """Return a normalized key for deduplication: lower, strip, remove punctuation, collapse whitespace."""
    if not n:
        return ""
    # Lowercase
    k = n.lower()
    # Replace non-alphanumeric (keep spaces) with space
    k = re.sub(r'[^0-9a-z\s]', ' ', k)
    # Collapse whitespace
    k = re.sub(r'\s+', ' ', k)
    return k.strip()

# Normalized existing sets for reliable deduplication
existing_wods_norm = {normalize_name(n) for n in existing_wods}
existing_movements_norm = {normalize_name(n) for n in existing_movements}

# Process crossfit_wods.csv
cf_path = IMPORTS_DIR / 'crossfit_wods.csv'
if cf_path.exists():
    with open(cf_path, newline='', encoding='utf-8') as f:
        reader = csv.DictReader(f)
        for row in reader:
            title = row.get('title','').strip()
            raw = row.get('wodRaw','').strip()
            url = row.get('url','').strip()
            if not title:
                continue
            key = normalize_name(title)
            # Only include canonical benchmark WODs to keep the set small.
            CANONICAL_WODS = [
                'fran','cindy','diane','helen','grace','isabel','annie','nancy','karen',
                'jackie','amanda','lynne','mary','eva','kelly','murph','dt','filthy fifty',
                'fight gone bad','king kong','the seven','angie','barbara','chelsea',
                'elizabeth','jt','randy','nate','jason','michael','daniel','tommy v',
                'bradley','roy','garrett','griff','joshie','mcghee','nutts','the chief',
                'the ghost','bull','nick','jack'
            ]
            CANONICAL_KEYS = {normalize_name(n) for n in CANONICAL_WODS}

            if not key or key in existing_wods_norm or key not in CANONICAL_KEYS:
                continue
            regime, score_type = detect_regime_and_score(raw)
            desc = clean_description(raw)
            entry = {
                'name': title,
                'source': 'CrossFit.com',
                'type': 'CrossFit',
                'regime': regime,
                'score_type': score_type,
                'description': desc,
                'url': url,
                'notes': '',
                'is_standard': 'FALSE',
                'created_by_email': ''
            }
            wods.append((key, entry))

# Process Performance_fromWodify6Nov2025.csv
perf_path = IMPORTS_DIR / 'Performance_fromWodify6Nov2025.csv'
if perf_path.exists():
    with open(perf_path, newline='', encoding='utf-8') as f:
        reader = csv.DictReader(f)
        for row in reader:
            comp_type = (row.get('Component Type') or '').strip()
            comp_name = (row.get('Component Name') or '').strip()
            comp_desc = (row.get('Component Description') or '').strip()
            perf_type = (row.get('Performance Result Type') or '').strip()
            if not comp_name:
                continue
            # If this is a metcon, add to wods (avoid duplicates)
            if comp_type.lower() in ('metcon','metcon ' ,'metcon'):
                key = normalize_name(comp_name)
                if not key or key in existing_wods_norm or key not in CANONICAL_KEYS:
                    continue
                regime, score_type_guess = detect_regime_and_score(comp_desc)
                # prefer the perf_type if available
                score_type = ''
                if perf_type:
                    score_type = perf_type
                elif score_type_guess:
                    score_type = score_type_guess
                entry = {
                    'name': comp_name,
                    'source': row.get('Location Name') or 'Wodify',
                    'type': 'Metcon',
                    'regime': regime,
                    'score_type': score_type,
                    'description': clean_description(comp_desc),
                    'url': '',
                    'notes': '',
                    'is_standard': 'FALSE',
                    'created_by_email': ''
                }
                wods.append((key, entry))
            else:
                # treat as potential movement
                # map type
                map_type = comp_type.lower() if comp_type else 'general'
                key = normalize_name(comp_name)
                if not key or key in existing_movements_norm:
                    continue
                # Some comp_descs are URLs; use as description
                entry = {
                    'name': comp_name,
                    'type': map_type,
                    'description': clean_description(comp_desc) or '',
                    'is_standard': 'FALSE',
                    'created_by_email': ''
                }
                movements.append((key, entry))

# Reduce to unique wods by normalized key, prefer entry with longer description
wod_map = {}
for key, entry in wods:
    if key in wod_map:
        # prefer longer description
        if len(entry.get('description','') or '') > len(wod_map[key].get('description','') or ''):
            wod_map[key] = entry
    else:
        wod_map[key] = entry
wods_unique = list(wod_map.values())

mov_map = {}
for key, entry in movements:
    if key in mov_map:
        if len(entry.get('description','') or '') > len(mov_map[key].get('description','') or ''):
            mov_map[key] = entry
    else:
        mov_map[key] = entry
mov_unique = list(mov_map.values())

# Write outputs
with open(more_wods_path, 'w', newline='', encoding='utf-8') as f:
    writer = csv.DictWriter(f, fieldnames=wod_fields)
    writer.writeheader()
    for e in wods_unique:
        writer.writerow(e)

with open(more_mov_path, 'w', newline='', encoding='utf-8') as f:
    writer = csv.DictWriter(f, fieldnames=mov_fields)
    writer.writeheader()
    for e in mov_unique:
        writer.writerow(e)

print(f'Wrote {len(wods_unique)} wods to {more_wods_path}')
print(f'Wrote {len(mov_unique)} movements to {more_mov_path}')
