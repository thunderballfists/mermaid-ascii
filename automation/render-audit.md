# Legacy text rendering audit

Date: 2026-10-02. Baseline: `59b6368` (M9).

The original characterization audit left production code and all goldens
unchanged. The bounded corrections below describe subsequent repairs. Passing
audit tests retain ASCII compatibility defects; they are not a claim that every
possible graph is correct. ASCII outputs and goldens remain unchanged.

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

Repair 6 separates repeated connections with the existing A* search and cardinal
ports. Unique node pairs are routed first so additional edges can avoid occupied
head/label ports, and prior routes receive a traversal penalty rather than being
blocked. Repeated searches remain in the nonnegative grid. Actual drawable spans
select label segments; content/gap bands, not structural border bands, expand.
After mapping, labels reserve node/head/cap and prior-label footprints, trying
other positions on their own route or adding clearance before final coordinates.
Stored label positions follow subgraph offsets. ASCII uses the original pipeline.
Label origins are calculated before rendering, so rejected negative positions
cannot panic. Bent alternate routes reserve drawable endpoint cells even in
existing gap bands. Repeated pairs with saturated corner nodes receive left/top
grid margins before routing, making the otherwise unavailable cardinal ports usable.

Tests cover two/three labelled parallel edges, labelled opposed edges, LR/TD,
all head styles, branching and grouped diagrams, and node-interior preservation.
Only two affected Unicode goldens change: `unicode_heads_basic.txt` and
`edge_styles_unicode_basic.txt` now retain the four requested heads instead of
collapsing multiple edges to two. Their rectangular trailing blank cells are
intentional. The existing finite cardinal-port model is retained; this is not
an unrestricted multi-port routing engine.

Repair 7 reserves group margins in content/gap bands before drawing coordinates
are finalized. Adjacent aligned nodes entering/leaving nested groups account for
every enclosing frame and its title overhead. Long titles expand a member's
content band before placement, rather than extending the frame over an external
node. Group-title footprints join label collision checks, so newly enlarged
vertical gaps cannot place an edge label beneath a title. The legacy frame-spacing
fallback remains for layouts that these bounded reservations do not separate.
Grouped incident edges whose visible heads share one physical port use the same
alternate-route search, avoiding collisions exposed by the corrected gaps. The
Unicode canvas includes group offsets and is resized after translation, retaining
outer routes instead of relying on node/frame drawing to grow it implicitly.

Tests cover LR/TD, padding 1/2/5, nested/sibling/disconnected groups, nested-only
parents, long CJK titles, labelled nested-group transitions, shared headed ports,
outer routes after group offsets, child/member
containment, external-node exclusion, and node-interior preservation. The 104
Unicode configurations in the original audit matrix now have no expected defects.
All existing ASCII goldens remain unchanged; repair 7 changes no goldens.
Interleaved groups whose members already surround unrelated nodes were the next
confirmed limitation of the original node-placement algorithm.

## Unicode group-lane correction (2026-10-05)

Repair 8 detects member envelopes containing foreign nodes and intersecting
unrelated group envelopes after the original grid mapping. Only those diagrams
receive secondary-axis lanes, before band sizing and routing. Primary graph ranks
are unchanged. The existing group tree supplies deepest node ownership; direct
nodes retain their relative order at each rank, while child groups receive their
own contiguous lane blocks. The original grid is rebuilt for the existing router.

Existing group-clearance reservations cover lane boundaries even when the nearest
members lie at different ranks. Bent edges in reassigned diagrams reserve drawable
endpoint cells, including existing narrow bands, so attachments cannot fall back
inside node boxes. Already noninterleaved layouts take the original
mapping path, and ASCII never enters this correction. No backend, dependency,
public API, source-file move, or golden change is introduced.

Regression tests cover LR/TD and padding 1/2/5 for sibling and nested interleaving,
external nodes, labelled cross-group edges, member/external-node containment,
child containment, unrelated-frame separation, primary-rank preservation, grid
occupancy, idempotence, and ASCII isolation. This is not an optimal cluster layout
or an unrestricted-port router; it keeps the existing primary-rank assignment.

Independent review compared 240 grouped diagrams against `bf4272b`: no new
interior intrusions, lost labels/titles, reduced head counts, panics, membership
failures, or sibling-frame overlap were observed. Four baseline interior defects
in diagrams not needing lane separation remain outside this repair. Unit tests,
vet, standard-Go/TinyGo WASM builds, and six LR/TD/padding WASM containment checks
pass. No ASCII or Unicode goldens change, and no new Pages or interactive-browser
validation is claimed.

