# TR-005: Interface Visual Design System

| Field | Value |
|---|---|
| **ID** | TR-005 |
| **Status** | Verified |
| **Verifies** | [TD-006 Interface Visual Design System][td-006] |
| **Owner** | Platform team |
| **Created** | 2026-10-22 |
| **Last Updated** | 2026-10-22 |

> This report records what was checked, how it was checked and what the check returned. The design it verifies is specified in [TD-006][td-006].

## 1. Scope

The verification covers the visual redesign of the single-page application: the type system, the colour system, the measurement primitives, the sheet and section structure, and the seven screens.

The redesign changes no behaviour. Every value the platform publishes before the redesign is the value it publishes after it, and every action a reader could take before it is an action they can take after it. The checks below therefore verify the rendering and the structure, not the numbers themselves.

## 2. Method

| Aspect | How it was checked |
|---|---|
| Rendered geometry | The Chromium instance attached to BrowserOS neo, reading `getBoundingClientRect()` and `getComputedStyle()` on the live DOM |
| Type system | `document.fonts` and the computed `font-family` of the body |
| Structure | The accessibility tree, the heading levels and the element counts of each screen |
| Behaviour | The existing frontend test suite, and the repository's own quality gate |

The preview server was `b-eval serve` on the loopback interface with a seeded data directory, so every check ran against the real binary with the real report builder.

## 3. The Evidence Set

Two reports were used, both produced by the platform itself rather than hand-written:

| Report | Content | Why it was used |
|---|---|---|
| `run-plan-demo-demo-good-3408113d` | Score 0.5000, interval [0.0000, 1.0000], 8 valid trials | A non-zero reading with an interval that spans the whole scale |
| `run-plan-partial-partial-agent-7f4b3d8d` | Score 0.0000, interval [0.0000, 0.0000], three metrics with no coverage | A zero reading and the missing-coverage state |

The second report was produced by running an agent that solved none of its tasks, which is what makes the three `avg_*_per_success` metrics unmeasurable: their denominator is the number of successful trials, and there were none.

## 4. Checks

### 4.1 Type

| # | Check | Expected | Measured | Result |
|---|---|---|---|---|
| 1 | The body typeface is the vendored one | Archivo | `Archivo, ui-sans-serif, system-ui, -apple-system, "Segoe UI", sans-serif` | Pass |
| 2 | The variable weight axis is present | one file covering 100–900 | `Archivo 100 900 loaded` | Pass |
| 3 | The monospace family is present for machine strings | IBM Plex Mono | `Plex Mono 400 loaded`, `Plex Mono 500 unloaded` (unused at the default weight) | Pass |
| 4 | The font is served by the binary, not by a CDN | a request to the application's own origin | `GET /assets/archivo-latin-var-*.woff2` → `200`, 34,940 bytes | Pass |
| 5 | The base size is the specified one | 15px | `15px` | Pass |

The fonts are vendored into `web/src/assets/fonts/` because the platform is deployed to a host in mainland China, where a font CDN is not reliably reachable. Check 4 confirms the binary serves them itself.

### 4.2 Colour and Structure

| # | Check | Expected | Measured | Result |
|---|---|---|---|---|
| 6 | The frame is the dark bezel | `rgb(20, 24, 27)` | `rgb(20, 24, 27)` | Pass |
| 7 | The measuring surface is the paper ground | `rgb(247, 247, 245)` | `rgb(247, 247, 245)` | Pass |
| 8 | The signal colour is the teal, not a dashboard blue | `rgb(13, 106, 110)` | `rgb(13, 106, 110)` | Pass |
| 9 | A sheet has no drop shadow | `none` on every sheet | `none` on all five sheets of the report | Pass |
| 10 | The report is built from sections and sheets, not from stacked cards | sheets and section headings | 5 sheets, each with a heading above it | Pass |

### 4.3 The Measurement Primitives

| # | Check | Expected | Measured | Result |
|---|---|---|---|---|
| 11 | The point mark of the main score lies inside its track | left and right edges within the track | point `977.0–980.0`, track `583.5–1371.5`, on the reading 0.5000 | Pass for that reading — the mark overhung the track at the two ends of the scale, which [FIX-004][fix-004] corrected and [TR-007][tr-007] re-measured |
| 12 | The interval band lies inside its track | left and right edges within the track | band `584.5–1372.5` before the fix, `584.5–1371.5` after | **Fail, then Pass** |
| 13 | The threshold notch sits at the pinned position | 70% of the track | 69.7% of 328.3px | Pass |
| 14 | A reading of zero is still visible | the fill is at least two pixels wide | 2.0px | Pass |
| 15 | The main score draws one interval, not two | exactly one band on the score | 1 band on the scale, 0 bands inside the interval label | Pass |
| 16 | A zero-width interval is still drawn | a band with a minimum width | `range__band--point`, 3px minimum | Pass |
| 17 | A missing value is drawn as a hatch, not as a number | a diagonal hatch, 22×8px | `repeating-linear-gradient(-45deg, ...)`, 22px × 8px, on all 6 no-coverage cells | Pass |

