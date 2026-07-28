#!/usr/bin/env python3
"""
Belcanto Product — deterministic corpus validator.

Read-only. Offline. Standard library only. Run from the repository root.

    python3 .claude/scripts/validate.py [path]

Exit codes
    0  no blocking errors
    1  validation errors found
    2  validator execution or configuration failure

This validator implements only checks computable from repository files.
Judgement-based checks (DES-004 level V3) are out of scope and are listed
as skipped. Governing rules are cited per check.
"""

import os
import re
import sys
from collections import defaultdict, Counter

# ── configuration derived from the corpus ────────────────────────────────────

# meta/003 §6.1 — mandatory frontmatter
REQUIRED_FIELDS = [
    "Document Id", "Title", "Layer", "Status", "Version", "Last Updated",
    "Reading Order", "Authority Rank", "Authority Scope",
    "Owners", "Depends On", "Required By", "Defines", "Must Not Define",
]
SCALAR_FIELDS = REQUIRED_FIELDS[:9]
LIST_FIELDS = REQUIRED_FIELDS[9:]

# DES-001 §6.1 — lifecycle states
ALL_STATES = {"Scaffold", "Draft", "Review", "Approved", "Canonical",
              "Disputed", "Superseded"}
HISTORICAL = {"Superseded"}          # retired: link-checked, not an active owner

# meta/003 §6.3 — dependency-status rule
FORBIDDEN_DEPS = {
    "Approved":  {"Scaffold", "Draft"},
    "Canonical": {"Scaffold", "Draft", "Review"},
}

REGISTRY_READING_ORDER = "meta/001-reading-order.md"
REGISTRY_OWNERSHIP = "meta/007-ownership-map.md"
REGISTRY_DEVIATIONS = "meta/011-conformance-statement.md"

SKIP_DIRS = {".git", ".claude", "node_modules"}

# ── finding model ────────────────────────────────────────────────────────────

ERROR, WARN, KNOWN = "ERROR", "WARN", "KNOWN"


class Findings:
    def __init__(self):
        self.items = []
        self.skipped = []

    def add(self, severity, check, path, line, message, rule=""):
        self.items.append((severity, check, path, line or 0, message, rule))

    def skip(self, check, reason):
        self.skipped.append((check, reason))

    def by(self, severity):
        return [f for f in self.items if f[0] == severity]


# ── frontmatter parsing ──────────────────────────────────────────────────────

def parse_doc(path):
    """Return (fields, body_offset) or (None, 0) when there is no frontmatter."""
    try:
        text = open(path, encoding="utf-8").read()
    except OSError as exc:
        raise RuntimeError(f"cannot read {path}: {exc}")
    if not text.startswith("---\n"):
        return None, 0
    end = text.find("\n---", 4)
    if end == -1:
        return None, 0
    head = text[4:end]
    fields, current = {}, None
    for raw in head.split("\n"):
        if raw.startswith("  - ") and current:
            fields.setdefault(current, []).append(raw[4:].strip())
            continue
        m = re.match(r"^([A-Za-z][A-Za-z ]*?):\s*(.*)$", raw)
        if m:
            key, val = m.group(1), m.group(2).strip()
            current = key
            fields[key] = val if val else []
    return fields, head.count("\n") + 2


def real_refs(values):
    """List entries that are genuine document paths (placeholders excluded)."""
    if not isinstance(values, list):
        return []
    return [v for v in values if v.endswith(".md") and not v.startswith("(")]


def collect(root="."):
    docs = {}
    plain = []
    for base, dirs, files in os.walk(root):
        dirs[:] = sorted(d for d in dirs if d not in SKIP_DIRS)
        for name in sorted(files):
            if not name.endswith(".md"):
                continue
            rel = os.path.relpath(os.path.join(base, name), root).replace(os.sep, "/")
            fields, _ = parse_doc(rel)
            if fields is None:
                plain.append(rel)
            else:
                docs[rel] = fields
    return docs, plain


def is_active(fields):
    return fields.get("Status") not in HISTORICAL


def is_legacy(fields):
    """Objectively pre-standard: carries none of the graph fields.

    Conservative, deterministic proxy for the DEV-0001 scope. The deviation
    register (meta/011 §6.4) states its scope in prose and is not reliably
    machine-readable — see the disclosure printed in the summary.
    """
    return not any(k in fields for k in ("Depends On", "Required By", "Defines"))


# ── checks ───────────────────────────────────────────────────────────────────

LINK_RE = re.compile(r"\]\(([^)]+)\)")


