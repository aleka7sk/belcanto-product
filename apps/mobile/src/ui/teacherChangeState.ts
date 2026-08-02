import type {
  AssignedTeacherSummary,
  Lesson,
  StudentDirectoryItem,
} from "@/api";

export interface TeacherLessonState {
  sourceTeacherAccountId: string;
  lessons: Lesson[];
  selectedLessonIds: string[];
  loading: boolean;
}

export function createTeacherLessonState(): TeacherLessonState {
  return {
    sourceTeacherAccountId: "",
    lessons: [],
    selectedLessonIds: [],
    loading: false,
  };
}

export function selectTeacherLessonSource(
  state: TeacherLessonState,
  sourceTeacherAccountId: string,
): TeacherLessonState {
  if (state.sourceTeacherAccountId === sourceTeacherAccountId) return state;
  return {
    sourceTeacherAccountId,
    lessons: [],
    selectedLessonIds: [],
    loading: sourceTeacherAccountId !== "",
  };
}

export function startTeacherLessonLoad(
  state: TeacherLessonState,
  sourceTeacherAccountId: string,
): TeacherLessonState {
  if (state.sourceTeacherAccountId !== sourceTeacherAccountId) return state;
  return { ...state, lessons: [], selectedLessonIds: [], loading: true };
}

export function resolveTeacherLessonLoad(
  state: TeacherLessonState,
  sourceTeacherAccountId: string,
  lessons: readonly Lesson[],
): TeacherLessonState {
  if (state.sourceTeacherAccountId !== sourceTeacherAccountId) return state;
  return { ...state, lessons: [...lessons] };
}

export function finishTeacherLessonLoad(
  state: TeacherLessonState,
  sourceTeacherAccountId: string,
): TeacherLessonState {
  if (state.sourceTeacherAccountId !== sourceTeacherAccountId) return state;
  return { ...state, loading: false };
}

export function sourceTeachers(
  students: readonly StudentDirectoryItem[],
  selectedSourceTeacher: AssignedTeacherSummary | null,
): AssignedTeacherSummary[] {
  const teachers = Array.from(
    new Map(
      students.map((student) => [
        student.primaryTeacher.accountId,
        student.primaryTeacher,
      ]),
    ).values(),
  );
  if (
    selectedSourceTeacher !== null &&
    !teachers.some(
      (teacher) => teacher.accountId === selectedSourceTeacher.accountId,
    )
  ) {
    teachers.push(selectedSourceTeacher);
  }
  return teachers;
}
