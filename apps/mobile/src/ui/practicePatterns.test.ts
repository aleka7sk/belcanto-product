import { HOMEWORK_STATUSES } from "../api/contracts";
import {
  HOMEWORK_STATUS_TONE,
  uploadAccent,
  uploadPercent,
} from "./patterns/practicePatterns";
import { semantic } from "./tokens";

describe("practice patterns (Page 23)", () => {
  it("maps every approved homework status to a tone", () => {
    for (const status of HOMEWORK_STATUSES) {
      expect(HOMEWORK_STATUS_TONE[status]).toBeDefined();
    }
    expect(HOMEWORK_STATUS_TONE.completed).toBe("success");
    expect(HOMEWORK_STATUS_TONE.cancelled).toBe("danger");
    expect(HOMEWORK_STATUS_TONE.expired).toBe("muted");
  });

  it("renders real byte progress, clamped and floored", () => {
    expect(uploadPercent(0, 100)).toBe(0);
    expect(uploadPercent(82, 100)).toBe(82);
    expect(uploadPercent(999, 100)).toBe(100);
    expect(uploadPercent(1, 3)).toBe(33);
    expect(uploadPercent(5, 0)).toBe(0);
  });

  it("cycles upload accents violet then cyan (327:463/469)", () => {
    expect(uploadAccent(0)).toBe(semantic.accentViolet);
    expect(uploadAccent(1)).toBe(semantic.accentCyan);
    expect(uploadAccent(2)).toBe(semantic.accentViolet);
  });
});