def check_v1_links(docs, plain, f, root="."):
    """V1 — internal markdown links resolve. DES-005 CR1."""
    for path in sorted(list(docs) + plain):
        base = os.path.dirname(path)
        for num, line in enumerate(open(path, encoding="utf-8"), 1):
            for m in LINK_RE.finditer(line):
                target = m.group(1).split("#")[0].strip()
                if not target or target.startswith(("http://", "https://", "mailto:")):
                    continue
                resolved = os.path.normpath(os.path.join(root, base, target))
                if not os.path.exists(resolved):
                    f.add(ERROR, "V1", path, num,
                          f"link target does not exist: {target}", "DES-005 CR1")
    f.skip("V1-anchors", "heading anchors not verified — Cyrillic heading slugs "
                         "cannot be derived deterministically")


def parse_reading_registry(f):
    """Rows of meta/001 §6.1: | pos | `path` | note |. Returns {path: (pos, historical)}."""
    entries = {}
    if not os.path.exists(REGISTRY_READING_ORDER):
        f.add(ERROR, "V2", REGISTRY_READING_ORDER, 0, "reading-order registry missing")
        return entries
    row = re.compile(r"^\|\s*([\d.]+)\s*\|\s*(~~)?`([^`]+)`(~~)?\s*\|")
    for num, line in enumerate(open(REGISTRY_READING_ORDER, encoding="utf-8"), 1):
        m = row.match(line)
        if m:
            entries[m.group(3)] = (m.group(1), bool(m.group(2)), num)
    return entries


def check_v2_reading(docs, f):
    """V2 — reading positions. meta/006 V1.7, V4.1; DES-004 V4.1."""
    registry = parse_reading_registry(f)

    seen = defaultdict(list)
    for path, fields in docs.items():
        if not is_active(fields):
            continue
        pos = fields.get("Reading Order")
        if isinstance(pos, str) and pos:
            seen[pos].append(path)
    for pos, paths in sorted(seen.items()):
        if len(paths) > 1:
            for p in sorted(paths):
                f.add(ERROR, "V2", p, 0,
                      f"duplicate active reading position {pos}: {', '.join(sorted(paths))}",
                      "meta/006 V1.7")

    for path, (pos, historical, num) in sorted(registry.items()):
        if "*" in path:
            continue          # directory pattern, e.g. features/catalog/F-*.md
        if not os.path.exists(path):
            f.add(ERROR, "V2", REGISTRY_READING_ORDER, num,
                  f"registry entry points to a missing file: {path}", "DES-004 V4.1")
            continue
        fields = docs.get(path)
        if not fields:
            continue
        declared = fields.get("Reading Order")
        if isinstance(declared, str) and declared and declared != pos:
            f.add(ERROR, "V2", path, 0,
                  f"reading position {declared} contradicts registry {pos}",
                  "meta/006 V1.7")
        if historical and is_active(fields):
            f.add(WARN, "V2", path, 0,
                  "registry marks the entry historical but the document is active",
                  "DES-001 §6.1")
        if not historical and not is_active(fields):
            f.add(WARN, "V2", REGISTRY_READING_ORDER, num,
                  f"{path} is {fields.get('Status')} but the registry entry is not marked historical",
                  "DES-001 §6.5")

    for path, fields in sorted(docs.items()):
        if is_active(fields) and path not in registry and path not in (
                REGISTRY_READING_ORDER,):
            f.add(WARN, "V2", path, 0,
                  "active document is absent from the reading-order registry",
                  "meta/006 V4.1")


def check_v3_ownership(docs, f):
    """V3 — one active owner per concept. meta/004 D1; DES-009 A1."""
    owners = defaultdict(list)
    for path, fields in docs.items():
        if not is_active(fields):
            continue
        for concept in fields.get("Defines", []) or []:
            if concept.startswith("(") or concept.lower().startswith("ничего"):
                continue
            owners[concept.strip().lower()].append(path)

    duplicates = {c: sorted(p) for c, p in owners.items() if len(p) > 1}
    for concept, paths in sorted(duplicates.items()):
        for p in paths:
            f.add(KNOWN, "V3", p, 0,
                  f"concept claimed by {len(paths)} active documents: "
                  f"\"{concept}\" — also {', '.join(x for x in paths if x != p)}",
                  "meta/004 D1")

    if os.path.exists(REGISTRY_OWNERSHIP):
        row = re.compile(r"^\|[^|]*\|\s*`([^`]+\.md)`")
        for num, line in enumerate(open(REGISTRY_OWNERSHIP, encoding="utf-8"), 1):
            m = row.match(line)
            if not m:
                continue
            target = m.group(1)
            if target.endswith("/*") or "*" in target:
                continue
            if not os.path.exists(target):
                f.add(ERROR, "V3", REGISTRY_OWNERSHIP, num,
                      f"ownership registry points to a missing file: {target}",
                      "DES-004 V2.8")
            elif target in docs and not is_active(docs[target]):
                f.add(ERROR, "V3", REGISTRY_OWNERSHIP, num,
                      f"ownership registry points to a retired document: {target} "
                      f"({docs[target].get('Status')})", "PD-0026")
            if target.startswith(".claude"):
                f.add(ERROR, "V3", REGISTRY_OWNERSHIP, num,
                      f".claude path registered as a normative owner: {target}",
                      "PD-0026")


