import { ApiClient, ApiTransportError } from "./client";

function mockResponse(status: number, body?: unknown): Response {
  return {
    ok: status >= 200 && status < 300,
    status,
    text: async () => (body === undefined ? "" : JSON.stringify(body)),
  } as Response;
}

function fetchMock(response: Response): jest.MockedFunction<typeof fetch> {
  return jest.fn(async () => response) as unknown as jest.MockedFunction<
    typeof fetch
  >;
}

describe("ApiClient", () => {
  it("uses the stable sign-in route without authorization", async () => {
    const fetch = fetchMock(
      mockResponse(200, {
        accessToken: "A".repeat(43),
        refreshToken: "R".repeat(43),
        accessExpiresAt: "2026-08-01T10:00:00Z",
        refreshExpiresAt: "2026-09-01T10:00:00Z",
      }),
    );
    const api = new ApiClient({ baseUrl: "https://api.example/", fetch });
    await api.signIn({ phone: "+77000000000", password: "secret" });
    expect(fetch).toHaveBeenCalledTimes(1);
    const [url, init] = fetch.mock.calls[0]!;
    expect(url).toBe("https://api.example/v1/sessions");
    expect(init?.method).toBe("POST");
    expect(init?.headers).toMatchObject({
      Accept: "application/json",
      "Content-Type": "application/json",
    });
    expect(init?.headers).not.toHaveProperty("Authorization");
  });

  it("sends bearer and stable idempotency headers for student creation", async () => {
    const fetch = fetchMock(
      mockResponse(201, {
        studentId: "student_1",
        accountId: "account_1",
        onboardingState: "awaiting_first_minute",
      }),
    );
    const api = new ApiClient({ baseUrl: "https://api.example", fetch });
    await api.createStudent(
      "access",
      {
        fullName: "Student",
        phone: "+77000000000",
        enrollmentReference: "enrollment_1",
        teacherAccountId: "teacher_1",
        locale: "ru-KZ",
        timezone: "Asia/Almaty",
        adultConfirmed: true,
      },
      "intent-key",
    );
    const [, init] = fetch.mock.calls[0]!;
    expect(init?.headers).toMatchObject({
      Authorization: "Bearer access",
      "Idempotency-Key": "intent-key",
    });
  });

  it("decodes the exact API error envelope", async () => {
    const fetch = fetchMock(
      mockResponse(403, {
        error: { code: "FORBIDDEN", message: "denied", requestId: "request_1" },
      }),
    );
    const api = new ApiClient({ baseUrl: "https://api.example", fetch });
    await expect(api.listStaff("access", "Teacher")).rejects.toMatchObject({
      status: 403,
      code: "FORBIDDEN",
      requestId: "request_1",
    });
  });

  it("uses the exact staff role query and onboarding queue routes", async () => {
    const fetch = fetchMock(mockResponse(200, []));
    const api = new ApiClient({ baseUrl: "https://api.example", fetch });

    await api.listStaff("access", "Administrator");
    await api.listStudentOnboarding("access");

    expect(fetch.mock.calls[0]?.[0]).toBe(
      "https://api.example/v1/staff?role=Administrator",
    );
    expect(fetch.mock.calls[1]?.[0]).toBe(
      "https://api.example/v1/student-onboarding",
    );
    for (const [, init] of fetch.mock.calls) {
      expect(init?.method).toBe("GET");
      expect(init?.headers).toMatchObject({ Authorization: "Bearer access" });
      expect(init?.headers).not.toHaveProperty("Content-Type");
    }
  });

  it("does not invoke fetch for an already-aborted signal", async () => {
    const fetch = fetchMock(mockResponse(204));
    const api = new ApiClient({ baseUrl: "https://api.example", fetch });
    const controller = new AbortController();
    controller.abort("caller");
    await expect(api.signOut("access", controller.signal)).rejects.toBeInstanceOf(
      ApiTransportError,
    );
    expect(fetch).not.toHaveBeenCalled();
  });

  it("rejects insecure non-local API origins", () => {
    expect(() => new ApiClient({ baseUrl: "http://api.example" })).toThrow(
      "must use HTTPS",
    );
    expect(() => new ApiClient({ baseUrl: "http://localhost:8080" })).not.toThrow();
  });

  it("allows private HTTP origins only behind the development option", () => {
    for (const baseUrl of [
      "http://localhost:8080",
      "http://10.0.2.2:8080",
      "http://192.168.1.24:8080",
      "http://172.20.0.4:8080",
    ]) {
      expect(
        () =>
          new ApiClient({
            baseUrl,
            allowInsecureDevelopmentOrigin: true,
          }),
      ).not.toThrow();
    }
    for (const baseUrl of ["http://api.example", "http://8.8.8.8"]) {
      expect(
        () =>
          new ApiClient({
            baseUrl,
            allowInsecureDevelopmentOrigin: true,
          }),
      ).toThrow("must use HTTPS");
    }
  });

  it("requires a clean API origin", () => {
    for (const baseUrl of [
      "https://user:pass@api.example",
      "https://api.example/v1",
      "https://api.example?tenant=one",
      "https://api.example#fragment",
    ]) {
      expect(() => new ApiClient({ baseUrl })).toThrow(
        "must be an origin without credentials or path",
      );
    }
  });

  it("rejects invalid path identifiers before fetch", () => {
    const fetch = fetchMock(mockResponse(204));
    const api = new ApiClient({ baseUrl: "https://api.example", fetch });
    expect(() =>
      api.revokeInvitation("access", "not/a/backend/id", "intent-key"),
    ).toThrow("valid backend identifier");
    expect(fetch).not.toHaveBeenCalled();
  });

  it("rejects a valid body delivered with the wrong success status", async () => {
    const completeFetch = fetchMock(mockResponse(200));
    const completeApi = new ApiClient({
      baseUrl: "https://api.example",
      fetch: completeFetch,
    });
    await expect(
      completeApi.completeActivation(
        {
          token: "T".repeat(43),
          phone: "+77000000000",
          password: "correct horse battery staple",
        },
        "intent-key",
      ),
    ).rejects.toMatchObject({ code: "UNEXPECTED_RESPONSE", status: 200 });

    const studentFetch = fetchMock(
      mockResponse(200, {
        studentId: "student_1",
        accountId: "account_1",
        onboardingState: "awaiting_first_minute",
      }),
    );
    const studentApi = new ApiClient({
      baseUrl: "https://api.example",
      fetch: studentFetch,
    });
    await expect(
      studentApi.createStudent(
        "access",
        {
          fullName: "Student",
          phone: "+77000000000",
          enrollmentReference: "BEL-001",
          teacherAccountId: "teacher_1",
          adultConfirmed: true,
        },
        "intent-key",
      ),
    ).rejects.toMatchObject({ code: "UNEXPECTED_RESPONSE", status: 200 });
  });

  it("rejects mismatched HTTP error status and error code", async () => {
    const fetch = fetchMock(
      mockResponse(401, {
        error: {
          code: "FORBIDDEN",
          message: "denied",
          requestId: "request_1",
        },
      }),
    );
    const api = new ApiClient({ baseUrl: "https://api.example", fetch });
    await expect(api.bootstrap("access")).rejects.toMatchObject({
      status: 401,
      code: "UNEXPECTED_RESPONSE",
      requestId: "request_1",
    });
  });

  it("rejects a globally valid error not declared for the route", async () => {
    const fetch = fetchMock(
      mockResponse(409, {
        error: {
          code: "CONFLICT",
          message: "conflict",
          requestId: "request_2",
        },
      }),
    );
    const api = new ApiClient({ baseUrl: "https://api.example", fetch });
    await expect(api.bootstrap("access")).rejects.toMatchObject({
      status: 409,
      code: "UNEXPECTED_RESPONSE",
      requestId: "request_2",
    });
  });

  it("applies the runtime activation-link trust policy to invitation results", async () => {
    const body = {
      invitationId: "invitation_1",
      studentId: "student_1",
      status: "issued",
      expiresAt: "2026-09-01T10:00:00Z",
      activationLink: `https://app.example/activate#token=${"T".repeat(43)}`,
    };
    const trusted = new ApiClient({
      baseUrl: "https://api.example",
      fetch: fetchMock(mockResponse(201, body)),
      activationLinkPolicy: {
        allowedHttpsOrigins: ["https://app.example"],
      },
    });
    await expect(
      trusted.issueInvitation("access", "student_1", "intent-key"),
    ).resolves.toMatchObject({ invitationId: "invitation_1" });

    const untrusted = new ApiClient({
      baseUrl: "https://api.example",
      fetch: fetchMock(mockResponse(201, body)),
    });
    await expect(
      untrusted.issueInvitation("access", "student_1", "intent-key"),
    ).rejects.toMatchObject({ code: "UNEXPECTED_RESPONSE" });
  });
});
