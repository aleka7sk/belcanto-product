import { Redirect, Slot } from "expo-router";

import { evaluateRouteGuard } from "@/access";
import { useSession } from "@/session";
import { LoadingScreen } from "@/ui/components";
import { SessionRecoveryScreen } from "@/ui/screens/SessionRecoveryScreen";

export default function ProtectedLayout() {
  const { state } = useSession();
  if (state.phase === "blocked") return <SessionRecoveryScreen />;
  const decision = evaluateRouteGuard(state, { kind: "authenticated" });
  if (decision === "pending") return <LoadingScreen />;
  if (decision !== "allow") return <Redirect href="/sign-in" />;
  return <Slot />;
}
