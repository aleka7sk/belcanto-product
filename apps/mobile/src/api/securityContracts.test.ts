import { ApiClient } from "./client";
import {
  decodeActivationProgress,
  decodeRecoveryCodes,
  decodeSecurityEventsPage,
  decodeSessionDevices,
  decodeSignInOutcome,
  decodeTwofaEnrollment,
  decodeTwofaStatus,
  decodeVerifiedContacts,
} from "./contracts";

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

describe("session security contracts (Page 32)", () => {
  const device = {
    sessionId: "session_1",
    deviceLabel: "iPhone 17",
    platform: "ios",
    createdAt: "2026-08-01T10:00:00Z",
    lastSeenAt: "2026-08-02T10:00:00Z",
    current: true,
  };

  it("decodes the session inventory and keeps optional metadata optional", () => {
    const devices = decodeSessionDevices([
      device,
      { sessionId: "session_2", createdAt: "2026-07-01T10:00:00Z", current: false },
    ]);
    expect(devices).toHaveLength(2);
    expect(devices[0]).toMatchObject({ deviceLabel: "iPhone 17", platform: "ios" });
    expect(devices[1]).not.toHaveProperty("deviceLabel");
  });

  it("rejects duplicate session ids and a second current session", () => {
    expect(() => decodeSessionDevices([device, device])).toThrow(
      "SessionDeviceList",
    );
    expect(() =>
      decodeSessionDevices([
        device,
        { ...device, sessionId: "session_2" },
      ]),
    ).toThrow("SessionDeviceList");
  });

  it("rejects unknown platforms and unexpected keys", () => {
    expect(() =>
      decodeSessionDevices([{ ...device, platform: "windows" }]),
    ).toThrow("SessionDeviceList");
    expect(() =>
      decodeSessionDevices([{ ...device, extra: true }]),
    ).toThrow("SessionDeviceList");
  });

  it("decodes the security feed and enforces newest-first ordering", () => {
    const page = decodeSecurityEventsPage({
      events: [
        {
          id: 9,
          action: "OtherSessionsRevoked",
          decision: "allow",
          targetType: "account",
          targetId: "account_1",
          recordedAt: "2026-08-02T10:00:00Z",
        },
        {
          id: 4,
          action: "RefreshTokenReuseDetected",
          decision: "deny",
          reasonCode: "inactive_or_reused_refresh_token",
          recordedAt: "2026-08-01T10:00:00Z",
        },
      ],
      nextCursor: "djE6NA",
    });
    expect(page.events.map((event) => event.id)).toEqual([9, 4]);
    expect(page.nextCursor).toBe("djE6NA");
    expect(() =>
      decodeSecurityEventsPage({
        events: [
          { id: 4, action: "SessionCreated", decision: "allow", recordedAt: "2026-08-01T10:00:00Z" },
          { id: 9, action: "SessionCreated", decision: "allow", recordedAt: "2026-08-02T10:00:00Z" },
        ],
      }),
    ).toThrow("SecurityEventsPage");
    expect(() =>
      decodeSecurityEventsPage({
        events: [
          { id: 1, action: "SomethingElse", decision: "allow", recordedAt: "2026-08-01T10:00:00Z" },
        ],
      }),
    ).toThrow("SecurityEventsPage");
  });

  it("lists sessions over the exact bearer route", async () => {
    const fetch = fetchMock(mockResponse(200, [device]));
    const api = new ApiClient({ baseUrl: "https://api.example", fetch });
    const devices = await api.listMySessions("access");
    expect(devices[0]?.current).toBe(true);
    const [url, init] = fetch.mock.calls[0]!;
    expect(url).toBe("https://api.example/v1/me/sessions");
    expect(init?.method).toBe("GET");
    expect(init?.headers).toMatchObject({ Authorization: "Bearer access" });
  });

  it("builds the per-session revoke path and validates identifiers before fetch", async () => {
    const fetch = fetchMock(mockResponse(204));
    const api = new ApiClient({ baseUrl: "https://api.example", fetch });
    await api.revokeMySession("access", "session_2", {
      currentPassword: "current-password",
    });
    const [url] = fetch.mock.calls[0]!;
    expect(url).toBe("https://api.example/v1/me/sessions/session_2/revoke");
    expect(() =>
      api.revokeMySession("access", "../evil", { currentPassword: "x" }),
    ).toThrow(TypeError);
    expect(fetch).toHaveBeenCalledTimes(1);
  });

  it("builds security-event queries and rejects invalid limits before fetch", async () => {
    const fetch = fetchMock(mockResponse(200, { events: [] }));
    const api = new ApiClient({ baseUrl: "https://api.example", fetch });
    await api.listSecurityEvents("access", { cursor: "djE6NA", limit: 25 });
    const [url] = fetch.mock.calls[0]!;
    expect(url).toBe(
      "https://api.example/v1/me/security-events?cursor=djE6NA&limit=25",
    );
    expect(() => api.listSecurityEvents("access", { limit: 51 })).toThrow(
      TypeError,
    );
    expect(fetch).toHaveBeenCalledTimes(1);
  });

  it("keeps password recovery public and one-shot", async () => {
    const fetch = fetchMock(mockResponse(204));
    const api = new ApiClient({ baseUrl: "https://api.example", fetch });
    await api.requestPasswordReset({ phone: "+77000000001" });
    const [url, init] = fetch.mock.calls[0]!;
    expect(url).toBe("https://api.example/v1/password-resets");
    expect(init?.headers).not.toHaveProperty("Authorization");
    await api.completePasswordReset({
      token: "T".repeat(43),
      newPassword: "Rotated-password-123!",
    });
    const [completeUrl] = fetch.mock.calls[1]!;
    expect(completeUrl).toBe("https://api.example/v1/password-resets/complete");
  });
});

