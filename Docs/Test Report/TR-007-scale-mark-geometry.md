# TR-007: Scale Mark Geometry

| Field | Value |
|---|---|
| **ID** | TR-007 |
| **Status** | Verified |
| **Verifies** | [FIX-004 Scale Mark Overhung the Track at the Ends of the Scale][fix-004] |
| **Owner** | Platform team |
| **Created** | 2026-10-22 |
| **Last Updated** | 2026-10-22 |

> This report records what was checked, how it was checked and what the check returned. The defect it verifies is specified in [FIX-004][fix-004].

## 1. Scope

The verification covers the geometry of the measurement primitives of the interface: the track, the interval band, the point mark that carries a reading, the gauge fill and the threshold notch.

The claim under verification is one the design system makes about itself: a mark never leaves the track it sits on, so that a reading at either end of a scale cannot be mistaken for a reading that ran off the end.

## 2. Method

| Aspect | How it was checked |
|---|---|
| Rendered geometry | The Chromium instance attached to BrowserOS neo, reading `getBoundingClientRect()` on the live DOM |
| The whole range of the scale | Three reports, read at 0.0000, 0.6667 and 1.0000 |
| The deployment | The same measurement against the running platform on the host |

The check is a measurement rather than a unit test, because the defect was a box the browser resolves: a percentage of an absolutely positioned element is a percentage of the padding box, and the defect existed only because the track was drawn in a different box from the one the marks were placed against.

## 3. The Evidence Set

| Report | Reading | Where it was measured |
|---|---|---|
| `run-plan-partial-partial-agent-7f4b3d8d` | 0.0000 | The local preview server, which holds the report |
| `run-plan-demo-demo-good-3408113d` | 0.6667 | The local preview server |
| `run-plan-demo-demo-good-3408113d` | 1.0000 | The deployment on the host |

The local preview holds a copy of both reports so that the two ends of the scale and its middle are all in one run.

## 4. Checks

### 4.1 The Point Mark

| # | Reading | Expected | Measured | Result |
|---|---|---|---|---|
| 1 | 0.0000 | the mark inside the track, flush with its start | mark `583.5–586.5`, track `583.5–1371.5` | Pass |
| 2 | 0.6667 | the mark inside the track | mark `1107.3–1110.3`, track `583.5–1371.5` | Pass |
| 3 | 1.0000 | the mark inside the track, flush with its end | mark `1368.5–1371.5`, track `583.5–1371.5` | Pass |

Before the fix, check 3 measured mark `1371.0–1374.0` against track `583.5–1371.5`: the mark stood 2.5 pixels past the end of the scale. Check 1 was measured after the fix and the mark is flush with the start of the track rather than overhanging it.

### 4.2 The Interval Band

| # | Check | Expected | Measured | Result |
|---|---|---|---|---|
| 4 | The band of the main score lies inside its track | within the track | band `1368.5–1371.5`, track `583.5–1371.5` | Pass |
| 5 | Every band of the metric table lies inside its track | within the track | 6 bands, all within their 90-pixel tracks | Pass |
| 6 | A band at the end of the scale does not overhang it | flush with the end | the last metric band ends at 90px of a 90-pixel track | Pass |
| 7 | A zero-width interval is still drawn | a band with a minimum width | 3px at the end of the track | Pass |

### 4.3 The Gauge

| # | Check | Expected | Measured | Result |
|---|---|---|---|---|
| 8 | The threshold notch lies inside its track | within the track | the notch at 69.7% of a 328.3-pixel track, inside it | Pass |
| 9 | The threshold notch sits at the pinned position | 70% of the track | 69.7% | Pass |
| 10 | Every gauge fill lies inside its track | within the track | 3 fills, all within their tracks | Pass |
| 11 | A reading of zero is still visible | the fill is at least two pixels wide | 2.0px on both failure-rate gauges | Pass |

### 4.4 The Repository Gates

| # | Check | Result |
|---|---|---|
| 12 | `web-typecheck` | passed |
| 13 | `web-test` | passed, 18 tests in 5 files |
| 14 | `web-build` | passed |
| 15 | `go-fmt`, `go-vet`, `go-test`, `go-build` | passed |
| 16 | `docs-english`, `docs-links`, `docs-index` | passed |
| 17 | `e2e-smoke` | passed |

## 5. Defects Found and Fixed

| # | Defect | Cause | Fix |
|---|---|---|---|
| 1 | The mark hung over the track at the two ends of the scale | The percentage of an absolutely positioned mark is resolved against the padding box, and the track was drawn in the content box | The track is the positioning context, and the mark is clamped inside it |

## 6. Result

The defect is verified as fixed by [FIX-004][fix-004]. One defect was found during the verification and it was fixed. The 17 checks above pass, and the scale was measured at 0.0000, 0.6667 and 1.0000, so the claim holds across the range and not at one point of it.

## 7. What This Report Does Not Claim

- It does not claim that the mark is exactly on the reading at the two ends of the scale. A mark clamped inside the track at a reading of 0.0000 is drawn flush with the start of the track, one half of its width away from the exact position. The offset is 1.5 pixels of an 788-pixel track.
- It does not cover the small screen layout, where the track is narrower and the same error would be proportionally larger. The measurements were taken at the default viewport.
- It does not claim that the primitive is accessible on its own: the tracks carry `aria-hidden` and the values are published as text beside them.

[fix-004]: <../Fix/reporting/scale-mark-overhung-the-track-at-the-ends-of-the-scale.md>
