# Visual Design (draft)

Purpose: define a clean, fitness-oriented visual language and a small set of UI components to keep the static site consistent and easy to implement.

## Tone & audience
- Tone: confident, clear, slightly technical — target audience is self-hosters: devs, gym owners, coaches, CrossFit enthusiasts who want control/privacy.
- Voice: benefit-first, concise technical details available via Deploy/Docs pages.

## Brand basics
- Name: ActaLog (use as display name unless you prefer another)
- Logo: (ask: do you already have a logo or mark?)

## Color palette (proposal)
- Primary: #0f172a (very dark navy) — headings, nav
- Accent: #ef4444 (vibrant red) — CTAs, highlights (CrossFit association)
- Secondary: #06b6d4 (teal) — badges/links
- Background: #ffffff / light neutral
- Muted text: #6b7280

## Typography
- Headings: Inter / system sans-family (bold)
- Body: Inter / system sans-family (regular)
- Sizes: clear scale — H1 40–48px, H2 28–32px, body 16px

## Layout & components
- Header: repo link (GitHub), primary CTA (Deploy), secondary CTA (Docs)
- Hero block: left text, right screenshot card (or centered on small screens)
- Feature cards: icon + short headline + 1–2 sentence blurb
- Code snippet blocks: monospace styling with copy-to-clipboard
- Footer: license, links to docs, GitHub, donate/sponsor link if applicable

## Imagery & assets
- Use screenshots of the UI and import CSV example. Prefer real UI screenshots in `web/dev-dist`/`public` or captured from a running instance.
- Small set of simple SVG icons: deploy, docker, database, import, mobile (PWA)

## Accessibility
- Contrast: ensure 4.5:1 for body/callouts
- Keyboard focus styles and semantic HTML
- Mobile-first responsive behavior

## Responsive behavior
- Single-column stack on small screens
- Two-column hero on medium/large screens
- Images scale to device width, avoid horizontal scroll

## SEO & microdata
- Add standard meta tags, Open Graph and Twitter card
- Use JSON-LD for softwareProject schema (name, repo, license, keywords)


---

Questions for you (short):
1. Do you have an existing logo or color preference?
2. Who is the primary audience: "devs/self-hosters" or "coaches/gym owners" or both?
3. Any must-have screenshots or demo links I should include now?

