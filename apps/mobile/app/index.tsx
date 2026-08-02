import { Redirect } from "expo-router";

import { evaluateRouteGuard } from "@/access";
import { useSession } from "@/session";
import { LoadingScreen } from "@/ui/components";
import { WelcomeScreen } from "@/ui/screens/WelcomeScreen";

export default function EntryRoute() {
  const { state } = useSession();
  if (state.phase === "blocked") return <Redirect href="/(protected)" />;
  const decision = evaluateRouteGuard(state, { kind: "authenticated" });
  if (decision === "pending") return <LoadingScreen />;
  if (decision === "sign_in_required") return <WelcomeScreen />;
  if (decision === "forbidden") return <Redirect href="/sign-in" />;
  return <Redirect href="/(protected)" />;
}
