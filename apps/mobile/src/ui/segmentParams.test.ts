import { parseOverviewSegment } from "./screens/OwnerOverviewScreen";
import { parseStudentsSegment } from "./screens/teacher/TeacherStudentsScreen";

describe("segment URL parameters drive segmented screens", () => {
  it("parses the teacher students segment with a safe default", () => {
    expect(parseStudentsSegment(undefined)).toBe("students");
    expect(parseStudentsSegment("review")).toBe("review");
    expect(parseStudentsSegment("analytics")).toBe("analytics");
    expect(parseStudentsSegment("junk")).toBe("students");
    expect(parseStudentsSegment(["review", "analytics"])).toBe("review");
  });

  it("parses the owner overview segment with a safe default", () => {
    expect(parseOverviewSegment(undefined)).toBe("overview");
    expect(parseOverviewSegment("analytics")).toBe("analytics");
    expect(parseOverviewSegment("governance")).toBe("governance");
    expect(parseOverviewSegment("junk")).toBe("overview");
  });
});
