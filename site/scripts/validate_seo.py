#!/usr/bin/env python3
"""Simple SEO / Open Graph validator for static HTML files in site/.

Checks each HTML file for:
- <title> presence and reasonable length (10-70 chars)
- meta description presence and reasonable length (50-160 chars)
- Open Graph tags: og:title, og:description, og:image
- Twitter card tag: twitter:card

Exits with code 0 if all checks pass, otherwise non-zero.
"""
import sys
from html.parser import HTMLParser
from pathlib import Path


class HeadParser(HTMLParser):
    def __init__(self):
        super().__init__()
        self.in_title = False
        self.title = ''
        self.metas = {}

    def handle_starttag(self, tag, attrs):
        if tag.lower() == 'title':
            self.in_title = True
        if tag.lower() == 'meta':
            d = dict(attrs)
            # meta can have name, property, content
            key = d.get('property') or d.get('name')
            if key:
                self.metas[key.lower()] = d.get('content', '')

    def handle_endtag(self, tag):
        if tag.lower() == 'title':
            self.in_title = False

    def handle_data(self, data):
        if self.in_title:
            self.title += data.strip()


def check_file(path: Path):
    text = path.read_text(encoding='utf8')
    # parse only head to be faster
    head_start = text.lower().find('<head')
    head_end = text.lower().find('</head>')
    chunk = text
    if head_start != -1 and head_end != -1:
        chunk = text[head_start:head_end+7]

    p = HeadParser()
    p.feed(chunk)

    errors = []
    warnings = []

    title = p.title.strip()
    if not title:
        errors.append('missing <title>')
    elif not (10 <= len(title) <= 70):
        warnings.append(f'title length {len(title)} (recommended 10-70)')

    desc = p.metas.get('description', '')
    if not desc:
        errors.append('missing meta description')
    elif not (50 <= len(desc) <= 160):
        warnings.append(f'description length {len(desc)} (recommended 50-160)')

    # OG tags
    og_title = p.metas.get('og:title', '')
    og_desc = p.metas.get('og:description', '')
    og_image = p.metas.get('og:image', '')
    if not og_title:
        warnings.append('missing og:title')
    if not og_desc:
        warnings.append('missing og:description')
    if not og_image:
        warnings.append('missing og:image')

    # twitter
    twitter = p.metas.get('twitter:card', '')
    if not twitter:
        warnings.append('missing twitter:card')

    return title, errors, warnings


def main():
    base = Path(__file__).resolve().parent.parent
    htmls = sorted(base.glob('*.html'))
    if not htmls:
        print('no HTML files found in site/ to validate')
        return 1

    had_errors = False
    for f in htmls:
        title, errors, warnings = check_file(f)
        if errors or warnings:
            print(f'== {f.name} ==')
            if title:
                print(' title:', title)
            for e in errors:
                print(' ERROR:', e)
            for w in warnings:
                print(' WARN :', w)
            print()
        had_errors = had_errors or bool(errors)

    if had_errors:
        print('SEO validation failed: required meta tags missing')
        return 2
    print('SEO validation passed (no required errors). Warnings may still exist.')
    return 0


if __name__ == '__main__':
    raise SystemExit(main())
