import { Pressable, StyleSheet, Text, View } from "react-native";

import type { NavigationRole, TabDefinition, TabKey } from "../navigation/tabs";
import { tabsForRole } from "../navigation/tabs";
import { Icon } from "./icons";
import { opacities, radius, semantic, space, strokes } from "./tokens";

/**
 * RoleBottomNav — implementation of «Bottom Navigation · Production»
 * (Figma 310:20542, Page 35): one accessible role-aware navigation shell
 * for all four workspaces. Active is positional within the role mapping;
 * disabled modules stay visible but inert (no empty slots, HOF-03).
 * The shell reflows: label growth under large font scales increases item
 * height instead of clipping (§3 accessibility contract).
 */
export type RoleBottomNavProps = {
  role: NavigationRole;
  active: TabKey;
  label: (key: TabKey) => string;
  onSelectTab: (tab: TabDefinition) => void;
  isTabEnabled?: (tab: TabDefinition) => boolean;
};

export function RoleBottomNav({
  role,
  active,
  label,
  onSelectTab,
  isTabEnabled,
}: RoleBottomNavProps) {
  const tabs = tabsForRole(role);
  return (
    <View accessibilityRole="tablist" style={styles.shell}>
      {tabs.map((tab) => {
        const selected = tab.key === active;
        const enabled = isTabEnabled ? isTabEnabled(tab) : true;
        const tabLabel = label(tab.key);
        return (
          <Pressable
            accessibilityLabel={tabLabel}
            accessibilityRole="tab"
            accessibilityState={{ disabled: !enabled, selected }}
            disabled={!enabled}
            key={tab.key}
            onPress={() => onSelectTab(tab)}
            style={[
              styles.item,
              selected ? styles.itemActive : null,
              enabled ? null : styles.itemDisabled,
            ]}
            testID={`nav-${role.toLowerCase()}-${tab.key}`}
          >
            <Icon
              color={selected ? semantic.textOnAction : semantic.iconDefault}
              name={tab.icon}
              size={20}
            />
            <Text style={[styles.label, selected ? styles.labelActive : null]}>
              {tabLabel}
            </Text>
          </Pressable>
        );
      })}
    </View>
  );
}

const styles = StyleSheet.create({
  shell: {
    alignItems: "center",
    alignSelf: "center",
    backgroundColor: semantic.bgRaised,
    borderColor: semantic.borderDefault,
    borderRadius: radius.xl,
    borderWidth: strokes.default,
    flexDirection: "row",
    gap: space.s1,
    justifyContent: "center",
    maxWidth: 366,
    minHeight: 68,
    padding: space.s1,
    width: "100%",
  },
  item: {
    alignItems: "center",
    borderRadius: radius.lg,
    flex: 1,
    gap: space.s1,
    justifyContent: "center",
    minHeight: 56,
    paddingHorizontal: space.s1,
    paddingVertical: space.s2,
  },
  itemActive: {
    backgroundColor: semantic.bgAction,
  },
  itemDisabled: {
    opacity: opacities.disabled,
  },
  label: {
    color: semantic.textMuted,
    fontFamily: "Onest_600SemiBold",
    fontSize: 10,
    letterSpacing: -0.1,
    lineHeight: 12,
    textAlign: "center",
  },
  labelActive: {
    color: semantic.textOnAction,
  },
});
