import { decodeRescheduleRequest, decodeRescheduleRequests } from "./contracts";

describe("reschedule request contracts (flows J/K/L)", () => {
  const request = {
    id: "resch_1",
    occurrenceId: "cocc_1",
    kind: "reschedule",
    proposedStartsAt: "2026-08-12T10:00:00Z",
    reason: "Совпадает со школьной олимпиадой",
    status: "pending",
    requestedBy: { accountId: "account_s", fullName: "Алишер Беков" },
    createdAt: "2026-08-06T09:00:00Z",
    version: 0,
  };

  it("binds the proposed time to the reschedule kind", () => {
    expect(decodeRescheduleRequest(request).kind).toBe("reschedule");
    expect(() =>
      decodeRescheduleRequest({ ...request, kind: "cancellation" }),
    ).toThrow("RescheduleRequest");
    const { proposedStartsAt: _omitted, ...withoutTime } = request;
    expect(() => decodeRescheduleRequest(withoutTime)).toThrow(
      "RescheduleRequest",
    );
    expect(
      decodeRescheduleRequest({ ...withoutTime, kind: "cancellation" }).kind,
    ).toBe("cancellation");
  });

  it("binds decidedAt to decided statuses exactly", () => {
    expect(() =>
      decodeRescheduleRequest({ ...request, status: "approved" }),
    ).toThrow("RescheduleRequest");
    expect(
      decodeRescheduleRequest({
        ...request,
        status: "approved",
        decidedAt: "2026-08-06T12:00:00Z",
        decisionNote: "Согласовано",
      }).status,
    ).toBe("approved");
    expect(() =>
      decodeRescheduleRequest({
        ...request,
        decidedAt: "2026-08-06T12:00:00Z",
      }),
    ).toThrow("RescheduleRequest");
  });

  it("rejects duplicate ids and unknown keys in the list", () => {
    expect(decodeRescheduleRequests([request])).toHaveLength(1);
    expect(() => decodeRescheduleRequests([request, request])).toThrow(
      "RescheduleRequestList",
    );
    expect(() =>
      decodeRescheduleRequest({ ...request, extra: 1 }),
    ).toThrow("RescheduleRequest");
  });
});
