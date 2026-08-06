import {
  decodeHomeworkAssignment,
  decodeMediaAccess,
  decodeMediaObject,
} from "./contracts";

describe("homework and media contracts (domain/homework.md, DEC-006)", () => {
  const media = {
    id: "med_1",
    kind: "audio",
    contentType: "audio/m4a",
    byteSize: 96,
    uploadedBytes: 96,
    status: "ready",
    createdAt: "2026-08-06T15:00:00Z",
    updatedAt: "2026-08-06T15:02:00Z",
  };

  const homework = {
    id: "hw_1",
    occurrenceId: "cocc_1",
    studentId: "student_1",
    teacher: { accountId: "account_t", fullName: "Елена Орлова" },
    status: "completed",
    goal: "Три повтора последних 8 секунд в темпе 75%",
    readinessCriteria: "Мягкий вход в финальную ноту",
    dueAt: "2026-08-10T13:00:00Z",
    tasks: [
      { id: "hwt_1", position: 1, title: "Разминка", status: "done" },
      {
        id: "hwt_2",
        position: 2,
        title: "Припев в 80% темпа",
        recommendedMinutes: 5,
        skillArea: "Дыхание",
        status: "pending",
      },
    ],
    attachments: [media],
    submissions: [
      {
        id: "sub_2",
        attempt: 2,
        note: "Контрольный дубль",
        media: [],
        submittedAt: "2026-08-07T10:00:00Z",
      },
      {
        id: "sub_1",
        attempt: 1,
        media: [media],
        submittedAt: "2026-08-06T18:00:00Z",
      },
    ],
    feedback: [
      {
        id: "fb_2",
        submissionId: "sub_2",
        teacher: { accountId: "account_t", fullName: "Елена Орлова" },
        decision: "accepted",
        body: "Работа принята",
        evidenceArea: "Дыхание",
        evidenceNote: "Держит фразу без подъёма плеч",
        createdAt: "2026-08-07T12:00:00Z",
      },
      {
        id: "fb_1",
        submissionId: "sub_1",
        teacher: { accountId: "account_t", fullName: "Елена Орлова" },
        decision: "needs_revision",
        body: "Финальные 8 секунд требуют внимания",
        nextStep: "Три раза в 75% темпа",
        createdAt: "2026-08-06T20:00:00Z",
      },
    ],
    version: 6,
    createdAt: "2026-08-06T16:00:00Z",
    updatedAt: "2026-08-07T12:00:00Z",
  };

  it("accepts the full revision loop shape", () => {
    const decoded = decodeHomeworkAssignment(homework);
    expect(decoded.submissions[0]?.attempt).toBe(2);
    expect(decoded.feedback[0]?.decision).toBe("accepted");
  });

  it("binds ready media to a finished upload", () => {
    expect(decodeMediaObject(media).status).toBe("ready");
    expect(() =>
      decodeMediaObject({ ...media, uploadedBytes: 64 }),
    ).toThrow("MediaObject");
    expect(() =>
      decodeMediaObject({ ...media, uploadedBytes: 120 }),
    ).toThrow("MediaObject");
  });

  it("rejects evidence attached to a needs_revision review (DEC-006)", () => {
    expect(() =>
      decodeHomeworkAssignment({
        ...homework,
        feedback: [
          {
            ...homework.feedback[1],
            evidenceArea: "Дыхание",
            evidenceNote: "оценка",
          },
          homework.feedback[0],
        ],
      }),
    ).toThrow("HomeworkAssignment");
  });

  it("requires a preserved reason on cancellation", () => {
    expect(() =>
      decodeHomeworkAssignment({
        ...homework,
        status: "cancelled",
        feedback: [],
        cancelReason: undefined,
      }),
    ).toThrow("HomeworkAssignment");
  });

  it("keeps completion tied to an accepted review", () => {
    expect(() =>
      decodeHomeworkAssignment({
        ...homework,
        feedback: [homework.feedback[1]],
      }),
    ).toThrow("HomeworkAssignment");
  });

  it("orders submissions newest-first and task positions ascending", () => {
    expect(() =>
      decodeHomeworkAssignment({
        ...homework,
        submissions: [...homework.submissions].reverse(),
      }),
    ).toThrow("HomeworkAssignment");
    expect(() =>
      decodeHomeworkAssignment({
        ...homework,
        tasks: [...homework.tasks].reverse(),
      }),
    ).toThrow("HomeworkAssignment");
  });

  it("rejects a score smuggled into any layer", () => {
    expect(() =>
      decodeHomeworkAssignment({ ...homework, score: 5 }),
    ).toThrow("HomeworkAssignment");
    expect(() =>
      decodeHomeworkAssignment({
        ...homework,
        tasks: [{ ...homework.tasks[0], score: 5 }, homework.tasks[1]],
      }),
    ).toThrow("HomeworkAssignment");
  });

  it("validates the sealed access link shape", () => {
    expect(
      decodeMediaAccess({
        url: "/v1/media/med_1/content?token=abc",
        expiresAt: "2026-08-06T15:10:00Z",
      }).url,
    ).toContain("token=");
    expect(() =>
      decodeMediaAccess({ url: "/elsewhere", expiresAt: "2026-08-06T15:10:00Z" }),
    ).toThrow("MediaAccess");
  });
});
