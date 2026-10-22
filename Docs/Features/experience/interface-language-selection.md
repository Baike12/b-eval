# FEAT-010: Interface Language Selection

| Field | Value |
|---|---|
| **ID** | FEAT-010 |
| **Status** | Implemented |
| **Owner** | Platform team |
| **Priority** | P1 |
| **Implements** | `F-EXP-01` |
| **Created** | 2026-10-22 |
| **Last Updated** | 2026-10-22 |
| **Source** | Functional design chapter 07 — User Flows and Feature Inventory |

## Summary

The interface speaks the language of the reader. Chinese is the language it speaks until a reader says otherwise, English is one click away, and the choice is remembered in the reader's browser. Every label, heading, table header, chip, placeholder, empty state and notice follows the choice, and so do the closed vocabularies the server publishes — a run status, a coverage state, a termination reason.

## Problem

**Current behaviour:** Every screen of the interface is written in English. The navigation, the table headers, the confidence section, the failure analysis and every empty state are English sentences, and the reader who does not read English cannot read a report at all.

**Pain:** The platform's readers are the people who write the tasks and read the reports. A reader working in Chinese has to translate a dense measurement interface — a confidence section, a decomposition by domain, a failure classification — in their head, and the translation of a measurement is where numbers get misread. The cost falls on every reader of the platform and on every task author who has to explain the interface to a colleague.

**Evidence:** The interface language was raised as a requirement by the platform's own readers. No measurement of the cost exists; the pain is the reading of a report in a language the reader does not speak.

## Proposed Solution

The interface carries a language switch in its header and speaks Chinese by default. Pressing the switch changes every screen at once, without a reload and without losing the address the reader is on. The choice is stored in the reader's browser and restored on the next visit, so the switch is asked once and not on every page. The content the platform holds — a task instruction, a metric definition, an agent name, the primary criterion of a task — is data rather than interface chrome and is shown as the platform published it, in every language.

## Scope

### In Scope

- The language of every piece of interface text: navigation, headings, labels, table headers, chips, placeholders, buttons, empty states, notices and the text of a render failure.
- The closed vocabularies the server publishes: run status, self-check status, coverage state, usage reporting, aggregation method, metric unit, termination reason, failure class and decomposition dimension.
- A switch in the header that offers Chinese and English by the names those languages give themselves.
- The language of the document itself: the `lang` attribute, the page title, the font the browser picks and the line-breaking rules it applies.
- The persistence of the choice in the reader's browser.

### Out of Scope

- **Translating the content the platform stores.** A task instruction, a metric definition, an agent name and a primary criterion are the data of the evaluation, not the chrome around it. Translating them would change what the platform published and would make a report exported from the platform differ from the report on the screen.
- **Translating the wire contract.** The API and the exported report keep their canonical vocabulary. A report is the same report in every language, and a script that reads the API is not broken by the language of the reader.
- **A per-account language.** The platform has no accounts yet, and the language of the interface is a property of the reader's session rather than of the data the platform holds.
- **Languages beyond Chinese and English, and right-to-left layout.** Both are one catalogue entry and one layout rule away, and neither is required by a reader today.

## Requirements

- **R1** The interface must speak Chinese by default, before the reader has expressed any preference.
- **R2** The reader must be able to change the language of the interface from any screen, without a reload and without losing the address they are on.
- **R3** The change must apply to every piece of interface text at once: no screen, no section and no notice may be left in the previous language.
- **R4** The choice must be remembered in the reader's browser and restored on the next visit.
- **R5** The closed vocabularies the server publishes must be presented in the language of the interface.
- **R6** The document must declare its language to the browser, so that the font, the line breaking and the screen reader follow the language of the interface.
- **R7** A value that no vocabulary covers must be shown exactly as the server published it.
- **R8** Every message of the interface must exist in both languages; a message that exists in one language only must fail the build.

## Acceptance Criteria

- [ ] **AC1** (R1) A browser that has stored no choice shows the whole interface in Chinese, including the navigation, the table headers, the empty states and the run status chips.
- [ ] **AC2** (R2) The header carries a control that offers Chinese and English under the names those languages give themselves, and pressing it changes the language of the screen that is open.
- [ ] **AC3** (R3) After the switch, no visible text of the screen remains in the previous language.
- [ ] **AC4** (R4) The choice survives a reload of the page and a visit to another address of the platform.
- [ ] **AC5** (R5) A run status, a coverage state and a termination reason are shown in the language of the interface.
- [ ] **AC6** (R6) The `lang` attribute of the document carries the chosen language and the page title follows it.
- [ ] **AC7** (R7) A value that no vocabulary covers is displayed unchanged.
- [ ] **AC8** (R8) The build fails when a message exists in one language only.

## Impact

| Area | Effect |
|---|---|
| Users | Chinese-speaking readers get the interface in their own language. English-speaking readers are unaffected: English is the second language and the content of the platform is unchanged in both. |
| Existing behaviour | Every screen is rendered through a catalogue instead of through literals. The English wording of a label is the wording the screen already showed, except for the closed vocabularies, which now read as words rather than as identifiers. |
| Operational cost | One catalogue per language, about 190 messages. The bundle grows by the size of the Chinese catalogue, a few kilobytes, because the catalogue is plain text in the bundle. |

## Risks and Mitigations

| Risk | Likelihood | Mitigation |
|---|---|---|
| A message is added to one language only and the reader sees a key. | Med | The Chinese catalogue is typed against the English one, so a missing key is a compile error and not a blank label. |
| A translation changes the meaning of a measurement. | Low | The values that carry a number — the threshold, the coverage, the counts — are not translated; only the words around them are. |
| The reader's choice is lost because the browser refuses storage. | Low | The choice still applies to the session in memory, and the interface falls back to the default language on the next visit. |

## Rollout

The feature ships with the interface it applies to. The default language changes for every reader who has stored no choice; a reader who has stored a choice keeps it. The rollback is the removal of the switch and of the catalogue, which leaves the English literals in place.

## References

- Tech Design: [TD-007 Interface Language Selection][td-007].
- Test Report: [TR-010 Interface Language Selection][tr-010].
- The requirement this feature implements: [FEAT-007 User Flows, Feature Inventory and Permissions][feat-007] (`F-EXP-01`).

## Open Questions

| # | Question | Owner | Blocking? |
|---|---|---|---|
| 1 | The catalogue is written by hand; the platform has no translation tooling, and a review of a Chinese catalogue against the English one is a human task. | Platform team | No |

<!-- Link definitions stay at the bottom of the file. -->

[td-007]: <../../Tech Design/TD-007-interface-language-selection.md>
[tr-010]: <../../Test Report/TR-010-interface-language-selection.md>
[feat-007]: <./user-flows-and-feature-inventory.md>
