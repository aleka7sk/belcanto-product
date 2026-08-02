import {
  createIdempotentSubmission,
  interpretBootstrap,
  prepareCompleteActivation,
  prepareCreateStudent,
  prepareFirstMinute,
  prepareGrantDelegation,
  prepareInvitation,
  prepareCreateLesson,
  prepareReassignPrimaryTeachers,
  prepareReplaceLessonTeachers,
  prepareRevokeInvitation,
  prepareSignIn,
  type IdempotencyKeyFactory,
} from ".";
import type { BootstrapView } from "@/api/contracts";

const token = "A".repeat(43);

describe("pure request controllers", () => {
  it("prepares an Owner delegation without exposing idempotency in form data", () => {
    const draft = {
      administratorAccountId: " admin_1 ",
      reason: " operational delegation ",
      expiresAt: "2026-08-02T00:00:00Z",
      currentPassword: "owner-password",
    };
    expect(draft).not.toHaveProperty("idempotencyKey");
    expect(
      prepareGrantDelegation(
        draft,
        "intent_1",
        new Date("2026-08-01T00:00:00Z"),
      ),
    ).toMatchObject({
      ok: true,
      value: {
        idempotencyKey: "intent_1",
        body: { administratorAccountId: "admin_1" },
      },
    });
  });

  it("normalizes Student creation and applies backend defaults", () => {
    expect(
      prepareCreateStudent(
        {
          fullName: " Student ",
          phone: "+7 (700) 000-00-00",
          enrollmentReference: " enrollment_1 ",
          teacherAccountId: " teacher_1 ",
          locale: "",
          timezone: "",
          adultConfirmed: true,
        },
        "intent_2",
      ),
    ).toEqual({
      ok: true,
      value: {
        body: {
          fullName: "Student",
          phone: "+77000000000",
          enrollmentReference: "enrollment_1",
          teacherAccountId: "teacher_1",
          locale: "ru-KZ",
          timezone: "Asia/Almaty",
          adultConfirmed: true,
        },
        idempotencyKey: "intent_2",
      },
    });
  });

  it("fails client-side at the finalized Student input bounds", () => {
    expect(
      prepareCreateStudent(
        {
          fullName: "x".repeat(201),
          phone: "+77000000000",
          enrollmentReference: "x".repeat(101),
          teacherAccountId: "not/a/backend/id",
          locale: "not_a_locale",
          timezone: "x".repeat(101),
          adultConfirmed: true,
        },
        "intent_bounds",
      ),
    ).toMatchObject({
      ok: false,
      issues: expect.arrayContaining([
        { field: "fullName", code: "too_long" },
        { field: "enrollmentReference", code: "too_long" },
        { field: "teacherAccountId", code: "invalid_format" },
        { field: "locale", code: "invalid_format" },
        { field: "timezone", code: "too_long" },
      ]),
    });
  });

  it("validates all First Minute fields and optimistic version", () => {
    const result = prepareFirstMinute(
      {
        studentId: "student_1",
        whatWorked: "x".repeat(501),
        currentFocus: "focus",
        nextStep: "next",
        expectedVersion: -1,
      },
      "intent_3",
    );
    expect(result).toMatchObject({
      ok: false,
      issues: expect.arrayContaining([
        { field: "whatWorked", code: "too_long" },
        { field: "expectedVersion", code: "invalid_value" },
      ]),
    });
  });

  it("keeps invitation actions explicit", () => {
    expect(
      prepareInvitation({ studentId: "student_1" }, "intent_4", "reissue"),
    ).toEqual({
      ok: true,
      value: {
        studentId: "student_1",
        idempotencyKey: "intent_4",
        mode: "reissue",
      },
    });
    expect(
      prepareRevokeInvitation({ invitationId: "inv_1" }, "intent_5"),
    ).toMatchObject({ ok: true, value: { invitationId: "inv_1" } });
  });

  it("converts human Almaty lesson time and keeps explicit Students", () => {
    expect(
      prepareCreateLesson(
        {
          title: " Индивидуальный урок ",
          startsOn: "10.08.2026",
          startsAtTime: "18:00",
          durationMinutes: "60",
          location: " Класс 2 ",
          teacherAccountId: "teacher_1",
          studentIds: ["student_1", "student_2"],
        },
        "lesson_intent",
        new Date("2026-08-02T00:00:00Z"),
      ),
    ).toEqual({
      ok: true,
      value: {
        body: {
          title: "Индивидуальный урок",
          startsAt: "2026-08-10T13:00:00.000Z",
          durationMinutes: 60,
          location: "Класс 2",
          teacherAccountId: "teacher_1",
          studentIds: ["student_1", "student_2"],
        },
        idempotencyKey: "lesson_intent",
      },
    });
  });

  it("uses stable server-clock mode and preserves exact assignment versions", () => {
    const draft = {
      students: [
        { studentId: "student_2", expectedAssignmentVersion: 5 },
        { studentId: "student_1", expectedAssignmentVersion: 3 },
      ],
      newTeacherAccountId: "teacher_2",
      effectiveImmediately: true,
      effectiveOn: "",
      effectiveAtTime: "",
    };
    const first = prepareReassignPrimaryTeachers(
      draft,
      "reassign_intent",
      new Date("2026-08-02T10:00:00Z"),
    );
    const retry = prepareReassignPrimaryTeachers(
      draft,
      "reassign_intent",
      new Date("2026-08-02T11:00:00Z"),
    );
    expect(first).toMatchObject({
      ok: true,
      value: {
        body: {
          students: [
            { studentId: "student_2", expectedAssignmentVersion: 5 },
            { studentId: "student_1", expectedAssignmentVersion: 3 },
          ],
          effectiveMode: "immediate",
        },
      },
    });
    expect(retry).toEqual(first);
    if (first.ok) expect(first.value.body).not.toHaveProperty("effectiveFrom");
  });

  it("converts scheduled reassignment time in Almaty", () => {
    expect(
      prepareReassignPrimaryTeachers(
        {
          students: [{ studentId: "student_1", expectedAssignmentVersion: 3 }],
          newTeacherAccountId: "teacher_2",
          effectiveImmediately: false,
          effectiveOn: "10.08.2026",
          effectiveAtTime: "09:00",
        },
        "scheduled_intent",
        new Date("2026-08-02T10:00:00Z"),
      ),
    ).toMatchObject({
      ok: true,
      value: {
        body: {
          effectiveMode: "scheduled",
          effectiveFrom: "2026-08-10T04:00:00.000Z",
        },
      },
    });
  });

  it("rejects empty selections and preserves exact Lesson concurrency guards", () => {
    expect(
      prepareReplaceLessonTeachers(
        { lessons: [], newTeacherAccountId: "teacher_2" },
        "replace_empty",
      ),
    ).toMatchObject({ ok: false, issues: [{ field: "lessons", code: "required" }] });
    expect(
      prepareReplaceLessonTeachers(
        {
          lessons: [
            {
              lessonId: "lesson_9",
              expectedVersion: 8,
              expectedPreviousTeacherAccountId: "teacher_9",
            },
            {
              lessonId: "lesson_2",
              expectedVersion: 1,
              expectedPreviousTeacherAccountId: "teacher_2",
            },
          ],
          newTeacherAccountId: "teacher_2",
        },
        "replace_intent",
      ),
    ).toMatchObject({
      ok: true,
      value: {
        body: {
          lessons: [
            {
              lessonId: "lesson_9",
              expectedVersion: 8,
              expectedPreviousTeacherAccountId: "teacher_9",
            },
            {
              lessonId: "lesson_2",
              expectedVersion: 1,
              expectedPreviousTeacherAccountId: "teacher_2",
            },
          ],
        },
      },
    });
  });

  it("enforces the backend limit of 100 selected objects", () => {
    const studentIds = Array.from({ length: 101 }, (_, index) => `student_${index}`);
    expect(
      prepareCreateLesson(
        {
          title: "Урок",
          startsOn: "10.08.2026",
          startsAtTime: "18:00",
          durationMinutes: "60",
          location: "",
          teacherAccountId: "teacher_1",
          studentIds,
        },
        "lesson_limit",
        new Date("2026-08-02T00:00:00Z"),
      ),
    ).toMatchObject({
      ok: false,
      issues: expect.arrayContaining([{ field: "studentIds", code: "invalid_value" }]),
    });
    expect(
      prepareReassignPrimaryTeachers(
        {
          students: studentIds.map((studentId) => ({
            studentId,
            expectedAssignmentVersion: 1,
          })),
          newTeacherAccountId: "teacher_2",
          effectiveImmediately: true,
          effectiveOn: "",
          effectiveAtTime: "",
        },
        "reassign_limit",
      ),
    ).toMatchObject({
      ok: false,
      issues: expect.arrayContaining([{ field: "students", code: "invalid_value" }]),
    });
    expect(
      prepareReplaceLessonTeachers(
        {
          lessons: studentIds.map((_, index) => ({
            lessonId: `lesson_${index}`,
            expectedVersion: index,
            expectedPreviousTeacherAccountId: "teacher_1",
          })),
          newTeacherAccountId: "teacher_2",
        },
        "replace_limit",
      ),
    ).toMatchObject({
      ok: false,
      issues: expect.arrayContaining([{ field: "lessons", code: "invalid_value" }]),
    });
  });

  it("prepares activation but never creates a session", () => {
    const result = prepareCompleteActivation(
      {
        token,
        phone: "+77000000000",
        password: "correct horse battery staple",
        passwordConfirmation: "correct horse battery staple",
      },
      "intent_6",
    );
    expect(result).toMatchObject({
      ok: true,
      value: { body: { token, phone: "+77000000000" } },
    });
    if (result.ok) expect(result.value).not.toHaveProperty("session");
  });

  it("normalizes sign-in independently from activation password rules", () => {
    expect(
      prepareSignIn({ phone: "+7 700 000 00 00", password: "existing" }),
    ).toEqual({
      ok: true,
      value: { phone: "+77000000000", password: "existing" },
    });
  });

  it("fails bootstrap interpretation when Student scope is inconsistent", () => {
    expect(
      interpretBootstrap({
        accountId: "account_1",
        roles: ["Student"],
        accessProfiles: [],
        permissions: [],
      }),
    ).toEqual({ ready: false, reason: "student_identity_missing" });

    expect(
      interpretBootstrap(
        {
          accountId: "account_1",
          roles: ["Student"],
          accessProfiles: [],
          permissions: [],
          studentId: "student_1",
        } as unknown as BootstrapView,
      ),
    ).toEqual({ ready: false, reason: "student_full_name_missing" });

    expect(
      interpretBootstrap(
        {
          accountId: "account_1",
          roles: ["Student"],
          accessProfiles: [],
          permissions: [],
          studentId: "student_1",
          fullName: "Student",
        } as unknown as BootstrapView,
      ),
    ).toEqual({ ready: false, reason: "first_minute_missing" });

    expect(
      interpretBootstrap({
        accountId: "account_1",
        roles: ["Owner"],
        accessProfiles: [],
        permissions: [],
      }),
    ).toMatchObject({ ready: true });
  });
});

describe("idempotent submission boundary", () => {
  it("keeps one key for retries and rotates only after completion", () => {
    let sequence = 0;
    const factory: IdempotencyKeyFactory = {
      create: () => `key_${++sequence}`,
    };
    const submission = createIdempotentSubmission(
      (input: string, idempotencyKey) => ({ input, idempotencyKey }),
      factory,
    );
    expect(submission.prepare("first").idempotencyKey).toBe("key_1");
    expect(submission.prepare("retry").idempotencyKey).toBe("key_1");
    submission.succeeded();
    expect(submission.prepare("next-intent").idempotencyKey).toBe("key_2");
  });
});
