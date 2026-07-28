---
name: artifact-value-gate
description: Apply before creating any permanent artifact in this repository — document, concept, decision record, layer, script, skill, framework, policy or abstraction. Decides create / extend an existing owner / record a decision / add a rule / automate a check / proceed to implementation / stop. Use whenever a task would add something permanent.
---

# Artifact Value Gate

Operationalises the central invariant of `docs/project/project-constitution.md`:

> Every new artifact must either reduce uncertainty about implementation or directly enable implementation. Otherwise, do not create it.

## Route to (do not restate)

| Question | Document |
|---|---|
| Engineering philosophy, simplicity rule | `docs/project/project-constitution.md` |
| Entry criteria, E0 non-duplication test | `meta/standard/002-entry-criteria.md` |
| Responsibility budget | `meta/standard/010-responsibility-budget.md` |
| Who owns which concept | `meta/007-ownership-map.md` |
| Lifecycle and decision rules | `meta/standard/001-document-lifecycle.md`, `governance/decisions/` |

Read `meta/007` every time. Do not recall ownership from memory.

## When NOT to use

Editing inside an artifact that already exists and already has an owner · answering a question · applying an already-recorded decision.

## Gates

**Gate 1 — implementation value.**
*Which specific implementation decision becomes possible, safer or unambiguous because this exists?*
- Names a concrete decision → continue
- Cannot name one → **STOP, reject**

"Completeness", "symmetry", "might be useful", "another project has one" are **not** implementation value.

**Gate 2 — existing owner.** Read `meta/007`.
- Owner exists → **extend that document**, exit
- No owner → continue
- Two plausible owners → **STOP**, escalate via `meta/002` §6.3

**Gate 3 — adjacent owner.** Can a neighbouring document carry this within its responsibility budget?
- Yes → **extend it**, exit
- No → continue

**Gate 4 — artifact form.** Exactly one:

| Form | Test | Destination |
|---|---|---|
| Corpus document | defines product truth or policy | corpus → `corpus-authoring` |
| Decision record | changes normative behaviour · resolves a real contradiction · affects several downstream artifacts · costly to reverse · must outlive this task | `governance/decisions/` |
| Rule | one line, always active, no procedure | `CLAUDE.md` |
| Deterministic check | computable from files without judgement | `.claude/scripts/validate.py` |
| Skill | reusable procedure with real gates | `.claude/skills/` |
| Nothing | routine implementation choice | proceed to implementation |

A routine choice does not become a decision record because a decision was made. If nobody would revisit it in six months, it is not a record.

**Gate 5 — permanent cost.** State each: cognitive load · ownership · maintenance · review effort · discoverability · AI context cost · migration cost · duplicate-authority risk.
- Gate 1's value clearly exceeds the sum → create
- Otherwise → **STOP**, name the dominating cost

**Gate 6 — registration and lifecycle.** For a corpus artifact name: location · owner · reading position · dependencies · registries to update in the same change · lifecycle entry point.
- Any of these requires inventing a layer or a conflicting owner → **STOP**

## Output

Recommendation · deciding gate · the concrete implementation value from gate 1 · chosen form · permanent costs · next routing skill or existing owner.

**Do not create the artifact while running this skill.**

## Stop conditions

Gate 1 or 5 fails · two owners claim the concept · creation would reopen an accepted decision · the artifact belongs to a workflow stage not yet reached.

## Anti-patterns

- Symmetry as justification — *"that layer has an overview, so this one needs one"*
- Speculative abstraction with no named consumer
- A decision record for a routine implementation choice
- A new document where `meta/007` already names an owner
- Enforcing a deterministic check in prose instead of `validate.py`
- An artifact created only to close a cosmetic gap
- A skill for a workflow stage the project has not reached

## Examples

**Create.** *"We need a capability map."* → G1 makes capability→goal traceability decidable · G2 no owner · G4 corpus document · G5 accepted → **create**, route to `corpus-authoring`.

**Reject.** *"Add an engineering-principles document."* → G1 names no implementation decision it unblocks → **STOP**. Record the gap instead.

**Extend.** *"We need a rule for numeric prefixes after a split."* → G2 `language/002-naming-rules.md` owns naming → **extend it**.

**Escalate.** *"Is retention a business goal?"* → G4 decision record, but the choice is the Product Owner's, not an inference → **STOP**, escalate.
