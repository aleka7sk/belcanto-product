import { LinearGradient } from "expo-linear-gradient";
import {
  ActivityIndicator,
  Image,
  KeyboardAvoidingView,
  Platform,
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  TextInput,
  View,
  useWindowDimensions,
  type KeyboardTypeOptions,
  type ScrollViewProps,
  type TextInputProps,
  type ViewStyle,
} from "react-native";
import { forwardRef, useState, type ReactNode } from "react";
import { useSafeAreaInsets } from "react-native-safe-area-context";

import authBadge from "../../assets/images/auth-badge.png";
import authGlow from "../../assets/images/auth-glow.png";
import authSuccessBadge from "../../assets/images/auth-success-badge.png";

import {
  colors,
  fonts,
  gradients,
  metrics,
  radii,
  spacing,
  typeScale,
} from "./tokens";

export function AppSurface({ children }: { children: ReactNode }) {
  return <View style={styles.surface}>{children}</View>;
}

export interface PremiumScrollScreenProps {
  children: ReactNode;
  keyboardAware?: boolean;
  gutter?: number;
  contentStyle?: ViewStyle;
  scrollProps?: Omit<ScrollViewProps, "contentContainerStyle">;
}

export function PremiumScrollScreen({
  children,
  keyboardAware = false,
  gutter = metrics.authGutter,
  contentStyle,
  scrollProps,
}: PremiumScrollScreenProps) {
  const insets = useSafeAreaInsets();
  const { height } = useWindowDimensions();
  const content = (
    <ScrollView
      keyboardShouldPersistTaps="handled"
      showsVerticalScrollIndicator={false}
      {...scrollProps}
      contentContainerStyle={[
        styles.scrollContent,
        {
          minHeight: height,
          paddingTop: insets.top + spacing.lg,
          paddingBottom: insets.bottom + spacing.xxl,
          paddingHorizontal: gutter,
        },
        contentStyle,
      ]}
    >
      <View style={styles.contentColumn}>{children}</View>
    </ScrollView>
  );
  return (
    <AppSurface>
      {keyboardAware ? (
        <KeyboardAvoidingView
          behavior={Platform.OS === "ios" ? "padding" : undefined}
          style={styles.flex}
        >
          {content}
        </KeyboardAvoidingView>
      ) : (
        content
      )}
    </AppSurface>
  );
}

export function AmbientGlow() {
  const insets = useSafeAreaInsets();
  return (
    <View
      accessible={false}
      pointerEvents="none"
      style={[styles.ambientGlow, { top: -(insets.top + spacing.lg) }]}
    >
      <Image
        accessibilityIgnoresInvertColors
        resizeMode="stretch"
        source={authGlow}
        style={styles.authGlowImage}
      />
    </View>
  );
}

export function BrandBadge({
  kind = "brand",
  large = false,
}: {
  kind?: "brand" | "success";
  large?: boolean;
}) {
  const size = large ? metrics.confirmationBadge : metrics.authBadge;
  return (
    <View
      accessibilityElementsHidden
      importantForAccessibility="no-hide-descendants"
      style={[styles.badge, { borderRadius: size / 2, height: size, width: size }]}
    >
      <Image
        accessibilityIgnoresInvertColors
        resizeMode="contain"
        source={kind === "brand" ? authBadge : authSuccessBadge}
        style={styles.badgeImage}
      />
      <Text style={[styles.badgeText, large && styles.badgeTextLarge]}>
        {kind === "brand" ? "B" : "✓"}
      </Text>
    </View>
  );
}

interface PremiumTextFieldProps {
  label: string;
  value: string;
  onChangeText(value: string): void;
  placeholder?: string;
  error?: string | undefined;
  helper?: string | undefined;
  secureTextEntry?: boolean;
  keyboardType?: KeyboardTypeOptions;
  autoComplete?: TextInputProps["autoComplete"];
  textContentType?: TextInputProps["textContentType"];
  autoCapitalize?: TextInputProps["autoCapitalize"];
  returnKeyType?: TextInputProps["returnKeyType"];
  onSubmitEditing?: TextInputProps["onSubmitEditing"];
  editable?: boolean;
  multiline?: boolean;
  testID?: string;
}

