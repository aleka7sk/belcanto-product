import {
  BACKEND_INPUT_LIMITS,
  backendIdentifierIssue,
  backendLocaleIssue,
  backendTextIssue,
  backendTimezoneIssue,
  isValidIdempotencyKey,
  utf8ByteLength,
} from "./backend";

describe("backend input parity", () => {
  it("mirrors the finalized service bounds", () => {
    expect(BACKEND_INPUT_LIMITS).toEqual({
      identifierBytes: 128,
      idempotencyKeyBytes: 128,
      fullNameRunes: 200,
      enrollmentReferenceRunes: 100,
      reasonRunes: 500,
      firstMinuteRunes: 500,
      localeBytes: 35,
      timezoneBytes: 100,
    });
  });

  it("counts UTF-8 bytes and Unicode code points independently", () => {
    expect(utf8ByteLength("Ән")).toBe(4);
    expect(backendTextIssue("🙂".repeat(200), 200)).toBeNull();
    expect(backendTextIssue("🙂".repeat(201), 200)).toBe("too_long");
    expect(backendTextIssue("line\nfeed", 200)).toBe("invalid_format");
    expect(utf8ByteLength("\ud800")).toBeNull();
  });

  it("matches identifier and visible-ASCII idempotency constraints", () => {
    expect(backendIdentifierIssue("teacher_1")).toBeNull();
    expect(backendIdentifierIssue("teacher/1")).toBe("invalid_format");
    expect(isValidIdempotencyKey("intent:one_1")).toBe(true);
    expect(isValidIdempotencyKey("intent key")).toBe(false);
    expect(isValidIdempotencyKey("x".repeat(129))).toBe(false);
  });

  it("applies BCP 47, IANA and byte-length checks", () => {
    expect(backendLocaleIssue("ru-KZ")).toBeNull();
    expect(backendLocaleIssue("not_a_locale")).toBe("invalid_format");
    expect(backendTimezoneIssue("Asia/Almaty")).toBeNull();
    expect(backendTimezoneIssue("Mars/Olympus_Mons")).toBe("invalid_format");
    expect(backendTimezoneIssue("x".repeat(101))).toBe("too_long");
  });
});