def check_v4_symmetry(docs, f):
    """V4 — Depends On / Required By reciprocity. meta/003 §6.1."""
    for path, fields in sorted(docs.items()):
        if not is_active(fields):
            continue
        for dep in real_refs(fields.get("Depends On")):
            if dep not in docs or not is_active(docs[dep]):
                continue
            if path not in (docs[dep].get("Required By") or []):
                sev = KNOWN if (is_legacy(docs[dep]) or is_legacy(fields)) else WARN
                note = " [matches DEV-0001 scope: legacy metadata]" if sev == KNOWN else ""
                f.add(sev, "V4", dep, 0,
                      f"missing reciprocal Required By: {path}{note}",
                      "meta/003 §6.1")


def check_v5_cycles(docs, f):
    """V5 — no cycles in the active dependency graph. DES-005 CR4."""
    graph = {p: [d for d in real_refs(fl.get("Depends On"))
                 if d in docs and is_active(docs[d])]
             for p, fl in docs.items() if is_active(fl)}
    colour, cycles = {}, []

    def walk(node, stack):
        colour[node] = 1
        stack.append(node)
        for nxt in sorted(graph.get(node, [])):
            if colour.get(nxt) == 1:
                cycles.append(stack[stack.index(nxt):] + [nxt])
            elif colour.get(nxt, 0) == 0:
                walk(nxt, stack)
        stack.pop()
        colour[node] = 2

    for node in sorted(graph):
        if colour.get(node, 0) == 0:
            walk(node, [])
    for cycle in cycles:
        f.add(ERROR, "V5", cycle[0], 0,
              "dependency cycle: " + " -> ".join(cycle), "DES-005 CR4")


def check_v6_dep_status(docs, f):
    """V6 — permitted dependency states. meta/003 §6.3."""
    for path, fields in sorted(docs.items()):
        status = fields.get("Status")
        forbidden = FORBIDDEN_DEPS.get(status)
        if not forbidden:
            continue
        for dep in real_refs(fields.get("Depends On")):
            dep_status = docs.get(dep, {}).get("Status")
            if dep_status in forbidden:
                f.add(ERROR, "V6", path, 0,
                      f"{status} document depends on {dep} ({dep_status})",
                      "meta/003 §6.3")


def check_v7_fields(docs, plain, f):
    """V7 — mandatory frontmatter. meta/003 §6.1."""
    for path, fields in sorted(docs.items()):
        missing = [k for k in REQUIRED_FIELDS if k not in fields]
        if missing:
            sev = KNOWN if is_legacy(fields) else WARN
            note = " [matches DEV-0001 scope: legacy metadata]" if sev == KNOWN else ""
            f.add(sev, "V7", path, 0,
                  f"missing mandatory field(s): {', '.join(missing)}{note}",
                  "meta/003 §6.1")
        status = fields.get("Status")
        if status and status not in ALL_STATES:
            f.add(ERROR, "V7", path, 0,
                  f"unknown Status value: {status}", "DES-001 §6.1")
        version = fields.get("Version")
        if isinstance(version, str) and version and not re.fullmatch(r"\d+\.\d+\.\d+", version):
            f.add(ERROR, "V7", path, 0,
                  f"Version is not semver: {version}", "DES-004 V1.3")
        updated = fields.get("Last Updated")
        if isinstance(updated, str) and updated and not re.fullmatch(r"\d{4}-\d{2}-\d{2}", updated):
            f.add(ERROR, "V7", path, 0,
                  f"Last Updated is not ISO-8601: {updated}", "DES-004 V1.5")
    for path in sorted(plain):
        f.add(KNOWN, "V7", path, 0,
              "no frontmatter — outside the document contract", "meta/003 §6.1")


