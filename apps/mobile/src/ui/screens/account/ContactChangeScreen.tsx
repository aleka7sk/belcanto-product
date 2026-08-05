import { router, useLocalSearchParams } from "expo-router";
import { useState } from "react";

import { useApiClient, type ContactKind } from "@/api";
import { useMessage } from "@/i18n";
import { useSession } from "@/session";
import { InlineNotice, PremiumTextField } from "../../components";
import {
  AccountBanner,
  AccountScreenShell,
  BlockAction,
  ScreenHeading,
  StatusCard,
  StatusRow,
} from "../../patterns/accountPatterns";
import { apiErrorMessage } from "../../viewModels";
import { AccountNav, useAccountResource } from "./shared";

/**
 * ACC-03 · Изменить контакт (Figma 365:189). Two-step flow: re-auth with
 * the current password sends a 6-digit code to the new contact; the code
 * confirms ownership and replaces the verified contact.
 */
export function ContactChangeScreen() {
  const message = useMessage();
  const api = useApiClient();
  const { runAuthenticated } = useSession();
  const params = useLocalSearchParams<{ kind?: string }>();
  const kind: ContactKind = params.kind === "phone" ? "phone" : "email";

  const contacts = useAccountResource((accessToken) =>
    api.listMyContacts(accessToken),
  );
  const current = contacts.value?.find((contact) => contact.kind === kind);

  const [step, setStep] = useState<"request" | "confirm">("request");
  const [value, setValue] = useState("");
  const [password, setPassword] = useState("");
  const [code, setCode] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const start = async () => {
    setBusy(true);
    setError(null);
    try {
      await runAuthenticated((accessToken) =>
        api.startContactChange(accessToken, {
          kind,
          value: value.trim(),
          currentPassword: password,
        }),
      );
      setStep("confirm");
    } catch (cause) {
      setError(apiErrorMessage(cause));
    } finally {
      setBusy(false);
    }
  };

  const confirm = async () => {
    setBusy(true);
    setError(null);
    try {
      await runAuthenticated((accessToken) =>
        api.confirmContactChange(accessToken, { code: code.trim() }),
      );
      router.back();
    } catch (cause) {
      setError(apiErrorMessage(cause));
    } finally {
      setBusy(false);
    }
  };

  return (
    <AccountScreenShell navigation={<AccountNav />} testID="account-contact-change">
      <ScreenHeading
        eyebrow={message("acc03.eyebrow")}
        subtitle={message("acc03.subtitle")}
        title={message(kind === "email" ? "acc03.title.email" : "acc03.title.phone")}
      />
      <StatusRow
        status={
          current
            ? message("acc03.current.confirmed")
            : message("acc03.current.none")
        }
        subtitle={current?.value ?? message("acc03.current.none")}
        title={message(kind === "email" ? "acc03.current.email" : "acc03.current.phone")}
        tone={current ? "success" : "muted"}
      />
      {step === "request" ? (
        <>
          <PremiumTextField
            keyboardType={kind === "email" ? "email-address" : "phone-pad"}
            label={message(kind === "email" ? "acc03.new.email" : "acc03.new.phone")}
            onChangeText={setValue}
            testID="contact-new-value"
            value={value}
          />
          <PremiumTextField
            label={message("acc03.password")}
            onChangeText={setPassword}
            secureTextEntry
            testID="contact-password"
            value={password}
          />
        </>
      ) : (
        <>
          <StatusCard
            body={message("acc03.code.body", { contact: value.trim() })}
            status={message("acc03.code.hint")}
            title={message("acc03.code.title")}
            tone="info"
          />
          <PremiumTextField
            keyboardType="number-pad"
            label={message("acc03.code.label")}
            onChangeText={setCode}
            testID="contact-code"
            value={code}
          />
        </>
      )}
      <AccountBanner
        body={message("acc03.banner.body")}
        title={message("acc03.banner.title")}
      />
      {error !== null ? (
        <InlineNotice body={error} title={message("common.retry")} tone="error" />
      ) : null}
      {step === "request" ? (
        <BlockAction
          busy={busy}
          disabled={value.trim().length === 0 || password.length === 0}
          label={message("acc03.sendCode")}
          onPress={() => void start()}
          testID="contact-start"
        />
      ) : (
        <BlockAction
          busy={busy}
          disabled={code.trim().length === 0}
          label={message(
            kind === "email" ? "acc03.submit.email" : "acc03.submit.phone",
          )}
          onPress={() => void confirm()}
          testID="contact-confirm"
        />
      )}
      <BlockAction
        kind="secondary"
        label={message("common.cancel")}
        onPress={() => router.back()}
      />
    </AccountScreenShell>
  );
}
