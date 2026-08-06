import { LinearGradient } from "expo-linear-gradient";
import type { ReactElement, ReactNode } from "react";
import {
  Pressable,
  StyleSheet,
  Text,
  View,
  type RefreshControlProps,
  type ScrollViewProps,
  type ViewStyle,
} from "react-native";

import { Button } from "../button";
import { ChevronIcon, Icon, type IconName } from "../icons";
import { Screen } from "../screen";
import {
  gradients,
  opacities,
  radius,
  semantic,
  sizes,
  space,
  strokes,
  typeStyles,
} from "../tokens";

/**
 * Account & Security pattern kit (Figma Page 32). One vocabulary for the
 * 31 ACC/AUTH states: screen heading, status card, status row, settings
 * row (List Row 135:147), banner and block actions. The screen frame
 * itself lives in the unified Screen scaffold.
 */

export type StatusTone = "success" | "warning" | "danger" | "info" | "muted";

const toneColor: Record<StatusTone, string> = {
  success: semantic.feedbackSuccess,
  warning: semantic.feedbackWarning,
  danger: semantic.feedbackDanger,
  info: semantic.accentCyan,
  muted: semantic.textMuted,
};

/**
 * @deprecated Thin delegate over the unified Screen scaffold — kept so
 * the account-area screens migrate name-by-name without churn.
 */
export function AccountScreenShell({
  children,
  navigation,
  refreshControl,
  scrollProps,
  keyboardAware,
  testID,
}: {
  children: ReactNode;
  navigation?: ReactNode | undefined;
  refreshControl?: ReactElement<RefreshControlProps> | undefined;
  scrollProps?:
    | Omit<ScrollViewProps, "contentContainerStyle" | "refreshControl">
    | undefined;
  keyboardAware?: boolean | undefined;
  testID?: string | undefined;
}) {
  return (
    <Screen
      keyboardAware={keyboardAware}
      navigation={navigation}
      refreshControl={refreshControl}
      scrollProps={scrollProps}
      testID={testID}
    >
      {children}
    </Screen>
  );
}

export function ScreenHeading({
  eyebrow,
  title,
  subtitle,
}: {
  eyebrow: string;
  title: string;
  subtitle: string;
}) {
  return (
    <View accessibilityRole="header" style={styles.heading}>
      <Text style={styles.eyebrow}>{eyebrow.toUpperCase()}</Text>
      <Text style={styles.title}>{title}</Text>
      <Text style={styles.subtitle}>{subtitle}</Text>
    </View>
  );
}

/** «Card · …» — 18-radius feature card with a toned status line. */
export function StatusCard({
  title,
  body,
  status,
  tone = "success",
  accent = false,
}: {
  title: string;
  body: string;
  status?: string | undefined;
  tone?: StatusTone | undefined;
  accent?: boolean | undefined;
}) {
  return (
    <View style={[styles.statusCard, accent ? styles.cardAccentBorder : null]}>
      <Text style={styles.statusCardTitle}>{title}</Text>
      <Text style={styles.statusCardBody}>{body}</Text>
      {status ? (
        <Text style={[styles.statusLine, { color: toneColor[tone] }]}>
          {status}
        </Text>
      ) : null}
    </View>
  );
}

/** «Row · …» — compact detail row with a toned status line. */
export function StatusRow({
  title,
  subtitle,
  status,
  tone = "muted",
  onPress,
  accessibilityHint,
  testID,
}: {
  title: string;
  subtitle: string;
  status?: string | undefined;
  tone?: StatusTone | undefined;
  onPress?: (() => void) | undefined;
  accessibilityHint?: string | undefined;
  testID?: string | undefined;
}) {
  const content = (
    <>
      <Text style={styles.statusRowTitle}>{title}</Text>
      <Text style={styles.statusRowSubtitle}>{subtitle}</Text>
      {status ? (
        <Text style={[styles.statusLine, { color: toneColor[tone] }]}>
          {status}
        </Text>
      ) : null}
    </>
  );
  if (!onPress) {
    return (
      <View style={styles.statusRow} testID={testID}>
        {content}
      </View>
    );
  }
  return (
    <Pressable
      accessibilityHint={accessibilityHint}
      accessibilityLabel={title}
      accessibilityRole="button"
      onPress={onPress}
      style={({ pressed }) => [styles.statusRow, pressed && styles.pressed]}
      testID={testID}
    >
      {content}
    </Pressable>
  );
}

