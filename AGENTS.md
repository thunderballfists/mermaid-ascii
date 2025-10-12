---
alwaysApply: true
---
# Project: mermaid-ascii — WASM demo & UI

## Hard Rules
- For every change, add/adjust tests; `go test ./...` must pass.
- Keep ASCII outputs byte-for-byte unchanged.
- Extended/Unicode changes affect `useAscii=false` only.
- Small diffs; commit after each passing milestone.

## Build & Test
- Unit tests: `go test ./...`
- WASM: `tinygo build -o dist/main.wasm -target wasm ./wasm`
- Dev site: copy `site/*` → `dist/`, copy `wasm_exec.js`, then `python3 -m http.server` in `dist/`.

## Target Layout
- pkg/lib/…        (Render API)
- wasm/…           (TinyGo wrapper; exports `mermaidAsciiRender`)
- site/{index.html,styles.css,app.js,examples.js,worker.js}
- dist/            (build output)
- .github/workflows/pages.yml  (Pages CI)

## DoD
- `go test ./...` green.
- `make wasm && make serve` (or commands above) yields a working demo.
- CI builds and deploys to Pages.
