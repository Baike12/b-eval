# Fix Index

> Index of every bug fix document in b-eval.

Each fix is one document, written to [`../Fix-Template.md`](../Fix-Template.md), filed in one
folder per category, and linked to the Test Report that verifies it.

**Tech Design documents are not indexed here.** A fix that needs a technical design links it
from its own `References` section; a fix that changes the behaviour of a feature is
cross-referenced from that feature.

| Category | What it covers |
|---|---|
| [Agent Integration](#agent-integration) | The agent adapter protocol and the self-check that gates evaluation |
| [Reporting](#reporting) | The report the platform publishes, the contract it publishes it under, and the geometry that carries it |
| [Experience](#experience) | The journey a reader has through the interface, and what the interface does when it cannot answer |

---

## Agent Integration

| ID | Fix | Summary |
|---|---|---|
| [FIX-002](agent-integration/blackbox-cli-empty-action-sequence-not-flagged.md) | Black-box CLI adapter recorded an empty action sequence instead of a protocol violation | A black-box agent that produced no observable output was recorded as a legitimate zero-action trial and silently counted as completed; the trial-close validation now rejects an empty action sequence with no matching violation. |

---

## Reporting

| ID | Fix | Summary |
|---|---|---|
| [FIX-003](reporting/report-null-collection-blanks-the-report-screen.md) | The run report screen rendered blank when a report carried an empty collection | The report builders returned Go nil slices, which the JSON encoder writes as `null`, so the client read a length off `null` and the whole screen went blank for every run with no failed trial; the builders now publish a list, and a test asserts that no value of a report is null. |
| [FIX-004](reporting/scale-mark-overhung-the-track-at-the-ends-of-the-scale.md) | The reading mark hung over the track at the two ends of the scale | The mark was placed against the padding box while the track was drawn in the content box, so a reading of 1.0000 stood 2.5 pixels past the end of its scale; the track is now the positioning context of its marks and the mark is clamped inside it. |
| [FIX-006](reporting/error-response-published-the-server-file-path.md) | An error response published the server's own file path | A request for a record the store does not hold was answered with the store's own error, which named the absolute path of the file it looked for, and the web client showed that path on the screen; the response now names the object that was looked for and the path stays in the server's log. |

---

## Experience

| ID | Fix | Summary |
|---|---|---|
| [FIX-005](experience/unknown-address-rendered-an-empty-document.md) | An address that names nothing rendered an empty document | The route table had no catch-all, so a stale link or a mistyped address produced an empty page with no navigation and no message; the table now ends with a route that states that the address names nothing and offers the way back. |
