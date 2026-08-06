import { LinearGradient } from "expo-linear-gradient";
import {
  ActivityIndicator,
  Image,
  Pressable,
  StyleSheet,
  Text,
  TextInput,
  View,
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

import { Button } from "./button";
import { Screen } from "./screen";
import {
  gradients,
  metrics,
  opacities,
  radius,
  semantic,
  sizes,
  space,
  strokes,
  typeStyles,
} from "./tokens";

/*
 * Shared building blocks, fully on the semantic token layer. The only
 * legacy remainder is `metrics` for the B.0 auth decoration dimensions
 * (BrandBadge, AmbientGlow) — replaced when the Page-32 auth re-skin
 * lands.
 */

export function AppSurface({ children }: { children: ReactNode }) {
  return <View style={styles.surface}>{children}</View>;
}

export interface PremiumScrollScreenProps {
  children: ReactNode;
  keyboardAware?: boolean;
  gutter?: number;
  contentStyle?: ViewStyle;
  scrollProps?: Omit<ScrollViewProps, "contentContainerStyle">;
  /** Fixed bottom navigation host (never scrolls with content). */
  navigation?: ReactNode | undefined;
  testID?: string | undefined;
}

/**
 * @deprecated Thin delegate over the unified Screen scaffold; screens
 * migrate to Screen directly, contour by contour.
 */
export function PremiumScrollScreen({
  children,
  keyboardAware = false,
  gutter = space.s5,
  contentStyle,
  scrollProps,
  navigation,
  testID,
}: PremiumScrollScreenProps) {
  const { refreshControl, ...restScrollProps } = scrollProps ?? {};
  return (
    <Screen
      contentGap={0}
      contentStyle={contentStyle}
      gutter={gutter}
      keyboardAware={keyboardAware}
      navigation={navigation}
      refreshControl={refreshControl}
      scrollProps={restScrollProps}
      testID={testID}
    >
      {children}
    </Screen>
  );
}

export function AmbientGlow() {
  const insets = useSafeAreaInsets();
  return (
    <View
      accessible={false}
      pointerEvents="none"
      style={[styles.ambientGlow, { top: -(insets.top + space.s6) }]}
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
              placeholderTextColor={semantic.textMuted}
              returnKeyType={returnKeyType}
              secureTextEntry={secureTextEntry && !revealed}
              selectionColor={semantic.accentViolet}
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
                hitSlop={space.s1}
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
  return (
    <Button
      accessibilityHint={accessibilityHint}
      busy={busy}
      disabled={disabled}
      kind="primary"
      label={label}
      onPress={onPress}
      shape="pill"
      testID={testID}
    />
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
    <Button
      disabled={disabled}
      kind="secondary"
      label={label}
      onPress={onPress}
      shape="pill"
      tone={tone}
    />
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
    <Button
      align={align === "right" ? "right" : "stretch"}
      kind="text"
      label={label}
      onPress={onPress}
      shape="pill"
    />
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

/**
 * ErrorNotice — a load/action failure the member can actually retry.
 * Every resource screen renders this instead of a bare InlineNotice, so
 * «Повторить» is a working control, not a promise in a title.
 */
export function ErrorNotice({
  title,
  body,
  actionLabel,
  onAction,
  testID,
}: {
  title: string;
  body: string;
  actionLabel?: string | undefined;
  onAction?: (() => void) | undefined;
  testID?: string | undefined;
}) {
  return (
    <View style={styles.errorStack} testID={testID}>
      <InlineNotice body={body} title={title} tone="error" />
      {onAction !== undefined && actionLabel !== undefined ? (
        <Button kind="secondary" label={actionLabel} onPress={onAction} shape="block" />
      ) : null}
    </View>
  );
}

export function LoadingScreen({ label = "Загружаем пространство Belcanto" }) {
  return (
    <AppSurface>
      <View accessibilityLabel={label} accessibilityRole="progressbar" style={styles.loading}>
        <BrandBadge />
        <ActivityIndicator color={semantic.accentViolet} size="large" />
        <Text style={styles.loadingText}>{label}</Text>
      </View>
    </AppSurface>
  );
}

/**
 * @deprecated Legacy type helper for the remaining B.0 screens; colors
 * already semantic, metric literals preserved for visual parity until
 * each consumer migrates to typeStyles.
 */
export const uiStyles = StyleSheet.create({
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
  },
  screenTitle: {
    color: semantic.textPrimary,
    fontFamily: "Onest_800ExtraBold",
    fontSize: 28,
    lineHeight: 34,
  },
  sectionTitle: {
    color: semantic.textPrimary,
    fontFamily: "Onest_600SemiBold",
    fontSize: 16,
    lineHeight: 21,
  },
  cardTitle: {
    color: semantic.textPrimary,
    fontFamily: "Onest_700Bold",
    fontSize: 19,
    lineHeight: 23,
  },
  body: {
    color: semantic.textSecondary,
    ...typeStyles.bodyS,
  },
  supporting: {
    color: semantic.textMuted,
    ...typeStyles.caption,
  },
});

const styles = StyleSheet.create({
  surface: { backgroundColor: semantic.bgCanvas, flex: 1 },
  ambientGlow: {
    height: metrics.authGlowHeight,
    position: "absolute",
    right: -space.s5,
    width: metrics.authGlowWidth,
  },
  authGlowImage: { height: "100%", width: "100%" },
  badge: { alignItems: "center", justifyContent: "center", overflow: "hidden" },
  badgeImage: { height: "100%", position: "absolute", width: "100%" },
  fieldStack: { gap: space.s1 },
  fieldStackInvalid: { minHeight: sizes.inputErrorMin },
  badgeText: {
    color: semantic.textOnAction,
    fontFamily: "Onest_800ExtraBold",
    fontSize: 34,
    includeFontPadding: false,
    lineHeight: 43,
    textAlign: "center",
  },
  badgeTextLarge: { fontFamily: "Onest_700Bold", fontSize: 36, lineHeight: 46 },
  field: {
    backgroundColor: semantic.bgRaised,
    borderColor: semantic.borderDefault,
    borderRadius: radius.md,
    borderWidth: strokes.hairline,
    justifyContent: "center",
    minHeight: sizes.inputMin,
    paddingHorizontal: space.s3,
    paddingVertical: space.s2,
  },
  multilineField: { minHeight: 116 },
  fieldFocused: { borderColor: semantic.borderAccent },
  fieldInvalid: { borderColor: semantic.feedbackDanger },
  disabled: { opacity: opacities.disabled },
  fieldLabel: {
    color: semantic.textMuted,
    ...typeStyles.labelM,
  },
  fieldLabelFocused: { color: semantic.textAccent },
  fieldLabelInvalid: { color: semantic.feedbackDanger },
  inputRow: {
    alignItems: "center",
    flexDirection: "row",
    minHeight: 30,
    position: "relative",
  },
  input: {
    color: semantic.textPrimary,
    flex: 1,
    fontFamily: typeStyles.bodyS.fontFamily,
    fontSize: typeStyles.bodyS.fontSize,
    includeFontPadding: false,
    lineHeight: typeStyles.bodyS.lineHeight,
    minHeight: 30,
    padding: 0,
  },
  multilineInput: { minHeight: 72, paddingTop: space.s1, textAlignVertical: "top" },
  inputWithAction: { paddingRight: space.s2 },
  revealAction: {
    alignItems: "center",
    justifyContent: "center",
    minHeight: sizes.touchMin,
    minWidth: sizes.touchMin,
    position: "absolute",
    right: -space.s1,
    top: -space.s2,
  },
  revealText: {
    color: semantic.textAccent,
    ...typeStyles.labelM,
  },
  fieldError: {
    color: semantic.feedbackDanger,
    ...typeStyles.caption,
  },
  fieldHelper: {
    color: semantic.textMuted,
    ...typeStyles.caption,
  },
  card: {
    backgroundColor: semantic.bgRaised,
    borderColor: semantic.borderGlass,
    borderRadius: 20,
    borderWidth: strokes.hairline,
    padding: space.s4,
  },
  featureCard: {
    borderColor: semantic.borderGlass,
    borderRadius: radius.xl,
    borderWidth: strokes.hairline,
    overflow: "hidden",
    padding: space.s4,
  },
  notice: {
    backgroundColor: semantic.bgRaised,
    borderColor: semantic.borderGlass,
    borderLeftColor: semantic.accentViolet,
    borderLeftWidth: 4,
    borderRadius: 18,
    borderWidth: strokes.hairline,
    gap: space.s2,
    padding: space.s4,
  },
  noticeError: { borderLeftColor: semantic.feedbackDanger },
  noticeSuccess: { borderLeftColor: semantic.feedbackSuccess },
  noticeTitle: {
    color: semantic.textPrimary,
    ...typeStyles.labelL,
  },
  noticeBody: {
    color: semantic.textSecondary,
    ...typeStyles.caption,
  },
  errorStack: { gap: space.s3 },
  loading: {
    alignItems: "center",
    flex: 1,
    gap: space.s6,
    justifyContent: "center",
    padding: space.s5,
  },
  loadingText: {
    color: semantic.textSecondary,
    ...typeStyles.bodyS,
    textAlign: "center",
  },
});