export const PremiumTextField = forwardRef<TextInput, PremiumTextFieldProps>(
  function PremiumTextField(
    {
      label,
      value,
      onChangeText,
      placeholder,
      error,
      helper,
      secureTextEntry = false,
      keyboardType,
      autoComplete,
      textContentType,
      autoCapitalize = "none",
      returnKeyType,
      onSubmitEditing,
      editable = true,
      multiline = false,
      testID,
    },
    ref,
  ) {
    const [focused, setFocused] = useState(false);
    const [revealed, setRevealed] = useState(false);
    const invalid = error !== undefined;
    return (
      <View
        style={[
          styles.fieldStack,
          invalid && styles.fieldStackInvalid,
          !editable && styles.disabled,
        ]}
      >
        <View
          style={[
            styles.field,
            multiline && styles.multilineField,
            invalid && styles.fieldInvalid,
            focused && !invalid && styles.fieldFocused,
          ]}
        >
          <Text
            style={[
              styles.fieldLabel,
              focused && !invalid && styles.fieldLabelFocused,
              invalid && styles.fieldLabelInvalid,
            ]}
          >
            {label}
          </Text>
          <View style={styles.inputRow}>
            <TextInput
              ref={ref}
              accessibilityHint={error}
              accessibilityLabel={label}
              accessibilityState={{ disabled: !editable }}
              autoCapitalize={autoCapitalize}
              autoComplete={autoComplete}
              editable={editable}
              keyboardType={keyboardType}
              multiline={multiline}
              onBlur={() => setFocused(false)}
              onChangeText={onChangeText}
              onFocus={() => setFocused(true)}
              onSubmitEditing={onSubmitEditing}
              placeholder={placeholder}
              placeholderTextColor={colors.textMuted}
              returnKeyType={returnKeyType}
              secureTextEntry={secureTextEntry && !revealed}
              selectionColor={colors.violet}
              style={[
                styles.input,
                secureTextEntry && styles.inputWithAction,
                multiline && styles.multilineInput,
              ]}
              testID={testID}
              textContentType={textContentType}
              value={value}
            />
            {secureTextEntry ? (
              <Pressable
                accessibilityLabel={revealed ? "Скрыть пароль" : "Показать пароль"}
                accessibilityRole="button"
                hitSlop={spacing.xs}
                onPress={() => setRevealed((current) => !current)}
                style={styles.revealAction}
              >
                <Text style={styles.revealText}>
                  {revealed ? "Скрыть" : "Показать"}
                </Text>
              </Pressable>
            ) : null}
          </View>
        </View>
        {invalid ? (
          <Text accessibilityLiveRegion="polite" style={styles.fieldError}>
            {error}
          </Text>
        ) : helper ? (
          <Text style={styles.fieldHelper}>{helper}</Text>
        ) : null}
      </View>
    );
  },
);

export interface PrimaryButtonProps {
  label: string;
  onPress(): void;
  busy?: boolean;
  disabled?: boolean;
  accessibilityHint?: string;
  testID?: string;
}

export function PrimaryButton({
  label,
  onPress,
  busy = false,
  disabled = false,
  accessibilityHint,
  testID,
}: PrimaryButtonProps) {
  const unavailable = busy || disabled;
  return (
    <Pressable
      accessibilityHint={accessibilityHint}
      accessibilityLabel={label}
      accessibilityRole="button"
      accessibilityState={{ busy, disabled: unavailable }}
      disabled={unavailable}
      onPress={onPress}
      style={({ pressed }) => [
        styles.primaryButton,
        pressed && !unavailable && styles.primaryButtonPressed,
        unavailable && styles.disabledButton,
      ]}
      testID={testID}
    >
      {busy ? (
        <ActivityIndicator color={colors.textOnAction} />
      ) : (
        <Text style={styles.primaryButtonText}>{label}</Text>
      )}
    </Pressable>
  );
}

export function SecondaryButton({
  label,
  onPress,
  disabled = false,
  tone = "default",
}: {
  label: string;
  onPress(): void;
  disabled?: boolean;
  tone?: "default" | "danger";
}) {
  return (
    <Pressable
      accessibilityLabel={label}
      accessibilityRole="button"
      accessibilityState={{ disabled }}
      disabled={disabled}
      onPress={onPress}
      style={({ pressed }) => [
        styles.secondaryButton,
        pressed && styles.secondaryButtonPressed,
        disabled && styles.disabled,
      ]}
    >
      <Text
        style={[
          styles.secondaryButtonText,
          tone === "danger" && styles.dangerText,
        ]}
      >
        {label}
      </Text>
    </Pressable>
  );
}