describe("two-factor and activation contracts (AUTH-01..10, ACC-03/06)", () => {
  it("decodes the sign-in union and rejects mixed or empty shapes", () => {
    const tokens = {
      accessToken: "A".repeat(43),
      refreshToken: "R".repeat(43),
      accessExpiresAt: "2026-08-01T10:00:00Z",
      refreshExpiresAt: "2026-09-01T10:00:00Z",
    };
    expect(decodeSignInOutcome({ tokens }).tokens?.accessToken).toBe(tokens.accessToken);
    const challenge = decodeSignInOutcome({
      twofaChallenge: "C".repeat(43),
      twofaExpiresAt: "2026-08-01T10:05:00Z",
    });
    expect(challenge.twofaChallenge).toBe("C".repeat(43));
    expect(() => decodeSignInOutcome({})).toThrow("SignInOutcome");
    expect(() =>
      decodeSignInOutcome({ tokens, twofaChallenge: "C".repeat(43) }),
    ).toThrow("SignInOutcome");
  });

  it("decodes activation progress and requires a kind for verified contacts", () => {
    const progress = decodeActivationProgress({
      invitationId: "inv_1",
      kind: "staff_activation",
      displayName: "Шугыла Замещающая",
      expiresAt: "2026-08-08T10:00:00Z",
      passwordSet: true,
      contactKind: "email",
      contactMasked: "s******@example.kz",
      contactVerified: true,
      twofaEnrolled: false,
      completed: false,
    });
    expect(progress.contactKind).toBe("email");
    expect(() =>
      decodeActivationProgress({
        invitationId: "inv_1",
        kind: "staff_activation",
        displayName: "X",
        expiresAt: "2026-08-08T10:00:00Z",
        passwordSet: true,
        contactVerified: true,
        twofaEnrolled: false,
        completed: false,
      }),
    ).toThrow("ActivationProgressView");
  });

  it("decodes verified contacts and rejects duplicate kinds", () => {
    const email = {
      id: "contact_1",
      kind: "email",
      value: "owner@belcanto.kz",
      verifiedAt: "2026-08-01T10:00:00Z",
    };
    expect(decodeVerifiedContacts([email])).toHaveLength(1);
    expect(() =>
      decodeVerifiedContacts([email, { ...email, id: "contact_2" }]),
    ).toThrow("VerifiedContactList");
  });

  it("keeps enabled twofa status coupled to its confirmation timestamp", () => {
    expect(
      decodeTwofaStatus({
        enabled: true,
        confirmedAt: "2026-08-01T10:00:00Z",
        recoveryCodesRemaining: 10,
      }).enabled,
    ).toBe(true);
    expect(() =>
      decodeTwofaStatus({ enabled: true, recoveryCodesRemaining: 10 }),
    ).toThrow("TwofaStatus");
  });

  it("requires an otpauth provisioning URI and exactly ten unique recovery codes", () => {
    expect(() =>
      decodeTwofaEnrollment({ secret: "S".repeat(32), provisioningUri: "https://x" }),
    ).toThrow("TwofaEnrollment");
    const normalized = Array.from({ length: 10 }, (_, index) =>
      `ABCD-EFGH-${["JK", "MN", "PQ", "RS", "TU", "VW", "XY", "Z2", "34", "56"][index]}`,
    );
    expect(decodeRecoveryCodes({ recoveryCodes: normalized })).toHaveLength(10);
    expect(() =>
      decodeRecoveryCodes({ recoveryCodes: normalized.slice(0, 9) }),
    ).toThrow("RecoveryCodesResponse");
    expect(() =>
      decodeRecoveryCodes({
        recoveryCodes: [...normalized.slice(0, 9), normalized[0]!],
      }),
    ).toThrow("RecoveryCodesResponse");
  });
});
