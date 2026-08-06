import { useLocalSearchParams } from "expo-router";

import { useMessage } from "@/i18n";
import { useSession } from "@/session";
import { InlineNotice } from "../../components";
import { AccountScreenShell, ScreenHeading } from "../../patterns/accountPatterns";
import { AccountNav } from "../account/shared";
import { GoalSection } from "./GrowthSections";

/**
 * STU-GROWTH-04 «Цель» (Figma 320:179) — its own screen. The active
 * goal, its history and the teacher's goal actions live here; the
 * progress overview links in instead of stacking the forms inline.
 */
export function GoalScreen() {
  const message = useMessage();
  const { state } = useSession();
  const params = useLocalSearchParams<{ studentId?: string }>();
  const paramStudentId = typeof params.studentId === "string" ? params.studentId : null;
  const studentId = paramStudentId ?? state.bootstrap?.studentId ?? null;
  const roles = state.bootstrap?.roles ?? [];
  const canLead = roles.includes("Teacher") || roles.includes("Administrator");

  if (studentId === null) {
    return (
      <AccountScreenShell navigation={<AccountNav />} testID="goal-guard">
        <InlineNotice
          body={message("growth.guard.body")}
          title={message("growth.guard.title")}
          tone="error"
        />
      </AccountScreenShell>
    );
  }

  return (
    <AccountScreenShell keyboardAware navigation={<AccountNav />} testID="goal-screen">
      <ScreenHeading
        eyebrow={message("growth.eyebrow")}
        subtitle={message("goal.section.hint")}
        title={message("goal.section.eyebrow")}
      />
      <GoalSection canLead={canLead} studentId={studentId} />
    </AccountScreenShell>
  );
}
