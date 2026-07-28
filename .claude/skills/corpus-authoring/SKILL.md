---
name: corpus-authoring
description: Use when authoring or substantially modifying a normative document in this repository (product/, meta/, domain/, language/, experience/, features/, design/, roadmap/, governance/, docs/project/). Routes to the correct authoring-pipeline stage, determines the permitted status ceiling from actual repository state, and enforces stop conditions. Run only after artifact-value-gate returns create or extend.
---

# Corpus Authoring

Routing and gating. **The pipeline is specified in `meta/standard/017-authoring-pipeline.md` — read it; this skill does not contain it.**

## Route to (do not restate)

| Question | Document |
|---|---|
| Stages P0–P9, gates, recovery strategies, re-entry by change class | `meta/standard/017-authoring-pipeline.md` |
| Entry criteria E0–E8 | `meta/standard/002-entry-criteria.md` |
| Exit criteria X1–X12 | `meta/standard/003-exit-criteria.md` |
| Assertion classes, traceability, derivation table | `meta/standard/006-traceability.md` |
| Document class and its characteristic failure mode | `meta/standard/018-document-class-profiles.md` |
| Class per directory, role per stage, artifact storage | `meta/012-authoring-binding.md` |
| Concept ownership | `meta/007-ownership-map.md` |
| Frontmatter and section contract, dependency-status rule | `meta/003-document-template.md` |
| States, transitions, versioning | `meta/standard/001-document-lifecycle.md` |

Read only what the current stage needs. Loading all nine is the failure this skill prevents.

## When NOT to use

`PATCH` edits (wording, links, typos) · reading or answering · reviewing (use `corpus-review`) · before `artifact-value-gate` returned *create* or *extend*.

## Three separate questions — never collapse them

Upstream status does **not** simply stop work. Determine each independently, from the repository:

1. **Is production permitted?** Read the governing binding and pipeline rules.
2. **What is the highest status currently permitted?** — the *ceiling*.
3. **Is approval permitted?** Read `meta/003` §6.3 and `meta/standard/001` transitions against the *actual* `Status` of each dependency.

If drafting is allowed but approval is blocked: **do the work up to the ceiling, state the ceiling explicitly, and preserve the blocker.** Stop only when production itself is prohibited.

**Never hardcode any document's status.** Read `Status:` from the file every time.

## Procedure

1. **Classify the change** — new / `MAJOR` / `MINOR` / `PATCH`.
2. **Find the pipeline entry point** for that class — `DES-017` §6.11. A `MINOR` addition enters at the assertion stage, not at prose.
3. **Stage eligibility** — apply the three questions above. Record the ceiling.
4. **Ownership** — read `meta/007`. Owned elsewhere → extend that document. Two claimants → **STOP**, escalate.
5. **Source dossier** — list every candidate source with its version and actual `Status`. Apply admissibility below.
6. **Assertion registry** — every assertion: one sentence, class, specific source, version. **No prose before the owning role approves it.**
7. **Separate Product Owner decisions** — see below.
8. **Structure**, then **prose**. Prose states the registry and nothing more. A needed assertion discovered while writing → **return to the registry stage**; do not add it to prose.
9. **Self-check** — run `.claude/scripts/validate.py`; diff prose against the registry; produce the non-verification list, which must not be empty by default.
10. **Hand off** — set the status to the permitted ceiling and stop.

## Source admissibility

| State | Use as normative evidence |
|---|---|
| `Approved`, `Canonical` | yes |
| recorded decision | yes |
| `Review` | only with the risk stated explicitly |
| `Draft`, `Scaffold`, `Superseded`, `Disputed` | **no** |
| observation | only with source and confirmation date |
| hypothesis | never — mark it and register it |

An empty `Scaffold` is never evidence. When relevant material is inadmissible: exclude it from derivation and **state the exclusion in the document** where the corpus requires it. Never let it become truth by omission.

## Product Owner decisions

Stop and request one when an assertion cannot be derived from admissible sources, selects between materially different product futures, or changes accepted strategic intent. **Do not disguise a decision as an architectural inference.**

## Output

Only what the pipeline and the current ceiling permit: the document at its permitted status, its derivation table, a validator report, and an explicit list of what was not verified.

**This skill must never:** promote its own output to `Approved` · fabricate review independence · freeze a slice · lift an approval block · modify unrelated legacy documents.

## Stop conditions

Production itself prohibited · concept has two plausible owners · an assertion needs a Product Owner decision · a required source is inadmissible with no alternative · the change would reopen an accepted decision.

## Anti-patterns

- Citing an empty `Scaffold` as the source of a constraint
- Writing prose before the assertion registry is approved
- Adding a discovered assertion straight into prose
- Collapsing "approval blocked" into "stop all work"
- Hardcoding a document's status instead of reading it
- Fixing unrelated legacy documents encountered along the way
- Treating differing wording as a contradiction without first asking whether the two statements answer different questions

## Examples

**Proceed with a ceiling.** Upstream dependency is `Review`; `meta/003` §6.3 forbids `Approved` depending on `Draft`/`Scaffold` but not on `Review`. → produce, state the ceiling, preserve the blocker.

**Stop.** A required term registry is `Scaffold` and an unresolved ownership migration is open → production prohibited; name the blocker.

**Re-entry.** *"Add one more goal"* → `MINOR`, enters at the assertion stage, not at prose.

**Escalate.** An assertion has no admissible source and is not derivable → register an open question, **STOP**.
