# TD-006: Interface Visual Design System

| Field | Value |
|---|---|
| **ID** | TD-006 |
| **Status** | Implemented |
| **Implements** | [FEAT-007 User Flows, Feature Inventory and Permissions](../Features/experience/user-flows-and-feature-inventory.md) |
| **Owner** | Platform team |
| **Created** | 2026-10-22 |
| **Last Updated** | 2026-10-22 |

> This document specifies *how* the interface looks and behaves visually. The *what* and *why* live in the Feature document it implements.

## 1. Problem Restatement

The M0 slice ships a working single-page application: every screen renders, every number the platform publishes reaches the reader, and the report leads with its confidence section. What it does not have is a visual identity. The interface used the browser's default sans-serif, blue hyperlink colour, a flat white card for every kind of content, and the same border, radius and padding on a form, a table and a report section.

Three problems follow from that, and only the first is a matter of taste.

**Hierarchy is absent.** A report section, a form panel and a metric row are the same object. The reader has no way to tell what is a container and what is content, so the page reads as a stack of undifferentiated boxes.

**The platform's own thesis is invisible.** The platform holds that a number without its coverage is not a measurement, and that the confidence section must precede the score. That rule is implemented in the report's *order*, but nothing in the *rendering* of a number distinguishes a reading of zero from the absence of a reading. The interface contradicts the product in its visual language.

**A fixed 0-to-1 scale cannot show a failure rate.** The platform failure rate is compared against a threshold that is typically 0.05. Drawn on a fixed 0-to-1 scale, a rate of 0.01 is one hundredth of the track, and a reader cannot see where it sits relative to the limit that judges it.

## 2. Design Goals

| Goal | Consequence for the design |
|---|---|
| The interface reads as an instrument, not a document | A dark bezel frames a light measuring surface; type and rules carry the hierarchy, not boxes |
| Absence is visibly different from zero | Missing coverage is drawn as a hatched empty track, never as a number and never as an empty fill |
| Uncertainty is shown, not just stated | Every confidence interval is drawn on a scale, so its width is legible |
| The threshold a rate is judged by is on the scale | The failure-rate gauge pins the threshold and draws the rate against it |
| The platform is self-contained | The two typefaces are vendored into the repository and served from the binary |
| The screens work on a narrow window | The layout reflows to a single column and tables scroll horizontally |

## 3. Non-Goals

- The redesign changes no behaviour, no API shape and no data contract. Every screen performs the same actions as before and every value it publishes is the same value.
- The redesign does not add dark mode. One deliberately finished theme is better than two half-finished ones.
- The redesign does not introduce a component framework or a CSS-in-JS build step. The stylesheet is hand-written, because the number of distinct components is small and their geometry is load-bearing.
- The redesign does not change the Go backend, the metric registry or the report builder.

## 4. The Visual Argument

The subject matter is measurement. An evaluation platform produces readings, and a reading without its uncertainty is a claim rather than a measurement. The design takes that seriously and makes the reader's eye land on three things in order.

**The confidence, before the score.** The confidence section opens the report and is the only element on the page with a tinted header band, so it separates from the white sections that follow. Its gauges carry a threshold mark on the scale.

**The score, on a scale.** The main score is drawn as a large figure beside a 0-to-1 track. The confidence interval is a band on that track and the score is a point mark on it. A reader sees a 0.5 whose band spans the entire track and understands at once that the sample is too small to conclude anything. The same three numbers in a table would have said the same thing, but not at a glance.

**The absence, hatched.** A cell with no coverage is drawn as a diagonal hatch. The hatch is the drafting convention for a region that is not material, and it is used nowhere else in the interface, so it always means one thing: this was not measured.

Everything else is deliberately quiet. Sheets are white fields with a hairline edge and no shadow. Rules separate sections. Colour carries state and nothing else.

## 5. The Token System

### 5.1 Colour

Colour has three jobs in this interface: it marks state, it marks interaction, and it does not appear as decoration.

| Token | Value | Job |
|---|---|---|
| `--bezel` | `#14181b` | The instrument frame. It carries the navigation and nothing else. |
| `--paper` | `#f7f7f5` | The measuring surface the content sits on. |
| `--sheet` | `#ffffff` | A content field. It has an edge, never a shadow. |
| `--ink` | `#171a1c` | Primary text. |
| `--muted` | `#6c7577` | Secondary text and axis labels. |
| `--rule` | `#e3e5e4` | Hairlines between rows and sections. |
| `--signal` | `#0d6a6e` | The signal colour: a measured quantity, an active tab, a primary action. |
| `--danger` | `#a03223` | A quantity that failed its threshold, and the reportable verdict's opposite. |

The signal colour is a deep teal rather than the blue that a dashboard reaches for by default, because the interface is a measuring instrument and teal reads as an instrument trace. It is used for interaction and for the fill of a measured quantity, and for nothing else. A missing quantity is never teal; it is a hatch on a neutral track.

### 5.2 Type

Two families, both vendored into `web/src/assets/fonts/` and served from the embedded bundle:

| Family | Role | Reason |
|---|---|---|
| Archivo (variable, 100–900) | All prose, headings, labels and numbers | A grotesque with squared terminals and a narrow aperture, which reads as instrument labelling rather than as a consumer web font |
| IBM Plex Mono (400, 500) | Machine strings only: run identifiers, fingerprints, metric identifiers, formulas, interval bounds | These are values a reader compares character by character, and a fixed width is what makes them comparable |

