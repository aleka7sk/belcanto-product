import {
  createIdempotentSubmission,
  interpretBootstrap,
  prepareCompleteActivation,
  prepareCreateStudent,
  prepareFirstMinute,
  prepareGrantDelegation,
  prepareInvitation,
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
