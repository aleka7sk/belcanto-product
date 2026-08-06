import { router, useLocalSearchParams } from "expo-router";
import { useState } from "react";

import { ApiError, useApiClient, type EventOccurrence } from "@/api";
import { useMessage, type MessageFormatter } from "@/i18n";
import { useSession } from "@/session";
import { InlineNotice } from "../../components";
import {
  AccountBanner,
  AccountScreenShell,
  BlockAction,
  ScreenHeading,
} from "../../patterns/accountPatterns";
import {
  categoryAccent,
  EventDetailCard,
  RsvpControl,
  type RsvpControlState,
} from "../../patterns/eventPatterns";
import { semantic } from "../../tokens";
import { apiErrorMessage, formatBelcantoDate } from "../../viewModels";
import { AccountNav, useAccountResource } from "../account/shared";

type SeatAction =
  | "rsvp"
  | "cancel"
  | "join"
  | "leave"
  | "confirmOffer"
  | "declineOffer";

type SeatFailure = "lastSeat" | "generic" | null;

function controlState(
  event: EventOccurrence,
  busyAction: SeatAction | null,
  failure: SeatFailure,
): RsvpControlState {
  if (busyAction === "rsvp" || busyAction === "join" || busyAction === "confirmOffer") {
    return "reserving";
  }
  if (busyAction === "cancel" || busyAction === "leave" || busyAction === "declineOffer") {
    return "cancelling";
  }
  if (failure !== null) return "error";
  if (event.status !== "scheduled" || new Date(event.startsAt).getTime() <= Date.now()) {
    return "closed";
  }
  if (event.myRsvp === "confirmed") return "confirmed";
  if (event.myOffer !== undefined) return "spotOffered";
  if (event.myWaitlistPosition !== undefined) return "waitlisted";
  if (event.confirmedCount >= event.capacity) return "full";
  return "available";
}

function controlCopy(
  state: RsvpControlState,
  event: EventOccurrence,
  failure: SeatFailure,
  message: MessageFormatter,
): { title: string; subtitle: string } {
  switch (state) {
    case "available":
      return {
        title: message("evt.rsvp.available.title"),
        subtitle: message("evt.rsvp.available.sub", {
          confirmed: event.confirmedCount,
          capacity: event.capacity,
        }),
      };
    case "reserving":
      return {
        title: message("evt.rsvp.reserving.title"),
        subtitle: message("evt.rsvp.reserving.sub"),
      };
    case "confirmed":
      return {
        title: message("evt.rsvp.confirmed.title"),
        subtitle: message("evt.rsvp.confirmed.sub"),
      };
    case "cancelling":
      return {
        title: message("evt.rsvp.cancelling.title"),
        subtitle: message("evt.rsvp.cancelling.sub"),
      };
    case "full":
      return {
        title: message("evt.rsvp.full.title"),
        subtitle: message("evt.rsvp.full.sub"),
      };
    case "waitlisted":
      return {
        title: message("evt.rsvp.waitlisted.title", {
          position: event.myWaitlistPosition ?? 0,
        }),
        subtitle: message("evt.rsvp.waitlisted.sub"),
      };
    case "spotOffered":
      return {
        title: message("evt.rsvp.offered.title"),
        subtitle: message("evt.rsvp.offered.sub", {
          time: event.myOffer ? formatBelcantoDate(event.myOffer.expiresAt) : "",
        }),
      };
    case "offerExpired":
      return {
        title: message("evt.rsvp.expired.title"),
        subtitle: message("evt.rsvp.expired.sub"),
      };
    case "closed":
      return {
        title: message("evt.rsvp.closed.title"),
        subtitle: message("evt.rsvp.closed.sub"),
      };
    case "error":
      return {
        title: message("evt.rsvp.error.title"),
        subtitle:
          failure === "lastSeat"
            ? message("evt.rsvp.error.lastSeat")
            : message("evt.rsvp.error.retry"),
      };
    default:
      return {
        title: message("evt.rsvp.waitlistAvailable.title"),
        subtitle: message("evt.rsvp.waitlistAvailable.sub"),
      };
  }
}

