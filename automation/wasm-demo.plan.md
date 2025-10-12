# Milestones (execute in order; commit after each)

M1: Library API
- Add pkg/lib/lib.go (Options{UseASCII,PaddingX,PaddingY,Heads?}; Render(string,Options)(string,error))
- Wire cmd/root.go to use lib.Render; no behavior change.
Acceptance: go test ./... (ASCII goldens unchanged)

M2: TinyGo WASM
- Add wasm/wasm.go (//go:build js && wasm). Export mermaidAsciiRender(mermaid, optsJSON) via syscall/js.
Acceptance: tinygo build -o dist/main.wasm -target wasm ./wasm

M3: New static site (no template reuse)
- Add site/index.html, site/styles.css, site/app.js; textarea+controls (Extended, PaddingX/Y, Render/Clear/Copy/Download).
- Load wasm_exec.js + main.wasm; call mermaidAsciiRender.
Acceptance: local static server -> render “graph LR\nA-->B”

M4: GitHub Pages CI
- Add .github/workflows/pages.yml:
  - checkout
  - setup TinyGo
  - build wasm → dist/main.wasm
  - copy wasm_exec.js + site/* → dist/
  - `touch dist/.nojekyll`
  - upload + deploy pages
Acceptance: Actions green; Pages is live.

M5: Head styles in UI
- Options.Heads ("pointer"|"open"|"none"); default pointer.
- wasm: parse Heads from opts JSON.
- UI: <select id="heads"> wired.
Acceptance: switching heads affects Unicode heads; tests green.

M6: Examples + permalinks
- site/examples.js; 3–4 presets.
- URL params: m, x, y, ext, heads. Restore on load; update on render.
Acceptance: shared URL restores state.

M7: Web Worker + debounce
- site/worker.js loads WASM; handle {mermaid,opts} → {text|error}.
- Debounce 300ms; ignore stale responses.
Acceptance: UI stays responsive.

M8: Dark mode, a11y, shortcuts
- Polished styles; prefers-color-scheme.
- Ctrl/Cmd+Enter render; Ctrl/Cmd+S download; Esc focuses editor.
Acceptance: manual smoke; no console errors.

M9: Makefile
- `wasm`, `serve`, `test` targets.
Acceptance: make wasm && make serve works.
