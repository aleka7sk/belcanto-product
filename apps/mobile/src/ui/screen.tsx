import type { ReactElement, ReactNode } from "react";
import {
  FlatList,
  KeyboardAvoidingView,
  Platform,
  ScrollView,
  StyleSheet,
  View,
  type FlatListProps,
  type RefreshControlProps,
  type ScrollViewProps,
  type StyleProp,
  type ViewStyle,
} from "react-native";
import { useSafeAreaInsets } from "react-native-safe-area-context";

import { navigation as navTokens, semantic, space } from "./tokens";

/**
 * Screen — the one scaffold behind every screen of the application
 * (Page 37 handoff contract). It owns what screens must never re-invent:
 * the canvas, safe-area padding, the scroll container, the keyboard
 * behaviour and the fixed bottom-navigation host. The navigation slot is
 * rendered OUTSIDE the scroll view and pinned to the bottom edge with
 * the geometry of «Bottom Navigation · Production» (310:20542) — the bar
 * never scrolls with content and content always reserves room for it.
 */
export interface ScreenProps {
  children: ReactNode;
  /** Fixed bottom navigation (RoleBottomNav host). Never scrolls. */
  navigation?: ReactNode | undefined;
  /** Pull-to-refresh control, passed straight to the scroll container. */
  refreshControl?: ReactElement<RefreshControlProps> | undefined;
  /** Extra ScrollView props (test hooks, scroll events). */
  scrollProps?: Omit<ScrollViewProps, "contentContainerStyle" | "refreshControl"> | undefined;
  /** Wrap content in KeyboardAvoidingView — screens with text entry. */
  keyboardAware?: boolean | undefined;
  /** Horizontal content padding. Default space.s4 per production frames. */
  gutter?: number | undefined;
  /** Vertical gap between direct children. */
  contentGap?: number | undefined;
  /**
   * Suppress the top safe-area padding for screens with a full-bleed
   * hero that draws under the status bar (STU-HOME-01).
   */
  topInset?: boolean | undefined;
  contentStyle?: StyleProp<ViewStyle> | undefined;
  testID?: string | undefined;
}

/** Bottom padding reserving room for the pinned navigation host. */
export function contentBottomReserve(safeBottom: number, hasNavigation: boolean): number {
  if (!hasNavigation) {
    return safeBottom + space.s6;
  }
  return safeBottom + navTokens.bottomGap + navTokens.height + space.s5;
}

export function Screen({
  children,
  navigation,
  refreshControl,
  scrollProps,
  keyboardAware = false,
  gutter = space.s4,
  contentGap = space.s3,
  topInset = true,
  contentStyle,
  testID,
}: ScreenProps) {
  const insets = useSafeAreaInsets();
  const scroll = (
    <ScrollView
      keyboardShouldPersistTaps="handled"
      showsVerticalScrollIndicator={false}
      {...scrollProps}
      contentContainerStyle={[
        {
          gap: contentGap,
          paddingBottom: contentBottomReserve(insets.bottom, navigation !== undefined),
          paddingHorizontal: gutter,
          paddingTop: topInset ? insets.top + space.s6 : 0,
        },
        contentStyle,
      ]}
      refreshControl={refreshControl}
    >
      {children}
    </ScrollView>
  );
  return (
    <View style={styles.shell} testID={testID}>
      {keyboardAware ? (
        <KeyboardAvoidingView
          behavior={Platform.OS === "ios" ? "padding" : undefined}
          style={styles.flex}
        >
          {scroll}
        </KeyboardAvoidingView>
      ) : (
        scroll
      )}
      {navigation !== undefined ? (
        <View style={[styles.navigationHost, { bottom: insets.bottom + navTokens.bottomGap }]}>
          {navigation}
        </View>
      ) : null}
    </View>
  );
}

/**
 * ScreenList — the same scaffold with a virtualized list instead of a
 * ScrollView, for feeds and catalogs that grow with server data.
 */
export interface ScreenListProps<Item>
  extends Pick<
    FlatListProps<Item>,
    | "data"
    | "renderItem"
    | "keyExtractor"
    | "ListHeaderComponent"
    | "ListFooterComponent"
    | "ListEmptyComponent"
    | "refreshControl"
    | "onEndReached"
    | "onEndReachedThreshold"
  > {
  navigation?: ReactNode | undefined;
  gutter?: number | undefined;
  contentGap?: number | undefined;
  topInset?: boolean | undefined;
  testID?: string | undefined;
}

export function ScreenList<Item>({
  navigation,
  gutter = space.s4,
  contentGap = space.s3,
  topInset = true,
  testID,
  ...listProps
}: ScreenListProps<Item>) {
  const insets = useSafeAreaInsets();
  return (
    <View style={styles.shell} testID={testID}>
      <FlatList
        keyboardShouldPersistTaps="handled"
        showsVerticalScrollIndicator={false}
        {...listProps}
        contentContainerStyle={{
          gap: contentGap,
          paddingBottom: contentBottomReserve(insets.bottom, navigation !== undefined),
          paddingHorizontal: gutter,
          paddingTop: topInset ? insets.top + space.s6 : 0,
        }}
      />
      {navigation !== undefined ? (
        <View style={[styles.navigationHost, { bottom: insets.bottom + navTokens.bottomGap }]}>
          {navigation}
        </View>
      ) : null}
    </View>
  );
}

const styles = StyleSheet.create({
  flex: { flex: 1 },
  navigationHost: {
    alignItems: "center",
    left: navTokens.sideInset,
    position: "absolute",
    right: navTokens.sideInset,
  },
  shell: { backgroundColor: semantic.bgCanvas, flex: 1 },
});