/**
 * Event detail with the full seat lifecycle (Figma STU-EVENT-02..12).
 * Every mutation goes through the atomic seat API and re-renders from
 * the returned occurrence view; the last-seat CONFLICT is a first-class
 * state (STU-EVENT-10), never a silent retry.
 */
export function EventDetailScreen() {
  const message = useMessage();
  const api = useApiClient();
  const { runAuthenticated } = useSession();
  const params = useLocalSearchParams<{ occurrenceId?: string }>();
  const occurrenceId = typeof params.occurrenceId === "string" ? params.occurrenceId : "";
  const event = useAccountResource((accessToken) =>
    api.getEvent(accessToken, occurrenceId),
  );
  const [busyAction, setBusyAction] = useState<SeatAction | null>(null);
  const [failure, setFailure] = useState<SeatFailure>(null);
  const [failureText, setFailureText] = useState<string | null>(null);
  const [confirmingCancel, setConfirmingCancel] = useState(false);
  const [justCancelled, setJustCancelled] = useState(false);

  const run = async (action: SeatAction, operation: (accessToken: string) => Promise<EventOccurrence>) => {
    setBusyAction(action);
    setFailure(null);
    setFailureText(null);
    try {
      await runAuthenticated(operation);
      await event.reload();
      if (action === "cancel") {
        setJustCancelled(true);
        setConfirmingCancel(false);
      } else {
        setJustCancelled(false);
      }
    } catch (cause) {
      if (action === "rsvp" && cause instanceof ApiError && cause.code === "CONFLICT") {
        setFailure("lastSeat");
      } else {
        setFailure("generic");
        setFailureText(apiErrorMessage(cause));
      }
      await event.reload();
    } finally {
      setBusyAction(null);
    }
  };

  const view = event.value;
  if (view === null) {
    return (
      <AccountScreenShell navigation={<AccountNav active="schedule" />} testID="event-detail-loading">
        {event.error !== null ? (
          <InlineNotice
            body={apiErrorMessage(event.error)}
            title={message("common.retry")}
            tone="error"
          />
        ) : null}
      </AccountScreenShell>
    );
  }

  const accent = categoryAccent(view.categoryId);
  const state = controlState(view, busyAction, failure);
  const copy = controlCopy(state, view, failure, message);
  const seatsText = message("evt.seats.of", {
    confirmed: view.confirmedCount,
    capacity: view.capacity,
  });

  if (confirmingCancel) {
    return (
      <AccountScreenShell navigation={<AccountNav active="schedule" />} testID="event-cancel-confirm">
        <ScreenHeading
          eyebrow={message("evt.cancel.eyebrow")}
          subtitle={`${view.title} · ${formatBelcantoDate(view.startsAt)}`}
          title={message("evt.cancel.title")}
        />
        <EventDetailCard
          accent={semantic.feedbackWarning}
          body={message("evt.cancel.card.body")}
          status={message("evt.cancel.card.status")}
          statusColor={semantic.accentCyan}
          title={message("evt.cancel.card.title")}
        />
        <BlockAction
          busy={busyAction === "cancel"}
          label={message("evt.cancel.confirm")}
          onPress={() =>
            void run("cancel", (accessToken) =>
              api.cancelEventRsvp(accessToken, view.id),
            )
          }
          testID="event-cancel-confirm-action"
        />
        <BlockAction
          kind="secondary"
          label={message("evt.cancel.stay")}
          onPress={() => setConfirmingCancel(false)}
        />
      </AccountScreenShell>
    );
  }

  return (
    <AccountScreenShell navigation={<AccountNav active="schedule" />} testID="event-detail">
      <ScreenHeading
        eyebrow={view.categoryName}
        subtitle={formatBelcantoDate(view.startsAt)}
        title={view.title}
      />
      {view.description ? (
        <EventDetailCard
          accent={accent}
          body={view.description}
          title={message("evt.detail.about.title")}
        />
      ) : null}
      <EventDetailCard
        accent={semantic.accentCyan}
        body={
          view.roomId
            ? message("evt.detail.host.room", {
                host: view.host.fullName,
                room: view.roomId,
              })
            : view.host.fullName
        }
        title={message("evt.detail.host.title")}
      />
      <EventDetailCard
        accent={semantic.accentGold}
        body={seatsText}
        status={message("evt.detail.conditions.date", {
          date: formatBelcantoDate(view.startsAt),
        })}
        statusColor={semantic.textGold}
        title={message("evt.detail.conditions.title")}
      />
      {justCancelled ? (
        <InlineNotice
          body={message("evt.cancelled.body")}
          title={message("evt.cancelled.title")}
          tone="success"
        />
      ) : null}
      <RsvpControl
        onPress={
          state === "available"
            ? () =>
                void run("rsvp", (accessToken) => api.rsvpToEvent(accessToken, view.id))
            : state === "error" && failure !== "lastSeat"
              ? () =>
                  void run("rsvp", (accessToken) =>
                    api.rsvpToEvent(accessToken, view.id),
                  )
              : undefined
        }
        state={state}
        subtitle={copy.subtitle}
        testID="event-rsvp-control"
        title={copy.title}
      />
      {failureText !== null ? (
        <InlineNotice body={failureText} title={message("common.retry")} tone="error" />
      ) : null}
      {state === "full" || failure === "lastSeat" ? (
        <>
          {failure === "lastSeat" ? (
            <EventDetailCard
              accent={semantic.feedbackWarning}
              body={message("evt.lastSeat.card.body")}
              status={message("evt.lastSeat.card.status")}
              statusColor={semantic.accentCyan}
              title={message("evt.lastSeat.card.title")}
            />
          ) : (
            <EventDetailCard
              accent={semantic.accentCyan}
              body={message("evt.waitlist.how.body")}
              status={message("evt.waitlist.how.status")}
              statusColor={semantic.accentCyan}
              title={message("evt.waitlist.how.title")}
            />
          )}
          <RsvpControl
            onPress={() =>
              void run("join", (accessToken) =>
                api.joinEventWaitlist(accessToken, view.id),
              )
            }
            state="waitlistAvailable"
            subtitle={message("evt.rsvp.waitlistAvailable.sub")}
            testID="event-waitlist-control"
            title={message("evt.rsvp.waitlistAvailable.title")}
          />
        </>
      ) : null}
      {state === "waitlisted" ? (
        <>
          <EventDetailCard
            accent={semantic.accentCyan}
            body={message("evt.waitlist.pending.body")}
            title={message("evt.waitlist.pending.title")}
          />
          <BlockAction
            busy={busyAction === "leave"}
            kind="secondary"
            label={message("evt.waitlist.leave")}
            onPress={() =>
              void run("leave", (accessToken) =>
                api.leaveEventWaitlist(accessToken, view.id),
              )
            }
            testID="event-waitlist-leave"
          />
        </>
      ) : null}
      {state === "spotOffered" && view.myOffer !== undefined ? (
        <>
          <EventDetailCard
            accent={semantic.accentGold}
            body={message("evt.offer.card.body")}
            title={message("evt.offer.card.title", {
              time: formatBelcantoDate(view.myOffer.expiresAt),
            })}
          />
          <BlockAction
            busy={busyAction === "confirmOffer"}
            label={message("evt.offer.confirm")}
            onPress={() =>
              void run("confirmOffer", (accessToken) =>
                api.confirmSpotOffer(accessToken, view.myOffer!.id),
              )
            }
            testID="event-offer-confirm"
          />
          <BlockAction
            busy={busyAction === "declineOffer"}
            kind="secondary"
            label={message("evt.offer.decline")}
            onPress={() =>
              void run("declineOffer", (accessToken) =>
                api.declineSpotOffer(accessToken, view.myOffer!.id),
              )
            }
            testID="event-offer-decline"
          />
        </>
      ) : null}
      {state === "confirmed" ? (
        <BlockAction
          kind="secondary"
          label={message("evt.detail.cancel.action")}
          onPress={() => setConfirmingCancel(true)}
          testID="event-cancel-open"
        />
      ) : null}
      {state === "closed" ? (
        <BlockAction
          kind="secondary"
          label={message("resched.openSchedule")}
          onPress={() => router.push("/(protected)/schedule")}
        />
      ) : null}
      <AccountBanner
        body={message("evt.catalog.body")}
        title={message("evt.catalog.eyebrow")}
      />
    </AccountScreenShell>
  );
}
