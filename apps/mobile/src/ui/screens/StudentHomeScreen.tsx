import { LinearGradient } from "expo-linear-gradient";
import { router } from "expo-router";
import { useCallback, useEffect, useState } from "react";
import {
  ImageBackground,
  StyleSheet,
  Text,
  View,
} from "react-native";
import { useSafeAreaInsets } from "react-native-safe-area-context";

import type { FirstMinute, IsoDateTime, Lesson } from "@/api";
import { useApiClient } from "@/api";
import { useSession } from "@/session";
import stageHero from "../../../assets/images/welcome-stage.png";
import {
  ErrorNotice,
  FeatureCard,
  PremiumCard,
  SecondaryButton,
  uiStyles,
} from "../components";
import { LessonCard } from "../lessonComponents";
import { Screen } from "../screen";
import { gradients, metrics, semantic, space, typeStyles } from "../tokens";
import { apiErrorMessage, formatBelcantoDate } from "../viewModels";
import { RoleNav } from "./account/shared";

export function StudentHomeScreen({
  fullName,
  studentId,
  firstMinute,
  onOpenStaff,
}: {
  fullName: string;
  studentId: string;
  firstMinute: FirstMinute;
  onOpenStaff?: (() => void) | undefined;
}) {
  const insets = useSafeAreaInsets();
  const api = useApiClient();
  const { signOut, runAuthenticated } = useSession();
  const [lessons, setLessons] = useState<Lesson[]>([]);
  const [loadingLessons, setLoadingLessons] = useState(true);
  const [lessonError, setLessonError] = useState<string | null>(null);
  const loadLessons = useCallback(async () => {
    setLoadingLessons(true);
    setLessonError(null);
    const now = new Date();
    const to = new Date(now.getTime() + 30 * 24 * 60 * 60 * 1000);
    try {
      const result = await runAuthenticated((accessToken) =>
        api.listLessons(accessToken, {
          from: now.toISOString() as IsoDateTime,
          to: to.toISOString() as IsoDateTime,
          studentId,
        }),
      );
      setLessons(
        [...result].sort(
          (left, right) =>
            new Date(left.startsAt).getTime() - new Date(right.startsAt).getTime(),
        ),
      );
    } catch (error) {
      setLessonError(apiErrorMessage(error));
    } finally {
      setLoadingLessons(false);
    }
  }, [api, runAuthenticated, studentId]);

  useEffect(() => {
    let active = true;
    queueMicrotask(() => { if (active) void loadLessons(); });
    return () => { active = false; };
  }, [loadLessons]);
  const leave = async () => {
    await signOut();
    router.replace("/");
  };
  return (
    <Screen
      contentGap={0}
      gutter={0}
      navigation={<RoleNav active="today" />}
      testID="student-home"
      topInset={false}
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
              { paddingTop: insets.top + space.s2 },
            ]}
          >
            <Text style={styles.brand}>BELCANTO</Text>
            <View style={styles.headerSpacer} />
            <Text accessibilityRole="header" style={styles.name}>
              {fullName}
            </Text>
            <Text style={styles.eyebrow}>ВАШ ПЕРВЫЙ ОРИЕНТИР</Text>
          </View>

          <View style={styles.content}>
            <View style={styles.sectionHeader}>
              <Text style={uiStyles.sectionTitle}>Ближайший урок</Text>
              <Text style={uiStyles.supporting}>Реальное расписание школы</Text>
            </View>
            {loadingLessons ? (
              <PremiumCard><Text style={uiStyles.body}>Загружаем расписание…</Text></PremiumCard>
            ) : null}
            {lessonError ? (
              <ErrorNotice
                actionLabel="Повторить"
                body={lessonError}
                onAction={() => void loadLessons()}
                title="Расписание не загрузилось"
              />
            ) : null}
            {!loadingLessons && !lessonError && lessons[0] ? (
              <LessonCard
                lesson={lessons[0]}
                onPress={() =>
                  router.push({
                    pathname: "/(protected)/lesson/[lessonId]",
                    params: { lessonId: lessons[0]!.id },
                  })
                }
              />
            ) : null}
            {!loadingLessons && !lessonError && lessons.length === 0 ? (
              <PremiumCard>
                <Text style={uiStyles.sectionTitle}>Уроков пока нет</Text>
                <Text style={[uiStyles.body, styles.emptyBody]}>
                  Здесь появится ближайшее занятие, когда школа добавит его в расписание.
                </Text>
              </PremiumCard>
            ) : null}

            <PremiumCard>
              <Text style={uiStyles.cardTitle}>Ваша первая минута в Belcanto</Text>
              <Text style={[uiStyles.supporting, styles.cardIntro]}>
                Педагог уже отметил главное, чтобы приложение началось с вашей
                реальной точки роста.
              </Text>
              <FocusRow
                body={firstMinute.whatWorked}
                color={semantic.accentViolet}
                title="Что уже получилось"
              />
              <FocusRow
                body={firstMinute.currentFocus}
                color={semantic.accentCyan}
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
    </Screen>
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
  column: {
    alignSelf: "center",
    flex: 1,
    maxWidth: 430 /* B.0 column cap */,
    width: "100%",
  },
  hero: { height: metrics.homeHeroHeight, left: 0, position: "absolute", right: 0, top: 0 },
  header: { height: metrics.homeHeroHeight, paddingHorizontal: space.s5 },
  brand: {
    color: semantic.textPrimary,
    fontFamily: "Onest_700Bold",
    fontSize: 11,
    letterSpacing: 2,
    lineHeight: 15,
  },
  headerSpacer: { flex: 1 },
  name: {
    color: semantic.textPrimary,
    fontFamily: "Onest_800ExtraBold",
    fontSize: 27,
    lineHeight: 32,
  },
  eyebrow: {
    color: semantic.textGold,
    fontFamily: "Onest_600SemiBold",
    fontSize: 10,
    letterSpacing: 1,
    lineHeight: 13,
    marginBottom: space.s5,
    marginTop: space.s1,
  },
  content: { gap: space.s4, paddingHorizontal: space.s4, paddingTop: space.s4 },
  sectionHeader: { gap: space.s1, marginBottom: -space.s2 },
  cardIntro: { marginBottom: space.s4, marginTop: space.s2 },
  emptyBody: { marginTop: space.s2 },
  focusRow: {
    borderTopColor: semantic.borderGlass,
    borderTopWidth: 1,
    flexDirection: "row",
    gap: space.s3,
    paddingTop: space.s4,
    marginTop: space.s4,
  },
  focusRail: { borderRadius: 2, minHeight: 56, width: 4 },
  focusCopy: { flex: 1 },
  focusTitle: {
    color: semantic.textAccent,
    ...typeStyles.labelM,
  },
  focusBody: {
    color: semantic.textPrimary,
    fontFamily: "Onest_500Medium",
    fontSize: 14,
    lineHeight: 20,
    marginTop: space.s1,
  },
  nextEyebrow: {
    color: semantic.textGold,
    fontFamily: "Onest_600SemiBold",
    fontSize: 10,
    letterSpacing: 1,
    lineHeight: 13,
  },
  nextStep: {
    color: semantic.textPrimary,
    fontFamily: "Onest_700Bold",
    fontSize: 19,
    lineHeight: 23,
    marginTop: space.s3,
  },
  publishedAt: {
    color: semantic.textSecondary,
    marginTop: space.s6,
    ...typeStyles.caption,
  },
});
