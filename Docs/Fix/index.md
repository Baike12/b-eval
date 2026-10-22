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

---

## Agent Integration

| ID | Fix | Summary |
|---|---|---|
| [FIX-002](agent-integration/blackbox-cli-empty-action-sequence-not-flagged.md) | Black-box CLI adapter recorded an empty action sequence instead of a protocol violation | A black-box agent that produced no observable output was recorded as a legitimate zero-action trial and silently counted as completed; the trial-close validation now rejects an empty action sequence with no matching violation. |
