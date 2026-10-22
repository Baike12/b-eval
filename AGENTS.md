# AGENTS.md

Project-level working rules for this repository.

> The full operating rules — language policy, document templates, cross-reference
> requirements, identifier namespaces — live in [`Agent.md`](./Agent.md).
> **Read `Agent.md` before writing anything.** This file only adds the
> repository-specific details that sit alongside it.

## Version Control

- **Never run `git push`** — including force pushes and PR creation — without an explicit instruction in the current conversation. A local commit is fine; publishing is not.
- `vendor/` holds read-only third-party mirrors and is excluded by `.gitignore`. Never add it to version control.
- Commit messages are English, imperative mood, one line of subject.

## Where to Write What

| You are writing | Put it in | Template |
|---|---|---|
| A new capability / requirement | `Docs/Features/<category>/<feature-name>.md` | `Docs/Feature-Template.md` |
| A change to the functional design | The matching `Docs/Features/<category>/<feature-name>.md` document | `Docs/Feature-Template.md` |
| A defect and its fix | `Docs/Fix/<category>/<fix-name>.md` | `Docs/Fix-Template.md` |
| How something is built | `Docs/Tech Design/TD-<nnn>-<slug>.md` | — |
| Proof that something works | `Docs/Test Report/TR-<nnn>-<slug>.md` | — |

**The functional design is the body of `Docs/Features/`.** There is no separate chapter
series — each capability area of the platform is one Feature document, filed in one folder
per category. `FEAT-007` holds the functional requirement inventory (`F-XXX-NN`) that the
other features declare in their `Implements` field.

**File names carry the requirement name, not the ID** — `scoring/scoring-and-metrics.md`,
not `FEAT-004-scoring-and-metrics.md`. The ID lives in the document's metadata table.

## Indexes

`Docs/Features/index.md` and `Docs/Fix/index.md` are the entry points for their folders.

- **Every document in a folder must appear in that folder's `index.md`.** A document missing
  from the index is a defect.
- The index lists the document ID, its file name, and a one-line summary.
- **Only the documents in the folder are indexed.** Tech Design and Test Report documents are
  not indexed separately — each is reachable from the `References` section of the feature or
  fix it belongs to, so indexing them again would create a second place to keep in sync.
- Update the index in the same change that adds, removes or renames a document.

## Mandatory Cross-References

- Every `Docs/Features/FEAT-*.md` links to **both** its Tech Design document and its Test Report document.
- Every `Docs/Fix/FIX-*.md` links to its Test Report document.
- Use Markdown reference-style links with the definitions at the bottom of the file, wrapped in angle brackets because `Tech Design/` and `Test Report/` contain spaces.

## Language

**English only.** Source code, identifiers, comments, tests, fixtures, configuration, documentation, commit messages — all English. See `Agent.md` §1 for the full policy, the permitted exceptions, and the check command.

## Ground Rules

- Read `Docs/` before writing; extend an existing document rather than creating a near-duplicate.
- Functional design (what) is separate from technical design (how).
- No feature ships without a test report.
- Cite evidence: link the upstream project for a borrowed design, and give the formula and denominator for every metric.
