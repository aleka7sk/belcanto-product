---
name: validate
description: Run the deterministic corpus validator over the Belcanto repository and report structural findings. User-invoked only.
disable-model-invocation: true
---

# /validate

Runs the deterministic corpus validator. Read-only — it never modifies files.

## Usage

```
/validate            # whole repository
/validate <path>     # narrow to a directory
```

## Execute

From the repository root:

```bash
python3 .claude/scripts/validate.py
```

With a path argument, pass it through:

```bash
python3 .claude/scripts/validate.py <path>
```

## Exit codes

| Code | Meaning |
|---|---|
| `0` | no blocking errors |
| `1` | validation errors found |
| `2` | validator execution or configuration failure |

Warnings and accepted known findings do **not** make the run a success on their own — report them separately from errors.

## Report back

- pass or fail, and the exit code
- checks performed — V1 links · V2 reading positions · V3 ownership · V4 symmetry · V5 cycles · V6 dependency status · V7 mandatory fields · V8 registry consistency
- **errors** with exact file, line where available, and the governing rule
- **warnings**
- **known / accepted** findings, kept distinct from new ones
- checks skipped or only partially performed, with the reason the validator gives

Reproduce the validator's own file paths and rule citations. Do not summarise them away — the paths are the actionable part.

## Boundaries

- Do not modify any repository file
- Do not fix findings unless separately asked
- Do not suppress a finding because it looks pre-existing — the validator marks what it can match conservatively, and states where that matching is unreliable
- Judgement-based review is out of scope; use `corpus-review` for that
