import {
  ContractDecodeError,
  decodeApiErrorEnvelope,
  decodeBootstrapView,
  decodeFirstMinute,
  decodeInvitationResult,
  decodeLesson,
  decodeLessons,
  decodeReassignPrimaryTeachersResult,
  decodeReplaceLessonTeachersResult,
  decodeSessionTokens,
  decodeStaffMembers,
  decodeStudentDirectory,
  decodeStudentOnboardingItems,
  decodeVoid,
} from "./contracts";

describe("API contract decoders", () => {
  it("accepts explicit server permissions and access profiles", () => {
    expect(
      decodeBootstrapView({
        accountId: "account_1",
        roles: ["Administrator"],
        accessProfiles: ["StudentOnboardingManager.v1"],
        permissions: [
          "students.create",
          "student_onboarding.read",
          "lessons.read",
          "lessons.create",
          "lesson_teachers.replace",
          "student_primary_teachers.reassign",
        ],
      }),
    ).toMatchObject({
      roles: ["Administrator"],
      accessProfiles: ["StudentOnboardingManager.v1"],
      permissions: expect.arrayContaining([
        "students.create",
        "student_onboarding.read",
        "lessons.create",
      ]),
    });
  });

  it("rejects invitation authority on delegated Administrator bootstrap", () => {
    expect(() =>
      decodeBootstrapView({
        accountId: "account_1",
        roles: ["Administrator"],
        accessProfiles: ["StudentOnboardingManager.v1"],
        permissions: [
          "students.create",
          "student_onboarding.read",
          "student_invitations.issue",
          "lessons.read",
          "lessons.create",
          "lesson_teachers.replace",
          "student_primary_teachers.reassign",
        ],
      }),
    ).toThrow(ContractDecodeError);
  });

  it("requires a complete and internally consistent Student bootstrap", () => {
    expect(() =>
      decodeBootstrapView({
        accountId: "account_1",
        roles: ["Student"],
        accessProfiles: [],
        permissions: ["lessons.read"],
      }),
    ).toThrow(ContractDecodeError);
    expect(() =>
      decodeBootstrapView({
        accountId: "account_1",
        roles: ["Student"],
        accessProfiles: [],
        permissions: ["lessons.read"],
        studentId: "student_1",
        fullName: "Student",
        firstMinute: {
          studentId: "student_2",
          revision: 1,
          whatWorked: "worked",
          currentFocus: "focus",
          nextStep: "next",
          publishedAt: "2026-09-01T10:00:00Z",
        },
      }),
    ).toThrow(ContractDecodeError);
    expect(
      decodeBootstrapView({
        accountId: "account_1",
        roles: ["Student"],
        accessProfiles: [],
        permissions: ["lessons.read"],
        studentId: "student_1",
        fullName: "Student",
        firstMinute: {
          studentId: "student_1",
          revision: 1,
          whatWorked: "worked",
          currentFocus: "focus",
          nextStep: "next",
          publishedAt: "2026-09-01T10:00:00Z",
        },
      }),
    ).toMatchObject({ studentId: "student_1", fullName: "Student" });
    expect(() =>
      decodeBootstrapView({
        accountId: "account_1",
        roles: ["Teacher"],
        accessProfiles: [],
        permissions: ["lessons.read", "lessons.create"],
        studentId: "student_1",
        fullName: "Student",
        firstMinute: {
          studentId: "student_1",
          revision: 1,
          whatWorked: "worked",
          currentFocus: "focus",
          nextStep: "next",
          publishedAt: "2026-09-01T10:00:00Z",
        },
      }),
    ).toThrow(ContractDecodeError);
  });

  it("rejects unknown roles, permissions and invalid timestamps", () => {
    expect(() =>
      decodeBootstrapView({
        accountId: "account_1",
        roles: ["SuperAdmin"],
        accessProfiles: [],
        permissions: ["lessons.read", "lessons.create"],
      }),
    ).toThrow(ContractDecodeError);
    expect(() =>
      decodeSessionTokens({
        accessToken: "A".repeat(43),
        refreshToken: "R".repeat(43),
        accessExpiresAt: "2026-08-01",
        refreshExpiresAt: "2026-09-01T10:00:00Z",
      }),
    ).toThrow(ContractDecodeError);
    expect(() =>
      decodeBootstrapView({
        accountId: "account_1",
        roles: ["Owner"],
        accessProfiles: [],
        permissions: [],
        serverSurprise: true,
      }),
    ).toThrow(ContractDecodeError);
    expect(() =>
      decodeSessionTokens({
        accessToken: "A".repeat(43),
        refreshToken: "R".repeat(43),
        accessExpiresAt: "not-a-date",
        refreshExpiresAt: "2026-09-01T10:00:00Z",
      }),
    ).toThrow(ContractDecodeError);
  });

  it("decodes exact Lesson and Student directory contracts", () => {
    const lesson = {
      id: "lesson_1",
      title: "Индивидуальный урок",
      startsAt: "2026-08-10T13:00:00Z",
      durationMinutes: 60,
      location: "Класс 2",
      teacher: { accountId: "teacher_1", fullName: "Мария Орлова" },
      students: [{ studentId: "student_1", fullName: "Алина Соколова" }],
      status: "scheduled",
      version: 4,
    };
    expect(decodeLesson(lesson)).toEqual(lesson);
    expect(decodeLessons([lesson])).toHaveLength(1);
    expect(
      decodeStudentDirectory([
        {
          studentId: "student_1",
          fullName: "Алина Соколова",
          primaryTeacher: { ...lesson.teacher, status: "inactive" },
          primaryTeacherAssignmentVersion: 7,
        },
      ])[0],
    ).toMatchObject({
      studentId: "student_1",
      primaryTeacher: { status: "inactive" },
      primaryTeacherAssignmentVersion: 7,
    });
    expect(() => decodeLessons([lesson, lesson])).toThrow(ContractDecodeError);
  });

  it("requires result counts to match exact returned objects", () => {
    expect(() =>
      decodeReassignPrimaryTeachersResult({ reassignedCount: 1, assignments: [] }),
    ).toThrow(ContractDecodeError);
    expect(() =>
      decodeReplaceLessonTeachersResult({ updatedCount: 1, lessons: [] }),
    ).toThrow(ContractDecodeError);
  });

  it("decodes staff delegation context without deriving authority from it", () => {
    expect(
      decodeStaffMembers([
        {
          accountId: "administrator_1",
          fullName: "Administrator",
          roles: ["Administrator"],
          accessProfiles: ["StudentOnboardingManager.v1"],
          onboardingDelegationId: "delegation_1",
          onboardingDelegationExpiresAt: "2026-09-01T10:00:00Z",
        },
      ]),
    ).toEqual([
      expect.objectContaining({
        accountId: "administrator_1",
        roles: ["Administrator"],
        onboardingDelegationId: "delegation_1",
      }),
    ]);
  });

  it("decodes every onboarding state and rejects unknown state values", () => {
    const common = {
      studentId: "student_1",
      fullName: "Student",
      enrollmentReference: "BEL-001",
      teacherAccountId: "teacher_1",
      studentVersion: 1,
    };
    expect(
      decodeStudentOnboardingItems([
        { ...common, onboardingState: "awaiting_first_minute" },
        { ...common, onboardingState: "ready_to_invite" },
        {
          ...common,
          onboardingState: "invited",
          invitationId: "invitation_1",
          invitationExpiresAt: "2026-09-01T10:00:00Z",
        },
        { ...common, onboardingState: "activated" },
      ]).map((item) => item.onboardingState),
    ).toEqual([
      "awaiting_first_minute",
      "ready_to_invite",
      "invited",
      "activated",
    ]);
    expect(() =>
      decodeStudentOnboardingItems([
        { ...common, onboardingState: "invitation_expired" },
      ]),
    ).toThrow(ContractDecodeError);
  });

  it("rejects impossible staff and onboarding optional-field combinations", () => {
    expect(() =>
      decodeStaffMembers([
        {
          accountId: "administrator_1",
          fullName: "Administrator",
          roles: ["Administrator"],
          accessProfiles: [],
          onboardingDelegationExpiresAt: "2026-09-01T10:00:00Z",
        },
      ]),
    ).toThrow(ContractDecodeError);
    expect(() =>
      decodeStaffMembers([
        {
          accountId: "administrator_1",
          fullName: "Administrator",
          roles: ["Administrator"],
          accessProfiles: ["StudentOnboardingManager.v1"],
        },
      ]),
    ).toThrow(ContractDecodeError);
    expect(() =>
      decodeStaffMembers([
        {
          accountId: "administrator_1",
          fullName: "Administrator",
          roles: ["Administrator"],
          accessProfiles: ["StudentOnboardingManager.v1"],
          onboardingDelegationId: "not/a/backend/id",
        },
      ]),
    ).toThrow(ContractDecodeError);
    expect(() =>
      decodeStudentOnboardingItems([
        {
          studentId: "student_1",
          fullName: "Student",
          enrollmentReference: "BEL-001",
          teacherAccountId: "teacher_1",
          studentVersion: 1,
          onboardingState: "ready_to_invite",
          invitationId: "invitation_1",
          invitationExpiresAt: "2026-09-01T10:00:00Z",
        },
      ]),
    ).toThrow(ContractDecodeError);
  });

  it("requires the canonical requestId in API error envelopes", () => {
    expect(() =>
      decodeApiErrorEnvelope({
        error: { code: "FORBIDDEN", message: "denied" },
      }),
    ).toThrow(ContractDecodeError);
  });

  it("mirrors the full OpenAPI error vocabulary without adding a readiness route", () => {
    expect(
      decodeApiErrorEnvelope({
        error: {
          code: "UNAVAILABLE",
          message: "database unavailable",
          requestId: "request_ready_1",
        },
      }),
    ).toMatchObject({ error: { code: "UNAVAILABLE" } });
  });

  it("accepts only an absent 204 response body", () => {
    expect(decodeVoid(undefined)).toBeUndefined();
    expect(() => decodeVoid(null)).toThrow(ContractDecodeError);
  });

  it("requires First Minute revisions to be positive", () => {
    expect(() =>
      decodeFirstMinute({
        studentId: "student_1",
        revision: 0,
        whatWorked: "worked",
        currentFocus: "focus",
        nextStep: "next",
        publishedAt: "2026-09-01T10:00:00Z",
      }),
    ).toThrow(ContractDecodeError);
  });

  it("rejects duplicate authority values and malformed server tokens", () => {
    expect(() =>
      decodeBootstrapView({
        accountId: "account_1",
        roles: ["Owner", "Owner"],
        accessProfiles: [],
        permissions: [],
      }),
    ).toThrow(ContractDecodeError);
    expect(() =>
      decodeSessionTokens({
        accessToken: "short",
        refreshToken: "R".repeat(43),
        accessExpiresAt: "2026-08-01T10:00:00Z",
        refreshExpiresAt: "2026-09-01T10:00:00Z",
      }),
    ).toThrow(ContractDecodeError);
  });

  it("rejects impossible Gregorian dates", () => {
    for (const accessExpiresAt of [
      "2026-02-29T00:00:00Z",
      "2026-02-31T00:00:00Z",
      "2026-04-31T00:00:00Z",
      "2026-01-01T24:00:00Z",
    ]) {
      expect(() =>
        decodeSessionTokens({
          accessToken: "A".repeat(43),
          refreshToken: "R".repeat(43),
          accessExpiresAt,
          refreshExpiresAt: "2026-09-01T10:00:00Z",
        }),
      ).toThrow(ContractDecodeError);
    }
    expect(() =>
      decodeSessionTokens({
        accessToken: "A".repeat(43),
        refreshToken: "R".repeat(43),
        accessExpiresAt: "2028-02-29T00:00:00Z",
        refreshExpiresAt: "2028-09-01T10:00:00Z",
      }),
    ).not.toThrow();
  });

  it("accepts only canonical trusted invitation links", () => {
    const base = {
      invitationId: "invitation_1",
      studentId: "student_1",
      status: "issued",
      expiresAt: "2026-09-01T10:00:00Z",
    };
    expect(
      decodeInvitationResult(
        {
          ...base,
          activationLink: `https://app.example/activate#token=${"T".repeat(43)}`,
        },
        { allowedHttpsOrigins: ["https://app.example"] },
      ),
    ).toMatchObject({ invitationId: "invitation_1" });
    for (const activationLink of [
      `https://evil.example/activate#token=${"T".repeat(43)}`,
      `belcanto://activate/#token=${"T".repeat(43)}`,
      `https://app.example/activate?token=${"T".repeat(43)}`,
      `https://app.example/activate#token=${"T".repeat(43)}&extra=1`,
      "javascript:alert(1)",
    ]) {
      expect(() =>
        decodeInvitationResult(
          { ...base, activationLink },
          { allowedHttpsOrigins: ["https://app.example"] },
        ),
      ).toThrow(ContractDecodeError);
    }
  });
});
