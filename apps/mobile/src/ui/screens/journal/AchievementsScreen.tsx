import { useLocalSearchParams } from "expo-router";

import { useMessage } from "@/i18n";
import { useSession } from "@/session";
import { InlineNotice } from "../../components";
import { AccountScreenShell, ScreenHeading } from "../../patterns/accountPatterns";
import { AccountNav } from "../account/shared";
import { AchievementsSection } from "./GrowthSections";

/**
 * STU-GROWTH-08 «Достижения» (Figma 321:288) — its own screen. Awards
 * with their evidence, plus the teacher/catalog actions; the progress
 * overview links in instead of stacking three forms inline.
 */
export function AchievementsScreen() {
  const message = useMessage();
  const { state } = useSession();
  const params = useLocalSearchParams<{ studentId?: string }>();
  const paramStudentId = typeof params.studentId === "string" ? params.studentId : null;
  const studentId = paramStudentId ?? state.bootstrap?.studentId ?? null;
  const roles = state.bootstrap?.roles ?? [];
  const canLead = roles.includes("Teacher") || roles.includes("Administrator");
  const canManageCatalog = roles.includes("Owner") || roles.includes("Administrator");

  if (studentId === null) {
    return (
      <AccountScreenShell navigation={<AccountNav />} testID="achievements-guard">
        <InlineNotice
          body={message("growth.guard.body")}
          title={message("growth.guard.title")}
          tone="error"
        />
      </AccountScreenShell>
    );
  }

  return (
    <AccountScreenShell keyboardAware navigation={<AccountNav />} testID="achievements-screen">
      <ScreenHeading
        eyebrow={message("growth.eyebrow")}
        subtitle={message("ach.section.hint")}
        title={message("ach.section.title")}
      />
      <AchievementsSection
        canLead={canLead}
        canManageCatalog={canManageCatalog}
        studentId={studentId}
      />
    </AccountScreenShell>
  );
}
