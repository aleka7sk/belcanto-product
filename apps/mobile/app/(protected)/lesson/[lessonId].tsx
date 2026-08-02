import { useLocalSearchParams } from "expo-router";

import { useSession } from "@/session";
import { LoadingScreen } from "@/ui/components";
import { LessonDetailScreen } from "@/ui/screens/LessonDetailScreen";

export default function LessonDetailRoute() {
  const params = useLocalSearchParams<{ lessonId: string | string[] }>();
  const { state } = useSession();
  const lessonId = Array.isArray(params.lessonId) ? params.lessonId[0] : params.lessonId;
  const firstMinute = state.bootstrap?.firstMinute;
  if (!lessonId || firstMinute === undefined) return <LoadingScreen />;
  return <LessonDetailScreen lessonId={lessonId} firstMinute={firstMinute} />;
}