def check_v8_registries(docs, f):
    """V8 — registry / frontmatter consistency."""
    for path, fields in sorted(docs.items()):
        if not is_active(fields):
            continue
        for dep in real_refs(fields.get("Depends On")):
            if dep in docs and not is_active(docs[dep]):
                f.add(ERROR, "V8", path, 0,
                      f"active document declares a retired dependency: {dep} "
                      f"({docs[dep].get('Status')})", "DES-001 §6.5")
        for req in real_refs(fields.get("Required By")):
            if req in docs and not is_active(docs[req]):
                f.add(ERROR, "V8", path, 0,
                      f"active document declares a retired dependent: {req} "
                      f"({docs[req].get('Status')})", "DES-001 §6.5")
    f.skip("V8-deviations", "deviation scopes in meta/011 §6.4 are prose and are "
                            "not machine-readable; DEV-0001 matching uses a "
                            "conservative structural proxy only")


# ── reporting ────────────────────────────────────────────────────────────────

def report(f, docs, plain):
    order = {ERROR: 0, WARN: 1, KNOWN: 2}
    items = sorted(f.items, key=lambda x: (order[x[0]], x[1], x[2], x[3]))

    for severity, label in ((ERROR, "ERRORS"), (WARN, "WARNINGS"),
                            (KNOWN, "KNOWN / ACCEPTED")):
        group = [i for i in items if i[0] == severity]
        if not group:
            continue
        print(f"\n{label} ({len(group)})")
        print("-" * 72)
        for _, check, path, line, message, rule in group:
            loc = f"{path}:{line}" if line else path
            src = f"  [{rule}]" if rule else ""
            print(f"  {check}  {loc}\n      {message}{src}")

    if f.skipped:
        print(f"\nSKIPPED / PARTIAL ({len(f.skipped)})")
        print("-" * 72)
        for check, reason in f.skipped:
            print(f"  {check}: {reason}")

    errors = len(f.by(ERROR))
    warns = len(f.by(WARN))
    known = len(f.by(KNOWN))
    checks = ["V1 links", "V2 reading positions", "V3 ownership",
              "V4 symmetry", "V5 cycles", "V6 dependency status",
              "V7 mandatory fields", "V8 registry consistency"]
    failed = {i[1] for i in f.by(ERROR)}

    print("\n" + "=" * 72)
    print("SUMMARY")
    print("=" * 72)
    print(f"  documents inspected      {len(docs)} with frontmatter, {len(plain)} without")
    print(f"  checks passed            {len([c for c in checks if c.split()[0] not in failed])}/{len(checks)}")
    print(f"  errors                   {errors}")
    print(f"  warnings                 {warns}")
    print(f"  known / accepted         {known}")
    print(f"  skipped or partial       {len(f.skipped)}")
    print(f"  exit code                {1 if errors else 0}")
    if known:
        print("\n  Note: KNOWN entries include pre-existing findings recorded in\n"
              "  meta/011 §6.4. Deviation scopes there are prose; this validator\n"
              "  cannot machine-distinguish a new duplicate-Defines finding from a\n"
              "  pre-existing one without a baseline artifact, which is not authorised.")
    return 1 if errors else 0


def main(argv):
    root = argv[1] if len(argv) > 1 else "."
    if not os.path.isdir(root):
        print(f"validator: not a directory: {root}", file=sys.stderr)
        return 2
    if not os.path.exists(os.path.join(root, REGISTRY_READING_ORDER)):
        print(f"validator: run from the repository root "
              f"({REGISTRY_READING_ORDER} not found under {root})", file=sys.stderr)
        return 2

    cwd = os.getcwd()
    try:
        os.chdir(root)
        docs, plain = collect(".")
        f = Findings()
        check_v1_links(docs, plain, f)
        check_v2_reading(docs, f)
        check_v3_ownership(docs, f)
        check_v4_symmetry(docs, f)
        check_v5_cycles(docs, f)
        check_v6_dep_status(docs, f)
        check_v7_fields(docs, plain, f)
        check_v8_registries(docs, f)
        return report(f, docs, plain)
    except RecursionError:
        print("validator: dependency graph too deep to traverse", file=sys.stderr)
        return 2
    except Exception as exc:                                   # noqa: BLE001
        print(f"validator: {type(exc).__name__}: {exc}", file=sys.stderr)
        return 2
    finally:
        os.chdir(cwd)


if __name__ == "__main__":
    sys.exit(main(sys.argv))
