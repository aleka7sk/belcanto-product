import { Redirect } from "expo-router";

import { evaluateRouteGuard } from "@/access";
import { useSession } from "@/session";

export default function EntryRoute() {
  const { state } = useSession();
  const decision = evaluateRouteGuard(state, { kind: "authenticated" });
  if (decision === "pending" || decision === "forbidden") return null;
  if (decision === "sign_in_required") return <Redirect href="/sign-in" />;
  return <Redirect href="/(protected)" />;
}