/** Toggle row variant of StatusRow (ACC-12): value flips on press. */
export function ToggleRow({
  title,
  subtitle,
  value,
  onLabel,
  offLabel,
  lockedStatus,
  lockedTone = "warning",
  onToggle,
  testID,
}: {
  title: string;
  subtitle: string;
  value: boolean;
  onLabel: string;
  offLabel: string;
  lockedStatus?: string | undefined;
  lockedTone?: StatusTone | undefined;
  onToggle?: ((next: boolean) => void) | undefined;
  testID?: string | undefined;
}) {
  const locked = onToggle === undefined;
  const status = lockedStatus ?? (value ? onLabel : offLabel);
  const tone: StatusTone = lockedStatus ? lockedTone : value ? "success" : "muted";
  return (
    <Pressable
      accessibilityLabel={title}
      accessibilityRole="switch"
      accessibilityState={{ checked: value, disabled: locked }}
      disabled={locked}
      onPress={() => onToggle?.(!value)}
      style={({ pressed }) => [styles.statusRow, pressed && styles.pressed]}
      testID={testID}
    >
      <Text style={styles.statusRowTitle}>{title}</Text>
      <Text style={styles.statusRowSubtitle}>{subtitle}</Text>
      <Text style={[styles.statusLine, { color: toneColor[tone] }]}>
        {status}
      </Text>
    </Pressable>
  );
}

/**
 * «List Row» (Figma component 135:147). Permission state controls whether
 * a row is shown, disabled with explanation, or omitted — callers omit the
 * row entirely when the capability is absent and pass `disabledReason`
 * when it exists but is unavailable here.
 */
export function SettingsRow({
  icon = "settings",
  title,
  subtitle,
  tail,
  emphasis = false,
  disabledReason,
  onPress,
  testID,
}: {
  icon?: IconName | undefined;
  title: string;
  subtitle: string;
  tail?: string | undefined;
  emphasis?: boolean | undefined;
  disabledReason?: string | undefined;
  onPress: () => void;
  testID?: string | undefined;
}) {
  const disabled = disabledReason !== undefined;
  return (
    <Pressable
      accessibilityHint={disabledReason}
      accessibilityLabel={title}
      accessibilityRole="button"
      accessibilityState={{ disabled }}
      disabled={disabled}
      onPress={onPress}
      style={({ pressed }) => [
        styles.settingsRow,
        emphasis && styles.settingsRowEmphasis,
        pressed && !disabled && styles.pressed,
        disabled && styles.disabled,
      ]}
      testID={testID}
    >
      <View style={styles.iconTile}>
        <Icon color={semantic.iconDefault} name={icon} size={sizes.iconLg} />
      </View>
      <View style={styles.settingsCopy}>
        <Text numberOfLines={1} style={styles.settingsTitle}>
          {title}
        </Text>
        <Text numberOfLines={2} style={styles.settingsSubtitle}>
          {disabledReason ?? subtitle}
        </Text>
      </View>
      <View style={styles.settingsTail}>
        {tail ? <Text style={styles.tailText}>{tail}</Text> : null}
        <ChevronIcon color={semantic.iconDefault} size={sizes.iconMd} />
      </View>
    </Pressable>
  );
}

/** «Banner · …» — accent-bordered callout with a toned title. */
export function AccountBanner({
  title,
  body,
  tone = "gold",
}: {
  title: string;
  body: string;
  tone?: "gold" | "warning" | "success" | "danger";
}) {
  const titleColor =
    tone === "gold"
      ? semantic.textGold
      : tone === "warning"
        ? semantic.feedbackWarning
        : tone === "success"
          ? semantic.feedbackSuccess
          : semantic.feedbackDanger;
  const borderColor =
    tone === "success"
      ? semantic.feedbackSuccess
      : tone === "danger"
        ? semantic.feedbackDanger
        : semantic.borderAccent;
  return (
    <View
      accessibilityRole="text"
      style={[styles.banner, { borderColor }]}
    >
      <Text style={[styles.bannerTitle, { color: titleColor }]}>{title}</Text>
      <Text style={styles.bannerBody}>{body}</Text>
    </View>
  );
}

/**
 * «Action · …» — 52-pt block action.
 * @deprecated Thin delegate over the unified Button; call sites migrate
 * to Button contour by contour.
 */
export function BlockAction({
  label,
  onPress,
  kind = "primary",
  busy = false,
  disabled = false,
  testID,
}: {
  label: string;
  onPress: () => void;
  kind?: "primary" | "secondary" | undefined;
  busy?: boolean | undefined;
  disabled?: boolean | undefined;
  testID?: string | undefined;
}) {
  return (
    <Button
      busy={busy}
      disabled={disabled}
      kind={kind}
      label={label}
      onPress={onPress}
      shape="block"
      testID={testID}
    />
  );
}

/** «Profile hero» (ACC-01/02): accent card with gradient initials avatar. */
export function ProfileHero({
  initials,
  name,
  context,
  status,
  statusTone = "success",
}: {
  initials: string;
  name: string;
  context: string;
  status?: string | undefined;
  statusTone?: StatusTone | undefined;
}) {
  return (
    <View style={styles.hero}>
      <LinearGradient
        colors={gradients.badge}
        end={{ x: 1, y: 1 }}
        start={{ x: 0, y: 0 }}
        style={styles.avatar}
      >
        <Text style={styles.avatarInitials}>{initials}</Text>
      </LinearGradient>
      <View style={styles.heroIdentity}>
        <Text numberOfLines={1} style={styles.heroName}>
          {name}
        </Text>
        <Text numberOfLines={1} style={styles.heroContext}>
          {context}
        </Text>
        {status ? (
          <Text style={[styles.statusLine, { color: toneColor[statusTone] }]}>
            {status}
          </Text>
        ) : null}
      </View>
    </View>
  );
}

