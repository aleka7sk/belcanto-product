import { semantic } from "./tokens";

/**
 * Domain color coding (owner decision 2026-08-06): each product area
 * carries one accent so screens stop looking identical — schedule
 * violet, practice cyan, community magenta, growth gold, school
 * operations gold, owner overview magenta. Values come from the
 * semantic token layer only.
 */
export type Domain =
  | "today"
  | "schedule"
  | "practice"
  | "community"
  | "teacher"
  | "growth"
  | "activity"
  | "operations"
  | "overview"
  | "account";

export function domainAccent(domain: Domain): string {
  switch (domain) {
    case "practice":
    case "teacher":
      return semantic.accentCyan;
    case "community":
      return semantic.accentMagenta;
    case "growth":
      return semantic.accentGold;
    case "operations":
      return semantic.accentGold;
    case "overview":
      return semantic.accentMagenta;
    case "today":
    case "schedule":
    case "activity":
    case "account":
      return semantic.accentViolet;
  }
}

/**
 * Deterministic person accent: the same name always renders the same
 * avatar color, picked from the four domain accents.
 */
export function personAccent(name: string): string {
  const palette = [
    semantic.accentViolet,
    semantic.accentCyan,
    semantic.accentGold,
    semantic.accentMagenta,
  ];
  let hash = 0;
  for (const char of name) {
    hash = (hash * 31 + char.codePointAt(0)!) >>> 0;
  }
  return palette[hash % palette.length]!;
}
