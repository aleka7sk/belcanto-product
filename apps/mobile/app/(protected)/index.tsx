import { router, useLocalSearchParams } from "expo-router";

import { canOpenOperationalWorkspace } from "@/access";
import { useSession } from "@/session";
import { LoadingScreen } from "@/ui/components";
import { StaffWorkspaceScreen } from "@/ui/screens/StaffWorkspaceScreen";
import { StudentHomeScreen } from "@/ui/screens/StudentHomeScreen";

export default function ProtectedEntryRoute() {
  const params = useLocalSearchParams<{ workspace?: string | string[] }>();
  const { state } = useSession();
  const bootstrap = state.bootstrap;
  if (bootstrap === null) return <LoadingScreen />;
  const hasStaffWorkspace = canOpenOperationalWorkspace(bootstrap);
  const hasStudentWorkspace =
    bootstrap.roles.includes("Student") &&
    bootstrap.fullName !== undefined &&
    bootstrap.firstMinute !== undefined;
  const requestedWorkspace = Array.isArray(params.workspace)
    ? params.workspace[0]
    : params.workspace;
  if (hasStudentWorkspace && (!hasStaffWorkspace || requestedWorkspace === "student")) {
    return (
      <StudentHomeScreen
        firstMinute={bootstrap.firstMinute}
        fullName={bootstrap.fullName}
        studentId={bootstrap.studentId}
        onOpenStaff={
          hasStaffWorkspace
            ? () => router.replace({ pathname: "/(protected)", params: { workspace: "staff" } })
            : undefined
        }
      />
    );
  }
  return (
    <StaffWorkspaceScreen
      onOpenStudent={
        hasStudentWorkspace
          ? () => router.replace({ pathname: "/(protected)", params: { workspace: "student" } })
          : undefined
      }
    />
  );
}
