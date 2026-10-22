# TD-002: Task Forms and the Task Metadata Contract

| Field | Value |
|---|---|
| **ID** | TD-002 |
| **Status** | Draft |
| **Implements** | [FEAT-002 Capability Taxonomy and Task System](../Features/task-system/capability-taxonomy-and-task-system.md) |
| **Owner** | Platform team |
| **Created** | 2026-10-22 |
| **Last Updated** | 2026-10-22 |

> This document specifies *how* the task form and the metadata contract are built. The *what* and *why* live in the Feature document it implements.

## 1. Problem Restatement

The metadata contract of FEAT-002 §2.2 requires nine field groups, of which two — `Initial state` and `Interaction contract` — only make sense for a task that runs against a live environment. A batch of question-and-answer data has neither. Two ways out exist: relax the contract so the groups become optional, or add a second form so that each task declares which contract it satisfies.

Relaxing the groups to optional was rejected: an optional group means the platform must handle "task with no action space" and "task with an action space" in the same code path, and every downstream consumer — the runner, the scorer, the report — needs a branch whose only purpose is to tolerate a missing field. Worse, it makes "this task has no environment" indistinguishable from "the author forgot to fill in the environment".

## 2. Design Goals

| Goal | Consequence for the design |
|---|---|
| A question-and-answer data set is a first-class task set | A second form, `static-qa`, with its own contract and its own gate subset. |
| Both forms score on one scale | `static-qa` and `environment-agentic` share the Trial → Trajectory → Verdict → Metric pipeline; only the environment bring-up differs. |
| No field is maintained that nobody reads | A task is rejected when it carries a field group its form does not use. |
| The form is checkable, not descriptive | Form enforcement is a validation rule with a rejection message, not a convention. |

## 3. Non-Goals

- The form does not change what a Trial is, what a Verdict is, or what a Metric is.
- The form does not decide how a task is scored. The scorer catalogue is form-independent; a `static-qa` task uses the rule-criteria and rubric scorers, an `environment-agentic` task may use any of the six.
- The form is not a version tag. A task's form is a property of the task, not of the release.

## 4. Form Model

```go
type TaskForm string

const (
    // TaskFormEnvAgentic runs the agent against a live environment and reads
    // the verdict from the environment's final state.
    TaskFormEnvAgentic TaskForm = "environment-agentic"

    // TaskFormStaticQA gives the agent a prompt and reads the verdict from the
    // answer text. No environment is brought up.
    TaskFormStaticQA TaskForm = "static-qa"
)
```

The form is stored on the task and is immutable once the task has been admitted. Changing a task's form means publishing a new task version: the old version's scores refer to a task that behaved differently.

## 5. Metadata Contract

### 5.1 Field groups

The contract is expressed as a set of field groups, each with a per-form requirement. A requirement is one of:

| Requirement | Meaning |
|---|---|
| `required` | The group must be present and non-empty. |
| `optional` | The group may be absent. |
| `forbidden` | The group must be absent. |
| `partial` | Only a named subset of the group's fields is required. |

| Field group | `environment-agentic` | `static-qa` |
|---|---|---|
| `identity` | required | required |
| `capability` | required | required |
| `description` | required | required |
| `budget` | required | required |
| `initial_state` | required | forbidden |
| `interaction_contract` | required | partial — `termination_conditions` and `multi_turn` only |
| `judgement` | required | required |
| `reference_material` | optional | forbidden |
| `meta` | required | required |

### 5.2 Identity group

`task_id`, `display_name`, `benchmark`, `task_set_version`, `form`.

`task_id` is globally unique and never reused. `form` is written once at creation; the store refuses a write that changes it.

### 5.3 Capability group

`level1_dimensions` (at least one), `level2_labels`, `difficulty`, `domains`.

### 5.4 Description group

`instruction` is the text handed to the agent. `evaluator_note` is optional and never reaches the agent.

### 5.5 Budget group

`steps`, `tokens`, `duration_seconds`, `cost_usd`, `errors`. A cap of zero means "no cap for this dimension".

### 5.6 Initial state group (`environment-agentic` only)

`snapshot_id`, `initial_data`, `initial_files`, `initial_credentials`, `seed`.

### 5.7 Interaction contract group

`action_space`, `observation_shape`, `termination_conditions`, `multi_turn`.

For `static-qa` only `termination_conditions` and `multi_turn` are accepted; `action_space` and `observation_shape` are forbidden.

### 5.8 Judgement group

`scorers` (a list of scorer specs), `combination` (`and` / `or`), `weights`, `primary` (the index of the primary criterion), `target_state` **or** `reference_answer`.

`reference_answer` is the `static-qa` counterpart of `target_state`: the text the answer is compared against. It is a judgement condition and is never included in the prompt.

### 5.9 Reference material group (`environment-agentic` only)

`reference_solution`, `reference_trajectory`, `policy_document`.

### 5.10 Meta group

`author`, `source`, `difficulty_basis`, `estimated_human_minutes`, `estimated_agent_minutes`, `created_at`, `updated_at`, `deprecated`.

