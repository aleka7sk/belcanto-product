import type {
  AssignedTeacherSummary,
  IsoDateTime,
  Lesson,
  StudentDirectoryItem,
} from "@/api";

import {
  finishTeacherLessonLoad,
  resolveTeacherLessonLoad,
  selectTeacherLessonSource,
  sourceTeachers,
  type TeacherLessonState,
} from "./teacherChangeState";

const teacherA: AssignedTeacherSummary = {
  accountId: "teacher_a",
  fullName: "Активный педагог",
  status: "active",
};
const inactiveTeacher: AssignedTeacherSummary = {
  accountId: "teacher_inactive",
  fullName: "Прежний педагог",
  status: "inactive",
};

function lesson(id: string, teacher: AssignedTeacherSummary): Lesson {
  return {
    id,
    title: `Занятие ${id}`,
    startsAt: "2026-08-10T09:00:00+05:00" as IsoDateTime,
    durationMinutes: 60,
    teacher,
    students: [{ studentId: "student_1", fullName: "Ученик" }],
    status: "scheduled",
    version: 1,
  };
}

describe("Teacher change state", () => {
  it("atomically clears the outgoing Teacher snapshot and rejects its late result", () => {
    const lessonA = lesson("lesson_a", teacherA);
    const loadedA: TeacherLessonState = {
      sourceTeacherAccountId: teacherA.accountId,
      lessons: [lessonA],
      selectedLessonIds: [lessonA.id],
      loading: false,
    };

    const loadingInactive = selectTeacherLessonSource(
      loadedA,
      inactiveTeacher.accountId,
    );
    expect(loadingInactive).toEqual({
      sourceTeacherAccountId: inactiveTeacher.accountId,
      lessons: [],
      selectedLessonIds: [],
      loading: true,
    });
    expect(
      resolveTeacherLessonLoad(loadingInactive, teacherA.accountId, [lessonA]),
    ).toBe(loadingInactive);

    const inactiveLesson = lesson("lesson_inactive", inactiveTeacher);
    const resolved = resolveTeacherLessonLoad(
      loadingInactive,
      inactiveTeacher.accountId,
      [inactiveLesson],
    );
    expect(
      finishTeacherLessonLoad(resolved, inactiveTeacher.accountId),
    ).toMatchObject({
      lessons: [inactiveLesson],
      loading: false,
    });
  });

  it("treats a same-Teacher selection as a no-op", () => {
    const loaded: TeacherLessonState = {
      sourceTeacherAccountId: teacherA.accountId,
      lessons: [lesson("lesson_a", teacherA)],
      selectedLessonIds: [],
      loading: false,
    };
    expect(selectTeacherLessonSource(loaded, teacherA.accountId)).toBe(loaded);
  });

  it("discovers an inactive source from assignments and preserves it after transfer", () => {
    const students: StudentDirectoryItem[] = [
      {
        studentId: "student_1",
        fullName: "Первый ученик",
        primaryTeacher: inactiveTeacher,
        primaryTeacherAssignmentVersion: 3,
      },
      {
        studentId: "student_2",
        fullName: "Второй ученик",
        primaryTeacher: teacherA,
        primaryTeacherAssignmentVersion: 1,
      },
    ];

    expect(sourceTeachers(students, null)).toEqual([
      inactiveTeacher,
      teacherA,
    ]);
    expect(sourceTeachers([], inactiveTeacher)).toEqual([inactiveTeacher]);
  });
});
