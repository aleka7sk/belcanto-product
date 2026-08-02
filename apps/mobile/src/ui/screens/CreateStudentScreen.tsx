import { router } from "expo-router";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import {
  AccessibilityInfo,
  Pressable,
  StyleSheet,
  Text,
  TextInput,
  View,
} from "react-native";

import { canCreateStudents } from "@/access";
import { useApiClient, type StaffMember } from "@/api";
import {
  createIntentIdempotency,
  prepareCreateStudent,
  type CreateStudentDraft,
} from "@/controllers";
import { useSession } from "@/session";
import {
  AmbientGlow,
  InlineNotice,
  PremiumScrollScreen,
  PremiumTextField,
  PrimaryButton,
  SecondaryButton,
  uiStyles,
} from "../components";
import {
  colors,
  fonts,
  metrics,
  radii,
  spacing,
  typeScale,
} from "../tokens";
import { apiErrorMessage, formIssueMap } from "../viewModels";

type CreateStudentErrors = Partial<
  Record<keyof CreateStudentDraft, string | undefined>
>;

export function CreateStudentScreen() {
  const api = useApiClient();
  const { state, runAuthenticated } = useSession();
  const bootstrap = state.bootstrap;
  const [teachers, setTeachers] = useState<StaffMember[]>([]);
  const [teachersError, setTeachersError] = useState<string | null>(null);
  const [loadingTeachers, setLoadingTeachers] = useState(true);
  const [fullName, setFullName] = useState("");
  const [phone, setPhone] = useState("");
  const [enrollmentReference, setEnrollmentReference] = useState("");
  const [teacherAccountId, setTeacherAccountId] = useState("");
  const [adultConfirmed, setAdultConfirmed] = useState(false);
  const [errors, setErrors] = useState<CreateStudentErrors>({});
  const [requestError, setRequestError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const fullNameRef = useRef<TextInput>(null);
  const phoneRef = useRef<TextInput>(null);
  const enrollmentRef = useRef<TextInput>(null);
  const idempotency = useMemo(() => createIntentIdempotency(), []);

  const loadTeachers = useCallback(async () => {
    if (
      bootstrap === null ||
      !canCreateStudents(bootstrap)
    ) {
      setLoadingTeachers(false);
      return;
    }
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
  }, [api, bootstrap, runAuthenticated]);

  useEffect(() => {
    let active = true;
    queueMicrotask(() => {
      if (active) void loadTeachers();
    });
    return () => {
      active = false;
    };
  }, [loadTeachers]);

  if (bootstrap === null) return null;
  if (!canCreateStudents(bootstrap)) {
    return (
      <PremiumScrollScreen contentStyle={styles.denied}>
        <InlineNotice
          body="Создавать учеников может владелец или администратор с действующим доступом суперадминистратора."
          title="Нет разрешения"
          tone="error"
        />
        <SecondaryButton label="Назад" onPress={() => router.back()} />
      </PremiumScrollScreen>
    );
  }

  const submit = async () => {
    setRequestError(null);
    const result = prepareCreateStudent(
      {
        fullName,
        phone,
        enrollmentReference,
        teacherAccountId,
        locale: "ru-KZ",
        timezone: "Asia/Almaty",
        adultConfirmed,
      },
      idempotency.key(),
    );
    if (!result.ok) {
      const nextErrors = formIssueMap(result.issues);
      setErrors(nextErrors);
      const first = nextErrors.fullName
        ? fullNameRef
        : nextErrors.phone
          ? phoneRef
          : enrollmentRef;
      first.current?.focus();
      AccessibilityInfo.announceForAccessibility("Проверьте данные ученика");
      return;
    }
    setErrors({});
    setSubmitting(true);
    try {
      const created = await runAuthenticated((accessToken) =>
        api.createStudent(
          accessToken,
          result.value.body,
          result.value.idempotencyKey,
        ),
      );
      idempotency.complete();
      AccessibilityInfo.announceForAccessibility("Ученик добавлен");
      router.replace({
        pathname: "/(protected)/student/[studentId]",
        params: { studentId: created.studentId },
      });
    } catch (error) {
      const message = apiErrorMessage(error);
      setRequestError(message);
      AccessibilityInfo.announceForAccessibility(message);
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <PremiumScrollScreen keyboardAware contentStyle={styles.content}>
      <AmbientGlow />
      <Text style={styles.brand}>BELCANTO</Text>
      <Text style={styles.eyebrow}>НОВЫЙ УЧЕНИК</Text>
      <Text accessibilityRole="header" style={styles.title}>
        Подготовить путь
      </Text>
      <Text style={styles.subtitle}>
        Сначала школа создаёт учебную запись. Ученик задаст пароль позже — по
        персональному приглашению.
      </Text>

      <View style={styles.form}>
        <PremiumTextField
          ref={fullNameRef}
          autoCapitalize="words"
          autoComplete="name"
          error={errors.fullName}
          label="Имя и фамилия"
          onChangeText={(value) => {
            setFullName(value);
            setErrors((current) => ({ ...current, fullName: undefined }));
          }}
          onSubmitEditing={() => phoneRef.current?.focus()}
          placeholder="Алина Соколова"
          returnKeyType="next"
          textContentType="name"
          value={fullName}
        />
        <PremiumTextField
          ref={phoneRef}
          autoComplete="tel"
          error={errors.phone}
          keyboardType="phone-pad"
          label="Телефон"
          onChangeText={(value) => {
            setPhone(value);
            setErrors((current) => ({ ...current, phone: undefined }));
          }}
          onSubmitEditing={() => enrollmentRef.current?.focus()}
          placeholder="+7 700 000 00 00"
          returnKeyType="next"
          textContentType="telephoneNumber"
          value={phone}
        />
        <PremiumTextField
          ref={enrollmentRef}
          autoCapitalize="characters"
          error={errors.enrollmentReference}
          label="Внутренний номер поступления"
          onChangeText={(value) => {
            setEnrollmentReference(value);
            setErrors((current) => ({ ...current, enrollmentReference: undefined }));
          }}
          placeholder="BC-2026-001"
          returnKeyType="done"
          value={enrollmentReference}
        />
      </View>

      <View style={styles.sectionHeader}>
        <Text style={uiStyles.sectionTitle}>Закреплённый педагог</Text>
        <Text style={uiStyles.supporting}>подготовит первый ориентир</Text>
      </View>
      {loadingTeachers ? <Text style={uiStyles.body}>Загружаем педагогов…</Text> : null}
      {teachersError ? (
        <View style={styles.stackGap}>
          <InlineNotice body={teachersError} title="Педагоги не загрузились" tone="error" />
          <SecondaryButton label="Повторить" onPress={() => void loadTeachers()} />
        </View>
      ) : null}
      <View style={styles.teacherList}>
        {teachers.map((teacher) => {
          const selected = teacher.accountId === teacherAccountId;
          return (
            <Pressable
              accessibilityLabel={`Педагог ${teacher.fullName}`}
              accessibilityRole="radio"
              accessibilityState={{ checked: selected }}
              key={teacher.accountId}
              onPress={() => {
                setTeacherAccountId(teacher.accountId);
                setErrors((current) => ({ ...current, teacherAccountId: undefined }));
              }}
              style={({ pressed }) => [
                styles.teacher,
                selected && styles.teacherSelected,
                pressed && styles.teacherPressed,
              ]}
            >
              <View style={[styles.radio, selected && styles.radioSelected]} />
              <View style={styles.teacherCopy}>
                <Text style={styles.teacherName}>{teacher.fullName}</Text>
                <Text style={uiStyles.supporting}>Педагог Belcanto</Text>
              </View>
            </Pressable>
          );
        })}
      </View>
      {errors.teacherAccountId ? (
        <Text accessibilityLiveRegion="polite" style={styles.errorText}>
          {errors.teacherAccountId}
        </Text>
      ) : null}

      <Pressable
        accessibilityLabel="Подтверждаю, что ученику исполнилось 18 лет"
        accessibilityRole="checkbox"
        accessibilityState={{ checked: adultConfirmed }}
        onPress={() => {
          setAdultConfirmed((value) => !value);
          setErrors((current) => ({ ...current, adultConfirmed: undefined }));
        }}
        style={styles.confirmation}
      >
        <View style={[styles.checkbox, adultConfirmed && styles.checkboxSelected]}>
          <Text style={styles.checkmark}>{adultConfirmed ? "✓" : ""}</Text>
        </View>
        <Text style={styles.confirmationText}>
          Подтверждаю, что ученику исполнилось 18 лет
        </Text>
      </Pressable>
      {errors.adultConfirmed ? (
        <Text accessibilityLiveRegion="polite" style={styles.errorText}>
          {errors.adultConfirmed}
        </Text>
      ) : null}

      {requestError ? (
        <InlineNotice body={requestError} title="Ученик не добавлен" tone="error" />
      ) : null}
      <View style={styles.actions}>
        <PrimaryButton busy={submitting} label="Добавить ученика" onPress={() => void submit()} />
        <SecondaryButton label="Отмена" onPress={() => router.back()} />
      </View>
    </PremiumScrollScreen>
  );
}

const styles = StyleSheet.create({
  denied: { gap: spacing.lg, justifyContent: "center", minHeight: 680 },
  content: { minHeight: 980 },
  brand: {
    color: colors.textPrimary,
    fontFamily: fonts.bold,
    ...typeScale.brand,
  },
  eyebrow: {
    color: colors.textGold,
    fontFamily: fonts.semibold,
    marginTop: metrics.workflowEyebrowTop,
    ...typeScale.eyebrow,
  },
  title: {
    color: colors.textPrimary,
    fontFamily: fonts.extrabold,
    marginTop: spacing.sm,
    ...typeScale.screenTitle,
  },
  subtitle: {
    color: colors.textSecondary,
    fontFamily: fonts.regular,
    marginTop: spacing.sm,
    ...typeScale.body,
  },
  form: { gap: spacing.field, marginTop: spacing.section },
  sectionHeader: { gap: spacing.xs, marginTop: spacing.section },
  stackGap: { gap: spacing.md },
  teacherList: { gap: spacing.sm, marginTop: spacing.md },
  teacher: {
    alignItems: "center",
    backgroundColor: colors.raised,
    borderColor: colors.borderGlass,
    borderRadius: radii.compactCard,
    borderWidth: metrics.borderWidth,
    flexDirection: "row",
    minHeight: 70,
    paddingHorizontal: spacing.lg,
  },
  teacherSelected: { borderColor: colors.violet },
  teacherPressed: { backgroundColor: colors.surface },
  radio: {
    borderColor: colors.textMuted,
    borderRadius: 8,
    borderWidth: 1,
    height: 16,
    width: 16,
  },
  radioSelected: { backgroundColor: colors.violet, borderColor: colors.violet },
  teacherCopy: { flex: 1, marginLeft: spacing.md },
  teacherName: {
    color: colors.textPrimary,
    fontFamily: fonts.semibold,
    ...typeScale.body,
  },
  confirmation: {
    alignItems: "center",
    backgroundColor: colors.raised,
    borderColor: colors.borderGlass,
    borderRadius: radii.compactCard,
    borderWidth: metrics.borderWidth,
    flexDirection: "row",
    minHeight: metrics.minimumTarget,
    marginTop: spacing.xxl,
    padding: spacing.md,
  },
  checkbox: {
    alignItems: "center",
    borderColor: colors.textMuted,
    borderRadius: 6,
    borderWidth: 1,
    height: 24,
    justifyContent: "center",
    width: 24,
  },
  checkboxSelected: { backgroundColor: colors.violet, borderColor: colors.violet },
  checkmark: { color: colors.textOnAction, fontFamily: fonts.bold, fontSize: 16 },
  confirmationText: {
    color: colors.textSecondary,
    flex: 1,
    fontFamily: fonts.regular,
    marginLeft: spacing.md,
    ...typeScale.supporting,
  },
  errorText: {
    color: colors.danger,
    fontFamily: fonts.regular,
    marginTop: spacing.xs,
    ...typeScale.label,
  },
  actions: { gap: spacing.md, marginTop: spacing.xxl },
});
