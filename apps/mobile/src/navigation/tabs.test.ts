import { ROLE_TABS, findTab, tabsForRole } from "./tabs";
import { kkKZ, ruKZ } from "../i18n/messages";

describe("role-aware bottom navigation contract (Figma 310:20542)", () => {
  it("gives every role exactly five positional slots", () => {
    for (const role of ["Student", "Teacher", "Administrator", "Owner"] as const) {
      const tabs = tabsForRole(role);
      expect(tabs).toHaveLength(5);
      expect(tabs.map((tab) => tab.slot)).toEqual([1, 2, 3, 4, 5]);
      expect(new Set(tabs.map((tab) => tab.key)).size).toBe(5);
    }
  });

  it("matches the canonical variant mapping including icon quirks", () => {
    expect(ROLE_TABS.Student.map((tab) => `${tab.key}:${tab.icon}`)).toEqual([
      "today:home",
      "schedule:calendar",
      "practice:mic",
      "community:users",
      "profile:trophy",
    ]);
    expect(ROLE_TABS.Teacher.map((tab) => `${tab.key}:${tab.icon}`)).toEqual([
      "today:home",
      "schedule:calendar",
      "students:users",
      "review:check",
      "community:mic",
    ]);
    expect(ROLE_TABS.Administrator.map((tab) => `${tab.key}:${tab.icon}`)).toEqual([
      "operations:home",
      "schedule:calendar",
      "people:users",
      "community:mic",
      "more:more",
    ]);
    expect(ROLE_TABS.Owner.map((tab) => `${tab.key}:${tab.icon}`)).toEqual([
      "overview:home",
      "analytics:trophy",
      "operations:calendar",
      "team:users",
      "more:more",
    ]);
  });

  it("has a localized shell label in both locales for every tab", () => {
    for (const role of ["Student", "Teacher", "Administrator", "Owner"] as const) {
      for (const tab of tabsForRole(role)) {
        expect(ruKZ[tab.labelKey]).toBeTruthy();
        expect(kkKZ[tab.labelKey]).toBeTruthy();
      }
    }
  });

  it("carries the Figma-sourced Kazakh shell overrides verbatim", () => {
    expect(kkKZ["nav.operations"]).toBe("Операциялар");
    expect(kkKZ["nav.schedule"]).toBe("Кесте");
    expect(kkKZ["nav.people"]).toBe("Адамдар");
    expect(kkKZ["nav.community"]).toBe("Қауымдастық");
    expect(kkKZ["nav.more"]).toBe("Тағы");
  });

  it("resolves tabs by key within a role only", () => {
    expect(findTab("Student", "practice")?.icon).toBe("mic");
    expect(findTab("Owner", "practice")).toBeNull();
  });
});
