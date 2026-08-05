import { ROLE_PRIORITY, resolveActiveRole } from "./activeRole";

describe("active working role (ACC-04/19/20, HOF-03)", () => {
  it("keeps the preferred role while the account still holds it", () => {
    expect(resolveActiveRole(["Student", "Teacher"], "Teacher")).toBe("Teacher");
    expect(resolveActiveRole(["Owner", "Teacher"], "Owner")).toBe("Owner");
  });

  it("falls back by priority when the preferred role is revoked", () => {
    expect(resolveActiveRole(["Student"], "Teacher")).toBe("Student");
    expect(resolveActiveRole(["Administrator", "Owner"], "Teacher")).toBe(
      "Administrator",
    );
  });

  it("defaults to the learning role first", () => {
    expect(resolveActiveRole(["Owner", "Teacher", "Student"], null)).toBe(
      "Student",
    );
    expect(ROLE_PRIORITY[0]).toBe("Student");
  });

  it("returns null when the account has no roles yet (AUTH-11)", () => {
    expect(resolveActiveRole([], null)).toBeNull();
    expect(resolveActiveRole([], "Student")).toBeNull();
  });
});
