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

import stageHero from "../../../assets/images/welcome-stage.png";
import { AppSurface, PrimaryButton } from "../components";
import {
  colors,
  fonts,
  gradients,
  metrics,
  spacing,
  typeScale,
} from "../tokens";

export function WelcomeScreen() {
  const insets = useSafeAreaInsets();
  const { height } = useWindowDimensions();
  return (
    <AppSurface>
      <ScrollView
        contentContainerStyle={[styles.scrollContent, { minHeight: height }]}
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
              colors={gradients.welcomeOverlay}
              end={{ x: 0, y: 1 }}
              locations={[0, 0.58, 1]}
              pointerEvents="none"
              start={{ x: 0, y: 0 }}
              style={StyleSheet.absoluteFill}
            />
          </ImageBackground>

          <View
            style={[
              styles.content,
              {
                paddingTop: insets.top + spacing.lg,
                paddingBottom: insets.bottom + spacing.xxl,
              },
            ]}
          >
            <Text style={styles.brand}>BELCANTO</Text>
            <View style={styles.heroSpacer} />
            <Text accessibilityRole="header" style={styles.title}>
              Ваш голос.{"\n"}Ваш путь. Ваша сцена.
            </Text>
            <Text style={styles.body}>
              Уроки, задания, обратная связь и история прогресса — в одном
              премиальном пространстве Belcanto.
            </Text>
            <View style={styles.flexSpacer} />
            <PrimaryButton label="Войти" onPress={() => router.push("/sign-in")} />
            <Text style={styles.footnote}>Доступ выдаёт школа после зачисления</Text>
          </View>
        </View>
      </ScrollView>
    </AppSurface>
  );
}

const styles = StyleSheet.create({
  scrollContent: { flexGrow: 1, width: "100%" },
  column: {
    alignSelf: "center",
    flex: 1,
    maxWidth: metrics.contentMaxWidth,
    minHeight: metrics.welcomeMinimumHeight,
    overflow: "hidden",
    width: "100%",
  },
  hero: {
    height: metrics.welcomeHeroHeight,
    left: 0,
    position: "absolute",
    right: 0,
    top: 0,
  },
  content: {
    flex: 1,
    minHeight: metrics.welcomeMinimumHeight,
    paddingHorizontal: metrics.authGutter,
  },
  brand: {
    color: colors.textPrimary,
    fontFamily: fonts.bold,
    ...typeScale.brand,
  },
  heroSpacer: { height: 382 },
  title: {
    color: colors.textPrimary,
    fontFamily: fonts.extrabold,
    ...typeScale.welcomeTitle,
  },
  body: {
    color: colors.textSecondary,
    fontFamily: fonts.regular,
    marginTop: spacing.section,
    maxWidth: 350,
    ...typeScale.bodyLarge,
  },
  flexSpacer: { flex: 1, minHeight: spacing.section },
  footnote: {
    color: colors.textMuted,
    fontFamily: fonts.regular,
    marginTop: spacing.lg,
    textAlign: "center",
    ...typeScale.label,
  },
});
