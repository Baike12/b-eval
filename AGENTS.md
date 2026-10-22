# AGENTS.md

The operating rules for this repository. **Read this file before writing anything** —
code, comments, documentation, commits. It is the single source of truth; there is no
second rules file.

---

## 1. Language

> **Everything written in this repository MUST be in English. English only.**

### 1.1 What this covers

| Surface | Rule |
|---|---|
| Source code | English |
| Identifiers (variables, functions, types, constants, packages, modules) | English |
| Code comments and docstrings | English |
| Tests, test names, test fixtures, sample data | English |
| Documentation under `Docs/` | English |
| Configuration files and their keys | English |
| Commit messages, branch names, PR titles and bodies | English |
| Issue titles and bodies | English |
| User-facing strings (UI labels, error messages, prompts) | English |
| Code and artefact file names | English |

### 1.2 What this forbids

**No Chinese. No other non-English language. Not anywhere.**

This includes text that "only developers will read" — a comment in a hot path, a string in a
fixture, a label in a snapshot file. Mixed-language repositories are a permanent maintenance
tax: they break search, confuse contributors, and make tooling (linters, spell-checkers, code
search, tokenizers) behave inconsistently.

### 1.3 Permitted exceptions

Only these may appear in their original form:

1. **Proper nouns and third-party names** — `SWE-bench`, `WebArena`, `τ-bench`. These are names, not prose.
2. **Direct quotations from external sources**, when the original wording is the point. Add a short English gloss if the quote carries meaning.
3. **Legal or contractual text** that must stay verbatim, with an English translation attached.
4. **The translation catalogues of the interface** — `web/src/i18n/`. The catalogue of a
   non-English locale is the product's own second language and a reader has to be able to read
   it. Its **keys stay English**; only the values are translated. Nothing outside that
   directory may hold the text of a translation.

A foreign-language *word* used because there is no established English equivalent is not an
exception — find the English term or write a description.

### 1.4 Enforcement

Before committing, verify no CJK or other non-Latin text has entered the tree:

```bash
# CJK-specific check (primary gate)
grep -rnP '[\x{4e00}-\x{9fff}\x{3400}-\x{4dbf}\x{3000}-\x{303f}\xff00-\xffef]' \
  --exclude-dir=vendor --exclude-dir=.git --exclude-dir=i18n .

# Broader non-ASCII check (secondary; catches Cyrillic, Arabic, etc.)
grep -rnP '[^\x00-\x7F]' --exclude-dir=vendor --exclude-dir=.git --exclude-dir=i18n .
```

The secondary check also flags typographic characters (curly quotes, em dashes, `…`) that are
legal in Markdown prose but should not appear inside code identifiers or source files.

`vendor/` is excluded from both checks — it is a read-only mirror of third-party code and
retains its original language. `web/src/i18n/` is excluded because it holds the translation
catalogues; the `docs-english` gate makes the same exemption and no other directory carries
one.

If a document is ever found to contain non-English text, that document is a defect: fix it in
the same change that introduced it.

---

## 2. Code Comments

> **A comment explains WHY, never WHAT. The code already says what.**

### 2.1 The rule

A comment earns its place only when it carries information the code itself cannot: a
constraint, a trade-off, a reason, a warning about a consequence. If the code makes the
reason obvious, there is nothing worth writing down.

**Write a comment when:**

- A decision looks arbitrary but is not — there is a constraint, a protocol or a measurement behind it.
- The code deliberately does something a reader would "fix" — a workaround, a pinned version, an ordering requirement.
- A non-obvious invariant holds — this value is always in range, this is always called before that.
- The obvious implementation is wrong here — a trap: an off-by-one, a swallowed exception, a silent fallback.
- A cost is being paid deliberately — a retry, a lock, a synchronous call — that a future reader might optimise away.
- The requirement lives in a design document and the code cannot carry the whole story.

**Never write a comment that:**

- Restates the code — `# increment i` above `i += 1`.
- Narrates the next line — `# now call the API`.
- Records where the code came from — "added by X in PR #123". That is what git history is for.
- Apologises or speculates — `# not sure why this works`, `# maybe refactor later`.
- Describes the change instead of the code — describe the final behaviour, not the diff.

### 2.2 What a good comment looks like

```python
# Good: states the constraint, the reason, and where the full rule lives.
# The judge may only return C / P / I; a free-form score would break the
# pass^k denominator, so partial credit is mapped here rather than upstream.
# See FEAT-004 "Scoring and Metrics".
JUDGE_SCORES = {"C": 1.0, "P": 0.5, "I": 0.0}
```

