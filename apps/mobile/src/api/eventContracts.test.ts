import {
  decodeEventCategories,
  decodeEventOccurrence,
  decodeEventSeries,
} from "./contracts";

describe("event contracts (DEC-001/003/101)", () => {
  const occurrence = {
    id: "evocc_1",
    categoryId: "evcat_1",
    categoryName: "Мастер-класс",
    title: "Мастер-класс по дыханию",
    startsAt: "2026-08-10T10:00:00Z",
    durationMinutes: 90,
    host: { accountId: "account_t", fullName: "Диана Садыкова" },
    capacity: 3,
    confirmedCount: 2,
    status: "scheduled",
    version: 0,
  };

  it("decodes the catalog entry and keeps caller projections optional", () => {
    const decoded = decodeEventOccurrence(occurrence);
    expect(decoded).not.toHaveProperty("myRsvp");
    expect(decoded).not.toHaveProperty("myOffer");
  });

  it("binds the seat invariants structurally", () => {
    expect(() =>
      decodeEventOccurrence({ ...occurrence, confirmedCount: 4 }),
    ).toThrow("EventOccurrence");
    expect(() =>
      decodeEventOccurrence({
        ...occurrence,
        myOffer: {
          id: "offer_1",
          occurrenceId: "evocc_OTHER",
          status: "pending",
          offeredAt: "2026-08-05T10:00:00Z",
          expiresAt: "2026-08-06T10:00:00Z",
        },
      }),
    ).toThrow("EventOccurrence");
    expect(() =>
      decodeEventOccurrence({
        ...occurrence,
        myWaitlistPosition: 1,
        myOffer: {
          id: "offer_1",
          occurrenceId: "evocc_1",
          status: "pending",
          offeredAt: "2026-08-05T10:00:00Z",
          expiresAt: "2026-08-06T10:00:00Z",
        },
      }),
    ).toThrow("EventOccurrence");
  });

  it("accepts a valid pending offer projection", () => {
    const decoded = decodeEventOccurrence({
      ...occurrence,
      myOffer: {
        id: "offer_1",
        occurrenceId: "evocc_1",
        status: "pending",
        offeredAt: "2026-08-05T10:00:00Z",
        expiresAt: "2026-08-06T10:00:00Z",
      },
    });
    expect(decoded.myOffer?.status).toBe("pending");
  });

  it("decodes categories and series with closed keys", () => {
    expect(
      decodeEventCategories([{ id: "evcat_1", name: "Концерт", status: "active" }]),
    ).toHaveLength(1);
    expect(() =>
      decodeEventSeries({
        id: "evser_1",
        categoryId: "evcat_1",
        title: "Ансамбль",
        host: { accountId: "account_t", fullName: "Диана" },
        capacity: 10,
        weekday: 7,
        startMinutes: 600,
        durationMinutes: 60,
        effectiveFrom: "2026-08-03",
        status: "active",
        version: 0,
      }),
    ).toThrow("EventSeries");
  });
});
