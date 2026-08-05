import { ApiClient } from "./client";
import {
  decodeDataExport,
  decodeDataExports,
  decodeDeletionRequest,
  decodePolicyVersions,
  decodePrivacySettings,
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

describe("policy and privacy contracts (ACC-10..16)", () => {
  const policy = {
    id: "polv_1",
    kind: "privacy",
    version: "2.0",
    title: "Политика конфиденциальности",
    bodyRef: "policies/privacy/2.0",
    effectiveFrom: "2026-07-01T00:00:00Z",
  };

  it("decodes the policy catalog and keeps acceptance optional", () => {
    const policies = decodePolicyVersions([
      { ...policy, acceptedAt: "2026-08-01T10:00:00Z" },
      { ...policy, id: "polv_2", kind: "terms" },
    ]);
    expect(policies).toHaveLength(2);
    expect(policies[0]).toMatchObject({ acceptedAt: "2026-08-01T10:00:00Z" });
    expect(policies[1]).not.toHaveProperty("acceptedAt");
  });

  it("rejects unknown policy kinds, duplicate ids and extra keys", () => {
    expect(() =>
      decodePolicyVersions([{ ...policy, kind: "marketing" }]),
    ).toThrow("PolicyVersionList");
    expect(() => decodePolicyVersions([policy, policy])).toThrow(
      "PolicyVersionList",
    );
    expect(() => decodePolicyVersions([{ ...policy, extra: 1 }])).toThrow(
      "PolicyVersionList",
    );
  });

  const settings = {
    communityProfileVisible: true,
    achievementsVisible: true,
    staffMessagesAllowed: false,
    mentionsAllowed: true,
    pushPreview: "hidden",
    version: 3,
  };

  it("decodes privacy settings and rejects unknown preview modes", () => {
    expect(decodePrivacySettings(settings)).toEqual(settings);
    expect(() =>
      decodePrivacySettings({ ...settings, pushPreview: "everything" }),
    ).toThrow("PrivacySettings");
    expect(() =>
      decodePrivacySettings({ ...settings, version: -1 }),
    ).toThrow("PrivacySettings");
    expect(() =>
      decodePrivacySettings({ ...settings, extra: true }),
    ).toThrow("PrivacySettings");
  });
});

describe("data rights contracts (ACC-17/18, DEC-104-safe)", () => {
  const openExport = {
    id: "export_1",
    status: "requested",
    requestedAt: "2026-08-01T10:00:00Z",
  };

  it("enforces the readiness invariant on exports", () => {
    expect(decodeDataExport(openExport)).toMatchObject({ status: "requested" });
    expect(
      decodeDataExport({
        ...openExport,
        status: "ready",
        readyAt: "2026-08-02T10:00:00Z",
        expiresAt: "2026-08-09T10:00:00Z",
      }),
    ).toMatchObject({ status: "ready" });
    expect(() =>
      decodeDataExport({ ...openExport, status: "ready" }),
    ).toThrow("DataExportRequest");
    expect(() =>
      decodeDataExport({ ...openExport, readyAt: "2026-08-02T10:00:00Z" }),
    ).toThrow("DataExportRequest");
  });

  it("rejects a second open export in the history", () => {
    expect(
      decodeDataExports([openExport, { ...openExport, id: "export_2", status: "cancelled" }]),
    ).toHaveLength(2);
    expect(() =>
      decodeDataExports([openExport, { ...openExport, id: "export_2", status: "processing" }]),
    ).toThrow("DataExportRequestList");
  });

  it("caps the export history at 10 entries", () => {
    const history = Array.from({ length: 11 }, (_, index) => ({
      id: `export_${index}`,
      status: "cancelled",
      requestedAt: "2026-08-01T10:00:00Z",
    }));
    expect(() => decodeDataExports(history)).toThrow("DataExportRequestList");
  });

  const deletion = {
    id: "delreq_1",
    status: "pending_review",
    requestedAt: "2026-08-01T10:00:00Z",
  };

  it("binds cancelledAt to the cancelled status exactly", () => {
    expect(decodeDeletionRequest(deletion)).toMatchObject({
      status: "pending_review",
    });
    expect(
      decodeDeletionRequest({
        ...deletion,
        status: "cancelled",
        cancelledAt: "2026-08-02T10:00:00Z",
      }),
    ).toMatchObject({ status: "cancelled" });
    expect(() =>
      decodeDeletionRequest({ ...deletion, status: "cancelled" }),
    ).toThrow("DeletionRequest");
    expect(() =>
      decodeDeletionRequest({
        ...deletion,
        cancelledAt: "2026-08-02T10:00:00Z",
      }),
    ).toThrow("DeletionRequest");
  });

  it("rejects statuses that could schedule erasure", () => {
    expect(() =>
      decodeDeletionRequest({ ...deletion, status: "scheduled" }),
    ).toThrow("DeletionRequest");
    expect(() =>
      decodeDeletionRequest({ ...deletion, status: "completed" }),
    ).toThrow("DeletionRequest");
  });
});

describe("privacy client routes", () => {
  it("sends the optimistic version on privacy updates", async () => {
    const settings = {
      communityProfileVisible: true,
      achievementsVisible: false,
      staffMessagesAllowed: true,
      mentionsAllowed: true,
      pushPreview: "title",
      version: 2,
    } as const;
    const mock = fetchMock(mockResponse(200, { ...settings, version: 3 }));
    const client = new ApiClient({ baseUrl: "https://api.test", fetch: mock });
    const updated = await client.updatePrivacySettings("access-token", settings);
    expect(updated.version).toBe(3);
    const [url, init] = mock.mock.calls[0]!;
    expect(String(url)).toBe("https://api.test/v1/me/privacy");
    expect(init?.method).toBe("PUT");
    expect(JSON.parse(String(init?.body))).toMatchObject({ version: 2 });
  });

  it("requires re-authentication material for deletion requests", async () => {
    const mock = fetchMock(
      mockResponse(201, {
        id: "delreq_9",
        status: "pending_review",
        requestedAt: "2026-08-01T10:00:00Z",
      }),
    );
    const client = new ApiClient({ baseUrl: "https://api.test", fetch: mock });
    const created = await client.createDeletionRequest("access-token", {
      currentPassword: "Owner-password-1!",
    });
    expect(created.status).toBe("pending_review");
    const [url, init] = mock.mock.calls[0]!;
    expect(String(url)).toBe("https://api.test/v1/me/deletion-request");
    expect(init?.method).toBe("POST");
  });
});
