import {
  decodeAssessment,
  decodeAssessmentChain,
  decodeAssessments,
} from "./contracts";

describe("assessment contracts (Page 27, domain/assessment.md)", () => {
  const draft = {
    id: "asmt_1",
    studentId: "student_1",
    author: { accountId: "teacher_1", fullName: "Коркем" },
    authorRole: "Teacher",
    type: "formative",
    contextType: "lesson",
    assessmentDate: "2026-08-01",
    summary: "Опора устойчива в коротких фразах.",
    strengths: "Стабильное окончание фразы.",
    visibility: "student_visible",
    status: "draft",
    evidence: [
      {
        id: "evd_1",
        kind: "observation",
        note: "Три повтора без подсказки.",
        addedAt: "2026-08-01T12:00:00Z",
      },
    ],
    version: 1,
    createdAt: "2026-08-01T11:00:00Z",
  };
  const published = {
    ...draft,
    id: "asmt_2",
    status: "published",
    publishedAt: "2026-08-01T12:30:00Z",
  };

  it("accepts the lifecycle shapes and rejects impossible pairings", () => {
    expect(decodeAssessment(draft).status).toBe("draft");
    expect(decodeAssessment(published).publishedAt).toBeDefined();
    expect(() =>
      decodeAssessment({ ...published, publishedAt: undefined }),
    ).toThrow("Assessment");
    expect(() =>
      decodeAssessment({ ...draft, status: "withdrawn" }),
    ).toThrow("Assessment");
    expect(() =>
      decodeAssessment({ ...draft, supersededById: "asmt_9" }),
    ).toThrow("Assessment");
  });

  it("requires published substance: summary plus one block or evidence", () => {
    expect(() =>
      decodeAssessment({
        ...published,
        summary: undefined,
      }),
    ).toThrow("Assessment");
    expect(() =>
      decodeAssessment({
        ...published,
        strengths: undefined,
        evidence: [],
      }),
    ).toThrow("Assessment");
  });

  it("keeps the history newest-first with unique identifiers", () => {
    const older = { ...published, id: "asmt_3", assessmentDate: "2026-07-01" };
    expect(decodeAssessments([published, older])).toHaveLength(2);
    expect(() => decodeAssessments([older, published])).toThrow("AssessmentList");
    expect(() => decodeAssessments([published, published])).toThrow("AssessmentList");
  });

  it("verifies the supersede chain pairing", () => {
    const replacement = {
      ...published,
      id: "asmt_4",
      summary: "Опора устойчива и в полном куплете.",
    };
    const replaced = {
      ...published,
      status: "superseded",
      supersededById: "asmt_4",
    };
    expect(decodeAssessmentChain([replaced, replacement])).toHaveLength(2);
    expect(() =>
      decodeAssessmentChain([replaced, { ...replacement, id: "asmt_5" }]),
    ).toThrow("AssessmentChain");
    expect(() => decodeAssessmentChain([replacement])).toThrow("AssessmentChain");
  });
});
