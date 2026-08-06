import {
  decodeActivityFeed,
  decodeNotificationPreferences,
} from "./contracts";

describe("activity contracts (Page 31)", () => {
  const feed = {
    unreadCount: 1,
    entries: [
      {
        id: "act_2",
        category: "learning",
        kind: "JournalPublished",
        targetType: "lesson_journal",
        targetId: "jrnl_1",
        payload: { journalId: "jrnl_1", studentId: "student_1", occurrenceId: "cocc_1" },
        occurredAt: "2026-08-06T15:00:00Z",
      },
      {
        id: "act_1",
        category: "important",
        kind: "AttendanceAbsenceRecorded",
        targetType: "lesson_attendance",
        targetId: "cocc_1:student_1",
        payload: { occurrenceId: "cocc_1", studentId: "student_1" },
        occurredAt: "2026-08-05T15:00:00Z",
        readAt: "2026-08-05T16:00:00Z",
      },
    ],
  };

  it("accepts the feed and keeps unread consistent with the list", () => {
    const decoded = decodeActivityFeed(feed);
    expect(decoded.unreadCount).toBe(1);
    expect(decoded.entries[0]?.payload.occurrenceId).toBe("cocc_1");
    expect(() =>
      decodeActivityFeed({ ...feed, unreadCount: 0 }),
    ).toThrow("ActivityFeed");
  });

  it("rejects unknown categories and duplicate entries", () => {
    expect(() =>
      decodeActivityFeed({
        ...feed,
        entries: [{ ...feed.entries[0], category: "marketing" }],
      }),
    ).toThrow("ActivityFeed");
    expect(() =>
      decodeActivityFeed({
        unreadCount: 2,
        entries: [feed.entries[0], feed.entries[0]],
      }),
    ).toThrow("ActivityFeed");
  });

  it("keeps preferences one-per-category with boolean push", () => {
    const preferences = [
      { category: "important", pushEnabled: true },
      { category: "learning", pushEnabled: false },
    ];
    expect(decodeNotificationPreferences(preferences)).toHaveLength(2);
    expect(() =>
      decodeNotificationPreferences([preferences[0], preferences[0]]),
    ).toThrow("NotificationPreferenceList");
    expect(() =>
      decodeNotificationPreferences([{ category: "important", pushEnabled: "yes" }]),
    ).toThrow("NotificationPreferenceList");
  });
});
