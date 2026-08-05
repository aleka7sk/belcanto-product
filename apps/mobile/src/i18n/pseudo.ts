/**
 * Deterministic pseudolocale expansion for layout tests (HOF-15:
 * pseudo +40%). The transform keeps placeholders intact, brackets the
 * string so truncation is visible, and pads to at least 140% of the
 * source length.
 */
const PADDING_UNIT = "·";

export function pseudoExpand(source: string): string {
  const protectedParts = source.split(/(\{[a-zA-Z0-9_]+\})/g);
  const body = protectedParts
    .map((part) =>
      part.startsWith("{") && part.endsWith("}")
        ? part
        : part.replace(/[aeiouаеёиоуыэюя]/gi, (vowel) => `${vowel}${vowel}`),
    )
    .join("");
  const target = Math.ceil(source.length * 1.4);
  const missing = Math.max(0, target - body.length);
  return `⟦${body}${PADDING_UNIT.repeat(missing)}⟧`;
}

export function pseudoExpansionRatio(source: string): number {
  if (source.length === 0) {
    return 1;
  }
  const expanded = pseudoExpand(source);
  return (expanded.length - 2) / source.length;
}
