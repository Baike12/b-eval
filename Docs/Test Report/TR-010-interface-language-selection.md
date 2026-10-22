# TR-010: Interface Language Selection

| Field | Value |
|---|---|
| **ID** | TR-010 |
| **Status** | Verified |
| **Verifies** | [TD-007 Interface Language Selection][td-007] |
| **Owner** | Platform team |
| **Created** | 2026-10-22 |
| **Last Updated** | 2026-10-22 |

> This report records what was checked, how it was checked and what the check returned. The design it verifies is specified in [TD-007][td-007].

## 1. Scope

The verification covers the language of the interface: the default language, the switch that changes it, the persistence of the choice, the translation of the closed vocabularies the server publishes, the language the document declares to the browser, and the key parity between the two catalogues.

It does not cover the translation of the content the platform stores. A task instruction, a metric definition and a primary criterion are data and are shown in the language the platform published them in, which is what the checks below assert.

## 2. Method

| Aspect | How it was checked |
|---|---|
| Default language, persistence, document language, message behaviour | The frontend test suite (`web/src/i18n/index.test.tsx`, `web/src/components/LocaleSwitch.test.tsx`) |
| Key parity between the catalogues | The `Messages` type at compile time, and a test that compares the flattened key sets |
| The English-only gate and its exemption | Two gate tests in `internal/gate/docs_test.go`, and the `docs-english` gate itself over the whole tree |
| The rendered interface | The Chromium instance attached to BrowserOS neo, reading the live DOM of the preview server (`b-eval serve` on the loopback interface with a seeded data directory) |
| The typography | The computed `font-family` and `letter-spacing` of the rendered elements |

The preview server carried the same two demonstration reports the visual verification used, so every screen was read against real data and not against a fixture.

The readings below name the catalogue key that produced the text rather than the text itself. The catalogue is the one place in the tree that holds a translation, and this report is written in English.

## 3. The Evidence Set

| Check | How it was checked | Result |
|---|---|---|
| AC1 — the interface is Chinese by default | The live DOM of the preview server, with no stored choice in the browser | The document carried `lang="zh-CN"`; the navigation read the five values of `nav.*`; the page title read the value of `document.title`; the stored choice was absent |
| AC2 — the switch offers both languages and changes the screen | The live DOM, after pressing the English option | The navigation, the page title, the section headings and the status labels changed to the English catalogue in one step; the pressed state moved from the Chinese endonym to the English one |
| AC3 — no text is left in the previous language | The live DOM of all eight screens after the switch | Every heading, table header, chip, placeholder and notice read in the new language |
| AC4 — the choice is remembered | The live DOM after a reload | The reloaded page opened in English; after the stored choice was removed, the next load opened in Chinese |
| AC5 — the closed vocabularies are translated | The live DOM of the run report | The run status was shown with the label of the `completed` entry, the coverage with the label of `no coverage`, the aggregation with the label of `macro`, the termination reason and the failure class with the label of `completed`, and the four decomposition dimensions with the labels of `domain`, `capability`, `difficulty` and `form` |
| AC6 — the document declares its language | `document.documentElement.lang` and `document.title` after every change | Both followed the chosen language |
| AC7 — an unknown value is shown unchanged | The test probe, which renders a value that no vocabulary covers | The value was returned as the server published it |
| AC8 — a message that exists in one language only fails the build | The type of the Chinese catalogue against the English one | The catalogue is typed against the reference, so a missing key is a compile error; the frontend type check passes with both catalogues complete |

## 4. The Screens That Were Read

Every screen was opened at the default language, and its text was read from the live DOM.

