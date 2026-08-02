import { LinearGradient } from "expo-linear-gradient";
import { router } from "expo-router";
import {
  ImageBackground,
  ScrollView,
  StyleSheet,
  Text,
  View,
  useWindowDimensions,
} from "react-native";
import { useSafeAreaInsets } from "react-native-safe-area-context";

import type { FirstMinute } from "@/api";
import { useSession } from "@/session";
import stageHero from "../../../assets/images/welcome-stage.png";
import {
  AppSurface,
  FeatureCard,
  PremiumCard,
  SecondaryButton,
  uiStyles,
} from "../components";
import {
  colors,
  fonts,
  gradients,
  metrics,
  spacing,
  typeScale,
} from "../tokens";
import { formatBelcantoDate } from "../viewModels";

export function StudentHomeScreen({
  fullName,
  firstMinute,
  onOpenStaff,
}: {
  fullName: string;
  firstMinute: FirstMinute;
  onOpenStaff?: (() => void) | undefined;
}) {
  const insets = useSafeAreaInsets();
  const { height } = useWindowDimensions();
  const { signOut } = useSession();
  const leave = async () => {
    await signOut();
    router.replace("/");
  };
  return (
    <AppSurface>
      <ScrollView
        contentContainerStyle={[styles.scroll, { minHeight: height }]}
        showsVerticalScrollIndicator={false}
      >
        <View style={styles.column}>
          <ImageBackground
            accessibilityIgnoresInvertColors
            accessible={false}
            resizeMode="cover"
            source={stageHero}
            style={styles.hero}
          >
            <LinearGradient
              colors={gradients.homeOverlay}
              locations={[0, 0.56, 1]}
              pointerEvents="none"
              style={StyleSheet.absoluteFill}
            />
          </ImageBackground>
          <View
            style={[
              styles.header,
              { paddingTop: insets.top + spacing.sm },
            ]}
          >
            <Text style={styles.brand}>BELCANTO</Text>
            <View style={styles.headerSpacer} />
            <Text accessibilityRole="header" style={styles.name}>
              {fullName}
            </Text>
            <Text style={styles.eyebrow}>ВАШ ПЕРВЫЙ ОРИЕНТИР</Text>
          </View>

          <View
            style={[
              styles.content,
              { paddingBottom: insets.bottom + spacing.xxl },
            ]}
          >
            <PremiumCard>
              <Text style={uiStyles.cardTitle}>Ваша первая минута в Belcanto</Text>
              <Text style={[uiStyles.supporting, styles.cardIntro]}>
                Педагог уже отметил главное, чтобы приложение началось с вашей
                реальной точки роста.
              </Text>
              <FocusRow
                body={firstMinute.whatWorked}
                color={colors.violet}
                title="Что уже получилось"
              />
              <FocusRow
                body={firstMinute.currentFocus}
                color={colors.cyan}
                title="Фокус сейчас"
              />
            </PremiumCard>

            <FeatureCard>
              <Text style={styles.nextEyebrow}>СЛЕДУЮЩИЙ ШАГ</Text>
              <Text style={styles.nextStep}>{firstMinute.nextStep}</Text>
              <Text style={styles.publishedAt}>
                Ориентир обновлён {formatBelcantoDate(firstMinute.publishedAt)}
              </Text>
            </FeatureCard>

            {onOpenStaff ? (
              <SecondaryButton
                label="Рабочее пространство"
                onPress={onOpenStaff}
              />
            ) : null}
            <SecondaryButton label="Выйти" onPress={() => void leave()} />
          </View>
        </View>
      </ScrollView>
    </AppSurface>
  );
}

function FocusRow({
  title,
  body,
  color,
}: {
  title: string;
  body: string;
  color: string;
}) {
  return (
    <View style={styles.focusRow}>
      <View accessible={false} style={[styles.focusRail, { backgroundColor: color }]} />
      <View style={styles.focusCopy}>
        <Text style={styles.focusTitle}>{title}</Text>
        <Text style={styles.focusBody}>{body}</Text>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  scroll: { flexGrow: 1, width: "100%" },
  column: {
    alignSelf: "center",
    flex: 1,
    maxWidth: metrics.contentMaxWidth,
    width: "100%",
  },
  hero: { height: metrics.homeHeroHeight, left: 0, position: "absolute", right: 0, top: 0 },
  header: { height: metrics.homeHeroHeight, paddingHorizontal: metrics.authGutter },
  brand: {
    color: colors.textPrimary,
    fontFamily: fonts.bold,
    ...typeScale.homeBrand,
  },
  headerSpacer: { flex: 1 },
  name: {
    color: colors.textPrimary,
    fontFamily: fonts.extrabold,
    ...typeScale.homeName,
  },
  eyebrow: {
    color: colors.textGold,
    fontFamily: fonts.semibold,
    marginBottom: spacing.xl,
    marginTop: spacing.xs,
    ...typeScale.eyebrow,
  },
  content: { gap: spacing.lg, paddingHorizontal: metrics.homeGutter, paddingTop: spacing.lg },
  cardIntro: { marginBottom: spacing.lg, marginTop: spacing.sm },
  focusRow: {
    borderTopColor: colors.borderGlass,
    borderTopWidth: 1,
    flexDirection: "row",
    gap: spacing.md,
    paddingTop: spacing.lg,
    marginTop: spacing.lg,
  },
  focusRail: { borderRadius: 2, minHeight: 56, width: 4 },
  focusCopy: { flex: 1 },
  focusTitle: {
    color: colors.textAccent,
    fontFamily: fonts.semibold,
    ...typeScale.label,
  },
  focusBody: {
    color: colors.textPrimary,
    fontFamily: fonts.medium,
    marginTop: spacing.xs,
    ...typeScale.body,
  },
  nextEyebrow: {
    color: colors.textGold,
    fontFamily: fonts.semibold,
    ...typeScale.eyebrow,
  },
  nextStep: {
    color: colors.textPrimary,
    fontFamily: fonts.bold,
    marginTop: spacing.md,
    ...typeScale.cardTitle,
  },
  publishedAt: {
    color: colors.textSecondary,
    fontFamily: fonts.regular,
    marginTop: spacing.xxl,
    ...typeScale.label,
  },
});
