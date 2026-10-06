# Changelog

## Unreleased

### Added

- Extended-mode border and edge colors: `classDef` label/border colors and
  source-indexed or default `linkStyle` stroke/label colors. Border attachments
  retain box colors; color composition does not change layout or glyphs.
- Independent minimum box padding in extended output: `--boxPaddingX` and
  `--boxPaddingY`, defaulting to the existing border padding of one cell.
  Vertical padding of zero removes blank rows in simple boxes.
- Opt-in `--uniformSpacing` equalizes visible gaps between node bands using the
  largest required gap on each axis. It does not equalize box sizes.
- Matching Go API, WASM JSON, demo controls, and shared URL settings for box
  padding and uniform spacing.
- Demo control tests and a smoke test of the built WASM renderer through the
  worker handler, also run by the Pages workflow.

### Fixed

All renderer repairs below apply to extended/Unicode output only.

- Measure and draw Unicode labels by terminal display cells, preserving CJK,
  combining marks, and emoji (Fix 1).
- Compose junctions from connectivity and reserve label/arrowhead clearance
  (Fix 2–3).
- Keep self-loop routes outside boxes and preserve heads on opposed edges
  (Fix 4–5).
- Preserve repeated edge labels and heads using separate routes (Fix 6).
- Reserve nested group/title clearance and separate interleaved group members
  into lanes without changing primary graph ranks (Fix 7–8).
- Reserve drawable cells at bent endpoints and assign distinct incident
  arrowhead ports where the existing port model permits (Fix 12–13).
- Keep labelled self-loop attachments on borders with compact or default box
  padding (Fix 14).
- Preserve spaces in arrow labels instead of exposing underlying line glyphs
  between words (Fix 15).

### Compatibility and validation

- ASCII output and its test goldens remain unchanged. Box padding overrides and
  uniform spacing are ignored in ASCII mode; uniform spacing defaults to off.
- Go tests, vet, demo control tests, and TinyGo/standard-Go WASM worker checks
  pass locally. Interactive browser inspection and a new remote Pages deployment
  have not been verified.
- These fixes retain the existing grid, rank assignment, and finite cardinal
  ports. See the [renderer audit](automation/render-audit.md) for regression
  coverage and remaining scope limits.
