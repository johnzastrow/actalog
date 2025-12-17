# Publishing this folder as a GitHub Pages site

This `site/` folder contains a small static marketing site for ActaLog. To publish it on GitHub Pages, choose one of the approaches below.

Option A — Publish from the `gh-pages` branch (recommended for custom folder):

1. Build or prepare the files in `site/` (already present in the repo).
2. Push `site/` contents to the `gh-pages` branch (example):

```bash
# from repository root
git subtree split -P site -b gh-pages
git push origin gh-pages --force
```

3. In the repository Settings → Pages, set Source to `gh-pages` branch (root).

Option B — Use GitHub Actions to deploy `site/` to `gh-pages` automatically

- Create a workflow that copies `site/` to the `gh-pages` branch or deploys to Pages. This is useful if you want CI-driven publishing.

Option C — Use `docs/` directory instead

- Rename or move `site/` → `docs/` and select `main` branch and `/docs` folder in Pages settings.

Notes
- If you want to use a custom domain, create a `CNAME` file at the root of `site/` containing the domain name.
- A `.nojekyll` file is already included to ensure files starting with an underscore and other static assets are not processed by Jekyll.

If you'd like, I can add a GitHub Actions workflow to automatically publish `site/` to `gh-pages` on push.