## 6. Form Enforcement

Validation runs at task creation and at task-set admission. The rule is declarative: a table of group → per-form requirement, plus a per-field rule for the `partial` case.

```
for each field group g in task:
    want := requirement[g][task.form]
    switch want:
    case forbidden:  if group present  -> reject "field group %q is not used by form %q"
    case required:   if group absent   -> reject "field group %q is required by form %q"
    case optional:   (no check)
    case partial:    for each field f in required_fields[g]:
                         if f absent -> reject "field %q is required by form %q"
                     for each field f in forbidden_fields[g]:
                         if f present -> reject "field %q is not used by form %q"
```

Rejection messages name the offending field or group. A task author who adds `initial_state` to a `static-qa` task is told exactly which field to remove, rather than being told "invalid task".

The requirement table is declared once, in the same package that validates against it, so the table and the validator cannot drift apart.

## 7. Form-Dependent Admission Gates

FEAT-002 §2.9 gives the per-form outcome of each of the eight gates. The implementation encodes it as a per-form gate list:

| Gate | `environment-agentic` | `static-qa` |
|---|---|---|
| 1 Solvability | reference solution executes | reference answer accepted by the scorer |
| 2 Scorer credibility | reference passes, wrong solution fails | reference answer passes, wrong answer fails |
| 3 Scorer non-triviality | required | required |
| 4 Platform fault separation | required | not applicable — no environment |
| 5 Difficulty credibility | required | required |
| 6 Scoring invisibility | required | not applicable — no environment |
| 7 Annotation completeness | required | required |
| 8 Reproducibility | required | required |

A gate that does not apply is recorded with outcome `not_applicable` and the reason `no environment`. It is **not** recorded as a pass, because "not applicable" and "passed" are different claims, and an admission report that conflates them overstates how much was actually checked.

## 8. Execution Difference

| Stage | `environment-agentic` | `static-qa` |
|---|---|---|
| Preparation | Bring up an isolated environment instance; load the snapshot; verify health | Nothing to bring up |
| Execution | The agent acts on the environment through the action space | The agent is given the prompt and produces one answer |
| Judgement | Replay the trajectory against the golden state, or read the artifact | Compare the answer text with the matchers and the rubric |
| Cleanup | Destroy the environment instance, reclaim credentials | Nothing to clean up |

A `static-qa` Trial is one agent call plus one judgement pass. It can still terminate as `model_failure`, `budget_exhausted`, `protocol_violation` or `cancelled`; it can never terminate as `platform_failure` for an environment reason.

## 9. Storage Shape

```json
{
  "task_id": "T-0001",
  "form": "static-qa",
  "display_name": "Refund eligibility",
  "benchmark": "support-faq",
  "task_set_version": "v1",
  "capability": { "level1_dimensions": ["C4"], "level2_labels": ["requirement-clarification"], "difficulty": "L1", "domains": ["D7"] },
  "description": { "instruction": "A customer asks ...", "evaluator_note": "" },
  "budget": { "steps": 1, "tokens": 4000, "duration_seconds": 60, "cost_usd": 0, "errors": 3 },
  "judgement": {
    "combination": "and",
    "primary": 0,
    "weights": [1.0],
    "reference_answer": "orders over $100 qualify for free returns",
    "scorers": [ { "type": "contains", "params": { "needles": ["30 days", "receipt"] }, "weight": 1.0, "primary": true } ]
  },
  "meta": { "author": "team", "created_at": "2026-10-22T00:00:00Z" }
}
```

The `environment-agentic` shape adds `initial_state`, `interaction_contract` and `reference_material`. The JSON decoder rejects unknown fields, so a typo in a group name is an error rather than a silently dropped value.

## 10. Alternatives Considered

| Alternative | Why rejected |
|---|---|
| Make `initial_state` and `interaction_contract` optional | Every downstream consumer needs a branch for "the field is absent", and the two situations "no environment by design" and "the author forgot" become indistinguishable. |
| Model a static-QA task as an `environment-agentic` task with a one-action environment | The environment machinery still runs per Trial, and the "environment" is a fiction that has to be maintained to satisfy the contract. |
| Keep one form and give static-QA tasks a sentinel snapshot | A sentinel is an unused field with a non-empty value; it violates the rule that a form carries only what it uses, and it defeats the form check. |
| Let the author declare a form but validate only on admission | The task sits in the store invalid until someone runs the gate; the error surfaces at the wrong time. |

## 11. Verification

Verified by [TR-003 M0 Vertical Slice][tr-003].

## 12. Open Points

- Whether a `static-qa` task set should support a per-item answer template (one prompt, many reference answers) as a separate form, or as a template instantiation of a single `static-qa` task. Deferred to M1, where task templating lands.

## References

- Feature: [FEAT-002 Capability Taxonomy and Task System][feat-002]
- Test Report: [TR-003 M0 Vertical Slice][tr-003]

[feat-002]: <../Features/task-system/capability-taxonomy-and-task-system.md>
[tr-003]: <../Test Report/TR-003-m0-vertical-slice.md>
