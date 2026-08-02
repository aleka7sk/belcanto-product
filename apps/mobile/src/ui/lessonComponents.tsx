import { LinearGradient } from "expo-linear-gradient";
import {
  Image,
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  View,
  type ImageSourcePropType,
} from "react-native";

import navHome from "../../assets/images/nav-home.png";
import navSchedule from "../../assets/images/nav-schedule.png";
import type { Lesson } from "@/api";

import { colors, fonts, gradients, metrics, radii, spacing, typeScale } from "./tokens";
import {
  dateKey,
  formatLessonDay,
  formatLessonTime,
  initials,
} from "./viewModels";

export function InitialsAvatar({ name, size = 48 }: { name: string; size?: number }) {
  return (
    <LinearGradient
      accessibilityLabel={`Педагог ${name}`}
      colors={gradients.badge}
      style={[styles.avatar, { borderRadius: size / 2, height: size, width: size }]}
    >
      <Text style={styles.avatarText}>{initials(name)}</Text>
    </LinearGradient>
  );
}

export function LessonCard({
  lesson,
  onPress,
}: {
  lesson: Lesson;
  onPress(): void;
}) {
  const supporting = [
    lesson.location,
    `${lesson.durationMinutes} мин`,
    lesson.teacher.fullName,
  ]
    .filter(Boolean)
    .join(" · ");
  return (
    <Pressable
      accessibilityHint="Открывает подробности урока"
      accessibilityLabel={`${lesson.title}, ${formatLessonDay(lesson.startsAt)}, ${formatLessonTime(lesson.startsAt)}`}
      accessibilityRole="button"
      onPress={onPress}
      style={({ pressed }) => [styles.lessonPressable, pressed && styles.pressed]}
    >
      <LinearGradient
        colors={gradients.feature}
        end={{ x: 1, y: 0.2 }}
        start={{ x: 0, y: 0 }}
        style={styles.lessonCard}
      >
        <View style={styles.lessonDate}>
          <Text style={styles.lessonDay}>{formatLessonDay(lesson.startsAt)}</Text>
          <Text style={styles.lessonTime}>{formatLessonTime(lesson.startsAt)}</Text>
        </View>
        <View style={styles.lessonCopy}>
          <Text numberOfLines={2} style={styles.lessonTitle}>
            {lesson.title}
          </Text>
          <Text numberOfLines={2} style={styles.lessonSupporting}>
            {supporting}
          </Text>
        </View>
        <Text accessibilityElementsHidden style={styles.chevron}>
          ›
        </Text>
      </LinearGradient>
    </Pressable>
  );
}

export function DateStrip({
  dates,
  selected,
  onSelect,
}: {
  dates: readonly Date[];
  selected: string;
  onSelect(value: string): void;
}) {
  const weekday = new Intl.DateTimeFormat("ru-RU", {
    weekday: "short",
    timeZone: "Asia/Almaty",
  });
  return (
    <ScrollView
      accessibilityLabel="Выбор дня"
      horizontal
      showsHorizontalScrollIndicator={false}
      contentContainerStyle={styles.dateStrip}
    >
      {dates.map((date) => {
        const key = dateKey(date);
        const active = key === selected;
        return (
          <Pressable
            accessibilityLabel={`${weekday.format(date)}, ${date.getDate()}`}
            accessibilityRole="button"
            accessibilityState={{ selected: active }}
            key={key}
            onPress={() => onSelect(key)}
            style={styles.dateTarget}
          >
            <View style={[styles.datePill, active && styles.datePillActive]}>
              <Text style={[styles.dateWeekday, active && styles.dateTextActive]}>
                {weekday.format(date).replace(".", "").toUpperCase()}
              </Text>
              <Text style={[styles.dateNumber, active && styles.dateTextActive]}>
                {date.getDate()}
              </Text>
            </View>
          </Pressable>
        );
      })}
    </ScrollView>
  );
}

export function SelectableRow({
  label,
  supporting,
  selected,
  onPress,
  kind = "checkbox",
}: {
  label: string;
  supporting?: string;
  selected: boolean;
  onPress(): void;
  kind?: "checkbox" | "radio";
}) {
  return (
    <Pressable
      accessibilityHint={supporting}
      accessibilityLabel={label}
      accessibilityRole={kind}
      accessibilityState={kind === "radio" ? { checked: selected } : { checked: selected }}
      onPress={onPress}
      style={({ pressed }) => [
        styles.selectionRow,
        selected && styles.selectionRowSelected,
        pressed && styles.pressed,
      ]}
    >
      <View style={[styles.selectionMark, kind === "radio" && styles.radioMark, selected && styles.selectionMarkSelected]}>
        <Text style={styles.selectionCheck}>{selected ? "✓" : ""}</Text>
      </View>
      <View style={styles.selectionCopy}>
        <Text style={styles.selectionLabel}>{label}</Text>
        {supporting ? <Text style={styles.selectionSupporting}>{supporting}</Text> : null}
      </View>
    </Pressable>
  );
}

const tabLabels = ["Сегодня", "Расписание", "Практика", "Сообщество", "Награды"];

