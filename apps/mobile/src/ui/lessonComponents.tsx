import { LinearGradient } from "expo-linear-gradient";
import { Pressable, StyleSheet, Text, View } from "react-native";

import type { Lesson } from "@/api";

import { ChevronIcon } from "./icons";
import { gradients, radius, semantic, sizes, space, strokes, typeStyles } from "./tokens";
import { formatLessonDay, formatLessonTime, initials } from "./viewModels";

export function InitialsAvatar({ name, size = sizes.avatarMd }: { name: string; size?: number }) {
  return (
    <LinearGradient
      accessibilityElementsHidden
      colors={gradients.badge}
      importantForAccessibility="no-hide-descendants"
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
        <ChevronIcon color={semantic.textAccent} size={sizes.iconLg} />
      </LinearGradient>
    </Pressable>
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
      accessibilityState={{ checked: selected }}
      onPress={onPress}
      style={({ pressed }) => [
        styles.selectionRow,
        selected && styles.selectionRowSelected,
        pressed && styles.pressed,
      ]}
    >
      <View
        style={[
          styles.selectionMark,
          kind === "radio" && styles.radioMark,
          selected && styles.selectionMarkSelected,
        ]}
      >
        <Text style={styles.selectionCheck}>{selected ? "✓" : ""}</Text>
      </View>
      <View style={styles.selectionCopy}>
        <Text style={styles.selectionLabel}>{label}</Text>
        {supporting ? <Text style={styles.selectionSupporting}>{supporting}</Text> : null}
      </View>
    </Pressable>
  );
}

const styles = StyleSheet.create({
  avatar: {
    alignItems: "center",
    borderColor: semantic.borderGlass,
    borderWidth: strokes.hairline,
    justifyContent: "center",
  },
  avatarText: {
    color: semantic.textPrimary,
    ...typeStyles.labelL,
  },
  lessonPressable: { borderRadius: radius.xl, overflow: "hidden" },
  pressed: { opacity: 0.78 },
  lessonCard: {
    alignItems: "center",
    borderColor: semantic.borderGlass,
    borderRadius: radius.xl,
    borderWidth: strokes.hairline,
    flexDirection: "row",
    minHeight: 104,
    padding: space.s4,
  },
  lessonDate: {
    alignItems: "center",
    borderRightColor: semantic.borderGlass,
    borderRightWidth: strokes.hairline,
    minWidth: 62,
    paddingRight: space.s3,
  },
  lessonDay: {
    color: semantic.textGold,
    textTransform: "uppercase",
    ...typeStyles.labelM,
  },
  lessonTime: {
    color: semantic.textPrimary,
    marginTop: space.s1,
    ...typeStyles.headingM,
  },
  lessonCopy: { flex: 1, paddingHorizontal: space.s3 },
  lessonTitle: { color: semantic.textPrimary, ...typeStyles.labelL },
  lessonSupporting: {
    color: semantic.textSecondary,
    marginTop: space.s2,
    ...typeStyles.caption,
  },
  selectionRow: {
    alignItems: "center",
    backgroundColor: semantic.bgRaised,
    borderColor: semantic.borderGlass,
    borderRadius: 18,
    borderWidth: strokes.hairline,
    flexDirection: "row",
    minHeight: 68,
    paddingHorizontal: space.s4,
  },
  selectionRowSelected: { borderColor: semantic.borderAccent },
  selectionMark: {
    alignItems: "center",
    borderColor: semantic.textMuted,
    borderRadius: 6,
    borderWidth: strokes.hairline,
    justifyContent: "center",
    minHeight: 22,
    minWidth: 22,
  },
  radioMark: { borderRadius: radius.pill },
  selectionMarkSelected: {
    backgroundColor: semantic.bgAction,
    borderColor: semantic.borderAccent,
  },
  selectionCheck: { color: semantic.textOnAction, ...typeStyles.labelL },
  selectionCopy: { flex: 1, marginLeft: space.s3 },
  selectionLabel: { color: semantic.textPrimary, ...typeStyles.labelL },
  selectionSupporting: {
    color: semantic.textMuted,
    marginTop: space.s1,
    ...typeStyles.caption,
  },
});