export function PatternGap({ style }: { style?: ViewStyle }) {
  return <View style={[{ height: space.s1 }, style]} />;
}

const styles = StyleSheet.create({
  heading: { gap: space.s1 },
  eyebrow: { color: semantic.textGold, ...typeStyles.labelM },
  title: { color: semantic.textPrimary, ...typeStyles.headingL },
  subtitle: { color: semantic.textSecondary, ...typeStyles.bodyS },
  statusCard: {
    backgroundColor: semantic.bgSurface,
    borderColor: semantic.borderDefault,
    borderRadius: 18,
    borderWidth: strokes.hairline,
    gap: space.s2,
    padding: space.s4,
  },
  cardAccentBorder: { borderColor: semantic.borderAccent },
  statusCardTitle: {
    color: semantic.textPrimary,
    fontFamily: "Onest_600SemiBold",
    fontSize: 18,
    lineHeight: 24,
  },
  statusCardBody: { color: semantic.textSecondary, ...typeStyles.bodyS },
  statusLine: { ...typeStyles.labelM },
  statusRow: {
    backgroundColor: semantic.bgSurface,
    borderColor: semantic.borderDefault,
    borderRadius: radius.lg,
    borderWidth: strokes.hairline,
    gap: space.s1,
    justifyContent: "center",
    minHeight: sizes.touchMin,
    padding: 14,
  },
  statusRowTitle: {
    color: semantic.textPrimary,
    fontFamily: "Onest_600SemiBold",
    fontSize: 15,
    lineHeight: 20,
  },
  statusRowSubtitle: {
    color: semantic.textSecondary,
    ...typeStyles.caption,
    letterSpacing: 0.2,
  },
  settingsRow: {
    alignItems: "center",
    backgroundColor: semantic.bgSurface,
    borderColor: semantic.borderDefault,
    borderRadius: radius.lg,
    borderWidth: strokes.hairline,
    flexDirection: "row",
    gap: space.s3,
    minHeight: sizes.touchMin + space.s5,
    padding: space.s3,
  },
  settingsRowEmphasis: { backgroundColor: semantic.bgRaised },
  iconTile: {
    alignItems: "center",
    backgroundColor: semantic.bgRaised,
    borderRadius: radius.md,
    height: 44,
    justifyContent: "center",
    width: 44,
  },
  settingsCopy: { flex: 1, gap: space.s1 },
  settingsTitle: { color: semantic.textPrimary, ...typeStyles.labelL },
  settingsSubtitle: { color: semantic.textSecondary, ...typeStyles.caption },
  settingsTail: { alignItems: "flex-end", gap: space.s1 },
  tailText: { color: semantic.textMuted, ...typeStyles.caption },
  banner: {
    backgroundColor: semantic.bgRaised,
    borderRadius: radius.lg,
    borderWidth: strokes.hairline,
    gap: 6,
    padding: 14,
  },
  bannerTitle: { fontFamily: "Onest_600SemiBold", fontSize: 14, lineHeight: 20 },
  bannerBody: {
    color: semantic.textSecondary,
    ...typeStyles.caption,
    letterSpacing: 0.2,
  },
  hero: {
    alignItems: "center",
    backgroundColor: semantic.bgSurface,
    borderColor: semantic.borderAccent,
    borderRadius: 20,
    borderWidth: strokes.hairline,
    flexDirection: "row",
    gap: space.s4,
    minHeight: 112,
    paddingHorizontal: space.s4,
    paddingVertical: space.s4,
  },
  avatar: {
    alignItems: "center",
    borderColor: semantic.borderAccent,
    borderRadius: radius.pill,
    borderWidth: 3,
    height: sizes.avatarLg,
    justifyContent: "center",
    width: sizes.avatarLg,
  },
  avatarInitials: {
    color: semantic.textOnAction,
    fontFamily: "Onest_700Bold",
    fontSize: 20,
    lineHeight: 26,
  },
  heroIdentity: { flex: 1, gap: space.s1 },
  heroName: {
    color: semantic.textPrimary,
    fontFamily: "Onest_600SemiBold",
    fontSize: 18,
    lineHeight: 24,
  },
  heroContext: {
    color: semantic.textSecondary,
    fontFamily: "Onest_400Regular",
    fontSize: 13,
    lineHeight: 18,
  },
  pressed: { opacity: 0.85 },
  disabled: { opacity: opacities.disabled },
});
