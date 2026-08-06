import { router } from "expo-router";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { AccessibilityInfo, StyleSheet, Text, View } from "react-native";

import {
  canReassignPrimaryTeachers,
  canReplaceLessonTeachers,
} from "@/access";
import {
  useApiClient,
  type AssignedTeacherSummary,
  type IsoDateTime,
  type Lesson,
  type StaffMember,
  type StudentDirectoryItem,
} from "@/api";
import {
  createIntentIdempotency,
  prepareReassignPrimaryTeachers,
  prepareReplaceLessonTeachers,
  type ReassignPrimaryTeachersCommand,
  type ReplaceLessonTeachersCommand,
} from "@/controllers";
import { useSession } from "@/session";
import {
  AmbientGlow,
  InlineNotice,
  PremiumCard,
  PremiumScrollScreen,
  PremiumTextField,
  PrimaryButton,
  SecondaryButton,
  uiStyles,
} from "../components";
import { SelectableRow } from "../lessonComponents";
import { createLatestRequestGuard } from "../latestRequest";
import {
  createTeacherLessonState,
  finishTeacherLessonLoad,
  resolveTeacherLessonLoad,
  selectTeacherLessonSource,
  sourceTeachers,
  startTeacherLessonLoad,
} from "../teacherChangeState";
import { colors, fonts, metrics, spacing, typeScale } from "../tokens";
import { apiErrorMessage, formatLessonDay, formatLessonTime } from "../viewModels";
import { RoleNav } from "./account/shared";

type Mode = "permanent" | "temporary";
type Review =
  | { mode: "permanent"; command: ReassignPrimaryTeachersCommand }
  | { mode: "temporary"; command: ReplaceLessonTeachersCommand };
type ErrorKey =
  | "students"
  | "lessons"
  | "newTeacherAccountId"
  | "effectiveOn"
  | "effectiveAtTime";