The families are self-hosted because the platform is deployed to a host in mainland China, where a font CDN is not reliably reachable. A page that falls back to a system font is a page whose measurements were rendered in a different typeface from the one that was designed.

Numerical columns use `font-variant-numeric: tabular-nums`, so digits align across rows without resorting to the monospace face for values that are not machine strings.

### 5.3 Scale and Rhythm

The base size is 15px with a line height of 1.55. The page title is 27px at weight 600, a section title 17px, and a label 12–13px in the muted colour. Spacing runs on a 4px rhythm: 14px between a heading and its content, 22px inside a sheet, 38px between sections.

## 6. The Measurement Primitives

Two primitives carry the design. Both are plain DOM with inline custom properties, because their geometry has to be exact.

### 6.1 The Track (`--from`, `--to`, `--at`)

A horizontal track with an optional band and an optional point mark. The band is the confidence interval; the point is the reading.

The track is the positioning context of its marks, so a percentage on a mark is a percentage of the track itself. A mark has a width of its own, so a reading at either end of the scale is clamped inside the track: a reader must be able to tell a reading of zero from a reading that was cut off by the edge. The clamp is `max(0px, min(...))` against the mark's own width, and the width is a property of the scale, so the wide mark of the main score and the narrow mark of a metric row obey the same rule.

The marks are positioned with the `--from`, `--to` and `--at` custom properties rather than with `left` and `width`, so that one rule governs every scale in the interface.

### 6.2 The Gauge

A gauge is a track with a fill and an optional notch. The fill is the observed rate; the notch is the threshold the rate is judged by.

The failure-rate gauge **pins** the threshold at 70% of the track rather than fitting the scale to the observed rate. A fitted scale would move the limit as the rate moves, and the reader would then be comparing the rate against a mark that moves with it. The scale maximum is the threshold divided by 0.7, so a rate below the limit always reads as below it and a rate above the limit visibly crosses it and turns the fill red.

The fill has a minimum width of two pixels. A rate of exactly zero is a fact, and it must not render as an absent bar.

## 7. Structure

```
┌──────────────────────────────────────────────────────────────┐
│  ▮ b-eval    agent evaluation platform      Dashboard  …    │  bezel, 56px, sticky
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  Run report                                                  │  page title, 27px
│  Run  …  Task set  …  Agent  …  Repetitions  …              │  facts, labelled fields
│                                                              │
│  ┌ Confidence            reportable ───────────────────────┐ │  tinted header band
│  │ Platform failure rate   0.0000 ────│──────             │ │  gauge with threshold notch
│  │ Judgement failure rate  0.0000 ──────│────             │ │
│  │ Valid sample size       8 / 8 ██████████████████        │ │
│  └─────────────────────────────────────────────────────────┘ │
│                                                              │
│  ┌ Main score ─────────────────────────────────────────────┐ │
│  │ 0.5000          ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓ │ │  band on a 0–1 track
│  │ [0.0000, 1.0000] normal                                   │ │
│  │ ────────────────────────────────────────────────────────  │ │
│  │ pass@k   pass^k   Continuous average   Coverage          │ │
│  └──────────────────────────────────────────────────────────┘ │
│                                                              │
│  Metrics, Breakdown, Tasks, Failure analysis — one sheet each │
└──────────────────────────────────────────────────────────────┘
```

A section is a heading and a sheet. A sheet is a white field with a hairline edge, and it is flat: no shadow, no gradient, no corner radius that says "app". A table inside a sheet is flush with the sheet's edges and carries its own row rules, so the table reads as a ruled sheet of data rather than as a widget.

Facts are rendered as labelled fields in a `dl`, not as a sentence joined by separators. A reader looking for the repetitions of a run should be able to find the field without reading the whole line.

## 8. The Screens

| Screen | Change |
|---|---|
| Dashboard | The status line became a four-cell readout strip. The run form and the recent runs became a sheet and a ruled list with a status chip. |
| Run report | The confidence section became the tinted band with three gauges. The main score became the figure with the interval on a scale. The four sections became sheets. |
| Runs | A ruled table with a status chip and a monospace identifier. |
| Task sets | One section per task set, its facts as labelled fields, its tasks as a ruled table. |
| Task detail | The badge row became a facts list; the contract fields became a pairs table with the budget and judgement broken into labelled fields. |
| Agents | The register form became a sheet; the registered table carries a self-check chip and a secondary self-check action. |
| Metric registry | The registry table and the registration form became two sections; a built-in metric is a filled chip and a registered one is an outlined chip. |

## 9. The Interface Contract

The redesign is bound by the platform's existing interface rules, and it satisfies them visually rather than only structurally:

| Rule | How the redesign satisfies it |
|---|---|
| The confidence section is always at the top of the report and can be neither collapsed nor hidden | It is the first element of the report and contains no disclosure control |
| A value with no coverage is visually distinguished from a normal value | The missing value is a hatched empty track and the word "no coverage", never a number and never an empty fill |
| A failed task shows its attribution | The failure analysis lists every failed trial with its reason and class |
| The main score is not a single overall capability number | The statement sits in the sheet's footer, directly under the reading it qualifies |

## 10. Verification

The visual claims in this document are verified in [TR-005 Interface Visual Design System][tr-005]. The claims include the measured geometry of the marks on the scale, the absence of a duplicate interval band on the main score, and the rendering of a reading of zero.

[tr-005]: <../Test Report/TR-005-interface-visual-design-system.md>