| Case | Observation | Scope / cause |
| --- | --- | --- |
| `graph LR; A<-->B`, padding 1 | Originally Unicode retained only `◄` because both heads shared one cell. | Corrected in Unicode mode by minimum clearance. ASCII still draws only a forward head, including at padding 5. |
| `A-->B; B-->A`, padding 1 | Originally only one head survived. | Now ASCII only. Unicode reserves two drawable cells for opposed heads; unlabelled opposite edges may still share the straight route. |
| `graph TD; A-->|a long edge label|B`, padding 1 | Originally the label overwrote the arrowhead. | Now ASCII only. Unicode reserves vertical label/head clearance before drawing. |
| `A-->B; A-.->C; A==>D` | Originally a dotted branch used `└` on a double vertical trunk that continued above and below it, losing the downward arm. | Corrected in Unicode mode: connectivity merging selects `╟`. ASCII is unchanged. |
| `A-->|one|B; A-->|two|B` | Originally only `two` and one head remained. | Now ASCII only. Unicode repeated edges have separate routes and collision-aware labels. |
| `A[Café]-->|é|B[東京]` | Labels become `CafÃ©`, `Ã©`, and byte-derived CJK mojibake; box widths are inflated. | Originally both modes; now ASCII only. Extended-mode measurement/drawing is corrected. |
| `A[é]-->|👩‍💻|B[🙂]` | Combining-mark and emoji labels are also corrupted. | Originally both modes; now ASCII only. Added during the spike to match its display-cell measurement probes. |
| Nested Outer/Inner groups, padding 1 | Originally frames touched/crossed node boxes and external D straddled Outer's frame. | Now ASCII only in this fixture. Unicode reserves nested-frame clearance before coordinates are finalized. |
| `graph LR; A-->A`, padding 1 | Originally an index-out-of-range panic. | Now ASCII only. Unicode outer path bands have drawable clearance; padding-2/3 ports also stay on the border rather than entering the node. |
| `graph LR; A---B` | ASCII emits an arrowhead for an undirected edge. | ASCII's `drawArrowHead` does not consult `arrowDir`; Unicode respects it. |

Semicolons in the table separate lines for readability; the executable sources
use actual newlines. Head suppression applies only to Unicode; ASCII ignores the
head-style option, preserving the current compatibility behavior.

## Opt-in box padding and uniform gaps (2026-10-05)

`--boxPaddingX` and `--boxPaddingY` control independent minimum padding inside
Unicode boxes. Defaults remain `(1, 1)`; omitted axes retain legacy `-p` behavior.
Setting Y to zero removes blank rows in simple boxes. Shared grid bands and
routing may enlarge a box. Zero-width labels retain a content cell so arrows
cannot overwrite them on a border.

`--uniformSpacing` expands visible gaps between adjacent node bands to the
largest required gap on each axis. It includes extra space in trailing border
bands and reruns after label placement enlarges a band. It equalizes gaps, not
box sizes or center-to-center pitch. Default spacing remains adaptive.

The Go API, WASM JSON options, and demo expose all three settings. Demo URLs
retain box padding as `bx`/`by` and uniform spacing as `u=1`. ASCII ignores these
options; its original goldens remain unchanged.

Validation: `go test ./...`, `go vet ./...`, `make test`, TinyGo and standard Go
WASM builds, and `node automation/wasm-smoke.cjs` pass locally. The smoke executes
the built renderer through the actual worker handler with Node worker API shims.
Five demo tests cover dispatch, URL round trips, zero values, mode toggling, and
reset. The local demo responds on port 5174. The Pages workflow now runs unit,
demo, and built-WASM checks; no remote build or deployment is claimed.

The follow-up repair milestones address all three cases found by broader probes:

- Fix 12 reserves a drawable cell outside every bent Unicode endpoint, preventing
  short segments from writing junctions inside boxes.
- Fix 13 assigns distinct incident arrowhead ports and reserves origin margins
  when a root needs another port. `A<-->B; B-->C; D-->C` now retains all four heads
  and D's label. Three Unicode snapshots change to preserve previously lost heads.
- Fix 14 applies physical label clearance and safe placement to labelled
  self-loops. `A-->|loop|A; A-->B; B-->C` now keeps the label and ports on borders
  with box padding `(0, 0)` as well as default padding.
- Fix 15 composites complete Unicode arrow-label spans, including spaces, so
  labels no longer show underlying dashes between words.

New regression matrices check both directions, small/default gaps, compact and
default boxes, all four head styles, and uniform/adaptive spacing. They assert
label/head preservation, node interiors, actual port locations, and equal visible
gaps where requested. The built-WASM worker smoke includes the repaired branch,
self-loop, and spaced-label cases. ASCII snapshots remain byte-for-byte unchanged.
These checks cover the reported repros; they do not establish correctness for
every possible graph or remove the existing limits on available ports.

## Earlier validation (2026-10-03)

`go test ./... -count=1`, `go vet ./...`, standard-Go WASM builds, and TinyGo 0.39.0
builds pass. The earlier missing `wasm-opt` blocker was resolved for validation
using a checksum-verified official Binaryen 133 archive under `/tmp`, without
changing repository dependencies or installing system packages:

```sh
WASMOPT=/tmp/mermaid-binaryen/binaryen-version_133/bin/wasm-opt make wasm
WASMOPT=/tmp/mermaid-binaryen/binaryen-version_133/bin/wasm-opt make serve
```

TinyGo runtime checks preserve all three parallel labels/heads, nested-group
labels, and exact ASCII output. The site's actual worker also passes 25 HTTP/WASM
checks against the local server across LR/TD and padding 1/2/5. Independent review
compared 300 repeated/cyclic and 240 grouped diagrams against their preceding
milestones, finding no newly lost labels/titles, reduced head counts, or panics.
Those samples do not establish correctness for arbitrary graphs.
Interactive browser inspection was blocked:
no browser connector was available and native Brave access was not approved.
The remote Pages workflow returned no runs, and the Pages API returned 404;
no deployment or push is claimed. Local passing builds do not establish Pages
deployment. Global layout optimization, unlimited ports, and TCSS integration remain
outside this renderer-repair series.

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
