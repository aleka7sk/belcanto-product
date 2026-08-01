import { Redirect, Slot } from "expo-router";

import { evaluateRouteGuard } from "@/access";
import { useSession } from "@/session";

export default function ProtectedLayout() {
  const { state } = useSession();
  const decision = evaluateRouteGuard(state, { kind: "authenticated" });
  if (decision === "pending") return null;
  if (decision !== "allow") return <Redirect href="/sign-in" />;
  return <Slot />;
}
