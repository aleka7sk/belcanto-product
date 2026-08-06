import type { IsoDateTime, ProgressEvidence } from "../api/contracts";
import {
  EVIDENCE_KINDS,
  groupEvidenceByArea,
  journalAreaAccent,
  RECAP_TONES,
} from "./patterns/journalPatterns";
import { semantic } from "./tokens";

const AREA_ACCENTS = [
  semantic.accentViolet,
  semantic.accentCyan,
  semantic.accentGold,
  semantic.accentMagenta,
  semantic.textAccent,
];

function evidence(id: string, area: string, recordedAt: string): ProgressEvidence {
  return {
    id,
    area,
    note: `note ${id}`,
    sourceKind: "lesson_journal",
    sourceId: `jrnl_1:${id}`,
    recordedAt: recordedAt as IsoDateTime,
  };
}

describe("journal patterns (Pages 22/26)", () => {
  it("assigns a stable constellation accent per area", () => {
    const first = journalAreaAccent("Дыхание");
    expect(journalAreaAccent("Дыхание")).toBe(first);
    for (const area of ["Голос", "Интонация", "Музыкальность", "Сцена", "Опора"]) {
      expect(AREA_ACCENTS).toContain(journalAreaAccent(area));
    }
  });

  it("groups evidence by named area preserving recency order (DEC-006)", () => {
    const groups = groupEvidenceByArea([
      evidence("evd_3", "Дыхание", "2026-08-06T10:00:00Z"),
      evidence("evd_2", "Сцена", "2026-08-01T10:00:00Z"),
      evidence("evd_1", "Дыхание", "2026-07-20T10:00:00Z"),
    ]);
    expect(groups.map((group) => group.area)).toEqual(["Дыхание", "Сцена"]);
    expect(groups[0]!.entries.map((entry) => entry.id)).toEqual(["evd_3", "evd_1"]);
    expect(groups[0]!.accent).toBe(journalAreaAccent("Дыхание"));
  });

  it("keeps recap tones on the design's three lifecycle colors (DEC-007)", () => {
    expect(RECAP_TONES.draft.border).toBe(semantic.feedbackWarning);
    expect(RECAP_TONES.published.border).toBe(semantic.borderAccent);
    expect(RECAP_TONES.published.label).toBe(semantic.accentCyan);
    expect(RECAP_TONES.offline.border).toBe(semantic.feedbackDanger);
  });

  it("tags evidence media kinds like the Evidence Card component", () => {
    expect(EVIDENCE_KINDS.audio).toEqual({ tag: "АУ", color: semantic.accentViolet });
    expect(EVIDENCE_KINDS.video).toEqual({ tag: "ВИ", color: semantic.accentMagenta });
    expect(EVIDENCE_KINDS.teacherNote).toEqual({ tag: "НТ", color: semantic.accentCyan });
  });
});
