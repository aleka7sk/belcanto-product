import {
  decodeCoreLessonSeries,
  decodeRooms,
  decodeSeriesGenerationResult,
} from "./contracts";

describe("core lesson series contracts (DEC-002/004)", () => {
  const series = {
    id: "clser_1",
    format: "individual",
    title: "Вокал · индивидуально",
    teacher: { accountId: "account_t", fullName: "Диана Садыкова" },
    roomId: "room_1",
    weekday: 0,
    startMinutes: 600,
    durationMinutes: 45,
    effectiveFrom: "2026-08-03",
    status: "active",
    version: 0,
    students: [{ studentId: "student_1", fullName: "Алишер Беков" }],
  };

  it("decodes a series and keeps optional fields optional", () => {
    const decoded = decodeCoreLessonSeries(series);
    expect(decoded.roomId).toBe("room_1");
    expect(decoded).not.toHaveProperty("effectiveUntil");
  });

  it("enforces the DEC-002 format constraint structurally", () => {
    expect(() =>
      decodeCoreLessonSeries({
        ...series,
        students: [
          { studentId: "student_1", fullName: "А" },
          { studentId: "student_2", fullName: "Б" },
        ],
      }),
    ).toThrow("CoreLessonSeries");
    expect(() =>
      decodeCoreLessonSeries({
        ...series,
        format: "group",
        students: [
          { studentId: "student_1", fullName: "А" },
          { studentId: "student_2", fullName: "Б" },
          { studentId: "student_3", fullName: "В" },
          { studentId: "student_4", fullName: "Г" },
        ],
      }),
    ).toThrow("CoreLessonSeries");
  });

  it("rejects out-of-range weekday and unknown keys", () => {
    expect(() => decodeCoreLessonSeries({ ...series, weekday: 7 })).toThrow(
      "CoreLessonSeries",
    );
    expect(() => decodeCoreLessonSeries({ ...series, extra: 1 })).toThrow(
      "CoreLessonSeries",
    );
  });

  it("decodes rooms and binds generation counts to ids", () => {
    expect(
      decodeRooms([
        { id: "room_1", name: "Зал на Абая", capacity: 3, status: "active", version: 0 },
      ]),
    ).toHaveLength(1);
    expect(
      decodeSeriesGenerationResult({
        seriesId: "clser_1",
        createdCount: 2,
        occurrenceIds: ["cocc_1", "cocc_2"],
      }).createdCount,
    ).toBe(2);
    expect(() =>
      decodeSeriesGenerationResult({
        seriesId: "clser_1",
        createdCount: 3,
        occurrenceIds: ["cocc_1", "cocc_2"],
      }),
    ).toThrow("SeriesGenerationResult");
  });
});
