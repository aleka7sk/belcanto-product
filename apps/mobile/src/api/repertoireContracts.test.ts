import { decodeStudentSong, decodeStudentSongs } from "./contracts";

describe("repertoire contracts (STU-GROWTH-07)", () => {
  const song = {
    id: "song_1",
    studentId: "student_1",
    title: "Easy On Me",
    artist: "Adele",
    stage: "technically_stable",
    stageNote: "Записать припев целиком",
    assignedBy: { accountId: "account_t", fullName: "Елена Орлова" },
    history: [
      {
        fromStage: "acquaintance",
        toStage: "technically_stable",
        note: "Куплет устойчив",
        changedAt: "2026-08-06T15:00:00Z",
      },
      {
        toStage: "acquaintance",
        changedAt: "2026-07-02T10:00:00Z",
      },
    ],
    version: 2,
    createdAt: "2026-07-02T10:00:00Z",
    updatedAt: "2026-08-06T15:00:00Z",
  };

  it("accepts the journey with newest-first history", () => {
    const decoded = decodeStudentSong(song);
    expect(decoded.stage).toBe("technically_stable");
    expect(decoded.history).toHaveLength(2);
  });

  it("requires the assignment entry and a stage-consistent head", () => {
    expect(() =>
      decodeStudentSong({ ...song, history: [] }),
    ).toThrow("StudentSong");
    expect(() =>
      decodeStudentSong({
        ...song,
        history: [song.history[1], { ...song.history[0], fromStage: "learning" }],
      }),
    ).toThrow("StudentSong");
    expect(() =>
      decodeStudentSong({ ...song, stage: "stage_ready" }),
    ).toThrow("StudentSong");
  });

  it("rejects unknown stages and smuggled scores", () => {
    expect(() =>
      decodeStudentSong({ ...song, stage: "grammy_ready" }),
    ).toThrow("StudentSong");
    expect(() => decodeStudentSong({ ...song, score: 5 })).toThrow("StudentSong");
  });

  it("keeps song ids unique in a list", () => {
    expect(decodeStudentSongs([song])).toHaveLength(1);
    expect(() => decodeStudentSongs([song, song])).toThrow("StudentSongList");
  });
});