```python
# Bad: restates the code.
# Map the judge scores
JUDGE_SCORES = {"C": 1.0, "P": 0.5, "I": 0.0}
```

```python
# Bad: narrates the next line.
# Retry the call three times
for attempt in range(3):
    ...
```

### 2.3 House rules

- **If a comment needs more than three lines, change the code instead.** A long comment is a
  sign the intent did not fit the design.
- **A comment must be true.** When the code around it changes, the comment changes in the same
  change. A stale comment is worse than no comment, because it gets believed.
- **One line, one reason.** Link out to the document that holds the full story
  (`# See FEAT-004`) rather than pasting the story inline.
- **Comments are written in the same language as the code** — English, per §1.
- **Docstrings follow the same rule.** State the contract, the invariants and the failure
  modes — not a narration of the steps.
- **Prefer naming over commenting.** If a name can carry the reason, name it that way and
  leave the comment out.

---

## 3. Documentation

### 3.1 Structure

```
Docs/
├── README.md                     # Index of all documentation
├── Feature-Template.md           # Template for feature requirement documents
├── Fix-Template.md               # Template for bug fix documents
├── Features/                     # Feature requirement documents — the functional design lives here
│   ├── index.md                  # Index of every feature document
│   ├── platform/
│   │   ├── platform-scope-and-positioning.md
│   │   └── delivery-roadmap-and-acceptance-criteria.md
│   ├── task-system/
│   ├── agent-integration/
│   ├── execution/
│   ├── scoring/
│   ├── reporting/
│   ├── security/
│   └── experience/
├── Fix/                          # Bug fix documents
│   ├── index.md                  # Index of every fix document
│   └── agent-integration/
├── Tech Design/                  # Technical design documents
│   └── TD-<nnn>-<slug>.md
└── Test Report/                  # Verification / test report documents
    └── TR-<nnn>-<slug>.md
```

**The functional design is the body of `Docs/Features/`.** There is no separate chapter
series: each capability area of the platform is one Feature document, written to
`Docs/Feature-Template.md`. `FEAT-007` holds the functional requirement inventory that the
other features reference.

### 3.2 Where to write what

| You are writing | Put it in | Template |
|---|---|---|
| A new capability / requirement | `Docs/Features/<category>/<feature-name>.md` | `Docs/Feature-Template.md` |
| A change to the functional design | The matching `Docs/Features/<category>/<feature-name>.md` document | `Docs/Feature-Template.md` |
| A defect and its fix | `Docs/Fix/<category>/<fix-name>.md` | `Docs/Fix-Template.md` |
| How something is built | `Docs/Tech Design/TD-<nnn>-<slug>.md` | — |
| Proof that something works | `Docs/Test Report/TR-<nnn>-<slug>.md` | — |

### 3.3 Categories

**One category of requirement lives in one folder.** `Docs/Features/` is organised by
capability area, and `Docs/Fix/` uses the same category names, so a fix sits next to the
feature area it affects.

| Category | Covers |
|---|---|
| `platform` | Scope, positioning, goals, roles, concept model, delivery plan |
| `task-system` | Capability taxonomy, task authoring, task sets, difficulty, validation gates |
| `agent-integration` | The agent adapter protocol and its self-check |
| `execution` | Evaluation plans, orchestration, budget, concurrency, resume, traces |
| `scoring` | Scorers, verdicts, metrics, aggregation |
| `reporting` | Artefacts, run reports, comparison, attribution, leaderboard |
| `security` | Anti-cheating, adversarial evaluation, quality assurance, contamination, platform health |
| `experience` | User journeys, screens, interaction principles, permissions, requirement inventory |

A new feature goes into the folder of the capability area it belongs to. If it does not fit
any existing category, add a new category folder and list it in `Docs/Features/index.md`.

### 3.4 File naming

**Document file names carry the requirement name, not the ID.**

- Feature document: `Docs/Features/<category>/<feature-name>.md` — e.g. `scoring/scoring-and-metrics.md`
- Fix document: `Docs/Fix/<category>/<fix-name>.md` — e.g. `agent-integration/blackbox-cli-empty-action-sequence-not-flagged.md`

No `FEAT-001-` / `FIX-002-` prefix in the file name. Lower-case, hyphen-separated, and the
same words as the document title. The ID (`FEAT-001`, `FIX-002`) lives in the document's
metadata table and is what cross-references cite.

