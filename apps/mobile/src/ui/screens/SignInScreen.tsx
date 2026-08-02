import { router } from "expo-router";
import { useRef, useState } from "react";
import {
  AccessibilityInfo,
  StyleSheet,
  Text,
  TextInput,
  View,
} from "react-native";

import { prepareSignIn, type SignInDraft } from "@/controllers";
import { useSession } from "@/session";
import {
  AmbientGlow,
  BrandBadge,
  InlineNotice,
  PremiumCard,
  PremiumScrollScreen,
  PremiumTextField,
  PrimaryButton,
  TextAction,
} from "../components";
import { colors, fonts, metrics, radii, spacing, typeScale } from "../tokens";
import { apiErrorMessage, formIssueMap } from "../viewModels";

type SignInErrors = Partial<Record<keyof SignInDraft, string | undefined>>;

export function SignInScreen() {
  const { signIn, state } = useSession();
  const [phone, setPhone] = useState("");
  const [password, setPassword] = useState("");
  const [errors, setErrors] = useState<SignInErrors>({});
  const [requestError, setRequestError] = useState<string | null>(null);
  const [supportVisible, setSupportVisible] = useState(false);
  const phoneRef = useRef<TextInput>(null);
  const passwordRef = useRef<TextInput>(null);
  const busy = state.phase === "authenticating" || state.phase === "bootstrapping";

  const submit = async () => {
    setRequestError(null);
    const result = prepareSignIn({ phone, password });
    if (!result.ok) {
      const nextErrors = formIssueMap(result.issues);
      setErrors(nextErrors);
      const first = nextErrors.phone ? phoneRef : passwordRef;
      first.current?.focus();
      AccessibilityInfo.announceForAccessibility("Проверьте поля входа");
      return;
    }
    setErrors({});
    try {
      await signIn(result.value);
      router.replace("/(protected)");
    } catch (error) {
      const message = apiErrorMessage(error, "sign_in");
      setRequestError(message);
      AccessibilityInfo.announceForAccessibility(message);
    }
  };

  return (
    <PremiumScrollScreen keyboardAware contentStyle={styles.content}>
      <AmbientGlow />
      <View style={styles.badgeWrap}>
        <BrandBadge />
      </View>
      <Text accessibilityRole="header" style={styles.title}>
        С возвращением
      </Text>
      <Text style={styles.subtitle}>Войдите в пространство обучения Belcanto</Text>

      <View style={styles.form}>
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
          onSubmitEditing={() => passwordRef.current?.focus()}
          placeholder="+7 700 000 00 00"
          returnKeyType="next"
          textContentType="telephoneNumber"
          value={phone}
        />
        <PremiumTextField
          ref={passwordRef}
          autoComplete="current-password"
          error={errors.password}
          label="Пароль"
          onChangeText={(value) => {
            setPassword(value);
            setErrors((current) => ({ ...current, password: undefined }));
          }}
          onSubmitEditing={() => void submit()}
          placeholder="Введите пароль"
          returnKeyType="done"
          secureTextEntry
          textContentType="password"
          value={password}
        />
      </View>

      {requestError ? (
        <View style={styles.requestNotice}>
          <InlineNotice body={requestError} title="Войти не получилось" tone="error" />
        </View>
      ) : null}

      <View style={styles.primaryAction}>
        <PrimaryButton busy={busy} label="Продолжить" onPress={() => void submit()} />
      </View>
      <View style={styles.forgotAction}>
        <TextAction
          align="right"
          label="Забыли пароль?"
          onPress={() => setSupportVisible(true)}
        />
      </View>

      {supportVisible ? (
        <InlineNotice
          body="Напишите школе по вашему обычному каналу — администратор проверит доступ и поможет восстановить вход."
          title="Связаться со школой"
        />
      ) : null}

      <PremiumCard style={styles.securityCard}>
        <Text style={styles.securityTitle}>Защищённый доступ</Text>
        <Text style={styles.securityBody}>
          Роль и доступ определяются аккаунтом школы.
        </Text>
      </PremiumCard>
      <TextAction
        label="Проблемы со входом? Связаться со школой"
        onPress={() => setSupportVisible(true)}
      />
    </PremiumScrollScreen>
  );
}

const styles = StyleSheet.create({
  content: { minHeight: 780 },
  badgeWrap: { alignItems: "center", marginTop: metrics.authBadgeTop },
  title: {
    color: colors.textPrimary,
    fontFamily: fonts.extrabold,
    marginTop: spacing.loose,
    ...typeScale.authTitle,
  },
  subtitle: {
    color: colors.textSecondary,
    fontFamily: fonts.regular,
    marginTop: spacing.xs,
    ...typeScale.body,
  },
  form: { gap: spacing.field, marginTop: metrics.authFormGap },
  requestNotice: { marginTop: spacing.field },
  primaryAction: { marginTop: spacing.section },
  forgotAction: { marginTop: metrics.authForgotGap },
  securityCard: {
    borderRadius: radii.compactCard,
    marginBottom: spacing.xxl,
    marginTop: spacing.loose,
    minHeight: 100,
    padding: spacing.field,
  },
  securityTitle: {
    color: colors.textPrimary,
    fontFamily: fonts.semibold,
    fontSize: 13,
    lineHeight: 17,
  },
  securityBody: {
    color: colors.textSecondary,
    fontFamily: fonts.regular,
    fontSize: 11,
    lineHeight: 17,
    marginTop: spacing.md,
  },
});
