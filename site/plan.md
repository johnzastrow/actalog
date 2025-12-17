# Static Site Plan (draft)

Goal: produce a concise, SEO-friendly static site hosted from the repository `site/` directory (GitHub Pages) that advertises ActaLog and helps users self-host and contribute.

## Primary objectives
- Communicate what ActaLog is and who it's for.
- Show key features and technical requirements for self-hosting.
- Provide clear deployment instructions and links to docs/repos.
- Improve discoverability with focused SEO and marketing copy.

## Top-level pages (recommended)
- Home — hero, one-line pitch, three main benefits, CTAs (Try / Deploy / Docs)
- Features — short blocks for: logging, templates, WODs, imports, seeds, analytics, PWA
- Deploy / Install — Docker Compose, Postgres/MariaDB, environment variables, quick start
- Docs & API — links to repository docs, migration notes, seeds
- Demo / Screenshots — screenshots, sample CSV import, example WODs
- Community / Contribute — GitHub, issues, how to contribute, license
- Changelog / Releases — link to release notes and db_versions

## SEO & metadata
- Primary focus keywords (from `site/keywords.md`) — pick top 8 for meta tags.
- Title format: "ActaLog — Self‑hostable workout logging (Open source)"
- Meta description: 155–160 chars, include "self‑hostable", "workout log", "WOD", and "Docker".
- Open Graph: hero screenshot, short description, repo link.

## Content outline (Home)
- Hero: 1-line problem + solution: "Self-host your workout logs — open source ActaLog"
- 3 benefits: privacy & control, easy deployment (Docker), fitness-first features (WODs, templates)
- Quick deploy CTA with link to Deploy page and GitHub repo
- Social proof / community links

## Technical notes for site build
- Static generator: simple handcrafted HTML+CSS + small JS (Vite is already in repo; consider a minimal build step) or plain HTML to keep Pages simple
- Use responsive layout, accessible HTML, semantic headings
- Host under GitHub Pages from `site/` branch/directory (already specified)

## Timeline & deliverables (proposed)
- Draft plan & design (this step) — done
- Produce content pages & assets — draft text, screenshots
- Implement templates & deploy config — create minimal build or direct HTML
- QA, SEO check, launch

## Next immediate steps
1. Finalize keyword priority and hero messaging.
2. Choose visual direction (see `site/design.md`).
3. Collect screenshots / logos.

