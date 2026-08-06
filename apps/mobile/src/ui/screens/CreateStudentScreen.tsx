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
import { semantic, sizes, space, strokes, typeStyles } from "../tokens";
import { apiErrorMessage, formIssueMap } from "../viewModels";
import { RoleNav } from "./account/shared";

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
      <PremiumScrollScreen contentStyle={styles.denied} navigation={<RoleNav active="people" />}>
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
    <PremiumScrollScreen keyboardAware contentStyle={styles.content} navigation={<RoleNav active="people" />}>
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
  denied: { gap: space.s4, justifyContent: "center", minHeight: 680 },
  content: { minHeight: 980 },
  brand: {
    color: semantic.textPrimary,
    fontFamily: "Onest_700Bold",
    fontSize: 13,
    letterSpacing: 2.4,
    lineHeight: 17,
  },
  eyebrow: {
    color: semantic.textGold,
    fontFamily: "Onest_600SemiBold",
    fontSize: 10,
    letterSpacing: 1,
    lineHeight: 13,
    marginTop: space.s10,
  },
  title: {
    color: semantic.textPrimary,
    fontFamily: "Onest_800ExtraBold",
    fontSize: 28,
    lineHeight: 34,
    marginTop: space.s2,
  },
  subtitle: {
    color: semantic.textSecondary,
    marginTop: space.s2,
    ...typeStyles.bodyS,
  },
  form: { gap: space.s5, marginTop: space.s8 },
  sectionHeader: { gap: space.s1, marginTop: space.s8 },
  stackGap: { gap: space.s3 },
  teacherList: { gap: space.s2, marginTop: space.s3 },
  teacher: {
    alignItems: "center",
    backgroundColor: semantic.bgRaised,
    borderColor: semantic.borderGlass,
    borderRadius: 18,
    borderWidth: strokes.hairline,
    flexDirection: "row",
    minHeight: 70,
    paddingHorizontal: space.s4,
  },
  teacherSelected: { borderColor: semantic.accentViolet },
  teacherPressed: { backgroundColor: semantic.bgSurface },
  radio: {
    borderColor: semantic.textMuted,
    borderRadius: 8,
    borderWidth: 1,
    height: 16,
    width: 16,
  },
  radioSelected: { backgroundColor: semantic.bgAction, borderColor: semantic.accentViolet },
  teacherCopy: { flex: 1, marginLeft: space.s3 },
  teacherName: {
    color: semantic.textPrimary,
    ...typeStyles.labelL,
  },
  confirmation: {
    alignItems: "center",
    backgroundColor: semantic.bgRaised,
    borderColor: semantic.borderGlass,
    borderRadius: 18,
    borderWidth: strokes.hairline,
    flexDirection: "row",
    minHeight: sizes.touchMin,
    marginTop: space.s6,
    padding: space.s3,
  },
  checkbox: {
    alignItems: "center",
    borderColor: semantic.textMuted,
    borderRadius: 6,
    borderWidth: 1,
    height: 24,
    justifyContent: "center",
    width: 24,
  },
  checkboxSelected: { backgroundColor: semantic.bgAction, borderColor: semantic.accentViolet },
  checkmark: { color: semantic.textOnAction, fontFamily: "Onest_700Bold", fontSize: 16 },
  confirmationText: {
    color: semantic.textSecondary,
    flex: 1,
    marginLeft: space.s3,
    ...typeStyles.caption,
  },
  errorText: {
    color: semantic.feedbackDanger,
    marginTop: space.s1,
    ...typeStyles.caption,
  },
  actions: { gap: space.s3, marginTop: space.s6 },
});
