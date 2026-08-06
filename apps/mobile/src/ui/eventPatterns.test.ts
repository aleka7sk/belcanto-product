import { categoryAccent } from "./patterns/eventPatterns";
import { semantic } from "./tokens";

describe("event pattern accents (Page 24 catalog)", () => {
  it("assigns a stable accent per category", () => {
    const first = categoryAccent("evcat_1");
    expect(categoryAccent("evcat_1")).toBe(first);
    const accents = new Set(
      ["a", "b", "c", "d", "e", "f", "g"].map((id) => categoryAccent(id)),
    );
    for (const accent of accents) {
      expect([
        semantic.accentViolet,
        semantic.accentGold,
        semantic.accentCyan,
        semantic.accentMagenta,
      ]).toContain(accent);
    }
  });
});
