import { resolveActiveTab, tabTarget } from "./routes";
import { ROLE_TABS, tabsForRole } from "./tabs";
import type { NavigationRole } from "./tabs";

const ROLES = Object.keys(ROLE_TABS) as NavigationRole[];

describe("tab route map", () => {
  it("routes every tab of every role somewhere", () => {
    for (const role of ROLES) {
      for (const tab of tabsForRole(role)) {
        expect(tabTarget(role, tab.key)).not.toBeNull();
      }
    }
  });

  it("sends staff schedule to the internal lessons workspace, not the student guard", () => {
    expect(tabTarget("Teacher", "schedule")?.pathname).toBe("/(protected)/lessons");
    expect(tabTarget("Administrator", "schedule")?.pathname).toBe("/(protected)/lessons");
    expect(tabTarget("Student", "schedule")?.pathname).toBe("/(protected)/schedule");
  });

  it("sends the teacher today tab to the Page-26 cockpit", () => {
    expect(tabTarget("Teacher", "today")?.pathname).toBe("/(protected)/teacher");
  });

  it("opens the staff workspace explicitly for the administrator people tab", () => {
    expect(tabTarget("Administrator", "people")).toEqual({
      pathname: "/(protected)",
      params: { workspace: "staff" },
    });
  });

  it("keeps segment deep links for review and analytics", () => {
    expect(tabTarget("Teacher", "review")?.params).toEqual({ segment: "review" });
    expect(tabTarget("Owner", "analytics")?.params).toEqual({ segment: "analytics" });
  });

  it("keeps the account area reachable for every role that shows it", () => {
    expect(tabTarget("Student", "profile")?.pathname).toBe("/(protected)/account");
    expect(tabTarget("Administrator", "more")?.pathname).toBe("/(protected)/account");
    expect(tabTarget("Owner", "more")?.pathname).toBe("/(protected)/account");
  });
});

describe("active tab resolution", () => {
  it("keeps a key the role actually has", () => {
    expect(resolveActiveTab("Student", "practice")).toBe("practice");
    expect(resolveActiveTab("Owner", "team")).toBe("team");
  });

  it("maps the account default onto the role's own account key", () => {
    expect(resolveActiveTab("Student", "profile")).toBe("profile");
    expect(resolveActiveTab("Administrator", "profile")).toBe("more");
    expect(resolveActiveTab("Owner", "profile")).toBe("more");
  });

  it("leaves the teacher account screens without a lying highlight", () => {
    expect(resolveActiveTab("Teacher", "profile")).toBe("profile");
    expect(
      tabsForRole("Teacher").some((tab) => tab.key === "profile"),
    ).toBe(false);
  });
});