### 3.5 Templates

- **Templates live directly in `Docs/`, never inside a sub-folder.** `Docs/Feature-Template.md` and `Docs/Fix-Template.md` are the single source of truth.
- Every document in `Docs/Features/` **must** be based on `Docs/Feature-Template.md`.
- Every document in `Docs/Fix/` **must** be based on `Docs/Fix-Template.md`.
- Deviating from a template requires a reason recorded in the document itself.
- Templates are deliberately small. A template that nobody fills in is worse than no template.

### 3.6 Cross-references (mandatory)

| Document | Must reference |
|---|---|
| `Docs/Features/FEAT-*.md` | Its **Tech Design** document **and** its **Test Report** document |
| `Docs/Fix/FIX-*.md` | Its **Test Report** document |
| `Docs/Tech Design/TD-*.md` | The feature or fix it implements |

A Feature document that has not been designed or verified yet **must still carry a
`References` section**, stating `Tech Design: Not yet created.` and
`Test Report: Not yet created.` for each missing document. An explicit gap is acceptable;
a missing section is not. Once the feature reaches `Implemented`, both entries must be
live links.

References use **Markdown reference-style links** with link definitions at the bottom of the file:

```markdown
## References

- Tech Design: [TD-001 Unified Agent Adapter Protocol][td-001]
- Test Report: [TR-001 Unified Agent Adapter Protocol][tr-001]

[td-001]: <../../Tech Design/TD-001-unified-agent-adapter-protocol.md>
[tr-001]: <../../Test Report/TR-001-unified-agent-adapter-protocol.md>
```

Angle brackets are **required** in the path because `Tech Design/` and `Test Report/` contain spaces.

A Feature document without both references is **incomplete**. A Test Report that is not
referenced from a Feature or Fix document is **orphaned**.

### 3.7 Indexes

**Every folder that holds documents has an `index.md`.**

- `Docs/Features/index.md` — lists every feature document, grouped by category.
- `Docs/Fix/index.md` — lists every fix document, grouped by category.
- The index is the entry point for that folder: a reader arrives there and finds every document.

**The index covers only the documents in its own folder.** Tech Design and Test Report
documents are **not** indexed separately — each one is reachable from the `References`
section of the feature or fix it belongs to, so indexing them again would create a second
place to keep in sync.

Rules:
- Every document in the folder **must** appear in the index. A document missing from the
  index is a defect.
- The index lists the document ID, its file name, and a one-line summary.
- The index is updated in the same change that adds or removes a document.
- Index entries use **inline links** — a table of links is more usable inline. Cross-references
  *inside* documents use reference-style links, as in §3.6.

### 3.8 Identifier namespaces

Two identifier families coexist and must not be confused:

| Prefix | Meaning | Defined in |
|---|---|---|
| `F-<module>-<nn>` | Functional requirement (capability that must exist) | `Docs/Features/experience/user-flows-and-feature-inventory.md` |
| `FEAT-<nnn>` | Feature document (a unit of delivered work) | `Docs/Features/` |
| `FIX-<nnn>` | Bug fix document | `Docs/Fix/` |
| `TD-<nnn>` | Technical design document | `Docs/Tech Design/` |
| `TR-<nnn>` | Test report document | `Docs/Test Report/` |

A Feature document **must** list the functional requirement IDs it implements, in an
`Implements` field. This is how a reader goes from "what must exist" to "how it is being
built" to "whether it works".

---

## 4. Version Control

- **Never run `git push`** — including force pushes and PR creation — without an explicit instruction in the current conversation. A local commit is fine; publishing is not.
- `vendor/` holds read-only third-party mirrors and is excluded by `.gitignore`. Never add it to version control.
- Commit messages are English, imperative mood, one line of subject, body only when it explains *why*.

---

## 5. Working Method

1. **Read before writing.** Check `Docs/` for an existing document on the topic; extend it rather than creating a near-duplicate.
2. **Functional design is separate from technical design.** The Feature documents describe *what* the system does. Technical choices live in `Docs/Tech Design/`.
3. **Every feature ships with a test report.** If work is not verified, it is not done.
4. **Cite the evidence.** When a design decision traces to an upstream project, link it. When a metric is defined, give the formula and the denominator.
5. **Prefer precise language.** No "probably", "maybe", "should be fine". State the rule and its threshold.
6. **Leave the code better explained than you found it** — not better commented, but better *understood*.