| Screen | Heading | What the read showed |
|---|---|---|
| Dashboard | `dashboard.title` | The status strip read `dashboard.taskSets`, `dashboard.agentsRegistered` and `dashboard.serverTime`; the form read `dashboard.taskSet`, `dashboard.agent`, `dashboard.repetitions` and `dashboard.aggregation`, with the two aggregation options of `dashboard.macroOption` and `dashboard.microOption` |
| Task sets | `taskSets.title` | The facts read `taskSets.version`, `taskSets.benchmark`, `taskSets.tasks` and `taskSets.frozen`; the table headers read the six entries of `taskSets.table` |
| Task detail | the task identifier of the platform | The facts read `taskDetail.form`, `taskDetail.benchmark`, `taskDetail.version` and `taskDetail.difficulty`; the sections read `taskDetail.instruction` and `taskDetail.contractFields`; the contract rows read the fourteen entries of `taskDetail.*` |
| Agents | `agents.title` | The form read `agents.agentID`, `agents.shape` and `agents.endpoint`; the table headers read `agents.agent`, `agents.shape`, `agents.selfCheck`, `agents.violations` and `agents.usageReporting`; the self-check chip read the label of the `passed` entry; the usage column read the label of the `ok` entry |
| Runs | `runs.title` | The table headers read the seven entries of `runs.*`; the rows read the labels of `macro` and `completed` |
| Metrics | `metrics.title` | The table headers read the eight entries of `metrics.*`; the rows read `common.builtIn`, the label of `ratio` and the labels of `macro` and `micro` |
| Run report | `report.title` | The sections read `confidence.title`, `report.mainScore`, `report.metrics`, `report.breakdown`, `report.tasks` and `failures.title`; the confidence section read `confidence.platformFailureRate`, `confidence.judgementFailureRate` and `confidence.validSampleSize`, and the note read `confidence.limitLabel` and `confidence.limitNote`; the failure section read `failures.classes`, `failures.reasons`, `failures.scorers` and `failures.failedTrials` |
| Unknown address | `notFound.title` | The screen named the address, read `notFound.empty` and `notFound.hint`, and offered `notFound.back` |

The content the platform holds stayed in the language it was published in: the task instruction, the metric names and definitions, the numerator and the denominator of every metric, and the primary criterion of the main score. That is the stated scope of the feature and not an omission.

## 5. The Deployed Platform

The feature was deployed to the platform's host as the image `b-eval:20261022-ui-i18n2`, served under the base path `/eval`, and the same readings were taken there through the nginx route.

| Check | Result |
|---|---|
| The default language of the deployed platform | The document carried `lang="zh-CN"`; the navigation read the five Chinese labels; the page title read the Chinese title; the stored choice was absent |
| The switch on the deployed platform | Pressing the English option changed the navigation, the title and every heading in one step and stored the choice; pressing the Chinese option changed them back |
| The run report on the deployed platform | The sections read the Chinese headings; the confidence section read the three Chinese gauge labels and the Chinese threshold note; the failure section read the Chinese headings |
| The error path on the deployed platform | The navigation stayed in Chinese and the notice showed the server's own message, which names the run that was asked for |
| The other containers of the host | Unchanged: the four pre-existing containers kept their uptime, and only the platform's own container was recreated |

## 6. The Typography
The rendered body carried the stack `Archivo, ui-sans-serif, system-ui, -apple-system, "Segoe UI", "PingFang SC", "Hiragino Sans GB", "Microsoft YaHei", "Noto Sans CJK SC", "Source Han Sans SC", sans-serif`: a Latin word comes from the self-hosted Archivo, a Chinese word from the system font of the reader.

The page title of the Chinese document carried `letter-spacing: normal`, where the Latin title carries `-0.02em`. A Chinese glyph is square and dense already, and the negative tracking of the Latin scale would crowd it.

## 7. The Gate

The `docs-english` gate scans every source file of the tree for CJK text. It passes over the repository with the Chinese catalogue in place, because the catalogue is the one text in the tree that is not English, and it fails on Chinese text anywhere else. Two gate tests hold that exemption: the catalogue passes, and a screen that carries Chinese text fails.

The repository's own rule was amended in the same change: `AGENTS.md` §1.3 now names the translation catalogue as the fourth permitted exception, and §1.4 excludes the same directory from the two manual checks.

## 8. What This Report Does Not Claim

- It does not claim that the English wording of every label is unchanged. The closed vocabularies now read as words rather than as identifiers, and the wording of a message may differ from the wording of the label it replaced.
- It does not claim that a translation is reviewed. The Chinese catalogue was written by hand and no second reader checked it against the English one; the guarantee the build gives is that no message is missing, not that no message is wrong.
- It does not claim that a browser that refuses its storage remembers the choice. The interface speaks the chosen language for the session and returns to the default on the next visit.

## 9. Verdict

Every acceptance criterion of [FEAT-010 Interface Language Selection][feat-010] was observed on the running platform, and the repository's own gate passes with the feature in place.

[td-007]: <../Tech Design/TD-007-interface-language-selection.md>
[feat-010]: <../Features/experience/interface-language-selection.md>
