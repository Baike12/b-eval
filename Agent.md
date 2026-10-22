# Agent.md

Operating rules for every AI agent and human contributor working in this repository.

---

## 1. Language Policy

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
| Code and artifact file names | English |

### 1.2 What this forbids

**No Chinese. No other non-English language. Not anywhere.**

This includes text that "only developers will read" — a comment in a hot path, a string in a fixture, a label in a snapshot file. Mixed-language repositories are a permanent maintenance tax: they break search, confuse contributors, and make tooling (linters, spell-checkers, code search, tokenizers) behave inconsistently.

### 1.3 Permitted exceptions

Only these may appear in their original form:

1. **Proper nouns and third-party names** — `SWE-bench`, `WebArena`, `τ-bench`. These are names, not prose.
2. **Direct quotations from external sources**, when the original wording is the point. Add a short English gloss if the quote carries meaning.
3. **Legal or contractual text** that must stay verbatim, with an English translation attached.

A foreign-language *word* used because there is no established English equivalent is not an exception — find the English term or write a description.

### 1.4 Enforcement

Before committing, verify no CJK or other non-Latin text has entered the tree:

```bash
# CJK-specific check (primary gate)
grep -rnP '[\x{4e00}-\x{9fff}\x{3400}-\x{4dbf}\x{3000}-\x{303f}\x{ff00}-\x{ffef}]' \
  --exclude-dir=vendor --exclude-dir=.git .

# Broader non-ASCII check (secondary; catches Cyrillic, Arabic, etc.)
grep -rnP '[^\x00-\x7F]' --exclude-dir=vendor --exclude-dir=.git .
```

The secondary check also flags typographic characters (curly quotes, em dashes, `…`) that are legal in Markdown prose but should not appear inside code identifiers or source files.

`vendor/` is excluded from both checks — it is a read-only mirror of third-party code and retains its original language.

### 1.5 Legacy documents pending English migration

None. All documents under `Docs/` are in English as of 2026-10-22. The eight functional-design chapters that predate this policy were migrated in place, and their file names were changed to English at the same time.

If a document is ever found to contain non-English text, that document is a defect: fix it in the same change that introduced it.

---

## 2. Documentation Structure

```
Docs/
├── README.md                     # Index of all documentation
├── Feature-Template.md           # Template for feature requirement documents
├── Fix-Template.md               # Template for bug fix documents
├── Features/                     # Feature requirement documents — the functional design lives here
│   ├── FEAT-001-platform-scope-and-positioning.md
│   ├── FEAT-002-capability-taxonomy-and-task-system.md
│   ├── FEAT-003-evaluation-execution-and-orchestration.md
│   ├── FEAT-004-scoring-and-metrics.md
│   ├── FEAT-005-reporting-analysis-and-comparison.md
│   ├── FEAT-006-security-robustness-and-quality-assurance.md
│   ├── FEAT-007-user-flows-and-feature-inventory.md
│   ├── FEAT-008-delivery-roadmap-and-acceptance-criteria.md
│   └── FEAT-009-unified-agent-adapter-protocol.md
├── Fix/                          # Bug fix documents
│   └── FIX-<nnn>-<slug>.md
├── Tech Design/                  # Technical design documents
│   └── TD-<nnn>-<slug>.md
└── Test Report/                  # Verification / test report documents
    └── TR-<nnn>-<slug>.md
```

**The functional design is the body of `Docs/Features/`.** There is no separate chapter
series: each capability area of the platform is one Feature document, written to
`Docs/Feature-Template.md`. `FEAT-007` holds the functional requirement inventory that the
other features reference.

### 2.1 Templates

- **Templates live directly in `Docs/`, never inside a sub-folder.** `Docs/Feature-Template.md` and `Docs/Fix-Template.md` are the single source of truth.
- Every document in `Docs/Features/` **must** be based on `Docs/Feature-Template.md`.
- Every document in `Docs/Fix/` **must** be based on `Docs/Fix-Template.md`.
- Deviating from a template requires a reason recorded in the document itself.
- Templates are deliberately small. A template that nobody fills in is worse than no template.

### 2.2 Cross-references (mandatory)

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

[td-001]: <../Tech Design/TD-001-unified-agent-adapter-protocol.md>
[tr-001]: <../Test Report/TR-001-unified-agent-adapter-protocol.md>
```

Angle brackets are **required** in the path because `Tech Design/` and `Test Report/` contain spaces.

A Feature document without both references is **incomplete**. A Test Report that is not referenced from a Feature or Fix document is **orphaned**.

### 2.3 Identifier namespaces

Two identifier families coexist and must not be confused:

| Prefix | Meaning | Defined in |
|---|---|---|
| `F-<module>-<nn>` | Functional requirement (capability that must exist) | `Docs/Features/FEAT-007` |
| `FEAT-<nnn>` | Feature document (a unit of delivered work) | `Docs/Features/` |
| `FIX-<nnn>` | Bug fix document | `Docs/Fix/` |
| `TD-<nnn>` | Technical design document | `Docs/Tech Design/` |
| `TR-<nnn>` | Test report document | `Docs/Test Report/` |

A Feature document **must** list the functional requirement IDs it implements, in a `Implements` field. This is how a reader goes from "what must exist" to "how it is being built" to "whether it works".

---

## 3. Version Control

- **Never run `git push`** — including force pushes and PR creation — without an explicit instruction in the current conversation. A local commit is fine; publishing is not.
- `vendor/` holds read-only third-party mirrors and is excluded by `.gitignore`. Never add it to version control.
- Commit messages are English, imperative mood, one line of subject, body only when it explains *why*.

---

## 4. Working Method

1. **Read before writing.** Check `Docs/` for an existing document on the topic; extend it rather than creating a near-duplicate.
2. **Functional design is separate from technical design.** `Docs/01..08` describe *what* the system does. Technical choices live in `Docs/Tech Design/`.
3. **Every feature ships with a test report.** If work is not verified, it is not done.
4. **Cite the evidence.** When a design decision traces to an upstream project, link it. When a metric is defined, give the formula and the denominator.
5. **Prefer precise language.** No "probably", "maybe", "should be fine". State the rule and its threshold.
