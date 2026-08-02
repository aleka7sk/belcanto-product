import { Redirect } from "expo-router";

import { evaluateRouteGuard } from "@/access";
import { useSession } from "@/session";
import { LoadingScreen } from "@/ui/components";
import { SignInScreen } from "@/ui/screens/SignInScreen";

export default function SignInRoute() {
  const { state } = useSession();
  if (state.phase === "blocked") return <Redirect href="/(protected)" />;
  const decision = evaluateRouteGuard(state, { kind: "authenticated" });
  if (decision === "pending") return <LoadingScreen label="Проверяем вход" />;
  if (decision === "allow") return <Redirect href="/(protected)" />;
  return <SignInScreen />;
}
