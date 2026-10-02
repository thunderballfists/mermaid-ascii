# Mermaid text-renderer feasibility spike

Date: 2026-10-02. Audit commit: `92f9d34`.

## Decision

**A TypeScript text renderer is feasible; a drop-in alternate Mermaid renderer
is not the path I recommend.** Reuse Mermaid's flowchart parsing through a small,
version-pinned adapter. Own cell measurement, layout options, orthogonal routing
constraints, and painting; evaluate ELK directly as the layout engine. Do not
round an SVG into characters or carry a fork of Mermaid's SVG composition code.

This is a feasibility result, not an approved production redesign. The Go
renderer, public API, dependencies, WASM/site files, and existing goldens are
unchanged. No legacy defect is fixed by these commits.

## Probe and reproduction

Throwaway code and its npm lockfile remain outside the repository at
`/tmp/mermaid-text-spike.KBhAze`. They are evidence for this experiment, not code
to promote into production. The temporary directory is not a durable artifact.

```sh
cd /tmp/mermaid-text-spike.KBhAze
npm test
npx tsc --noEmit --target esnext --module nodenext \
  --allowImportingTsExtensions --skipLibCheck spike.ts spike.test.ts
```

Observed versions: Node 26.9.0, Mermaid 12.1.0, Mermaid's resolved ELK 0.9.3,
happy-dom 20.14.5, string-width 8.1.0, TypeScript 7.0.2, @types/node 26.6.4.

The adapter calls `mermaid.parse` first to apply front-matter configuration, then
`mermaidAPI.getDiagramFromText(...).db.getData()`. It measures labels in display
cells, supplies node/edge/group-label dimensions, runs Mermaid's ELK core, and
paints a character grid. SVG measurement methods throw if called. Parsing uses
a DOM shim for Mermaid's sanitization: this is **not** proof of a DOM-free worker
implementation.

Two local copies of installed Mermaid bundles expose otherwise private exports:
`preprocessDiagram`, `prepareLayoutForElk`, and `runElkLayoutCore`. Only their
export lists change; no upstream layout algorithm is edited. Their chunk names,
database shape, measured-label fields, and helper context are not stable public
contracts. The prototype's declarations explicitly leave this private seam
loosely typed; passing the type checker does not establish API stability.

## Observations

| Check | Result |
| --- | --- |
| Front-matter title and `config.flowchart.nodeSpacing` | Preserved. The config value survives; the prototype uses its own explicit cell spacing rather than promising equivalent SVG spacing semantics. |
| Accented/CJK labels | Literal node and edge labels survive rendering. |
| Combining marks and ZWJ emoji | Measurement returns 1 cell for `é`, 2 for `👩‍💻`; CJK `東京` is 4. Terminal font/width differences still need real-terminal checks. |
| Bidirectional edge and two-edge cycle | Both heads survive, adjacent to their respective node borders. |
| Parallel `one`/`two` labels | Both labels and heads survive with separate raw ELK routes. Label placement is still visually crude. |
| Self-loop | Renders without a panic and attaches to the node. |
| Long vertical edge label | Label and arrowhead survive without an ownership collision. |
| Nested groups | Child boxes clear parent frames in this fixture, but the result is **211 columns × 121 rows** for four nodes: unacceptable terminal composition. |
| Mermaid classes / CSS declarations | Class identity remains separate from characters. Font/color declarations do not alter this prototype's grid; this is not CSS/TCSS rendering support. |
| Junctions | Direction-bit masks retain all connections. Stems are intentionally normalized to light Unicode: mixed stroke weights, dotted junctions, and other head styles are not implemented. |

The first geometry run passed 5/10 tests. Four fixtures failed because Mermaid's
post-layout SVG marker/intersection adjustments produced diagonal terminal
segments at cell scale. The front-matter title also needed separate extraction:
`getDiagramFromText` alone did not populate it in this probe.

Using the core's **raw ELK sections**, with hierarchy offsets, instead of its
SVG-adjusted `data.edges[].points` avoided those diagonal segments. This is a
deliberate bypass, not evidence that the whole Mermaid renderer is reusable.
The final 12 tests pass, including a characterization test that asserts the
known oversized nested result. They check label/head preservation, head
adjacency, node-interior route collisions, child-frame clearance, connectivity,
and forbidden SVG measurements on this small corpus. They are not a general
collision-free routing proof or production acceptance suite.

## Front matter and TCSS

Mermaid's preprocessing keeps its recognized title/config fields but does not
surface an arbitrary top-level `tcss` field through the probed result. A source
containing that field parsed, but `data.config.tcss` was absent. A future adapter
needs its own outer front-matter document model if custom metadata/styles must
be preserved. Do not discard those fields and claim front-matter support.

TCSS syntax, cascade, validation, and styling are **not implemented** here.
Retaining node IDs/classes is a useful selector seam, not a TCSS engine. A
production design must distinguish full Textual TCSS compatibility from a
documented TCSS-like subset. Style spans belong outside the character grid;
measurement-affecting styles must be resolved before placement/routing.

## Next boundary

If proceeding with TypeScript, first design a flowchart-only adapter and a
cell-native intermediate graph, then test a direct ELK adapter with terminal
spacing/ports and the audit fixtures. Keep the Go ASCII renderer as the
compatibility path. This spike does not establish an ASCII-compatible replacement,
other diagram types, browser-worker packaging, security equivalence, or complete
TCSS support.

## Verification

- `go test ./... -count=1`, `go vet ./...`, and `git diff --check` pass.
- The Go audit now covers 208 configurations, including combining-mark/emoji
  labels matching the spike's width probes. Known legacy corruption is
  characterized, not repaired. ASCII and Unicode goldens have no diff.
- The temporary spike's 12 tests and TypeScript check pass.
- A fresh `tinygo build -o /tmp/mermaid-spike-verification.wasm -target wasm
  ./wasm` fails with `could not find wasm-opt, set the WASMOPT environment
  variable to override`. No fresh WASM/demo acceptance is claimed; the existing
  distribution and toolchain were left alone.

Upstream references: [ELK core and SVG adjustments](https://github.com/mermaid-js/mermaid/blob/mermaid%4012.1.0/packages/mermaid/src/rendering-util/layout-algorithms/elk/render.ts),
[shared SVG-oriented renderer](https://github.com/mermaid-js/mermaid/blob/mermaid%4012.1.0/packages/mermaid/src/rendering-util/layout-algorithms/common/index.ts),
[parsing/preprocessing API](https://github.com/mermaid-js/mermaid/blob/mermaid%4012.1.0/packages/mermaid/src/mermaidAPI.ts).
