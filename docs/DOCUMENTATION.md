# Maintaining the documentation

Repository Markdown is the source of truth. The optional site renders those same pages and adds generated CLI/settings/interface references. Keep instructions useful directly on GitHub as well as in the site; no external publishing service is needed to read them.

## Where a change belongs

| Material | Source |
| --- | --- |
| Entry point | Root `README.md`. |
| Learning paths and guides | `docs/README.md` and `docs/*.md`. |
| Roadmap | Root `ROADMAP.md`. |
| Examples | `examples/`, indexed in `examples/README.md`. |
| Defaults and descriptions | Component `Environment()` data; `cmd/settings-doc` updates the generated section of `docs/CONFIGURATION.md`. |
| CLI reference | Actual CLI help, rendered by the generator. |
| Interface reference | Go documentation selected by `docs-site/scripts/generate.sh`. |
| Generated pages | `docs-site/generated/`; regenerate, do not edit by hand. |

The site builds ignored `docs-site/content/docs` and copies referenced repository images into ignored public output. Relative links become site routes or public repository files. Conversion does not edit the source Markdown.

## Write against implemented behavior

Check code and run tutorial commands. State when output is recorded, synthetic, historical or live. Keep identity separate from freshness, compatibility separate from quality, and findings separate from completion. Mark planned features explicitly.

Use reserved example domains and placeholders. Never copy private paths, infrastructure names, credentials or unrelated operational notes into public pages. A public-text guard catches known shapes; it cannot certify every value is publishable. Review the diff too.

## Refresh references

From the repository root:

```sh
go run ./cmd/settings-doc
sh docs-site/scripts/generate.sh docs-site/generated
node docs-site/scripts/build-content.mjs
node --test docs-site/scripts/*.test.mjs
node docs-site/scripts/guards.mjs
make check-links
```

Guards check conflict markers, page/image/anchor links, known private-text patterns and stale references. Staleness is compared with freshly generated output. Go/module dependencies must be available for generation. The GitHub checks run these content/guard steps on Linux and macOS as well as testing the scripts.

## Build and read the site locally

With Node and the committed lockfile:

```sh
npm ci --prefix docs-site
npm run build --prefix docs-site
npm run dev --prefix docs-site
```

Open the local address printed by the server. The production build also converts content and runs guards. Building does not publish it or change hosting. Inspect representative tutorial, table and reference pages when changing rendering/navigation.

Avoid relying on site-only components for an instruction GitHub readers need. Standard fenced commands and Markdown tables keep examples copyable in both places.

## Validate substantial changes

Run examples, including controls/failures where a guide makes an enforcement claim. Use isolated databases with clearly invented tutorial rows. Distinguish illustrative model settings from measured profiles. See [Contributing](../CONTRIBUTING.md) for repository checks and [Evaluation](EVALUATION.md) for quality claims.
