# FIX-006: An error response published the server's own file path

| Field | Value |
|---|---|
| **ID** | FIX-006 |
| **Status** | Fixed |
| **Owner** | Platform team |
| **Severity** | S3 Medium |
| **Priority** | P1 |
| **Affects** | `F-RPT-01` |
| **Detected in** | Deployment `20261022-ui-final`, 2026-10-22 |
| **Created** | 2026-10-22 |
| **Last Updated** | 2026-10-22 |

## Summary

The API answered a request for a record that does not exist with the store's own error, which names the absolute path of the file it looked for. The response published the deployment's directory layout, and the web client showed the message verbatim.

## Severity and Impact

**Technical impact:** The store's `readJSONFile` wrapped the operating system's error with the path it had opened. The API's not-found handlers passed that error straight into the response body, so `GET /api/runs/does-not-exist` answered `{"error":"read /app/data/runs/does-not-exist.json: open /app/data/runs/does-not-exist.json: no such file or directory"}`. The path is the deployment's own layout: where the records live on the host, how they are named, and that the platform keeps one file per record.

**Blast radius:** Every request for a task set, a task set version, a run, a run report, a run trial or a run comparison that named a record the store does not hold — 7 handler sites. On the deployment, `GET /api/runs/does-not-exist`, `GET /api/runs/does-not-exist/report` and `GET /api/tasksets/nope/v1` all published the path. The web client renders the API's message as the error notice, so the path reached the screen of every reader who opened a stale link.

**Detection:** Found during the verification of the deployed interface, when the run screen was opened at the address of a run that does not exist and the screen showed the server's file path.

## Symptom

Opening `http://<host>/eval/runs/does-not-exist` produced a red notice reading:

```
read /app/data/runs/does-not-exist.json: open /app/data/runs/does-not-exist.json: no such file or directory
```

The status code was 404, which was correct; the body was not.

## Reproduction

**Environment:** Deployment `20261022-ui-final` on the cloud host, data directory `/opt/b-eval/data` mounted at `/app/data`.

**Preconditions:** A record the store does not hold.

**Steps:**

1. Request a run that does not exist.
2. Read the response body.

**Expected:** The response states that the run does not exist.

**Actual:** The response carried the absolute path of the file the store looked for, twice.

## Root Cause

**Cause:** The store reports a missing file by wrapping the operating system's error, and the wrapper names the path. The API's not-found handlers passed the store's error into the response without translating it, so the internal detail became the published message. The client displays the API's error field as it is.

**Failure class:** Integration

**Why it was not caught:** The API's own tests covered the status code of a missing record, and the status code was right. Nothing asserted what the body says, so a body that leaked the deployment's layout passed every check. The agent handler already answered with a clean message of its own, and the difference between that handler and the others was not noticed.

## Fix

**Change:** The store now names a missing record with a sentinel error, `store.ErrNotFound`, and names every other failure with the record's file name rather than its path. The API translates the two cases apart: a record that is not there is a 404 that states what was looked for, and a record that cannot be read is a 500 whose internal detail goes to the server's log rather than to the response.

**Rationale:** The store knows the file and the API knows the object, so the object is what the response names. Splitting the two cases also makes the status code truthful: a record that exists but cannot be parsed is a server fault, not a missing record.

**Guardrail added:** `TestMissingRecordResponseDoesNotLeakTheDataDirectory` requests three absent records and asserts that the response names the data directory nowhere, carries no raw file error, and says what was missing.

## Verification

| Check | Method | Result |
|---|---|---|
| Original reproduction no longer reproduces | The same three addresses requested on the deployment | **Pass** — `{"error":"run \"nope\" does not exist"}` |
| Regression test added | `internal/api/api_test.go` → `TestMissingRecordResponseDoesNotLeakTheDataDirectory` | **Pass** — it fails against the code before the fix, on all three addresses |
| Adjacent behaviour unaffected | The whole Go suite and the end-to-end smoke gate | **Pass** — the smoke run creates, reads and reports a task set, an agent and a run through the same handlers |
| Blast radius re-scored | Every handler that reads a record, read in the source | **Pass** — 7 sites read a record; all 7 now go through the translation, and the one site that ignores the error reads a task set for a name only |

## Regression Prevention

- **Test added:** `TestMissingRecordResponseDoesNotLeakTheDataDirectory`, in `internal/api/api_test.go`, asserts the shape of the body of a missing record.
- **Run gate:** The test runs in the `go-test` gate of `scripts/gate.sh`, which every change must pass.
- **Detection change:** A record that exists but cannot be read is now logged by the server and answered with a 500, so a corrupt record no longer passes for a missing one.

## References

- Test Report: [TR-009 Error Response of a Missing Record][tr-009]
- Related Fix: [FIX-005 Unknown Address Rendered an Empty Document][fix-005]

## Open Questions

| # | Question | Owner | Blocking? |
|---|---|---|---|
| 1 | No functional requirement states that the API must not publish the deployment's own layout. The defect was fixed at the source, but the rule it follows is not written down, so a later endpoint can publish a path again without contradicting a requirement. | Platform team | No |

[tr-009]: <../../Test Report/TR-009-error-response-of-a-missing-record.md>
[fix-005]: <../../Fix/experience/unknown-address-rendered-an-empty-document.md>
