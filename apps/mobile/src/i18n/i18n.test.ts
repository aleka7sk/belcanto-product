import { kkKZ, ruKZ } from "./messages";
import { formatTemplate, selectPlural } from "./plural";
import { pseudoExpand, pseudoExpansionRatio } from "./pseudo";

describe("localization foundation (HOF-15)", () => {
  it("keeps both catalogs key-complete and free of placeholder gaps", () => {
    const ruKeys = Object.keys(ruKZ).sort();
    const kkKeys = Object.keys(kkKZ).sort();
    expect(kkKeys).toEqual(ruKeys);
    for (const value of [...Object.values(ruKZ), ...Object.values(kkKZ)]) {
      expect(value.trim().length).toBeGreaterThan(0);
    }
  });

  it("selects Russian plural categories per CLDR", () => {
    const forms = { one: "занятие", few: "занятия", many: "занятий", other: "занятия" };
    expect(selectPlural("ru-KZ", 1, forms)).toBe("занятие");
    expect(selectPlural("ru-KZ", 2, forms)).toBe("занятия");
    expect(selectPlural("ru-KZ", 5, forms)).toBe("занятий");
    expect(selectPlural("ru-KZ", 21, forms)).toBe("занятие");
    expect(selectPlural("ru-KZ", 111, forms)).toBe("занятий");
  });

  it("selects Kazakh forms with a mandatory other fallback (ICU {count} сабақ)", () => {
    const forms = { one: "сабақ", other: "сабақ" };
    expect(selectPlural("kk-KZ", 1, forms)).toBe("сабақ");
    expect(selectPlural("kk-KZ", 5, forms)).toBe("сабақ");
    expect(selectPlural("kk-KZ", 2, { other: "сабақ" })).toBe("сабақ");
  });

  it("substitutes template parameters without concatenation and keeps unknown ones visible", () => {
    expect(formatTemplate("Урок перенесён на {time}", { time: "19:00" })).toBe(
      "Урок перенесён на 19:00",
    );
    expect(formatTemplate("Осталось {count} мест", {})).toBe("Осталось {count} мест");
  });

  it("expands pseudolocale strings by at least 40 percent and keeps placeholders intact", () => {
    const source = "Урок перенесён на {time}";
    const expanded = pseudoExpand(source);
    expect(expanded.startsWith("⟦")).toBe(true);
    expect(expanded.endsWith("⟧")).toBe(true);
    expect(expanded).toContain("{time}");
    expect(pseudoExpansionRatio(source)).toBeGreaterThanOrEqual(1.4);
  });
});
