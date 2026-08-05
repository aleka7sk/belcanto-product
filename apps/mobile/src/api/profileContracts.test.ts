import { decodeProfileView } from "./contracts";

describe("profile contract (ACC-01/02)", () => {
  const profile = {
    accountId: "account_1",
    fullName: "Алишер Беков",
    tenantName: "Belcanto Алматы",
    roles: ["Student", "Teacher"],
    phone: "+77001234567",
  };

  it("decodes the profile with its role set", () => {
    expect(decodeProfileView(profile)).toEqual(profile);
  });

  it("rejects unknown roles, duplicates and extra keys", () => {
    expect(() =>
      decodeProfileView({ ...profile, roles: ["Student", "Guest"] }),
    ).toThrow("ProfileView");
    expect(() =>
      decodeProfileView({ ...profile, roles: ["Student", "Student"] }),
    ).toThrow("ProfileView");
    expect(() => decodeProfileView({ ...profile, extra: 1 })).toThrow(
      "ProfileView",
    );
  });

  it("requires the identity fields", () => {
    const { fullName: _omitted, ...withoutName } = profile;
    expect(() => decodeProfileView(withoutName)).toThrow("ProfileView");
  });
});
