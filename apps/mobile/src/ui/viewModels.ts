import { ApiError, ApiTransportError } from "@/api";
import type {
  FormIssue,
  FormIssueCode,
} from "@/forms/result";
import type {
  IsoDateTime,
  StudentOnboardingState,
} from "@/api/contracts";

const issueMessages: Record<FormIssueCode, string> = {
  required: "Заполните это поле",
  invalid_format: "Проверьте формат значения",
  must_be_future: "Укажите будущую дату",
  must_confirm: "Нужно подтвердить это условие",
  too_short: "Нужно не менее 15 символов",
  too_long: "Значение слишком длинное",
  mismatch: "Пароли не совпадают",
  invalid_value: "Выберите другое значение",
};

export function formIssueMap<Field extends string>(
  issues: readonly FormIssue<Field>[],
): Partial<Record<Field, string | undefined>> {
  const result: Partial<Record<Field, string | undefined>> = {};
  for (const issue of issues) {
    result[issue.field] ??= issueMessages[issue.code];
  }
  return result;
}

export function apiErrorMessage(
  error: unknown,
  context: "default" | "sign_in" = "default",
): string {
  if (error instanceof ApiTransportError) {
    return "Не удалось связаться с Belcanto. Проверьте интернет и попробуйте ещё раз.";
  }
  if (error instanceof ApiError) {
    switch (error.code) {
      case "UNAUTHENTICATED":
        return context === "sign_in"
          ? "Телефон или пароль не подошли."
          : "Сессия завершилась. Войдите снова.";
      case "FORBIDDEN":
        return "У этой учётной записи нет права выполнить действие.";
      case "ACTIVATION_INVALID":
        return "Приглашение больше не действует. Попросите школу открыть доступ заново.";
      case "CONFLICT":
      case "INVALID_STATE":
        return "Состояние уже изменилось. Обновите экран и повторите действие.";
      case "RATE_LIMITED":
        return "Слишком много попыток. Подождите немного и повторите.";
      case "UNAVAILABLE":
        return "Belcanto временно недоступен. Попробуйте ещё раз чуть позже.";
      default:
        return "Не получилось завершить действие. Попробуйте ещё раз.";
    }
  }
  return "Не получилось завершить действие. Попробуйте ещё раз.";
}

export const onboardingStateCopy: Record<
  StudentOnboardingState,
  { eyebrow: string; title: string; description: string }
> = {
  awaiting_first_minute: {
    eyebrow: "НУЖЕН ОРИЕНТИР",
    title: "Ожидает First Belcanto Minute",
    description: "Закреплённый педагог отмечает, что получилось, фокус и следующий шаг.",
  },
  ready_to_invite: {
    eyebrow: "ГОТОВ К ДОСТУПУ",
    title: "Можно отправлять приглашение",
    description: "Первый учебный ориентир уже подготовлен педагогом.",
  },
  invited: {
    eyebrow: "ПРИГЛАШЕНИЕ ОТПРАВЛЕНО",
    title: "Ожидает активации",
    description: "Ученик самостоятельно задаст пароль по персональной ссылке.",
  },
  activated: {
    eyebrow: "ДОСТУП АКТИВИРОВАН",
    title: "Ученик уже внутри Belcanto",
    description: "Учётная запись активна, первый ориентир доступен после входа.",
  },
};

export function formatBelcantoDate(value: IsoDateTime | string): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "дата не указана";
  return new Intl.DateTimeFormat("ru-KZ", {
    dateStyle: "medium",
    timeStyle: "short",
    timeZone: "Asia/Almaty",
  }).format(date);
}

export function roleLabel(roles: readonly string[]): string {
  if (roles.includes("Owner")) return "Владелец";
  if (roles.includes("Teacher")) return "Педагог";
  if (roles.includes("Administrator")) return "Администратор";
  return "Ученик";
}
