import { ApiClient } from "./client";
import {
  decodeSecurityEventsPage,
  decodeSessionDevices,
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
