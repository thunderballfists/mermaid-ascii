# Legacy text rendering audit

Date: 2026-10-02. Baseline: `59b6368` (M9).

The original characterization audit left production code and all goldens
unchanged. The bounded corrections below describe subsequent repairs. Passing
audit tests still reproduce the remaining defects; they do not mean all affected
output is correct. ASCII outputs and goldens remain unchanged.

## Reproduce

```sh
go test ./cmd -run TestRenderAuditMatrix -count=1 -v
go test ./... -count=1
```

`cmd/render_audit_test.go` exercises 13 small sources in both character modes,
with padding 1 and 5 and pointer/open/wide/none heads: 208 configurations.
Straight horizontal/vertical arrows and incoming branches are controls. Other
cases cover undirected/bidirectional edges, mixed stems, cycles, self-loops,
parallel labels, Unicode/grapheme labels, long vertical labels, and nested groups.

The checks are deliberately bounded to this corpus: head counts, literal label
preservation, group-border/node intersections, a mixed junction's local
connectivity, and captured render panics. They are not a general graph verifier.
Known-defect assertions must be updated explicitly when an intentional fix lands;
do not regenerate the existing goldens to erase a discrepancy.

`TestRenderAuditTextUIEmbeddingDemo` additionally characterizes the three-node
Unicode demo from the [native embedding spike](textui-embedding-spike.md) in both
modes at padding 5: two checks outside the 208-configuration matrix. Its ASCII
label corruption remains reproducible; extended mode now retains the labels.
No goldens are added or changed.

## Unicode measurement correction

The first repair milestone fixes label measurement and drawing only when
`useAscii=false`. Node and edge labels use `github.com/rivo/uniseg` v0.4.7 for
grapheme segmentation and terminal-cell width. Wide graphemes reserve continuation
cells without emitting extra spaces; combining marks and joined emoji stay intact.
Colored nodes wrap whole graphemes rather than individual UTF-8 bytes. Group
titles use cell-based centering, whole-grapheme clipping, and title-aware width.

Regression tests cover accents, combining marks, CJK, emoji/flags, zero-width
text, labels merged over routes, colored labels, and long group titles. Audit
expectations remove only the extended-mode label-loss defects. ASCII remains
byte-for-byte compatible, including its existing Unicode-label defects, and all
existing ASCII/extended goldens remained untouched in that milestone. Remaining
arrowhead, routing, and tight-padding group/self-loop defects are tracked below.

## Unicode junction correction

The second repair milestone unions directional connections before choosing a
junction glyph, only when `useAscii=false`. Light, heavy, double, rounded, and
dotted/dashed palette glyphs are recognized. Mixed light/double axes retain their
weights: a dotted branch on a continuous double trunk now uses `╟`, not `└`.
Same-axis weight conflicts take the stronger weight. Other mixed-axis combinations
fall back to a uniform stronger-weight glyph without losing connections; this is
not a per-arm style compositor. Identical dotted/rounded glyphs stay unchanged.
Arrowheads, ANSI-styled cells, colors, and routing keep their overlay behavior.
Explicit edge-label and group-title layers bypass stroke merging so box-drawing
characters used as text remain literal. Source-border attachment weight policy
is unchanged.

Regression tests check both merge orders, directional union across every recognized
glyph pair, mixed weights, palette aliases, ASCII isolation, literal box-drawing
labels/titles, and the rendered mixed-branch fixture at padding 1 and 5. The audit
matrix no longer expects the mixed-junction defect and checks the exact glyph.
Only three junction cells in `unicode_styles_stems.txt` change: two `└` become
`╟`, and one continuing double-trunk branch changes from `╚` to `╠`. No ASCII
goldens are changed.

Validation: `go test ./... -count=1`, `go vet ./...`, and a standard-Go WASM build
with a Node runtime junction check pass. TinyGo remains blocked by the previously
observed missing `wasm-opt`; no fresh TinyGo demo or Pages deployment is claimed.

## Unicode clearance correction

The third repair milestone (2026-10-03) reserves drawable cells during mapping,
before node coordinates and drawings are finalized, only when `useAscii=false`.
Straight bidirectional edges with visible heads reserve at least two cells, so
tight horizontal/vertical gaps retain both heads. Vertical label segments with
visible heads reserve at least three cells, placing the label on a separate row
between the endpoints. Existing horizontal label-width sizing is retained.

Only the physical clearance deficit is added; larger requested padding remains
unchanged. Coordinates refer to band centers, so endpoint-band growth accounts
for its half-span contribution. Unlabeled single-head edges, undirected edges,
and suppressed heads do not acquire these new minimums. Head glyph selection,
route selection, and ASCII behavior are unchanged. All existing ASCII and Unicode
goldens remain untouched in this milestone.