export function TextAction({
  label,
  onPress,
  align = "center",
}: {
  label: string;
  onPress(): void;
  align?: "center" | "right";
}) {
  return (
    <Pressable
      accessibilityLabel={label}
      accessibilityRole="button"
      onPress={onPress}
      style={[styles.textAction, align === "right" && styles.textActionRight]}
    >
      <Text style={styles.textActionText}>{label}</Text>
    </Pressable>
  );
}

export function PremiumCard({
  children,
  style,
  accessibilityLabel,
}: {
  children: ReactNode;
  style?: ViewStyle;
  accessibilityLabel?: string;
}) {
  return (
    <View accessibilityLabel={accessibilityLabel} style={[styles.card, style]}>
      {children}
    </View>
  );
}

export function FeatureCard({ children }: { children: ReactNode }) {
  return (
    <LinearGradient
      colors={gradients.feature}
      end={{ x: 1, y: 0 }}
      start={{ x: 0, y: 0 }}
      style={styles.featureCard}
    >
      {children}
    </LinearGradient>
  );
}

export function InlineNotice({
  title,
  body,
  tone = "info",
}: {
  title: string;
  body: string;
  tone?: "info" | "error" | "success";
}) {
  return (
    <View
      accessibilityLiveRegion={tone === "info" ? "none" : "polite"}
      style={[
        styles.notice,
        tone === "error" && styles.noticeError,
        tone === "success" && styles.noticeSuccess,
      ]}
    >
      <Text style={styles.noticeTitle}>{title}</Text>
      <Text style={styles.noticeBody}>{body}</Text>
    </View>
  );
}

export function LoadingScreen({ label = "Загружаем пространство Belcanto" }) {
  return (
    <AppSurface>
      <View accessibilityLabel={label} accessibilityRole="progressbar" style={styles.loading}>
        <BrandBadge />
        <ActivityIndicator color={colors.violet} size="large" />
        <Text style={styles.loadingText}>{label}</Text>
      </View>
    </AppSurface>
  );
}

export const uiStyles = StyleSheet.create({
  brand: {
    color: colors.textPrimary,
    fontFamily: fonts.bold,
    ...typeScale.brand,
  },
  eyebrow: {
    color: colors.textGold,
    fontFamily: fonts.semibold,
    ...typeScale.eyebrow,
  },
  screenTitle: {
    color: colors.textPrimary,
    fontFamily: fonts.extrabold,
    ...typeScale.screenTitle,
  },
  sectionTitle: {
    color: colors.textPrimary,
    fontFamily: fonts.semibold,
    ...typeScale.sectionTitle,
  },
  cardTitle: {
    color: colors.textPrimary,
    fontFamily: fonts.bold,
    ...typeScale.cardTitle,
  },
  body: {
    color: colors.textSecondary,
    fontFamily: fonts.regular,
    ...typeScale.body,
  },
  supporting: {
    color: colors.textMuted,
    fontFamily: fonts.regular,
    ...typeScale.supporting,
  },
});

