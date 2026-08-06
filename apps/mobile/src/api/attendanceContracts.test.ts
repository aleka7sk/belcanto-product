import { decodeAttendanceRecords } from "./contracts";

describe("attendance contracts (TCH-JOURNAL-01/02)", () => {
  const records = [
    {
      studentId: "student_1",
      studentName: "Аружан",
      status: "late",
      lateMinutes: 7,
      note: "Предупредила заранее",
      recordedAt: "2026-08-06T13:30:00Z",
      updatedAt: "2026-08-06T13:30:00Z",
    },
    {
      studentId: "student_2",
      studentName: "Дана",
      status: "absent",
      recordedAt: "2026-08-06T13:31:00Z",
      updatedAt: "2026-08-06T13:31:00Z",
    },
  ];

  it("accepts marks including a note-hidden absence (student view)", () => {
    const decoded = decodeAttendanceRecords(records);
    expect(decoded[0]?.lateMinutes).toBe(7);
    expect(decoded[1]?.note).toBeUndefined();
  });

  it("binds minutes to the late status in both directions", () => {
    expect(() =>
      decodeAttendanceRecords([{ ...records[0], lateMinutes: undefined }]),
    ).toThrow("AttendanceRecordList");
    expect(() =>
      decodeAttendanceRecords([{ ...records[1], lateMinutes: 5 }]),
    ).toThrow("AttendanceRecordList");
  });

  it("rejects duplicate students and smuggled fields", () => {
    expect(() => decodeAttendanceRecords([records[0], records[0]])).toThrow(
      "AttendanceRecordList",
    );
    expect(() =>
      decodeAttendanceRecords([{ ...records[0], score: 5 }]),
    ).toThrow("AttendanceRecordList");
  });
});
