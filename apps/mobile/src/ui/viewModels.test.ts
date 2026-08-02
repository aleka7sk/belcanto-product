import { ApiError, ApiTransportError } from "@/api";

import {
  apiErrorMessage,
  formIssueMap,
  onboardingStateCopy,
  roleLabel,
} from "./viewModels";

describe("premium UI view models", () => {
  it("keeps only the first message for each invalid field", () => {
    expect(
      formIssueMap([
        { field: "password", code: "too_short" },
        { field: "password", code: "invalid_value" },
        { field: "phone", code: "invalid_format" },
      ]),
    ).toEqual({
      password: "Нужно не менее 15 символов",
      phone: "Проверьте формат значения",
    });
  });

  it("does not expose backend error details to the interface", () => {
    expect(
      apiErrorMessage(
        new ApiError(401, "UNAUTHENTICATED", "secret detail"),
        "sign_in",
      ),
    ).toBe("Телефон или пароль не подошли.");
    expect(
      apiErrorMessage(new ApiError(401, "UNAUTHENTICATED", "secret detail")),
    ).toBe("Сессия завершилась. Войдите снова.");
    expect(apiErrorMessage(new ApiTransportError("socket detail"))).toContain(
      "Не удалось связаться",
    );
  });

  it("describes the exact B.0 queue states", () => {
    expect(onboardingStateCopy.ready_to_invite.eyebrow).toBe("ГОТОВ К ДОСТУПУ");
    expect(onboardingStateCopy.activated.title).toContain("Belcanto");
  });

  it("uses the highest operational role label", () => {
    expect(roleLabel(["Student", "Owner"])).toBe("Владелец");
    expect(roleLabel(["Administrator"])).toBe("Администратор");
  });
});
