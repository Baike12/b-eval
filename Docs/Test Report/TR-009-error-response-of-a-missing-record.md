# TR-009: Error Response of a Missing Record

| Field | Value |
|---|---|
| **ID** | TR-009 |
| **Status** | Verified |
| **Verifies** | [FIX-006 Error Response Published the Server File Path][fix-006] |
| **Owner** | Platform team |
| **Created** | 2026-10-22 |
| **Last Updated** | 2026-10-22 |

> This report records what was checked, how it was checked and what the check returned. The defect it verifies is specified in [FIX-006][fix-006].

## 1. Scope

The verification covers what the platform says when a reader asks for a record it does not hold: the status code, the message, and what that message discloses about the deployment.

The claim under verification is that the platform names the object that was looked for and discloses nothing about the host it runs on.

## 2. Method

| Aspect | How it was checked |
|---|---|
| The response | The three addresses requested against the running platform on the host, and against a server under test in the Go test suite |
| The screen | The same address opened in the Chromium instance attached to BrowserOS neo |
| The handlers | Every handler that reads a record, read in the source |

## 3. The Evidence Set

| Address | Object it names |
|---|---|
| `/eval/api/runs/nope` | A run |
| `/eval/api/runs/nope/report` | The report of a run |
| `/eval/api/tasksets/nope/v1` | A task set at a version |

## 4. Checks

### 4.1 The Response

| # | Check | Expected | Measured | Result |
|---|---|---|---|---|
| 1 | A missing run is answered with 404 | 404 | 404 | Pass |
| 2 | The message names the run that was looked for | the run's identifier | `run "nope" does not exist` | Pass |
| 3 | The message names no path | the data directory appears nowhere | no `/app/data` and no file name | Pass |
| 4 | The message carries no raw file error | no `no such file or directory` | absent | Pass |
| 5 | A missing task set is answered the same way | 404 and a message | `task set "nope" version "v1" does not exist` | Pass |
| 6 | A missing run report is answered the same way | 404 and a message | `run "nope" does not exist` | Pass |

Before the fix, checks 2 to 4 were false: the body was `{"error":"read /app/data/runs/does-not-exist.json: open /app/data/runs/does-not-exist.json: no such file or directory"}`, which named the path twice.

### 4.2 The Screen

| # | Check | Expected | Measured | Result |
|---|---|---|---|---|
| 7 | The run screen of a run that does not exist opens | the screen with an error notice | the screen opens, 997 bytes of markup, the navigation present | Pass |
| 8 | The notice carries the API's message | the message, verbatim | `run "does-not-exist" does not exist` | Pass |
| 9 | The notice carries no path | no path | no `/app/data` | Pass |

### 4.3 The Handlers and the Gates

| # | Check | Expected | Measured | Result |
|---|---|---|---|---|
| 10 | Every handler that reads a record translates the store's error | 7 sites | 7 sites go through the translation; the 8th ignores the error and reads a task set for a name only | Pass |
| 11 | A record that exists but cannot be read is a server fault | 500, with the detail in the server's log | 500, and the detail is written to the standard error log | Pass |
| 12 | `go-test` | passed, including `TestMissingRecordResponseDoesNotLeakTheDataDirectory` | passed | Pass |
| 13 | `e2e-smoke` | passed | passed | Pass |
| 14 | `go-fmt`, `go-vet`, `go-build` | passed | passed | Pass |
| 15 | `web-typecheck`, `web-test`, `web-build` | passed | passed | Pass |
| 16 | `docs-english`, `docs-links`, `docs-index` | passed | passed | Pass |

## 5. Defects Found and Fixed

| # | Defect | Cause | Fix |
|---|---|---|---|
| 1 | The response published the deployment's file path | The store wrapped the file error with the path, and the handlers passed it through | The store names the record, and the API names the object |

## 6. Result

The defect is verified as fixed by [FIX-006][fix-006]. One defect was found during the verification and it was fixed. The 16 checks above pass, and the response to a request for an absent record names the object and nothing else.

## 7. What This Report Does Not Claim

- It does not claim that no response of the platform can disclose the host's layout. It covers the handlers that read a record, which are the ones that carry the store's errors.
- It does not verify the operator's view of a corrupt record, which is the server's log line rather than the response.
- It does not claim that the message is the one a reader would choose. It claims that the message names the object and discloses nothing about the host.

[fix-006]: <../Fix/reporting/error-response-published-the-server-file-path.md>