export function TeacherChangeScreen() {
  const api = useApiClient();
  const { state, runAuthenticated } = useSession();
  const bootstrap = state.bootstrap;
  const [students, setStudents] = useState<StudentDirectoryItem[]>([]);
  const [teachers, setTeachers] = useState<StaffMember[]>([]);
  const [loading, setLoading] = useState(true);
  const [directoryError, setDirectoryError] = useState<string | null>(null);
  const [lessonLoadError, setLessonLoadError] = useState<string | null>(null);
  const [mode, setMode] = useState<Mode>("permanent");
  const [teacherLessonState, setTeacherLessonState] = useState(
    createTeacherLessonState,
  );
  const {
    sourceTeacherAccountId: currentTeacherAccountId,
    lessons,
    selectedLessonIds,
    loading: loadingLessons,
  } = teacherLessonState;
  const [selectedSourceTeacher, setSelectedSourceTeacher] =
    useState<AssignedTeacherSummary | null>(null);
  const [newTeacherAccountId, setNewTeacherAccountId] = useState("");
  const [selectedStudentIds, setSelectedStudentIds] = useState<string[]>([]);
  const [guidedStudentIds, setGuidedStudentIds] = useState<string[]>([]);
  const [effectiveImmediately, setEffectiveImmediately] = useState(true);
  const [effectiveOn, setEffectiveOn] = useState("");
  const [effectiveAtTime, setEffectiveAtTime] = useState("");
  const [errors, setErrors] = useState<
    Partial<Record<ErrorKey, string | undefined>>
  >({});
  const [requestError, setRequestError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [review, setReview] = useState<Review | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const idempotency = useMemo(() => createIntentIdempotency(), []);
  const lessonLoadGuard = useRef(createLatestRequestGuard());

  const allowed =
    bootstrap !== null &&
    canReplaceLessonTeachers(bootstrap) &&
    canReassignPrimaryTeachers(bootstrap);

  const loadDirectory = useCallback(async () => {
    if (!allowed) return;
    setLoading(true);
    setDirectoryError(null);
    try {
      const [studentDirectory, teacherDirectory] = await Promise.all([
        runAuthenticated((accessToken) => api.listStudents(accessToken)),
        runAuthenticated((accessToken) => api.listStaff(accessToken, "Teacher")),
      ]);
      setStudents(studentDirectory);
      setTeachers(teacherDirectory);
    } catch (error) {
      setDirectoryError(apiErrorMessage(error));
    } finally {
      setLoading(false);
    }
  }, [allowed, api, runAuthenticated]);

  const loadTeacherLessons = useCallback(async () => {
    if (!allowed || currentTeacherAccountId === "" || mode !== "temporary") {
      lessonLoadGuard.current.cancel();
      setTeacherLessonState((current) => ({
        ...current,
        lessons: [],
        selectedLessonIds: [],
        loading: false,
      }));
      setLessonLoadError(null);
      return;
    }
    const sourceTeacherAccountId = currentTeacherAccountId;
    const request = lessonLoadGuard.current.begin();
    setTeacherLessonState((current) =>
      startTeacherLessonLoad(current, sourceTeacherAccountId),
    );
    setLessonLoadError(null);
    const from = new Date();
    const to = new Date(from.getTime() + 365 * 24 * 60 * 60 * 1000);
    try {
      const result = await runAuthenticated((accessToken) =>
        api.listLessons(
          accessToken,
          {
            from: from.toISOString() as IsoDateTime,
            to: to.toISOString() as IsoDateTime,
            teacherAccountId: sourceTeacherAccountId,
          },
          request.signal,
        ),
      );
      if (request.isCurrent()) {
        setTeacherLessonState((current) =>
          resolveTeacherLessonLoad(current, sourceTeacherAccountId, result),
        );
      }
    } catch (error) {
      if (request.isCurrent()) setLessonLoadError(apiErrorMessage(error));
    } finally {
      if (request.isCurrent()) {
        setTeacherLessonState((current) =>
          finishTeacherLessonLoad(current, sourceTeacherAccountId),
        );
      }
    }
  }, [allowed, api, currentTeacherAccountId, mode, runAuthenticated]);

  useEffect(() => {
    let active = true;
    queueMicrotask(() => { if (active) void loadDirectory(); });
    return () => { active = false; };
  }, [loadDirectory]);
  useEffect(() => {
    let active = true;
    const guard = lessonLoadGuard.current;
    queueMicrotask(() => { if (active) void loadTeacherLessons(); });
    return () => {
      active = false;
      guard.cancel();
    };
  }, [loadTeacherLessons]);
  if (bootstrap === null) return null;
  if (!allowed) {
    return (
      <PremiumScrollScreen navigation={<RoleNav active="schedule" />}>
        <InlineNotice
          title="Нет разрешения"
          body="Временно заменить или постоянно переназначить педагога может только владелец или администратор."
          tone="error"
        />
        <SecondaryButton label="Назад" onPress={() => router.back()} />
      </PremiumScrollScreen>
    );
  }

  const currentTeachers = sourceTeachers(students, selectedSourceTeacher);
  const eligibleStudents = students.filter(
    (student) => student.primaryTeacher.accountId === currentTeacherAccountId,
  );
  const currentTeacher = selectedSourceTeacher;
  const newTeacher = teachers.find(
    (teacher) => teacher.accountId === newTeacherAccountId,
  );
  const visibleLessons = guidedStudentIds.length === 0
    ? lessons
    : lessons.filter((lesson) =>
        lesson.students.some((student) => guidedStudentIds.includes(student.studentId)),
      );

  const selectCurrentTeacher = (teacher: AssignedTeacherSummary) => {
    if (teacher.accountId === currentTeacherAccountId) return;
    lessonLoadGuard.current.cancel();
    setTeacherLessonState((current) =>
      selectTeacherLessonSource(current, teacher.accountId),
    );
    setSelectedSourceTeacher(teacher);
    setNewTeacherAccountId("");
    setSelectedStudentIds([]);
    setGuidedStudentIds([]);
    setLessonLoadError(null);
    setErrors({});
    setRequestError(null);
    setSuccess(null);
    setReview(null);
    idempotency.abandon();
  };
  const toggleStudent = (studentId: string) => {
    if (!selectedStudentIds.includes(studentId) && selectedStudentIds.length >= 100) {
      setErrors((current) => ({
        ...current,
        students: "Можно выбрать не более 100 учеников",
      }));
      return;
    }
    setSelectedStudentIds((current) =>
      current.includes(studentId)
        ? current.filter((candidate) => candidate !== studentId)
        : [...current, studentId],
    );
    setErrors((current) => ({ ...current, students: undefined }));
  };
  const toggleLesson = (lessonId: string) => {
    if (!selectedLessonIds.includes(lessonId) && selectedLessonIds.length >= 100) {
      setErrors((current) => ({
        ...current,
        lessons: "Можно выбрать не более 100 занятий",
      }));
      return;
    }
    setTeacherLessonState((current) => ({
      ...current,
      selectedLessonIds: current.selectedLessonIds.includes(lessonId)
        ? current.selectedLessonIds.filter((candidate) => candidate !== lessonId)
        : [...current.selectedLessonIds, lessonId],
    }));
    setErrors((current) => ({ ...current, lessons: undefined }));
  };

  const showIssues = (issues: readonly { field: string; code: string }[]) => {
    const next: Partial<Record<ErrorKey, string | undefined>> = {};
    for (const issue of issues) {
      const key = issue.field as ErrorKey;
      next[key] ??=
        issue.code === "required"
          ? "Сделайте явный выбор"
          : issue.code === "must_be_future"
            ? "Укажите будущие дату и время"
            : "Проверьте значение";
    }
    setErrors(next);
    AccessibilityInfo.announceForAccessibility("Проверьте параметры изменения");
  };

  const prepareReview = () => {
    setRequestError(null);
    setSuccess(null);
    if (mode === "permanent") {
      const selected = selectedStudentIds
        .map((studentId) => students.find((student) => student.studentId === studentId))
        .filter((student): student is StudentDirectoryItem => student !== undefined);
      const result = prepareReassignPrimaryTeachers(
        {
          students: selected.map((student) => ({
            studentId: student.studentId,
            expectedAssignmentVersion: student.primaryTeacherAssignmentVersion,
          })),
          newTeacherAccountId,
          effectiveImmediately,
          effectiveOn,
          effectiveAtTime,
        },
        idempotency.key(),
      );
      if (!result.ok) return showIssues(result.issues);
      setErrors({});
      setReview({ mode, command: result.value });
      return;
    }
    const selected = selectedLessonIds
      .map((lessonId) => lessons.find((lesson) => lesson.id === lessonId))
      .filter((lesson): lesson is Lesson => lesson !== undefined);
    const result = prepareReplaceLessonTeachers(
      {
        lessons: selected.map((lesson) => ({
          lessonId: lesson.id,
          expectedVersion: lesson.version,
          // Bind the mutation to the Teacher explicitly selected above the list.
          // If a stale response ever shows another Teacher's Lesson, the server
          // rejects it instead of silently moving the wrong Lesson.
          expectedPreviousTeacherAccountId: currentTeacherAccountId,
        })),
        newTeacherAccountId,
      },
      idempotency.key(),
    );
    if (!result.ok) return showIssues(result.issues);
    setErrors({});
    setReview({ mode, command: result.value });
  };

  const confirm = async () => {
    if (review === null) return;
    setSubmitting(true);
    setRequestError(null);
    try {
      if (review.mode === "permanent") {
        const result = await runAuthenticated((accessToken) =>
          api.reassignPrimaryTeachers(
            accessToken,
            review.command.body,
            review.command.idempotencyKey,
          ),
        );
        setSuccess(`Постоянно переназначено учеников: ${result.reassignedCount}. Существующие занятия не изменены.`);
        setGuidedStudentIds(
          review.command.body.students.map((student) => student.studentId),
        );
        setMode("temporary");
      } else {
        const result = await runAuthenticated((accessToken) =>
          api.replaceLessonTeachers(
            accessToken,
            review.command.body,
            review.command.idempotencyKey,
          ),
        );
        setSuccess(`Временная замена применена к занятиям: ${result.updatedCount}.`);
        setGuidedStudentIds([]);
      }
      idempotency.complete();
      setReview(null);
      setSelectedStudentIds([]);
      setTeacherLessonState((current) => ({
        ...current,
        selectedLessonIds: [],
      }));
      if (review.mode === "permanent") {
        await loadDirectory();
      } else {
        await Promise.all([loadDirectory(), loadTeacherLessons()]);
      }
      AccessibilityInfo.announceForAccessibility("Изменение сохранено");
    } catch (error) {
      const message = apiErrorMessage(error);
      setRequestError(message);
      AccessibilityInfo.announceForAccessibility(message);
    } finally {
      setSubmitting(false);
    }
  };

  if (review !== null) {
    const count =
      review.mode === "permanent"
        ? review.command.body.students.length
        : review.command.body.lessons.length;
    const total = review.mode === "permanent" ? eligibleStudents.length : visibleLessons.length;
    return (
      <PremiumScrollScreen navigation={<RoleNav active="schedule" />}>
        <AmbientGlow />
        <Text style={styles.brand}>BELCANTO</Text>
        <Text style={styles.eyebrow}>ПРОВЕРКА ИЗМЕНЕНИЯ</Text>
        <Text accessibilityRole="header" style={styles.title}>
          {review.mode === "permanent" ? "Постоянное переназначение" : "Временная замена"}
        </Text>
        <PremiumCard>
          <Text style={styles.summaryTeacher}>{currentTeacher?.fullName ?? "Текущий педагог"} → {newTeacher?.fullName ?? "Новый педагог"}</Text>
          <Text style={styles.summaryLine}>Выбрано: {count} из {total}</Text>
          {review.mode === "permanent" ? (
            <>
              <Text style={styles.summaryLine}>
                Начало: {effectiveImmediately ? "сразу" : `${effectiveOn}, ${effectiveAtTime}`}
              </Text>
              <Text style={styles.warning}>Существующие занятия не изменятся. При необходимости замените их отдельно.</Text>
            </>
          ) : (
            <Text style={styles.summaryLine}>Будут изменены только явно выбранные занятия. Если данные успели измениться, приложение попросит обновить список.</Text>
          )}
        </PremiumCard>
        {requestError ? <InlineNotice title="Изменение не сохранено" body={requestError} tone="error" /> : null}
        <PrimaryButton
          busy={submitting}
          label={review.mode === "permanent" ? "Подтвердить переназначение" : "Подтвердить временную замену"}
          onPress={() => void confirm()}
        />
        <SecondaryButton
          disabled={submitting}
          label="Вернуться к выбору"
          onPress={() => {
            setReview(null);
            idempotency.abandon();
          }}
        />
      </PremiumScrollScreen>
    );
  }

  return (
    <PremiumScrollScreen keyboardAware contentStyle={styles.content} navigation={<RoleNav active="schedule" />}>
      <AmbientGlow />
      <Text style={styles.brand}>BELCANTO</Text>
      <Text style={styles.eyebrow}>УПРАВЛЕНИЕ ПЕДАГОГОМ</Text>
      <Text accessibilityRole="header" style={styles.title}>Изменить педагога</Text>
      <Text style={styles.subtitle}>Два независимых действия с явным выбором объектов.</Text>

      <View style={styles.stack}>
        <SelectableRow
          kind="radio"
          label="Постоянно переназначить учеников"
          supporting="Изменяет закреплённого педагога с выбранного момента"
          selected={mode === "permanent"}
          onPress={() => {
            if (mode === "permanent") return;
            setMode("permanent");
            setGuidedStudentIds([]);
            setReview(null);
          }}
        />
        <SelectableRow
          kind="radio"
          label="Временно заменить на занятиях"
          supporting="Изменяет только явно выбранные будущие занятия"
          selected={mode === "temporary"}
          onPress={() => {
            if (mode === "temporary") return;
            setMode("temporary");
            setGuidedStudentIds([]);
            setReview(null);
          }}
        />
      </View>

      {directoryError ? <InlineNotice title="Справочник не загрузился" body={directoryError} tone="error" /> : null}
      {loading ? <Text style={uiStyles.body}>Загружаем справочник…</Text> : null}
      <View style={styles.sectionHeader}>
        <Text style={uiStyles.sectionTitle}>Текущий педагог</Text>
        <Text style={uiStyles.supporting}>фильтр учеников и занятий</Text>
      </View>
      <View style={styles.stack}>
        {currentTeachers.map((teacher) => (
          <SelectableRow
            key={teacher.accountId}
            kind="radio"
            label={teacher.fullName}
            supporting={
              teacher.status === "active"
                ? "Текущее активное закрепление"
                : "Педагог неактивен, закрепление сохранено"
            }
            selected={currentTeacherAccountId === teacher.accountId}
            onPress={() => selectCurrentTeacher(teacher)}
          />
        ))}
      </View>

      {currentTeacherAccountId !== "" ? (
        <>
          {mode === "permanent" ? (
            <>
              <View style={styles.sectionHeader}>
                <Text style={uiStyles.sectionTitle}>Ученики этого педагога</Text>
                <Text style={uiStyles.supporting}>выбрано {selectedStudentIds.length} из {eligibleStudents.length}</Text>
              </View>
              <View style={styles.stack}>
                {eligibleStudents.map((student) => (
                  <SelectableRow
                    key={student.studentId}
                    label={student.fullName}
                    selected={selectedStudentIds.includes(student.studentId)}
                    onPress={() => toggleStudent(student.studentId)}
                  />
                ))}
              </View>
              {errors.students ? <Text style={styles.error}>{errors.students}</Text> : null}
              <InlineNotice title="Важно" body="Постоянное переназначение не меняет уже созданные занятия." />
              <View style={styles.stack}>
                <SelectableRow kind="radio" label="Сразу" selected={effectiveImmediately} onPress={() => setEffectiveImmediately(true)} />
                <SelectableRow kind="radio" label="С выбранной даты" selected={!effectiveImmediately} onPress={() => setEffectiveImmediately(false)} />
              </View>
              {!effectiveImmediately ? (
                <View style={styles.dateFields}>
                  <View style={styles.fieldGrow}>
                    <PremiumTextField
                      error={errors.effectiveOn}
                      keyboardType="numbers-and-punctuation"
                      label="Дата"
                      onChangeText={(value) => { setEffectiveOn(value); setErrors((current) => ({ ...current, effectiveOn: undefined })); }}
                      placeholder="10.08.2026"
                      value={effectiveOn}
                    />
                  </View>
                  <View style={styles.fieldGrow}>
                    <PremiumTextField
                      error={errors.effectiveAtTime}
                      keyboardType="numbers-and-punctuation"
                      label="Время"
                      onChangeText={(value) => { setEffectiveAtTime(value); setErrors((current) => ({ ...current, effectiveAtTime: undefined })); }}
                      placeholder="09:00"
                      value={effectiveAtTime}
                    />
                  </View>
                </View>
              ) : null}
            </>
          ) : (
            <>
              {guidedStudentIds.length > 0 ? (
                <InlineNotice
                  title="Шаг 2 · существующие занятия"
                  body="Закрепление уже изменено. Теперь явно выберите только те будущие занятия перенесённых учеников, где нужна замена педагога. По умолчанию не выбрано ни одного."
                  tone="success"
                />
              ) : null}
              <View style={styles.sectionHeader}>
                <Text style={uiStyles.sectionTitle}>Будущие занятия</Text>
                <Text style={uiStyles.supporting}>выбрано {selectedLessonIds.length} из {visibleLessons.length}</Text>
              </View>
              {loadingLessons ? <Text style={uiStyles.body}>Загружаем занятия…</Text> : null}
              {lessonLoadError ? (
                <>
                  <InlineNotice
                    title="Занятия не загрузились"
                    body={lessonLoadError}
                    tone="error"
                  />
                  <SecondaryButton
                    disabled={loadingLessons}
                    label="Повторить загрузку занятий"
                    onPress={() => void loadTeacherLessons()}
                  />
                </>
              ) : null}
              <View style={styles.stack}>
                {visibleLessons.map((lesson) => (
                  <SelectableRow
                    key={lesson.id}
                    label={`${formatLessonDay(lesson.startsAt)}, ${formatLessonTime(lesson.startsAt)} · ${lesson.title}`}
                    supporting={lesson.students.map((student) => student.fullName).join(", ")}
                    selected={selectedLessonIds.includes(lesson.id)}
                    onPress={() => toggleLesson(lesson.id)}
                  />
                ))}
              </View>
              {errors.lessons ? <Text style={styles.error}>{errors.lessons}</Text> : null}
            </>
          )}

          <View style={styles.sectionHeader}>
            <Text style={uiStyles.sectionTitle}>Новый педагог</Text>
            <Text style={uiStyles.supporting}>явный выбор</Text>
          </View>
          <View style={styles.stack}>
            {teachers
              .filter((teacher) => teacher.accountId !== currentTeacherAccountId)
              .map((teacher) => (
                <SelectableRow
                  key={teacher.accountId}
                  kind="radio"
                  label={teacher.fullName}
                  selected={newTeacherAccountId === teacher.accountId}
                  onPress={() => { setNewTeacherAccountId(teacher.accountId); setErrors((current) => ({ ...current, newTeacherAccountId: undefined })); }}
                />
              ))}
          </View>
          {errors.newTeacherAccountId ? <Text style={styles.error}>{errors.newTeacherAccountId}</Text> : null}
          {success ? <InlineNotice title="Готово" body={success} tone="success" /> : null}
          <PrimaryButton label="Проверить изменение" onPress={prepareReview} />
          {guidedStudentIds.length > 0 ? (
            <SecondaryButton
              label="Пропустить — оставить расписание"
              onPress={() => {
                setGuidedStudentIds([]);
                router.replace("/(protected)/lessons");
              }}
            />
          ) : null}
        </>
      ) : null}
      <SecondaryButton label="Назад к занятиям" onPress={() => router.replace("/(protected)/lessons")} />
    </PremiumScrollScreen>
  );
}

const styles = StyleSheet.create({
  content: { minHeight: 1180 },
  brand: { color: colors.textPrimary, fontFamily: fonts.bold, ...typeScale.brand },
  eyebrow: { color: colors.textGold, fontFamily: fonts.semibold, marginTop: metrics.workflowEyebrowTop, ...typeScale.eyebrow },
  title: { color: colors.textPrimary, fontFamily: fonts.extrabold, marginTop: spacing.sm, ...typeScale.screenTitle },
  subtitle: { color: colors.textSecondary, fontFamily: fonts.regular, marginBottom: spacing.section, marginTop: spacing.sm, ...typeScale.body },
  stack: { gap: spacing.md },
  sectionHeader: { alignItems: "center", flexDirection: "row", justifyContent: "space-between", marginTop: spacing.section },
  error: { color: colors.danger, fontFamily: fonts.regular, ...typeScale.label },
  dateFields: { flexDirection: "row", gap: spacing.md },
  fieldGrow: { flex: 1 },
  summaryTeacher: { color: colors.textPrimary, fontFamily: fonts.bold, ...typeScale.cardTitle },
  summaryLine: { color: colors.textSecondary, fontFamily: fonts.regular, marginTop: spacing.md, ...typeScale.body },
  warning: { color: colors.textGold, fontFamily: fonts.semibold, marginTop: spacing.lg, ...typeScale.body },
});
