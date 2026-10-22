# FIX-004: The reading mark hung over the track at the two ends of the scale

| Field | Value |
|---|---|
| **ID** | FIX-004 |
| **Status** | Fixed |
| **Owner** | Platform team |
| **Severity** | S3 Medium |
| **Priority** | P2 |
| **Affects** | `F-RPT-01`, `F-RPT-02` |
| **Detected in** | Deployment `20261022-ui`, run `run-plan-demo-demo-good-3408113d`, 2026-10-22 |
| **Created** | 2026-10-22 |
| **Last Updated** | 2026-10-22 |

## Summary

The point mark that carries the reading on a scale was positioned against a one-pixel inset of the track, but the percentage of an absolutely positioned element is resolved against the padding box rather than the content box. The mark therefore hung two and a half pixels over the track at a reading of 1.0000, and a reader could not tell a reading at the end of the scale from a reading that ran off it.

## Severity and Impact

**Technical impact:** The track was drawn one pixel inside the element on each side, and the marks were placed with `left: calc(1px + var(--at))`. The percentage in that expression is a percentage of the padding box, which is the whole element, so the reading was placed on a scale two pixels wider than the track it was drawn on. The error grows with the width of the scale, and it is worst at the two ends, where the mark is furthest from the middle. On the report's main score the track is 788 pixels wide, and a reading of 1.0000 put the mark at 1371 to 1374 while the track ended at 1371.5.

**Blast radius:** Every scale in the interface at a reading at either end of the scale. The main score of the deployment's run is 1.0000, so the defect was visible on the report that the deployment published. The metric rows carry their interval on a 90-pixel track, where the same error is two pixels out of ninety. No number was wrong; the mark was drawn outside the scale that carries it.

**Detection:** Found during the verification of the deployed interface, when the report of a run with the score 1.0000 was measured and the mark was found outside the track. The earlier verification had measured the same check on a reading of 0.5000, in the middle of the scale, where the error is under half a pixel and was not visible.

## Symptom

On the report screen of the run with the score 1.0000, the point mark on the main score scale was measured at `1371.0–1374.0` while the track was measured at `583.5–1371.5`: the mark stood 2.5 pixels past the end of the scale.

## Reproduction

**Environment:** Deployment `20261022-ui` on the cloud host, run `run-plan-demo-demo-good-3408113d`, score 1.0000.

**Preconditions:** A reading at either end of a scale. The middle of the scale does not show the defect.

**Steps:**

1. Open the report screen of a run whose total score is 1.0000.
2. Read the bounding rectangle of the point mark and of the track it sits on.

**Expected:** The mark lies within the track, flush with its end at a reading of 1.0000 and with its start at a reading of 0.0000.

**Actual:** The mark lay 2.5 pixels past the end of the track.

## Root Cause

**Cause:** The percentage of an absolutely positioned element is resolved against the padding box of its containing block. The track was drawn with `padding: 0 1px` and `background-clip: content-box`, so the visible track was the content box, two pixels narrower than the padding box the marks were placed against. The one-pixel offset in `left` compensated for the inset on the left side and for nothing on the right side, and the percentage itself was scaled by the width of the wider box.

**Failure class:** Logic error

**Why it was not caught:** The check that the mark lies inside its track was run against a reading in the middle of the scale, where the mark sits near the middle and the two-pixel error is invisible at both edges. Nothing measured the same check at the two ends of the scale, and the scale is the one element of the design system whose correctness depends on the whole range of its input rather than on one point of it.

## Fix

**Change:** The track is now the positioning context of its marks: the element carries no padding, and the visible track is the element itself, so a percentage on a mark is a percentage of the track. The mark is a fixed width, so a reading at either end is clamped inside the track with `max(0px, min(...))`, which pulls the mark flush with the end instead of letting it hang over it. The zero-width interval band is clamped the same way, and the width of the mark is a property of the scale, so the large mark of the main score and the small mark of a metric row are clamped by the same rule.

**Rationale:** The defect was a mismatch between the box the track was drawn in and the box the marks were placed against. Making the track the positioning context removes the second box, and the clamp states the rule the design system already claims: a mark never leaves its track, so a reading of zero is distinguishable from a reading that ran off the end of the scale.

**Guardrail added:** The scale is measured at all three positions — 0.0000, 0.6667 and 1.0000 — in [TR-007][tr-007], so a regression at either end is measured rather than assumed.

## Verification

| Check | Method | Measured | Result |
|---|---|---|---|
| Original reproduction no longer reproduces | The report of the deployment's run, measured in the browser | mark `1368.5–1371.5`, track `583.5–1371.5` | **Pass** — the mark is flush with the end of the track |
| A reading of zero stays inside the track | The report of `run-plan-partial-partial-agent-7f4b3d8d`, measured in the browser | mark `583.5–586.5`, track `583.5–1371.5` | **Pass** — the mark is flush with the start of the track |
| A reading in the middle stays inside the track | The same report, measured in the browser | mark `1107.3–1110.3`, track `583.5–1371.5` | **Pass** |
| The interval band stays inside the track | The main score and all 6 metric rows, measured in the browser | the band of the main score at `1368.5–1371.5`; all 6 metric bands within their 90-pixel tracks | **Pass** |
| The threshold notch stays inside its track | The confidence gauge, measured in the browser | the notch at 69.7% of the track, inside it | **Pass** |
| Adjacent behaviour unaffected | The repository's gates | 18 frontend tests and the whole Go suite | **Pass** |

## Regression Prevention

- **Test added:** The scale geometry is measured at the two ends and in the middle of the scale in [TR-007][tr-007], which is a browser measurement rather than a unit test, because the defect is a box the browser resolves.
- **Run gate:** The `web-build` and `web-typecheck` gates cover the stylesheet's syntax, and the frontend test suite covers the components that draw the scale.
- **Detection change:** The claim in [TR-005][tr-005] that the mark lies inside its track is now stated for one reading only, and points at this document for the rest of the range.

## References

- Test Report: [TR-007 Scale Mark Geometry][tr-007]
- Related Design: [TD-006 Interface Visual Design System][td-006]
- Related Report: [TR-005 Interface Visual Design System][tr-005]

[tr-007]: <../../Test Report/TR-007-scale-mark-geometry.md>
[td-006]: <../../Tech Design/TD-006-interface-visual-design-system.md>
[tr-005]: <../../Test Report/TR-005-interface-visual-design-system.md>
