import { semantic } from "../tokens";
import { agendaAccentColor, agendaCardColors } from "./agendaEntry";
import { heroAccentColor } from "./premiumContextHero";

describe("PremiumContextHero role accents (Figma 300:2)", () => {
  it("maps each role variant to its canonical accent", () => {
    expect(heroAccentColor("Student")).toBe(semantic.accentViolet);
    expect(heroAccentColor("Profile")).toBe(semantic.accentViolet);
    expect(heroAccentColor("Teacher")).toBe(semantic.accentCyan);
    expect(heroAccentColor("Administrator")).toBe(semantic.accentGold);
    expect(heroAccentColor("Owner")).toBe(semantic.accentMagenta);
  });
});

describe("AgendaEntry variant matrix (Figma 306:2)", () => {
  it("colors upcoming entries by domain type", () => {
    expect(agendaAccentColor("individual", "upcoming")).toBe(semantic.accentViolet);
    expect(agendaAccentColor("group", "upcoming")).toBe(semantic.accentCyan);
    expect(agendaAccentColor("event", "upcoming")).toBe(semantic.accentMagenta);
  });

  it("lets state accents win over type accents", () => {
    for (const type of ["individual", "group", "event"] as const) {
      expect(agendaAccentColor(type, "completed")).toBe(semantic.textMuted);
      expect(agendaAccentColor(type, "cancelled")).toBe(semantic.feedbackDanger);
      expect(agendaAccentColor(type, "changed")).toBe(semantic.feedbackWarning);
      expect(agendaAccentColor(type, "now")).toBe(semantic.accentViolet);
    }
  });

  it("raises the card and switches the border for the active entry", () => {
    expect(agendaCardColors("now")).toEqual({
      backgroundColor: semantic.bgRaised,
      borderColor: semantic.borderAccent,
    });
    expect(agendaCardColors("cancelled").borderColor).toBe(semantic.feedbackDanger);
    expect(agendaCardColors("changed").borderColor).toBe(semantic.feedbackWarning);
    expect(agendaCardColors("upcoming")).toEqual({
      backgroundColor: semantic.bgSurface,
      borderColor: semantic.borderDefault,
    });
    expect(agendaCardColors("completed").borderColor).toBe(semantic.borderDefault);
  });
});
