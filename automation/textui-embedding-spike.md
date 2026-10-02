# Native TextUI graph embedding spike

Date: 2026-10-02. Follow-up to the Mermaid renderer spike.

## Result

**Embedding works through TextUI's existing component registry.** A throwaway
`<graph src="workflow.mmd"/>` component mounts an existing netext `GraphView`,
attaches native Textual node widgets, applies native TCSS, and dispatches node
selection through TextUI's existing action mechanism. No TUI engine, stylesheet
parser, graph layout algorithm, router, or glyph compositor was implemented.

**netext 0.5.0 is not yet a full Mermaid flowchart backend for this use case.**
Parallel edges, self-loops, compound group frames, and native style-driven node
remeasurement remain unmet requirements. Do not promote this spike as a complete
component or fix these by quietly writing another graph renderer.

The experiment leaves TextUI's production source, package metadata, lockfile,
and environment untouched. It also leaves the Go renderer, site/WASM, and all
existing ASCII/Unicode goldens unchanged. Only findings and an additional Go
characterization test for the same demo source are committed here.

## Run the throwaway demo

Code, pinned Python requirements, tests, and captured output are outside both
repositories at `/tmp/textui-graph-spike.oSGN45`. These are temporary probe
artifacts, not a durable application or an installed TextUI component.

```sh
cd /tmp/textui-graph-spike.oSGN45
.venv/bin/python run.py
.venv/bin/python -m pytest test_embedding.py -q -rx
PYTHONPATH=/Users/abeihl/Development/TextUI \
  .venv/bin/python -m pytest test_embedding.py -q -rx
```

`run.py` defaults to `demo.mmd`; it also accepts another Mermaid source path.
Ctrl+Q exits. A PTY launch and Ctrl+Q shutdown completed with exit status 0.
`capture.py` produces `demo.svg` and `demo.txt` using Textual's own screenshot
and compositor facilities. The captured graph is:

```text
╭──────╮           ╭──────╮           ╭──────╮
│ Café │──────é───▶│ 東京 │──────────▶│ Done │
╰──────╯           ╰──────╯           ╰──────╯
```

Python 3.12.9; textui-markup 0.7.0; Textual 8.2.8; netext 0.5.0; Rich 15.0.0;
PyYAML 6.0.3. The same embedding tests also run against the local TextUI source
at commit `c85637b`, rather than relying solely on the published wheel.

The source adapter reuses Mermaid 12.1.0 in the earlier temporary Node environment
at `/tmp/mermaid-text-spike.KBhAze`. `export-graph.mjs` uses Mermaid's parser and
database; it does not invoke its renderer or layout. It still needs Node and a
DOM shim, and uses the version-pinned, non-stable database seam. This is not a
pure-Python Mermaid parser or a proposed subprocess-per-render production API.

## Verified behavior

- Registration uses the existing `ComponentSpec`, `AttributeSpec`, and
  `DocumentLoader` extension mechanism; no TextUI loader changes are needed.
- A native `GraphView` mounts under the declared component ID. Nodes are native
  focusable Label subclasses, not characters masquerading as widgets.
- Front matter is read with PyYAML. Title, custom metadata, and an embedded
  `tcss` string are retained. The stylesheet is passed to TextUI's existing
  native stylesheet integration; invalid declarations raise `DocumentStyleError`.
- Mermaid node IDs/classes survive into native widget IDs/classes. Colors and
  borders are verified as computed native TCSS styles, and nodes have nonzero
  rendered regions at their backend-computed coordinates.
- Viewport resizing from 80x24 to 120x40 updates the backend's view dimensions
  and keeps attached node widgets rendered. This is distinct from node resizing.
- Pilot clicks and Enter on a focused node dispatch exactly one TextUI action
  per interaction, with the graph component and selected node ID.
- Straight, bidirectional, cycle, Unicode/grapheme, long-label, and mixed-stem
  fixtures retain their tested labels/head counts within an 80x24 graph budget.
  Rich-backed measurement gives both `Café` and `東京` 8-cell boxes, including
  padding/borders. Mixed-stem checks are not a general junction-connectivity proof.

## Compatibility adaptations

An unadapted first mount crashed in Textual's monochrome filter because netext
returned Rich segments with no style. The wrapper applies the native widget's
visual style using Textual's `Strip.apply_style`; it does not repaint the graph.

netext's node attachment API mounted children, but Textual's `ScrollView`
reported `is_container=False`, leaving those children with zero-sized regions.
The wrapper declares itself a container and makes attached widgets native
absolute-positioned children instead of left-docked children. Tests verify their
actual regions and backend-coordinate alignment, not just their presence in the
widget tree. These overrides require further compatibility coverage before a
production integration, particularly scrolling, clipping, and theme changes.

## Remaining failures

The final suite has **17 passes and 4 strict expected failures**. All four first
failed as ordinary acceptance probes; marking them expected records limitations,
not successful full-Mermaid acceptance.

| Requirement | Observed failure |
| --- | --- |
| Parallel `one`/`two` connections | Only `two` and one head render. netext stores edge buffers by endpoint pair. |
| `A-->A` self-loop | The box renders, but the loop and its head do not. |
| Nested Outer/Inner groups | Group labels become independent ordinary nodes; child nodes are not enclosed by compound frames. Group IDs/parents were retained for this containment check. |
| Native node width/padding changes | Textual's node grows to 24 columns while the backend still measures it as 8. Routes are not recomputed from the widget's native geometry. |

The demo component explicitly rejects parallel edges, self-loops, and subgraphs
instead of silently presenting those wrong results. It does not solve arbitrary
TCSS-driven sizing. Only simple boxed LR/TD flowcharts were evaluated; arbitrary
node components, shape fidelity, full Mermaid styling, live source reload,
other diagram types, and untrusted-input security were not established.

## Recommendation

Keep the integration-first direction: a native TextUI graph component, an input
adapter, and an existing graph backend. Treat the graph backend as replaceable
and preferably optional rather than making every core TextUI installation depend
on it. Mermaid should be an input format, not the widget's identity.

Before designing a production API, resolve the four failed requirements by
evaluating backend capabilities/upstream support. If intentionally shipping a
limited first version, explicitly agree on the unsupported features and sizing
contract. Do not introduce custom routing/group composition as an unnoticed
extension of this spike. Native TCSS and node-widget composition should remain
delegated to Textual/TextUI.

## Verification

- Embedding suite: 17 passed, 4 strict expected failures, against both the
  published TextUI wheel and local checkout.
- Existing local TextUI extension/style tests: 25 passed in the isolated
  environment. The full TextUI suite and other Python/platform versions were
  not run for this throwaway probe.
- `go test ./... -count=1`, `go vet ./...`, and `git diff --check` pass.
- Existing ASCII/Unicode goldens have no diff against the pre-audit baseline.
- The native PTY demo launches and exits normally; no fresh WASM/Pages
  acceptance is claimed. The previously recorded missing `wasm-opt` remains
  outside this spike's scope.

Upstream references: [netext GraphView](https://github.com/mahrz24/netext/blob/main/netext/textual_widget/widget.py),
[graph data/backend](https://github.com/mahrz24/netext/blob/main/netext/console_graph.py).
