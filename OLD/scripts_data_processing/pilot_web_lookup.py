#!/usr/bin/env python3
"""Pilot web lookup for candidate hero WOD names.

Usage: python scripts/pilot_web_lookup.py --max 30 --append --generate

This script reads the first N lines from `seeds/candidate_hero_wods.txt`,
tries WODwell slugs for each candidate, writes results to
`seeds/web_found_pilot.csv`, and (optionally) appends new entries to
`seeds/more_wods_curated.csv` and runs the generator to rebuild
`seeds/more_wods.csv`.
"""
import argparse
import csv
import os
import re
import subprocess
import sys
import time
from html import unescape
from urllib.error import HTTPError, URLError
from urllib.request import Request, urlopen


ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
SEEDS_DIR = os.path.join(ROOT, "seeds")
CANDIDATES = os.path.join(SEEDS_DIR, "candidate_hero_wods.txt")
PILOT_CSV = os.path.join(SEEDS_DIR, "web_found_pilot.csv")
CURATED = os.path.join(SEEDS_DIR, "more_wods_curated.csv")


def slugify(name: str) -> str:
    s = name.strip()
    # remove common leading prefixes added by parser
    s = re.sub(r'^(wod|workout)\s+', '', s, flags=re.I)
    s = s.replace('&', ' and ')
    # keep letters, numbers, spaces, hyphens
    s = re.sub(r"[^\w\s-]", '', s, flags=re.U)
    s = re.sub(r"[\s_]+", '-', s.strip())
    s = s.strip('-').lower()
    return s


def try_fetch(url, timeout=8):
    headers = {'User-Agent': 'ActaLog-pilot/1.0 (+https://github.com)'}
    req = Request(url, headers=headers)
    try:
        with urlopen(req, timeout=timeout) as resp:
            content = resp.read().decode('utf-8', errors='replace')
            return 200, content
    except HTTPError as e:
        return e.code, None
    except URLError as e:
        return None, None
    except Exception:
        return None, None


def extract_title_excerpt(html: str):
    title = ''
    m = re.search(r'<title>(.*?)</title>', html, flags=re.I | re.S)
    if m:
        title = unescape(re.sub(r'<.*?>', '', m.group(1))).strip()

    # simple first paragraph extraction
    p = ''
    m2 = re.search(r'<p[^>]*>(.*?)</p>', html, flags=re.I | re.S)
    if m2:
        p = re.sub(r'<.*?>', '', m2.group(1)).strip()
        p = unescape(p)
        if len(p) > 400:
            p = p[:400].rsplit(' ', 1)[0] + '...'

    return title, p


def read_candidates(max_n):
    out = []
    if not os.path.exists(CANDIDATES):
        print(f"Candidates file missing: {CANDIDATES}")
        return out
    with open(CANDIDATES, 'r', encoding='utf-8') as f:
        for line in f:
            s = line.strip()
            if s:
                out.append(s)
                if len(out) >= max_n:
                    break
    return out


def ensure_pilot_csv():
    header = ['candidate_name', 'slug', 'url', 'page_title', 'excerpt', 'found_source', 'status']
    new = not os.path.exists(PILOT_CSV)
    f = open(PILOT_CSV, 'a', newline='', encoding='utf-8')
    writer = csv.writer(f)
    if new:
        writer.writerow(header)
    return f, writer


def curated_names_set():
    s = set()
    if not os.path.exists(CURATED):
        return s
    try:
        with open(CURATED, 'r', encoding='utf-8') as f:
            r = csv.reader(f)
            header = next(r, None)
            for row in r:
                if row:
                    name = row[0].strip().lower()
                    s.add(re.sub(r"[^\w]", '', name))
    except Exception:
        pass
    return s


def append_to_curated(name, excerpt, url):
    # curated format: name,source,type,regime,score_type,description,url,notes,is_standard,created_by_email
    row = [name, 'WODwell', '', '', '', excerpt or '', url, '', 'TRUE', '']
    write_header = not os.path.exists(CURATED)
    with open(CURATED, 'a', newline='', encoding='utf-8') as f:
        w = csv.writer(f)
        if write_header:
            w.writerow(['name','source','type','regime','score_type','description','url','notes','is_standard','created_by_email'])
        w.writerow(row)


def main():
    p = argparse.ArgumentParser()
    p.add_argument('--max', type=int, default=30)
    p.add_argument('--append', action='store_true', help='Append new web-found entries to curated CSV')
    p.add_argument('--generate', action='store_true', help='Run generator after appending to regenerate more_wods.csv')
    p.add_argument('--delay', type=float, default=1.0, help='Seconds between requests')
    args = p.parse_args()

    candidates = read_candidates(args.max)
    if not candidates:
        print('No candidates found; exiting.')
        return

    pilot_f, writer = ensure_pilot_csv()
    existing_curated = curated_names_set() if args.append else set()

    try:
        for cand in candidates:
            base_slug = slugify(cand)
            tried = []
            # generate slug variants
            variants = [base_slug]
            if ' ' in base_slug:
                variants.append(base_slug.replace('-', ''))
            variants.append(re.sub(r'-+', '-', base_slug))
            variants = [v for v in dict.fromkeys(variants) if v]

            found = False
            found_url = ''
            found_slug = ''
            page_title = ''
            excerpt = ''
            status = 'not found'

            for s in variants:
                url = f'https://wodwell.com/wod/{s}/'
                tried.append(url)
                code, html = try_fetch(url)
                if code == 200 and html:
                    title, ex = extract_title_excerpt(html)
                    page_title = title
                    excerpt = ex
                    found = True
                    found_url = url
                    found_slug = s
                    status = 'found'
                    break
                # short sleep to be polite
                time.sleep(args.delay)

            writer.writerow([cand, found_slug or '', found_url or '', page_title or '', excerpt or '', 'wodwell', status])
            print(f'{cand} -> {status} {found_url}')

            if args.append and found:
                key = re.sub(r"[^\w]", '', cand.strip().lower())
                if key not in existing_curated:
                    append_to_curated(cand, excerpt, found_url)
                    existing_curated.add(key)
                    print(f'Appended to {CURATED}: {cand}')

    finally:
        pilot_f.close()

    if args.generate and args.append:
        print('Running generator to rebuild seeds/more_wods.csv...')
        try:
            subprocess.run([sys.executable, os.path.join('scripts','generate_high_quality_more_wods.py')], check=False)
        except Exception as e:
            print('Generator run failed:', e)


if __name__ == '__main__':
    main()
