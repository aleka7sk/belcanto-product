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
import { gradients, metrics, semantic, space, typeStyles } from "../tokens";

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
                paddingTop: insets.top + space.s4,
                paddingBottom: insets.bottom + space.s6,
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
    maxWidth: 430 /* B.0 column cap */,
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
    paddingHorizontal: space.s5,
  },
  brand: {
    color: semantic.textPrimary,
    fontFamily: "Onest_700Bold",
    fontSize: 13,
    letterSpacing: 2.4,
    lineHeight: 17,
  },
  heroSpacer: { height: 382 },
  title: {
    color: semantic.textPrimary,
    fontFamily: "Onest_800ExtraBold",
    fontSize: 34,
    lineHeight: 41,
  },
  body: {
    color: semantic.textSecondary,
    fontFamily: "Onest_400Regular",
    fontSize: 15,
    lineHeight: 23,
    marginTop: space.s8,
    maxWidth: 350,
  },
  flexSpacer: { flex: 1, minHeight: space.s8 },
  footnote: {
    color: semantic.textMuted,
    marginTop: space.s4,
    textAlign: "center",
    ...typeStyles.caption,
  },
});
