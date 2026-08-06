import { decodeLessonJournal, decodeProgressEvidence } from "./contracts";

describe("journal contracts (DEC-006/007)", () => {
  const journal = {
    id: "jrnl_1",
    occurrenceId: "cocc_1",
    studentId: "student_1",
    teacher: { accountId: "account_t", fullName: "Диана Садыкова" },
    status: "published",
    currentVersion: 2,
    versions: [
      {
        version: 2,
        whatWorked: "Опора устойчива",
        currentFocus: "Верхний регистр",
        nextStep: "Легато и стаккато",
        correctionNote: "Уточнение после разбора",
        publishedAt: "2026-08-06T15:00:00Z",
      },
      {
        version: 1,
        whatWorked: "Опора устойчива",
        currentFocus: "Верхний регистр",
        nextStep: "Легато",
        publishedAt: "2026-08-06T12:00:00Z",
      },
    ],
    updatedAt: "2026-08-06T15:00:00Z",
  };

  it("requires a correction note on every version after the first", () => {
    expect(decodeLessonJournal(journal).currentVersion).toBe(2);
    expect(() =>
      decodeLessonJournal({
        ...journal,
        versions: [{ ...journal.versions[0], correctionNote: undefined }, journal.versions[1]],
      }),
    ).toThrow("LessonJournal");
  });

  it("orders versions newest-first and binds published to a version", () => {
    expect(() =>
      decodeLessonJournal({ ...journal, versions: [...journal.versions].reverse() }),
    ).toThrow("LessonJournal");
    expect(() =>
      decodeLessonJournal({ ...journal, currentVersion: 0 }),
    ).toThrow("LessonJournal");
  });

  it("keeps the draft optional and closed", () => {
    const withDraft = decodeLessonJournal({
      ...journal,
      draft: { whatWorked: "а", currentFocus: "б", nextStep: "в" },
    });
    expect(withDraft.draft?.nextStep).toBe("в");
    expect(() =>
      decodeLessonJournal({
        ...journal,
        draft: { whatWorked: "а", currentFocus: "б", nextStep: "в", score: 5 },
      }),
    ).toThrow("LessonJournal");
  });

  it("decodes evidence without any numeric score field", () => {
    const evidence = decodeProgressEvidence([
      {
        id: "evd_1",
        area: "Дыхание",
        note: "Фраза 8 тактов",
        sourceKind: "lesson_journal",
        sourceId: "jrnl_1:1",
        recordedAt: "2026-08-06T12:00:00Z",
      },
    ]);
    expect(evidence[0]?.area).toBe("Дыхание");
    expect(() =>
      decodeProgressEvidence([
        {
          id: "evd_2",
          area: "Дыхание",
          note: "х",
          sourceKind: "lesson_journal",
          sourceId: "jrnl_1:1",
          recordedAt: "2026-08-06T12:00:00Z",
          score: 5,
        },
      ]),
    ).toThrow("ProgressEvidenceList");
  });
});