Check 12 failed on the first pass: the band was one pixel wider than the track, because the band's offset was measured against the padding box while its width was measured as the difference between the two ends, and the inset was counted once in the offset and not subtracted from the width. The fix subtracts the inset from the width. The measurement was taken after the fix and is reported above.

Check 15 failed on the first pass as well: the main score drew the interval on the scale and the interval label drew it again, so one interval appeared as two bands and the reader could not tell which carried the reading. The label now draws the bounds and the method only.

Check 11 was measured on a reading in the middle of the scale, and it does not hold for every reading. The mark was positioned against a one-pixel inset of the track, but the percentage of an absolutely positioned mark is resolved against the padding box rather than the content box, so the two ends of the scale were drawn at the wrong width and the mark hung over the track by two and a half pixels at a reading of 1.0000. The geometry is corrected in [FIX-004][fix-004] and the whole scale is measured in [TR-007][tr-007].

### 4.4 The Screens

| # | Check | Expected | Measured | Result |
|---|---|---|---|---|
| 18 | Every screen renders its title | a page title on each | Dashboard, Task sets, Agents, Runs, Metric registry, Run report | Pass |
| 19 | The confidence section opens the report | it is the first element | it precedes the main score | Pass |
| 20 | The confidence section cannot be collapsed | no disclosure control in the section | the existing test asserts no `details` and no `button`; it passes | Pass |
| 21 | A missing value is distinguished from a normal value | the word "no coverage" and a hatch | 6 cells on the partial report, each with both | Pass |
| 22 | The main score is not presented as a single capability number | the statement sits under the reading | present in the sheet's footer | Pass |
| 23 | The metric table shows the aggregation method next to the value | the method is a separate line | the value cell reads `0.0000` and `macro` on separate lines | Pass |

Check 23 failed on the first pass: the value and the aggregation method rendered as `0.0000macro`, because the aggregation rule that separated them was dropped when the stylesheet was rewritten. The rule was restored and the two are now separate lines.

### 4.5 The Repository Gates

| # | Check | Result |
|---|---|---|
| 24 | `web-typecheck` | passed |
| 25 | `web-test` | passed, 15 tests in 4 files |
| 26 | `web-build` | passed |
| 27 | `go-fmt`, `go-vet`, `go-test`, `go-build` | passed |
| 28 | `docs-english`, `docs-links`, `docs-index` | passed |
| 29 | `e2e-smoke` | passed |

The English gate does not read the vendored font binaries: it reads the `.md`, `.go`, `.ts`, `.tsx`, `.json`, `.yaml`, `.yml`, `.sh`, `.html` and `.css` extensions, and the fonts carry the `.woff2` extension.

## 5. Defects Found and Fixed

| # | Defect | Cause | Fix |
|---|---|---|---|
| 1 | The interval band overhung its track by one pixel | The band's offset was measured against the padding box and its width as a plain difference, so the inset was counted twice in one direction and not at all in the other | The width subtracts the inset |
| 2 | The main score drew its interval twice | The scale and the interval label each drew a band | The label draws the bounds and the method only |
| 3 | A reading of zero rendered as an absent bar | The fill had no minimum width | The fill is at least two pixels wide |
| 4 | The value and its aggregation method rendered as `0.0000macro` | The rule that separated them was dropped in the stylesheet rewrite | The rule was restored |
| 5 | The threshold mark sat at a position that moved with the rate | The scale was fitted to the observed rate, so the limit moved as the rate moved | The limit is pinned at 70% of the track |

## 6. Result

The redesign is verified against [TD-006][td-006]. Five defects were found during the verification and all five were fixed and re-measured. The 29 checks above pass.

## 7. What This Report Does Not Claim

- It does not claim that the design is optimal, only that it renders as specified and that the platform's interface rules are satisfied by it.
- It does not verify the platform's numbers. The report builder is covered by the Go test suite and by the end-to-end smoke gate, both of which pass.
- It does not cover dark mode, which the redesign does not implement.
- It does not claim that the mark of check 11 stays inside its track for every reading. The check was measured on one reading in the middle of the scale; the two ends of the scale were measured afterwards, and the defect they showed is recorded in [FIX-004][fix-004].

[td-006]: <../Tech Design/TD-006-interface-visual-design-system.md>
[fix-004]: <../Fix/reporting/scale-mark-overhung-the-track-at-the-ends-of-the-scale.md>
[tr-007]: <../Test Report/TR-007-scale-mark-geometry.md>
