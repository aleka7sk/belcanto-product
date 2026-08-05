import { useState } from "react";

import { useApiClient, type PolicyVersion } from "@/api";
import { useMessage } from "@/i18n";
import { useSession } from "@/session";
import { InlineNotice } from "../../components";
import {
  AccountBanner,
  AccountScreenShell,
  BlockAction,
  ScreenHeading,
  StatusRow,
} from "../../patterns/accountPatterns";
import { apiErrorMessage, formatBelcantoDate } from "../../viewModels";
import { AccountNav, useAccountResource } from "./shared";

/**
 * ACC-18 · Согласия и документы (Figma 366:939). A new policy version is
 * never applied silently: unaccepted documents surface with an explicit
 * accept action, and acceptance is idempotent server-side.
 */
export function PoliciesScreen() {
  const message = useMessage();
  const api = useApiClient();
  const { runAuthenticated } = useSession();
  const policies = useAccountResource((accessToken) => api.listPolicies(accessToken));
  const [busyPolicy, setBusyPolicy] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const accept = async (policy: PolicyVersion) => {
    setBusyPolicy(policy.id);
    setError(null);
    try {
      await runAuthenticated((accessToken) =>
        api.acceptPolicy(accessToken, { policyVersionId: policy.id }),
      );
      await policies.reload();
    } catch (cause) {
      setError(apiErrorMessage(cause));
    } finally {
      setBusyPolicy(null);
    }
  };

  const pendingPolicy = policies.value?.find((policy) => policy.acceptedAt === undefined);

  return (
    <AccountScreenShell navigation={<AccountNav />} testID="account-policies">
      <ScreenHeading
        eyebrow={message("acc18.eyebrow")}
        subtitle={message("acc18.subtitle")}
        title={message("acc18.title")}
      />
      {policies.value !== null && policies.value.length === 0 ? (
        <InlineNotice body={message("acc18.empty")} title={message("acc18.title")} />
      ) : null}
      {(policies.value ?? []).map((policy) => (
        <StatusRow
          key={policy.id}
          status={
            policy.acceptedAt !== undefined
              ? message("acc18.accepted")
              : message("acc18.pending")
          }
          subtitle={message("acc18.version", {
            version: policy.version,
            date: formatBelcantoDate(policy.effectiveFrom),
          })}
          testID={`policy-${policy.id}`}
          title={message(`acc18.kind.${policy.kind}` as const)}
          tone={policy.acceptedAt !== undefined ? "success" : "warning"}
        />
      ))}
      <AccountBanner
        body={message("acc18.banner.body")}
        title={message("acc18.banner.title")}
      />
      {error !== null ? (
        <InlineNotice body={error} title={message("common.retry")} tone="error" />
      ) : null}
      {pendingPolicy !== undefined ? (
        <BlockAction
          busy={busyPolicy === pendingPolicy.id}
          label={message("acc18.accept")}
          onPress={() => void accept(pendingPolicy)}
          testID="policy-accept"
        />
      ) : null}
    </AccountScreenShell>
  );
}
