import { Pressable, StyleSheet, Text, View } from "react-native";

import { radius, semantic, typeStyles } from "../tokens";

/**
 * DateChip — implementation of «Date Chip · Production» (Figma 333:231,
 * Page 35): weekly date chip whose selected/today/event states use
 * shape, border and an indicator dot — never color alone. The
 * screen-reader copy announces the full date and the activity count.
 */
export type DateChipProps = {
  date: Date;
  selected: boolean;
  today: boolean;
  itemCount: number;
  /** The day carries an event the member registered for. */
  eventDay: boolean;
  onPress: () => void;
  testID?: string | undefined;
};

const WEEKDAY = new Intl.DateTimeFormat("ru-RU", {
  weekday: "short",
  timeZone: "Asia/Almaty",
});

const FULL_DATE = new Intl.DateTimeFormat("ru-RU", {
  weekday: "long",
  day: "numeric",
  month: "long",
  timeZone: "Asia/Almaty",
});

const DAY_OF_MONTH = new Intl.DateTimeFormat("ru-RU", {
  day: "numeric",
  timeZone: "Asia/Almaty",
});

export function dateChipAccessibilityLabel(date: Date, itemCount: number): string {
  const base = FULL_DATE.format(date);
  if (itemCount === 0) {
    return `${base}, занятий нет`;
  }
  return `${base}, активностей: ${itemCount}`;
}

export function DateChip({
  date,
  selected,
  today,
  itemCount,
  eventDay,
  onPress,
  testID,
}: DateChipProps) {
  return (
    <Pressable
      accessibilityLabel={dateChipAccessibilityLabel(date, itemCount)}
      accessibilityRole="button"
      accessibilityState={{ selected }}
      onPress={onPress}
      style={[
        styles.chip,
        today && styles.chipToday,
        eventDay && !selected && !today && styles.chipEvent,
        selected && styles.chipSelected,
      ]}
      testID={testID}
    >
      <Text style={styles.weekday}>
        {WEEKDAY.format(date).replace(".", "").toUpperCase()}
      </Text>
      <Text style={styles.day}>{DAY_OF_MONTH.format(date)}</Text>
      {itemCount > 0 ? (
        <View
          style={[
            styles.indicator,
            { backgroundColor: eventDay ? semantic.accentCyan : semantic.accentViolet },
            selected && styles.indicatorOnAction,
          ]}
        />
      ) : null}
    </Pressable>
  );
}

const styles = StyleSheet.create({
  chip: {
    alignItems: "center",
    backgroundColor: semantic.bgRaised,
    borderColor: semantic.borderDefault,
    borderRadius: 18,
    borderWidth: 1,
    gap: 2,
    height: 64,
    justifyContent: "center",
    width: 48,
  },
  chipToday: {
    backgroundColor: semantic.bgSurface,
    borderColor: semantic.accentGold,
  },
  chipEvent: {
    borderColor: semantic.accentCyan,
  },
  chipSelected: {
    backgroundColor: semantic.bgAction,
    borderColor: semantic.borderAccent,
  },
  weekday: {
    color: semantic.textSecondary,
    textAlign: "center",
    ...typeStyles.labelM,
  },
  day: {
    color: semantic.textPrimary,
    textAlign: "center",
    ...typeStyles.headingM,
  },
  indicator: {
    borderRadius: radius.pill,
    height: 6,
    width: 6,
  },
  indicatorOnAction: {
    backgroundColor: semantic.textOnAction,
  },
});
