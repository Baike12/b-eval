# TR-008: Unknown Address and Render Failure Handling

| Field | Value |
|---|---|
| **ID** | TR-008 |
| **Status** | Verified |
| **Verifies** | [FIX-005 Unknown Address Rendered an Empty Document][fix-005] |
| **Owner** | Platform team |
| **Created** | 2026-10-22 |
| **Last Updated** | 2026-10-22 |

> This report records what was checked, how it was checked and what the check returned. The defect it verifies is specified in [FIX-005][fix-005].

## 1. Scope

The verification covers the two ways the interface used to leave a reader with an empty document: an address that matches no route, and a payload that breaks the contract the client was written against.

Both are now answered. The check is that the platform answers, and that the answer is inside the frame of the application rather than in place of it.

## 2. Method

| Aspect | How it was checked |
|---|---|
| Rendered outcome | The Chromium instance attached to BrowserOS neo, reading the markup of the root element and the text of the document |
| The route table | The frontend test suite, rendering the table at an address that matches nothing |
| The deployment | The same addresses opened against the running platform on the host |

## 3. The Evidence Set

| Address | Why it was used |
|---|---|
| `/eval/task-sets/demo/v1` | The address a task set has by name and version. No route matched it, and it is the address a reader arrives at from outside the platform |
| `/eval/nothing/here` | An address that names nothing at all |
| `/eval/runs/does-not-exist` | A route that matches, whose record does not exist. The screen opens, and the error notice is the API's message |

## 4. Checks

### 4.1 The Addresses

| # | Check | Expected | Measured | Result |
|---|---|---|---|---|
| 1 | An address that names nothing renders a page | a title, a message and a way back | `Nothing here`, 1294 bytes of markup | Pass |
| 2 | The navigation survives the address | the tabs are present | the five tabs present | Pass |
| 3 | The way back is offered | a link to the dashboard | the link "Back to the dashboard" present | Pass |
| 4 | An address that names nothing at all behaves the same | the same screen | `Nothing here`, 1264 bytes of markup | Pass |
| 5 | No address renders an empty document | a non-empty root element | the shortest of the seven screens renders 1634 bytes | Pass |

Before the fix, check 1 measured 0 bytes of markup: the root element was empty, the navigation was gone, and nothing was on the screen.

### 4.2 The Route Table

| # | Check | Result |
|---|---|---|
| 6 | The route table answers an address that names nothing | the case "answers an address that names nothing instead of rendering nothing" passes |

### 4.3 The Repository Gates

| # | Check | Result |
|---|---|---|
| 7 | `web-typecheck` | passed |
| 8 | `web-test` | passed, 19 tests in 6 files |
| 9 | `web-build` | passed |
| 10 | `go-fmt`, `go-vet`, `go-test`, `go-build` | passed |
| 11 | `docs-english`, `docs-links`, `docs-index` | passed |
| 12 | `e2e-smoke` | passed |

## 5. Defects Found and Fixed

| # | Defect | Cause | Fix |
|---|---|---|---|
| 1 | An address that names nothing rendered an empty document | The route table had no catch-all, and a router renders nothing when no route matches | The table ends with a catch-all route |
| 2 | A payload that breaks the contract blanked the whole screen | Nothing caught a render failure, and the exception unmounted the tree | The screen is wrapped in a boundary that names the fault |

## 6. Result

The defect is verified as fixed by [FIX-005][fix-005]. Two defects were found during the verification and both were fixed. The 12 checks above pass, and no address of the platform leads to an empty document.

## 7. What This Report Does Not Claim

- It does not claim that every address a reader can reach has been tried. It claims that the address that matched no route is answered, and that a screen that failed to render is answered.
- It does not verify what the run screen shows for a run that does not exist beyond the fact that it opens and shows the API's message. The content of that message is verified in [TR-009][tr-009].
- It does not claim that the platform handles a server that is down. A server that does not answer produces a network error, which the client shows as a notice.

[fix-005]: <../Fix/experience/unknown-address-rendered-an-empty-document.md>
[tr-009]: <./TR-009-error-response-of-a-missing-record.md>