export function StudentBottomNavigation({
  active,
  onOpenHome,
  onOpenSchedule,
}: {
  active: "home" | "schedule";
  onOpenHome(): void;
  onOpenSchedule(): void;
}) {
  const source: ImageSourcePropType = active === "home" ? navHome : navSchedule;
  return (
    <View style={styles.navShell}>
      <Image
        accessibilityIgnoresInvertColors
        resizeMode="contain"
        source={source}
        style={styles.navImage}
      />
      <View style={styles.navActions}>
        {tabLabels.map((label, index) => {
          const enabled = index < 2;
          const selected =
            (active === "home" && index === 0) ||
            (active === "schedule" && index === 1);
          return (
            <Pressable
              accessibilityLabel={label}
              accessibilityRole="tab"
              accessibilityState={{ disabled: !enabled, selected }}
              disabled={!enabled}
              key={label}
              onPress={index === 0 ? onOpenHome : onOpenSchedule}
              style={styles.navAction}
            />
          );
        })}
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  avatar: {
    alignItems: "center",
    borderColor: colors.borderGlass,
    borderWidth: metrics.borderWidth,
    justifyContent: "center",
  },
  avatarText: {
    color: colors.textPrimary,
    fontFamily: fonts.bold,
    ...typeScale.sectionTitle,
  },
  lessonPressable: { borderRadius: radii.feature, overflow: "hidden" },
  pressed: { opacity: 0.78 },
  lessonCard: {
    alignItems: "center",
    borderColor: colors.borderGlass,
    borderRadius: radii.feature,
    borderWidth: metrics.borderWidth,
    flexDirection: "row",
    minHeight: 104,
    padding: spacing.lg,
  },
  lessonDate: {
    alignItems: "center",
    borderRightColor: colors.borderGlass,
    borderRightWidth: metrics.borderWidth,
    minWidth: 62,
    paddingRight: spacing.md,
  },
  lessonDay: {
    color: colors.textGold,
    fontFamily: fonts.semibold,
    textTransform: "uppercase",
    ...typeScale.micro,
  },
  lessonTime: {
    color: colors.textPrimary,
    fontFamily: fonts.extrabold,
    marginTop: spacing.xs,
    ...typeScale.cardTitle,
  },
  lessonCopy: { flex: 1, paddingHorizontal: spacing.md },
  lessonTitle: { color: colors.textPrimary, fontFamily: fonts.bold, ...typeScale.sectionTitle },
  lessonSupporting: {
    color: colors.textSecondary,
    fontFamily: fonts.regular,
    marginTop: spacing.sm,
    ...typeScale.supporting,
  },
  chevron: { color: colors.textAccent, fontFamily: fonts.bold, fontSize: 26 },
  dateStrip: { gap: spacing.sm },
  dateTarget: {
    alignItems: "center",
    height: metrics.minimumTarget,
    justifyContent: "center",
    width: metrics.minimumTarget,
  },
  datePill: {
    alignItems: "center",
    borderRadius: radii.control,
    height: 48,
    justifyContent: "center",
    width: 42,
  },
  datePillActive: { backgroundColor: colors.violet },
  dateWeekday: { color: colors.textMuted, fontFamily: fonts.semibold, ...typeScale.micro },
  dateNumber: {
    color: colors.textPrimary,
    fontFamily: fonts.bold,
    marginTop: spacing.xxs,
    ...typeScale.body,
  },
  dateTextActive: { color: colors.textOnAction },
  selectionRow: {
    alignItems: "center",
    backgroundColor: colors.raised,
    borderColor: colors.borderGlass,
    borderRadius: radii.compactCard,
    borderWidth: metrics.borderWidth,
    flexDirection: "row",
    minHeight: 68,
    paddingHorizontal: spacing.lg,
  },
  selectionRowSelected: { borderColor: colors.violet },
  selectionMark: {
    alignItems: "center",
    borderColor: colors.textMuted,
    borderRadius: 6,
    borderWidth: 1,
    height: 22,
    justifyContent: "center",
    width: 22,
  },
  radioMark: { borderRadius: 11 },
  selectionMarkSelected: { backgroundColor: colors.violet, borderColor: colors.violet },
  selectionCheck: { color: colors.textOnAction, fontFamily: fonts.bold, fontSize: 14 },
  selectionCopy: { flex: 1, marginLeft: spacing.md },
  selectionLabel: { color: colors.textPrimary, fontFamily: fonts.semibold, ...typeScale.body },
  selectionSupporting: { color: colors.textMuted, fontFamily: fonts.regular, marginTop: spacing.xs, ...typeScale.supporting },
  navShell: {
    alignSelf: "center",
    aspectRatio: 354 / 76,
    maxWidth: 354,
    position: "relative",
    width: "100%",
  },
  navImage: { height: "100%", width: "100%" },
  navActions: {
    bottom: 0,
    flexDirection: "row",
    left: 0,
    position: "absolute",
    right: 0,
    top: 0,
  },
  navAction: { flex: 1, minHeight: metrics.minimumTarget },
});
