import type { ActivityEntry, IsoDateTime } from "@/api/contracts";
import { buildActivityRows } from "./screens/activity/ActivityScreen";

const entry = (id: string): ActivityEntry => ({
  id,
  kind: "JournalPublished",
  category: "learning",
  targetType: "journal",
  targetId: id,
  occurredAt: "2026-08-06T10:00:00Z" as IsoDateTime,
  payload: {},
});

describe("activity feed virtualized rows", () => {
  it("prefixes each non-empty day section with its header row", () => {
    const rows = buildActivityRows([entry("a")], [entry("b"), entry("c")]);
    expect(rows.map((row) => row.id)).toEqual([
      "section-today",
      "a",
      "section-earlier",
      "b",
      "c",
    ]);
    expect(rows[0]).toMatchObject({ rowKind: "section", labelKey: "act.today" });
    expect(rows[2]).toMatchObject({ rowKind: "section", labelKey: "act.earlier" });
  });

  it("omits empty sections entirely", () => {
    expect(buildActivityRows([], [])).toEqual([]);
    expect(buildActivityRows([], [entry("x")]).map((row) => row.id)).toEqual([
      "section-earlier",
      "x",
    ]);
  });
});
