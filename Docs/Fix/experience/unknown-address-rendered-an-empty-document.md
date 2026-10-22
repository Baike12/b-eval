# FIX-005: An address that names nothing rendered an empty document

| Field | Value |
|---|---|
| **ID** | FIX-005 |
| **Status** | Fixed |
| **Owner** | Platform team |
| **Severity** | S3 Medium |
| **Priority** | P2 |
| **Affects** | `F-HUM-01` |
| **Detected in** | Deployment `20261022-ui-final`, 2026-10-22 |
| **Created** | 2026-10-22 |
| **Last Updated** | 2026-10-22 |

## Summary

The route table had no catch-all, so an address that matched no route rendered no element at all and the document stayed empty. A reader who opened a stale link, or mistyped an address, was shown nothing and could not tell that the platform had an answer for them.

## Severity and Impact

**Technical impact:** React Router renders the matched route and nothing else. With no `*` route in the table, an unmatched address produced an empty document: no page title, no navigation, no message, and no entry in the console. The reader could not tell the address from a screen that was still loading, and the platform offered no way back.

**Blast radius:** Every address that names nothing. The table has a route for a task detail and none for a task set on its own, so `GET /task-sets/demo/v1` — the address a task set has by name and version — was the reachable case. No in-application link produced it, so the blast radius was the addresses a reader arrives at from outside the platform: a bookmark, a copied link, a hand-typed path.

**Detection:** Found during the verification of the deployed interface, when every route of the application was opened in turn and the task set address was found to render an empty document.

## Symptom

Opening `http://<host>/eval/task-sets/demo/v1` produced a document with an empty root element. The measurement of the root element's markup length was 0, against 1634 to 20246 for every other screen.

## Reproduction

**Environment:** Deployment `20261022-ui-final` on the cloud host.

**Preconditions:** An address that matches no route in the table.

**Steps:**

1. Open an address that names no task set, no task, no agent, no run and no metric.
2. Read the markup of the root element.

**Expected:** The platform states that the address names nothing and offers the way back.

**Actual:** The root element was empty. The page was blank.

## Root Cause

**Cause:** The route table declared one route per object of the platform and no route for the absence of one. React Router renders nothing when no route matches, and a layout element is only rendered when one of its child routes matches, so the navigation was lost as well.

**Failure class:** Logic error

**Why it was not caught:** The verification opened the addresses the application links to, and every one of them matched a route. Nothing opened an address the application does not link to, because the application had no reason to.

## Fix

**Change:** The route table ends with a catch-all route that renders a screen stating that the address names nothing, with a link back to the dashboard. The screen keeps the navigation, because it is rendered inside the layout.

**Rationale:** A reader who arrives at an address the platform does not have is a reader the platform has failed to serve, and the failure mode of a router is silence. The catch-all turns the silence into an answer, and because it is a route like any other, the navigation stays.

**Guardrail added:** A test in `web/src/routes/NotFound.test.tsx` renders the route table at the address of a task set, which matched no route before the fix, and asserts that the platform answers instead of rendering nothing.

## Verification

| Check | Method | Result |
|---|---|---|
| Original reproduction no longer reproduces | The task set address opened in the browser on the deployment | **Pass** — the screen states that the address names nothing, and the navigation is present |
| Regression test added | `web/src/routes/NotFound.test.tsx` | **Pass** |
| Adjacent behaviour unaffected | The whole route table opened in the browser | **Pass** — the six routes and the task detail route all render as before |
| Blast radius re-scored | Every route of the application opened at the deployed host | **Pass** — no address renders an empty document |

## Regression Prevention

- **Test added:** The case "answers an address that names nothing instead of rendering nothing", in `web/src/routes/NotFound.test.tsx`, renders the route table at an address that matches nothing.
- **Run gate:** The case runs in the `web-test` gate of `scripts/gate.sh`, which every change must pass.
- **Detection change:** The route table now ends with a catch-all, so an address that names nothing has one answer rather than none.

## References

- Test Report: [TR-008 Unknown Address and Render Failure Handling][tr-008]
- Related Feature: [FEAT-007 User Flows, Feature Inventory and Permissions][feat-007]

[tr-008]: <../../Test Report/TR-008-unknown-address-and-render-failure-handling.md>
[feat-007]: <../../Features/experience/user-flows-and-feature-inventory.md>
