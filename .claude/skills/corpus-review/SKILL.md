---
name: corpus-review
description: Use when reviewing a normative document in this repository, performing a P7 review, judging readiness for approval, or processing a human review verdict. Establishes reviewer identity and authority first, classifies findings, and gives minimum-change remediation. Never turns review into redesign and never permits author self-approval.
---

# Corpus Review

## Route to (do not restate)

| Question | Document |
|---|---|
| Exit criteria X1–X12, reviewer independence | `meta/standard/003-exit-criteria.md` |
| Validation levels V1–V4 | `meta/standard/004-validation-checklist.md` |
| Cross-reference checks CR1–CR14 | `meta/standard/005-cross-reference-checklist.md` |
| Eligible reviewer per document class | `meta/012-authoring-binding.md` |
| Class per directory, characteristic failure mode | `meta/012` §6.1, `meta/standard/018-document-class-profiles.md` |
| States, transitions, versioning | `meta/standard/001-document-lifecycle.md` |
| Conflict resolution order | `meta/002-authority-model.md` |

## Identity gate — run first, always

Determine from the repository and the request:

1. Who authored the document
2. Reviewer identity and type — human or agent
3. Document class → `meta/012` §6.1
4. Eligible reviewer role for that class → `meta/012` §6.4
5. Whether the corpus requires a **human** reviewer
6. Whether the reviewer is independent of the author
7. What authority this review therefore has

**The output must always state all seven.** Never label a review independent when it is not.

## Three modes

### Mode A — author self-review

The current agent authored or substantially authored the artifact.

May produce **only**: classified findings · verified-clean list · minimum remediation recommendations · a packet for independent human review.

Must **not**: issue a valid approval verdict · promote the document · call the review independent · freeze a slice · lift an approval block · claim approval became permitted because remediation was applied.

The permitted status ceiling comes from the corpus, not from the fact that findings were fixed. Read `meta/standard/003` §6.4 for what a self-review may carry.

### Mode B — independent agent review

Permitted only when the agent is independent of the author **and** the corpus permits an agent reviewer for this class and gate. Where the corpus requires a human, an independent agent review remains **advisory** — say so.

### Mode C — processing a human verdict

When a human verdict is supplied: verify role eligibility against `meta/012` §6.4 and the role catalog in `meta/007` §6.3 · verify independence · verify the stated scope and verdict form · apply it only when every requirement is satisfied · preserve the audit trail.

**Never infer a missing human confirmation.** Never treat a role named in a prompt as proof of eligibility — verify it from the corpus.

## Procedure

1. Read the target **in full** — never review from memory of authoring it
2. Determine class and its characteristic failure mode; look for that first
3. Run `.claude/scripts/validate.py`
4. **Open and verify every citation** — confirm quoted text and version; never accept a citation on plausibility
5. Internal consistency
6. Consistency with the approved corpus
7. Traceability — every normative assertion sourced
8. Ownership — nothing from `Must Not Define` defined
9. Attribution correctness where the document class requires it
10. Dependency graph — cycles, symmetry, dependency states
11. Distinguish ambiguity from contradiction. **A conflict resolvable by citing an authoritative source is not a decision.** Two statements answering *different questions* coexist — propose coexistence rather than forcing a choice
12. Classify every finding exactly once:

| Class | Test |
|---|---|
| **Blocking** | objectively prevents approval — a validation rule fails, a citation is false, an internal contradiction exists |
| **Major** | should be fixed; does not prevent approval |
| **Minor** | ambiguity or cosmetic |
| **Observation** | worth recording, no action |

13. For a Blocking finding give **only the minimum change**. Do not redesign
14. Preserve accepted decisions unless a corpus contradiction is proven

## Output

Review type and authority · the seven identity facts · findings classified · verified-clean list · the verdict or recommendation the mode permits · minimum remediation · required next reviewer action.

Do not edit the document before explicit authorisation, unless the governing task already authorised remediation.

## Stop conditions

Two approved documents of equal rank conflict and no citation resolves it · the review would reopen an accepted decision · a finding requires product strategy rather than correction · the reviewer is not eligible under `meta/012` §6.4.

## Anti-patterns

- Reviewing from memory of having authored the document
- Labelling a self-review independent
- Redesigning under cover of review
- Escalating a Minor finding to Blocking to force a rewrite
- Treating differing wording as contradiction without the different-questions test
- Approving while a dependency sits in a state the corpus forbids
- Accepting a role claim from the prompt without verifying it in the corpus

## Examples

**Blocking.** Goal A declares a dependency on goal B while B declares one on A, contradicting the document's own structure section. Minimum change: delete one dependency entry.

**Major, not Blocking.** A document cites an empty `Scaffold` as the source of a constraint an approved principle already carries. Over-attribution, not a false claim.

**Not a finding.** Two approved documents phrase the vision differently — one answers *what it is relative to the school*, the other *its position among the school's systems*. Different questions → **coexist**.

**Mode A stop.** The agent authored the document, applied its own findings, and is asked to approve. → produce the review packet; **refuse the promotion**; name the eligible human role.
