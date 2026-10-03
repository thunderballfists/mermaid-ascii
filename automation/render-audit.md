# Legacy text rendering audit

Date: 2026-10-02. Baseline: `59b6368` (M9).

This is a characterization audit, not a rendering fix. Production code and all
ASCII/Unicode goldens remain unchanged. Passing audit tests mean the observations
below are reproducible; they do not mean the affected output is correct.

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
existing ASCII/extended goldens remain untouched. Junction, arrowhead, routing,
and tight-padding group/self-loop defects below are still open.

## Confirmed defects

| Case | Observation | Scope / cause |
| --- | --- | --- |
| `graph LR; A<-->B`, padding 1 | Unicode ends up with only `◄`; the right head is overwritten. | Both heads share one cell in `drawArrowHead`. ASCII draws only a forward head even at padding 5. |
| `A-->B; B-->A`, padding 1 | Only one head survives. | Both modes route opposite edges through the same one-cell gap. Padding 5 retains both heads but shares the route. |
| `graph TD; A-->|a long edge label|B`, padding 1 | Label overwrites the arrowhead. | Both modes. Label drawings are merged after heads in `graph.draw`; label-aware minimum vertical clearance is missing. |
| `A-->B; A-.->C; A==>D` | A dotted branch uses `└` on a double vertical trunk that continues above and below it. The glyph drops the upward connection. | Unicode only. `mergeJunctions` and `junctionChars` handle light glyphs, not mixed-weight connectivity. |
| `A-->|one|B; A-->|two|B` | Only `two` and one head remain. | Both modes. The paths are identical; later labels/heads overwrite earlier edges. |
| `A[Café]-->|é|B[東京]` | Labels become `CafÃ©`, `Ã©`, and byte-derived CJK mojibake; box widths are inflated. | Originally both modes; now ASCII only. Extended-mode measurement/drawing is corrected. |
| `A[é]-->|👩‍💻|B[🙂]` | Combining-mark and emoji labels are also corrupted. | Originally both modes; now ASCII only. Added during the spike to match its display-cell measurement probes. |
| Nested Outer/Inner groups, padding 1 | Group frames touch/cross node boxes; the external D node straddles Outer's frame. | Both modes. Node placement precedes group extents; group border clearance is not reserved. Padding 5 avoids this specific overlap. |
| `graph LR; A-->A`, padding 1 | Index-out-of-range panic. | Both modes. Path/canvas allocation fails at tight spacing; `increaseGridSizeForPath` allocates half-padding (zero at 1), while `drawLine` assumes drawable coordinates. Padding 5 succeeds. |
| `graph LR; A---B` | ASCII emits an arrowhead for an undirected edge. | ASCII's `drawArrowHead` does not consult `arrowDir`; Unicode respects it. |

Semicolons in the table separate lines for readability; the executable sources
use actual newlines. Head suppression applies only to Unicode; ASCII ignores the
head-style option, preserving the current compatibility behavior.

## Example: lost vertical head

Source: `graph TD` followed by `A-->|a long edge label|B`; padding 1.

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
