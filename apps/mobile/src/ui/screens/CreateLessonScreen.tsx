import { router } from "expo-router";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { AccessibilityInfo, StyleSheet, Text, View } from "react-native";

import { canCreateLessons } from "@/access";
import {
  useApiClient,
  type IsoDateTime,
  type StaffMember,
  type StudentDirectoryItem,
} from "@/api";
import {
  createIntentIdempotency,
  prepareCreateLesson,
  type CreateLessonDraft,
} from "@/controllers";
import { useSession } from "@/session";
import { parseAlmatyLocalDateTime } from "@/validation/datetime";
import {
  AmbientGlow,
  InlineNotice,
  PremiumScrollScreen,
  PremiumTextField,
  PrimaryButton,
  SecondaryButton,
  uiStyles,
} from "../components";
import { SelectableRow } from "../lessonComponents";
import { createLatestRequestGuard } from "../latestRequest";
import { semantic, space, typeStyles } from "../tokens";
import { apiErrorMessage, formIssueMap } from "../viewModels";
import { RoleNav } from "./account/shared";

type Errors = Partial<Record<keyof CreateLessonDraft, string | undefined>>;

export function CreateLessonScreen() {
  const api = useApiClient();
  const { state, runAuthenticated } = useSession();
  const bootstrap = state.bootstrap;
  const [students, setStudents] = useState<StudentDirectoryItem[]>([]);
  const [teachers, setTeachers] = useState<StaffMember[]>([]);
  const [loadingDirectory, setLoadingDirectory] = useState(false);
  const [loadingTeachers, setLoadingTeachers] = useState(false);
  const [directoryError, setDirectoryError] = useState<string | null>(null);
  const [teachersError, setTeachersError] = useState<string | null>(null);
  const [title, setTitle] = useState("");
  const [startsOn, setStartsOn] = useState("");
  const [startsAtTime, setStartsAtTime] = useState("");
  const [durationMinutes, setDurationMinutes] = useState("60");
  const [location, setLocation] = useState("");
  const [teacherAccountId, setTeacherAccountId] = useState("");
  const [studentIds, setStudentIds] = useState<string[]>([]);
  const [errors, setErrors] = useState<Errors>({});
  const [requestError, setRequestError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const idempotency = useMemo(() => createIntentIdempotency(), []);
  const studentDirectoryGuard = useRef(createLatestRequestGuard());
  const manager =
    bootstrap?.roles.includes("Owner") === true ||
    bootstrap?.roles.includes("Administrator") === true;

  const startsAtTimestamp = parseAlmatyLocalDateTime(startsOn, startsAtTime);

  const loadTeachers = useCallback(async () => {
    if (bootstrap === null || !canCreateLessons(bootstrap) || !manager) return;
    setLoadingTeachers(true);
    setTeachersError(null);
    try {
      setTeachers(
        await runAuthenticated((accessToken) =>
          api.listStaff(accessToken, "Teacher"),
        ),
      );
    } catch (error) {
      setTeachersError(apiErrorMessage(error));
    } finally {
      setLoadingTeachers(false);
    }
  }, [api, bootstrap, manager, runAuthenticated]);

  const loadDirectory = useCallback(async () => {
    if (
      bootstrap === null ||
      !canCreateLessons(bootstrap) ||
      startsAtTimestamp === null
    ) {
      studentDirectoryGuard.current.cancel();
      setStudents([]);
      setLoadingDirectory(false);
      return;
    }
    const request = studentDirectoryGuard.current.begin();
    setLoadingDirectory(true);
    setDirectoryError(null);
    try {
      const asOf = new Date(startsAtTimestamp).toISOString() as IsoDateTime;
      const result = await runAuthenticated((accessToken) =>
        api.listStudents(accessToken, { asOf }, request.signal),
      );
      if (request.isCurrent()) setStudents(result);
    } catch (error) {
      if (request.isCurrent()) setDirectoryError(apiErrorMessage(error));
    } finally {
      if (request.isCurrent()) setLoadingDirectory(false);
    }
  }, [api, bootstrap, runAuthenticated, startsAtTimestamp]);

  useEffect(() => {
    let active = true;
    queueMicrotask(() => { if (active) void loadTeachers(); });
    return () => { active = false; };
  }, [loadTeachers]);
  useEffect(() => {
    let active = true;
    const guard = studentDirectoryGuard.current;
    queueMicrotask(() => { if (active) void loadDirectory(); });
    return () => {
      active = false;
      guard.cancel();
    };
  }, [loadDirectory]);
  if (bootstrap === null) return null;
  if (!canCreateLessons(bootstrap)) {
    return (
      <PremiumScrollScreen navigation={<RoleNav active="schedule" />}>
        <InlineNotice title="Нет разрешения" body="Создавать занятия может владелец, администратор или педагог." tone="error" />
        <SecondaryButton label="Назад" onPress={() => router.back()} />
      </PremiumScrollScreen>
    );
  }

  const toggleStudent = (studentId: string) => {
    if (!studentIds.includes(studentId) && studentIds.length >= 100) {
      setErrors((current) => ({
        ...current,
        studentIds: "Можно выбрать не более 100 учеников",
      }));
      return;
    }
    setStudentIds((current) =>
      current.includes(studentId)
        ? current.filter((candidate) => candidate !== studentId)
        : [...current, studentId],
    );
    setErrors((current) => ({ ...current, studentIds: undefined }));
  };

  const submit = async () => {
    setRequestError(null);
    const result = prepareCreateLesson(
      {
        title,
        startsOn,
        startsAtTime,
        durationMinutes,
        location,
        teacherAccountId: manager ? teacherAccountId : bootstrap.accountId,
        studentIds,
      },
      idempotency.key(),
    );
    if (!result.ok) {
      setErrors(formIssueMap(result.issues));
      AccessibilityInfo.announceForAccessibility("Проверьте параметры занятия");
      return;
    }
    setErrors({});
    setSubmitting(true);
    try {
      await runAuthenticated((accessToken) =>
        api.createLesson(accessToken, result.value.body, result.value.idempotencyKey),
      );
      idempotency.complete();
      AccessibilityInfo.announceForAccessibility("Занятие создано");
      router.replace("/(protected)/lessons");
    } catch (error) {
      const message = apiErrorMessage(error);
      setRequestError(message);
      AccessibilityInfo.announceForAccessibility(message);
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <PremiumScrollScreen keyboardAware contentStyle={styles.content} navigation={<RoleNav active="schedule" />}>
      <AmbientGlow />
      <Text style={styles.brand}>BELCANTO</Text>
      <Text style={styles.eyebrow}>НОВОЕ ЗАНЯТИЕ</Text>
      <Text accessibilityRole="header" style={styles.title}>Создать урок</Text>
      <Text style={styles.subtitle}>Занятие сразу появится в расписании выбранных учеников.</Text>

      <View style={styles.form}>
        <PremiumTextField
          autoCapitalize="sentences"
          error={errors.title}
          label="Название"
          onChangeText={(value) => { setTitle(value); setErrors((current) => ({ ...current, title: undefined })); }}
          placeholder="Индивидуальный урок"
          value={title}
        />
        <PremiumTextField
          autoCapitalize="none"
          error={errors.startsOn}
          helper="Дата по времени школы"
          keyboardType="numbers-and-punctuation"
          label="Дата"
          onChangeText={(value) => {
            const nextTimestamp = parseAlmatyLocalDateTime(value, startsAtTime);
            setStartsOn(value);
            if (nextTimestamp !== startsAtTimestamp) {
              setStudentIds([]);
              setStudents([]);
            }
            setErrors((current) => ({ ...current, startsOn: undefined }));
          }}
          placeholder="10.08.2026"
          value={startsOn}
        />
        <PremiumTextField
          error={errors.startsAtTime}
          helper="Время по Алматы"
          keyboardType="numbers-and-punctuation"
          label="Время"
          onChangeText={(value) => {
            const nextTimestamp = parseAlmatyLocalDateTime(startsOn, value);
            setStartsAtTime(value);
            if (nextTimestamp !== startsAtTimestamp) {
              setStudentIds([]);
              setStudents([]);
            }
            setErrors((current) => ({ ...current, startsAtTime: undefined }));
          }}
          placeholder="18:00"
          value={startsAtTime}
        />
        <PremiumTextField
          error={errors.durationMinutes}
          keyboardType="number-pad"
          label="Продолжительность, минут"
          onChangeText={(value) => { setDurationMinutes(value); setErrors((current) => ({ ...current, durationMinutes: undefined })); }}
          value={durationMinutes}
        />
        <PremiumTextField
          autoCapitalize="sentences"
          error={errors.location}
          label="Место или ссылка (необязательно)"
          onChangeText={(value) => { setLocation(value); setErrors((current) => ({ ...current, location: undefined })); }}
          placeholder="Класс 2"
          value={location}
        />
      </View>

      {directoryError ? (
        <View style={styles.stack}>
          <InlineNotice title="Справочник не загрузился" body={directoryError} tone="error" />
          <SecondaryButton label="Повторить" onPress={() => void loadDirectory()} />
        </View>
      ) : null}
      {teachersError ? (
        <View style={styles.stack}>
          <InlineNotice title="Педагоги не загрузились" body={teachersError} tone="error" />
          <SecondaryButton label="Повторить" onPress={() => void loadTeachers()} />
        </View>
      ) : null}
      {loadingDirectory ? <Text style={uiStyles.body}>Загружаем учеников на время занятия…</Text> : null}
      {loadingTeachers ? <Text style={uiStyles.body}>Загружаем активных педагогов…</Text> : null}

      <View style={styles.sectionHeader}>
        <Text style={uiStyles.sectionTitle}>Педагог</Text>
        <Text style={uiStyles.supporting}>{manager ? "выберите" : "только вы"}</Text>
      </View>
      {manager ? (
        <View style={styles.stack}>
          {teachers.map((teacher) => (
            <SelectableRow
              key={teacher.accountId}
              kind="radio"
              label={teacher.fullName}
              supporting="Педагог Belcanto"
              selected={teacherAccountId === teacher.accountId}
              onPress={() => { setTeacherAccountId(teacher.accountId); setErrors((current) => ({ ...current, teacherAccountId: undefined })); }}
            />
          ))}
        </View>
      ) : (
        <InlineNotice title="Педагог занятия" body="Занятие будет создано от вашего имени. Выбрать другого педагога нельзя." />
      )}
      {errors.teacherAccountId ? <Text style={styles.error}>{errors.teacherAccountId}</Text> : null}

      <View style={styles.sectionHeader}>
        <Text style={uiStyles.sectionTitle}>Ученики</Text>
        <Text style={uiStyles.supporting}>выбрано {studentIds.length} из 100</Text>
      </View>
      {startsAtTimestamp === null ? (
        <InlineNotice
          title="Сначала укажите дату и время"
          body="Состав учеников проверяется на точный момент начала занятия."
        />
      ) : null}
      <View style={styles.stack}>
        {students.map((student) => (
          <SelectableRow
            key={student.studentId}
            label={student.fullName}
            supporting={`Закреплённый педагог: ${student.primaryTeacher.fullName}`}
            selected={studentIds.includes(student.studentId)}
            onPress={() => toggleStudent(student.studentId)}
          />
        ))}
      </View>
      {startsAtTimestamp !== null && !loadingDirectory && students.length === 0 ? (
        <InlineNotice title="Нет доступных учеников" body="Справочник включает и учеников, которые ещё не активировали приложение." />
      ) : null}
      {errors.studentIds ? <Text style={styles.error}>{errors.studentIds}</Text> : null}
      {requestError ? <InlineNotice title="Занятие не создано" body={requestError} tone="error" /> : null}
      <View style={styles.stack}>
        <PrimaryButton busy={submitting} label="Создать занятие" onPress={() => void submit()} />
        <SecondaryButton label="Отмена" onPress={() => router.back()} />
      </View>
    </PremiumScrollScreen>
  );
}

const styles = StyleSheet.create({
  content: { minHeight: 1120 },
  brand: { color: semantic.textPrimary, fontFamily: "Onest_700Bold", fontSize: 13, letterSpacing: 2.4, lineHeight: 17 },
  eyebrow: { color: semantic.textGold, fontFamily: "Onest_600SemiBold", fontSize: 10, letterSpacing: 1, lineHeight: 13, marginTop: space.s10 },
  title: { color: semantic.textPrimary, fontFamily: "Onest_800ExtraBold", fontSize: 28, lineHeight: 34, marginTop: space.s2 },
  subtitle: { color: semantic.textSecondary, marginTop: space.s2, ...typeStyles.bodyS },
  form: { gap: space.s5, marginTop: space.s8 },
  sectionHeader: { alignItems: "center", flexDirection: "row", justifyContent: "space-between", marginTop: space.s8 },
  stack: { gap: space.s3 },
  error: { color: semantic.feedbackDanger, ...typeStyles.caption },
});