const styles = StyleSheet.create({
  flex: { flex: 1 },
  surface: { backgroundColor: colors.canvas, flex: 1 },
  scrollContent: { flexGrow: 1, width: "100%" },
  contentColumn: {
    alignSelf: "center",
    maxWidth: metrics.contentMaxWidth,
    width: "100%",
  },
  ambientGlow: {
    height: metrics.authGlowHeight,
    position: "absolute",
    right: -metrics.authGutter,
    width: metrics.authGlowWidth,
  },
  authGlowImage: { height: "100%", width: "100%" },
  badge: { alignItems: "center", justifyContent: "center", overflow: "hidden" },
  badgeImage: { height: "100%", position: "absolute", width: "100%" },
  fieldStack: { gap: spacing.xxs },
  fieldStackInvalid: { minHeight: metrics.errorInputMinHeight },
  badgeText: {
    color: colors.textOnAction,
    fontFamily: fonts.extrabold,
    fontSize: 34,
    includeFontPadding: false,
    lineHeight: 43,
    textAlign: "center",
  },
  badgeTextLarge: { fontFamily: fonts.bold, fontSize: 36, lineHeight: 46 },
  field: {
    backgroundColor: colors.raised,
    borderColor: colors.border,
    borderRadius: radii.input,
    borderWidth: metrics.borderWidth,
    justifyContent: "center",
    minHeight: metrics.inputMinHeight,
    paddingHorizontal: spacing.md,
    paddingVertical: 9,
  },
  multilineField: { minHeight: 116 },
  fieldFocused: { borderColor: colors.violet },
  fieldInvalid: { borderColor: colors.danger },
  disabled: { opacity: 0.42 },
  fieldLabel: {
    color: colors.textMuted,
    fontFamily: fonts.medium,
    ...typeScale.label,
  },
  fieldLabelFocused: { color: colors.textAccent },
  fieldLabelInvalid: { color: colors.danger },
  inputRow: {
    alignItems: "center",
    flexDirection: "row",
    minHeight: 30,
    position: "relative",
  },
  input: {
    color: colors.textPrimary,
    flex: 1,
    fontFamily: fonts.regular,
    fontSize: typeScale.body.fontSize,
    includeFontPadding: false,
    lineHeight: typeScale.body.lineHeight,
    minHeight: 30,
    padding: 0,
  },
  multilineInput: { minHeight: 72, paddingTop: spacing.xs, textAlignVertical: "top" },
  inputWithAction: { paddingRight: spacing.sm },
  revealAction: {
    alignItems: "center",
    height: metrics.minimumTarget,
    justifyContent: "center",
    minWidth: metrics.minimumTarget,
    position: "absolute",
    right: -spacing.xs,
    top: -9,
  },
  revealText: {
    color: colors.textAccent,
    fontFamily: fonts.semibold,
    ...typeScale.label,
  },
  fieldError: {
    color: colors.danger,
    fontFamily: fonts.regular,
    ...typeScale.label,
  },
  fieldHelper: {
    color: colors.textMuted,
    fontFamily: fonts.regular,
    ...typeScale.label,
  },
  primaryButton: {
    alignItems: "center",
    backgroundColor: colors.violet,
    borderRadius: radii.pill,
    height: metrics.minimumTarget,
    justifyContent: "center",
    paddingHorizontal: spacing.xl,
  },
  primaryButtonPressed: { backgroundColor: colors.violetPressed },
  disabledButton: { backgroundColor: colors.surface, opacity: 0.42 },
  primaryButtonText: {
    color: colors.textOnAction,
    fontFamily: fonts.semibold,
    fontSize: typeScale.body.fontSize,
    lineHeight: typeScale.body.lineHeight,
  },
  secondaryButton: {
    alignItems: "center",
    backgroundColor: colors.raised,
    borderColor: colors.borderGlass,
    borderRadius: radii.pill,
    borderWidth: metrics.borderWidth,
    minHeight: metrics.minimumTarget,
    justifyContent: "center",
    paddingHorizontal: spacing.xl,
  },
  secondaryButtonPressed: { backgroundColor: colors.surface },
  secondaryButtonText: {
    color: colors.textAccent,
    fontFamily: fonts.semibold,
    ...typeScale.body,
  },
  dangerText: { color: colors.danger },
  textAction: {
    alignItems: "center",
    justifyContent: "center",
    minHeight: metrics.minimumTarget,
    paddingHorizontal: spacing.sm,
  },
  textActionRight: { alignSelf: "flex-end" },
  textActionText: {
    color: colors.textAccent,
    fontFamily: fonts.semibold,
    ...typeScale.supporting,
  },
  card: {
    backgroundColor: colors.raised,
    borderColor: colors.borderGlass,
    borderRadius: radii.card,
    borderWidth: metrics.borderWidth,
    padding: spacing.lg,
  },
  featureCard: {
    borderColor: colors.borderGlass,
    borderRadius: radii.feature,
    borderWidth: metrics.borderWidth,
    overflow: "hidden",
    padding: spacing.lg,
  },
  notice: {
    backgroundColor: colors.raised,
    borderColor: colors.borderGlass,
    borderLeftColor: colors.violet,
    borderLeftWidth: 4,
    borderRadius: radii.compactCard,
    borderWidth: metrics.borderWidth,
    gap: spacing.sm,
    padding: spacing.lg,
  },
  noticeError: { borderLeftColor: colors.danger },
  noticeSuccess: { borderLeftColor: colors.success },
  noticeTitle: {
    color: colors.textPrimary,
    fontFamily: fonts.semibold,
    ...typeScale.body,
  },
  noticeBody: {
    color: colors.textSecondary,
    fontFamily: fonts.regular,
    ...typeScale.supporting,
  },
  loading: {
    alignItems: "center",
    flex: 1,
    gap: spacing.xxl,
    justifyContent: "center",
    padding: spacing.xl,
  },
  loadingText: {
    color: colors.textSecondary,
    fontFamily: fonts.medium,
    textAlign: "center",
    ...typeScale.body,
  },
});