Tests cover LR/TD bidirectional heads, all head styles, padding 1/2/5, forward/
backward/bidirectional/undirected vertical labels, terminal-width graphemes,
asymmetric padding, band-center arithmetic, repeated reservations, and literal
ASCII compatibility. Audit expectations remove only the corrected Unicode
bidirectional and vertical-label head-loss cases. Separate-edge cycles, parallel
routes, tight group framing, and self-loop allocation remain open.

Validation: `go test ./... -count=1`, `go vet ./...`, and standard-Go WASM/Node
runtime checks pass. TinyGo still fails because `wasm-opt` is missing; the demo
build and Pages deployment are not claimed as verified.

## Confirmed defects

Repair 4 reserves at least two cells for newly allocated Unicode outer path
bands. This fixes the padding-1 self-loop panic and padding-2/3 attachments that
were drawn inside the node. Routing and existing larger/default padding stay
unchanged. LR/TD tests cover padding 1/2/3/5, solid/thick/dotted/bidirectional/
undirected edges and all head styles, including node-interior preservation.
Existing ASCII and Unicode goldens are unchanged; the legacy ASCII padding-1
panic remains an explicit compatibility expectation.

Repair 5 extends the same two-head clearance to opposite directed edges sharing
a straight route. Both heads survive at padding 1 in LR/TD with every visible
head style; suppressed heads and larger/default padding retain their spacing.
No routing or golden outputs change. ASCII's one-cell head collision is retained.

| Case | Observation | Scope / cause |
| --- | --- | --- |
| `graph LR; A<-->B`, padding 1 | Originally Unicode retained only `◄` because both heads shared one cell. | Corrected in Unicode mode by minimum clearance. ASCII still draws only a forward head, including at padding 5. |
| `A-->B; B-->A`, padding 1 | Originally only one head survived. | Now ASCII only. Unicode reserves two drawable cells for opposed heads; unlabelled opposite edges may still share the straight route. |
| `graph TD; A-->|a long edge label|B`, padding 1 | Originally the label overwrote the arrowhead. | Now ASCII only. Unicode reserves vertical label/head clearance before drawing. |
| `A-->B; A-.->C; A==>D` | Originally a dotted branch used `└` on a double vertical trunk that continued above and below it, losing the downward arm. | Corrected in Unicode mode: connectivity merging selects `╟`. ASCII is unchanged. |
| `A-->|one|B; A-->|two|B` | Only `two` and one head remain. | Both modes. The paths are identical; later labels/heads overwrite earlier edges. |
| `A[Café]-->|é|B[東京]` | Labels become `CafÃ©`, `Ã©`, and byte-derived CJK mojibake; box widths are inflated. | Originally both modes; now ASCII only. Extended-mode measurement/drawing is corrected. |
| `A[é]-->|👩‍💻|B[🙂]` | Combining-mark and emoji labels are also corrupted. | Originally both modes; now ASCII only. Added during the spike to match its display-cell measurement probes. |
| Nested Outer/Inner groups, padding 1 | Group frames touch/cross node boxes; the external D node straddles Outer's frame. | Both modes. Node placement precedes group extents; group border clearance is not reserved. Padding 5 avoids this specific overlap. |
| `graph LR; A-->A`, padding 1 | Originally an index-out-of-range panic. | Now ASCII only. Unicode outer path bands have drawable clearance; padding-2/3 ports also stay on the border rather than entering the node. |
| `graph LR; A---B` | ASCII emits an arrowhead for an undirected edge. | ASCII's `drawArrowHead` does not consult `arrowDir`; Unicode respects it. |

Semicolons in the table separate lines for readability; the executable sources
use actual newlines. Head suppression applies only to Unicode; ASCII ignores the
head-style option, preserving the current compatibility behavior.

## Baseline example: lost vertical head

Source: `graph TD` followed by `A-->|a long edge label|B`; padding 1. This is the
original Unicode output; the clearance correction now separates the label and
head onto different rows.

```text
┌───────────────────┐
│                   │
│         A         │
│                   │
└─────────┬─────────┘
  a long edge label
┌───────────────────┐
│                   │
│         B         │
│                   │
└───────────────────┘
```

## Acceptance cases for the TypeScript feasibility spike

- Preserve label text and measure display cells rather than UTF-8 bytes.
- Keep both heads where two directed connections exist; reserve enough space.
- Preserve both parallel edge labels, with distinct routes.
- Reserve label and group-frame clearance before placement/routing.
- Derive junction glyphs from connectivity, not last-writer glyph replacement.
- Handle a self-loop without a panic.

The spike is exploratory, not an ASCII-compatible replacement. New layout may
differ; the production Go API/CLI and its existing ASCII snapshots stay intact.
TCSS behavior and a complete styling system are outside this probe; styles must
remain separate from the character grid when a production design is proposed.
