import { router, useLocalSearchParams } from "expo-router";

import { canOpenOperationalWorkspace } from "@/access";
import { useActiveRole } from "@/access/activeRole";
import { useSession } from "@/session";
import { LoadingScreen } from "@/ui/components";
import { NoRolesScreen } from "@/ui/screens/account/NoRolesScreen";
import { StaffWorkspaceScreen } from "@/ui/screens/StaffWorkspaceScreen";
import { StudentHomeScreen } from "@/ui/screens/StudentHomeScreen";

export default function ProtectedEntryRoute() {
  const params = useLocalSearchParams<{ workspace?: string | string[] }>();
  const { state } = useSession();
  const { preferredRole } = useActiveRole();
  const bootstrap = state.bootstrap;
  if (bootstrap === null) return <LoadingScreen />;
  if (bootstrap.roles.length === 0) return <NoRolesScreen />;
  const hasStaffWorkspace = canOpenOperationalWorkspace(bootstrap);
  const hasStudentWorkspace =
    bootstrap.roles.includes("Student") &&
    bootstrap.fullName !== undefined &&
    bootstrap.firstMinute !== undefined;
  const requestedWorkspace =
    (Array.isArray(params.workspace) ? params.workspace[0] : params.workspace) ??
    (preferredRole === "Student"
      ? "student"
      : preferredRole !== null
        ? "staff"
        : undefined);
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
