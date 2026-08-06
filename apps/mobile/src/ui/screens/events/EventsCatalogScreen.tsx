import { router } from "expo-router";
import { useState } from "react";
import { RefreshControl, ScrollView, StyleSheet, View } from "react-native";

import { useApiClient, type EventOccurrence, type IsoDateTime } from "@/api";
import { useMessage } from "@/i18n";
import { ErrorNotice, InlineNotice } from "../../components";
import { AccountScreenShell } from "../../patterns/accountPatterns";
import {
  categoryAccent,
  EventCard,
  FilterChip,
} from "../../patterns/eventPatterns";
import { PremiumContextHero } from "../../patterns/premiumContextHero";
import { semantic, space } from "../../tokens";
import { apiErrorMessage, formatBelcantoDate } from "../../viewModels";
import { AccountNav, useAccountResource } from "../account/shared";

const CATALOG_WINDOW_DAYS = 14;

function seatsLine(
  event: EventOccurrence,
  message: ReturnType<typeof useMessage>,
): string {
  const free = event.capacity - event.confirmedCount;
  if (free <= 0) return message("evt.seats.none");
  if (free <= 2) return message("evt.seats.left", { count: free });
  return message("evt.seats.of", {
    confirmed: event.confirmedCount,
    capacity: event.capacity,
  });
}

/** STU-EVENT-01 · Каталог событий (Figma 335:314). */
export function EventsCatalogScreen() {
  const message = useMessage();
  const api = useApiClient();
  const [categoryFilter, setCategoryFilter] = useState<string | null>(null);
  const events = useAccountResource((accessToken) => {
    const from = new Date();
    const to = new Date(from.getTime() + CATALOG_WINDOW_DAYS * 24 * 3600 * 1000);
    return api.listEvents(accessToken, {
      from: from.toISOString() as IsoDateTime,
      to: to.toISOString() as IsoDateTime,
    });
  });

  const list = events.value ?? [];
  const categories = new Map<string, string>();
  for (const event of list) {
    categories.set(event.categoryId, event.categoryName);
  }
  const visible =
    categoryFilter === null
      ? list
      : list.filter((event) => event.categoryId === categoryFilter);

  return (
    <AccountScreenShell
      navigation={<AccountNav active="schedule" />}
      refreshControl={
        <RefreshControl
          onRefresh={() => {
            void events.reload();
          }}
          refreshing={events.refreshing}
          tintColor={semantic.accentViolet}
        />
      }
      testID="events-catalog"
    >
      <PremiumContextHero
        body={message("evt.catalog.body")}
        eyebrow={message("evt.catalog.eyebrow")}
        metric={message("evt.catalog.metric", {
          formats: categories.size,
          count: list.length,
        })}
        role="Student"
        title={message("evt.catalog.title")}
      />
      {categories.size > 0 ? (
        <ScrollView
          contentContainerStyle={styles.filters}
          horizontal
          showsHorizontalScrollIndicator={false}
        >
          <FilterChip
            accent={semantic.textAccent}
            active={categoryFilter === null}
            label={message("evt.filter.all")}
            onPress={() => setCategoryFilter(null)}
            testID="events-filter-all"
          />
          {[...categories.entries()].map(([categoryId, name]) => (
            <FilterChip
              accent={categoryAccent(categoryId)}
              active={categoryFilter === categoryId}
              key={categoryId}
              label={name}
              onPress={() =>
                setCategoryFilter((current) =>
                  current === categoryId ? null : categoryId,
                )
              }
              testID={`events-filter-${categoryId}`}
            />
          ))}
        </ScrollView>
      ) : null}
      {events.error !== null ? (
        <ErrorNotice
          actionLabel={message("common.retry")}
          body={apiErrorMessage(events.error)}
          onAction={() => void events.reload()}
          title={message("evt.catalog.eyebrow")}
        />
      ) : null}
      {events.value !== null && visible.length === 0 ? (
        <InlineNotice
          body={message("evt.catalog.empty")}
          title={message("evt.catalog.eyebrow")}
        />
      ) : null}
      {visible.map((event) => (
        <EventCard
          accent={categoryAccent(event.categoryId)}
          category={event.categoryName}
          key={event.id}
          meta={[formatBelcantoDate(event.startsAt), event.roomId]
            .filter(Boolean)
            .join(" · ")}
          onPress={() =>
            router.push({
              pathname: "/(protected)/events/[occurrenceId]",
              params: { occurrenceId: event.id },
            })
          }
          seats={seatsLine(event, message)}
          testID={`event-card-${event.id}`}
          title={event.title}
        />
      ))}
      <View style={styles.footer} />
    </AccountScreenShell>
  );
}

const styles = StyleSheet.create({
  filters: { flexDirection: "row", gap: space.s2 },
  footer: { height: space.s2 },
});
