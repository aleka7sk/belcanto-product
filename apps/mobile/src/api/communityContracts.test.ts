import {
  decodeBlockedMembers,
  decodeCommunityFeed,
  decodeCommunityPost,
  decodeCommunityReport,
  decodeCommunityReports,
} from "./contracts";

describe("community contracts (Page 28)", () => {
  const author = { accountId: "acct_1", fullName: "Айгерим Ученица", role: "Student" };
  const tombstoneAuthor = { accountId: "", fullName: "", role: "" };
  const post = {
    id: "post_1",
    kind: "post",
    body: "Кто идёт на Open Stage?",
    audience: "school",
    commentsEnabled: true,
    pinned: false,
    status: "published",
    author,
    commentCount: 1,
    comments: [
      {
        id: "comment_1",
        author: { accountId: "acct_2", fullName: "Педагог Алия", role: "Teacher" },
        body: "Я иду!",
        status: "published",
        createdAt: "2026-08-06T12:00:00Z",
      },
    ],
    createdAt: "2026-08-06T10:00:00Z",
  };

  it("accepts a published thread and keeps the count honest", () => {
    const decoded = decodeCommunityPost(post);
    expect(decoded.commentCount).toBe(1);
    expect(decoded.comments?.[0]?.body).toBe("Я иду!");
    expect(() =>
      decodeCommunityPost({ ...post, commentCount: 2 }),
    ).toThrow("CommunityPost");
  });

  it("requires the tombstone to drop the words and the author together", () => {
    const hidden = {
      ...post,
      commentCount: 0,
      comments: [
        {
          id: "comment_1",
          author: tombstoneAuthor,
          status: "hidden",
          createdAt: "2026-08-06T12:00:00Z",
        },
      ],
    };
    expect(decodeCommunityPost(hidden).comments?.[0]?.status).toBe("hidden");
    expect(() =>
      decodeCommunityPost({
        ...hidden,
        comments: [{ ...hidden.comments[0], body: "утёкшие слова" }],
      }),
    ).toThrow("CommunityPost");
    expect(() =>
      decodeCommunityPost({
        ...post,
        comments: [{ ...post.comments[0], author: tombstoneAuthor }],
      }),
    ).toThrow("CommunityPost");
    // The moderator view keeps both the words and the author.
    expect(
      decodeCommunityPost({
        ...hidden,
        comments: [
          {
            ...hidden.comments[0],
            body: "скрытые слова",
            author: { accountId: "acct_2", fullName: "Педагог Алия", role: "Teacher" },
          },
        ],
      }).comments?.[0]?.body,
    ).toBe("скрытые слова");
  });

  it("requires a published announcement to carry its title", () => {
    const announcement = {
      ...post,
      id: "post_2",
      kind: "announcement",
      title: "Отчётный концерт",
      pinned: true,
      commentCount: 0,
      comments: [],
    };
    expect(decodeCommunityPost(announcement).pinned).toBe(true);
    expect(() =>
      decodeCommunityPost({ ...announcement, title: undefined }),
    ).toThrow("CommunityPost");
    expect(() =>
      decodeCommunityPost({ ...post, pinned: true }),
    ).toThrow("CommunityPost");
  });

  it("keeps the feed published-only with pinned announcements first", () => {
    const announcement = {
      ...post,
      id: "post_2",
      kind: "announcement",
      title: "Отчётный концерт",
      pinned: true,
      commentCount: 0,
    };
    const feedPost = { ...post, commentCount: 0 };
    delete (announcement as Record<string, unknown>).comments;
    delete (feedPost as Record<string, unknown>).comments;
    expect(decodeCommunityFeed([announcement, feedPost])).toHaveLength(2);
    expect(() =>
      decodeCommunityFeed([feedPost, announcement]),
    ).toThrow("CommunityFeed");
    expect(() =>
      decodeCommunityFeed([{ ...feedPost, status: "removed" }]),
    ).toThrow("CommunityFeed");
  });

  const newReport = {
    id: "report_1",
    targetType: "comment",
    targetId: "comment_1",
    reason: "abuse",
    status: "new",
    createdAt: "2026-08-06T13:00:00Z",
    targetExcerpt: "Я иду!",
  };

  it("pairs the review decision fields atomically", () => {
    expect(decodeCommunityReport(newReport).status).toBe("new");
    const reviewed = {
      ...newReport,
      status: "reviewed",
      decision: "hidden",
      decisionReason: "Нарушение правил сообщества",
      decidedAt: "2026-08-06T14:00:00Z",
    };
    expect(decodeCommunityReport(reviewed).decision).toBe("hidden");
    expect(() =>
      decodeCommunityReport({ ...reviewed, decidedAt: undefined }),
    ).toThrow("CommunityReport");
    expect(() =>
      decodeCommunityReport({ ...newReport, decision: "kept" }),
    ).toThrow("CommunityReport");
    expect(() =>
      decodeCommunityReport({ ...newReport, reason: "other" }),
    ).toThrow("CommunityReport");
  });

  it("orders the moderation queue new-first", () => {
    const reviewed = {
      ...newReport,
      id: "report_2",
      status: "reviewed",
      decision: "kept",
      decisionReason: "Контекст допустим",
      decidedAt: "2026-08-06T14:00:00Z",
    };
    expect(decodeCommunityReports([newReport, reviewed])).toHaveLength(2);
    expect(() =>
      decodeCommunityReports([reviewed, newReport]),
    ).toThrow("CommunityReportList");
  });

  it("keeps the block list unique", () => {
    expect(decodeBlockedMembers({ blocked: ["acct_2"] }).blocked).toEqual(["acct_2"]);
    expect(() =>
      decodeBlockedMembers({ blocked: ["acct_2", "acct_2"] }),
    ).toThrow("BlockedMembers");
  });
});
