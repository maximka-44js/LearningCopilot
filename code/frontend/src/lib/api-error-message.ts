import axios from "axios";

type ApiErrorPayload = {
  code?: unknown;
  kind?: unknown;
  message?: unknown;
  error?: unknown;
};

const safeServerKinds = new Set([
  "validation",
  "not_found",
  "conflict",
  "forbidden",
  "limit_exceeded",
]);

function isSafeMessage(value: unknown): value is string {
  return (
    typeof value === "string" &&
    value.trim().length > 0 &&
    value.trim().length <= 240 &&
    !/(\n|\r|https?:\/\/|\/var\/|\/home\/|stack trace|sql)/i.test(value)
  );
}

export function getApiErrorMessage(
  error: unknown,
  fallback = "Не удалось выполнить операцию. Попробуйте ещё раз."
): string {
  if (!axios.isAxiosError(error)) {
    return fallback;
  }

  if (!error.response) {
    if (error.code === "ECONNABORTED" || error.code === "ETIMEDOUT") {
      return "Сервер отвечает слишком долго. Проверьте соединение и повторите попытку.";
    }

    return "Не удалось связаться с сервером. Проверьте соединение и повторите попытку.";
  }

  const status = error.response.status;
  const payload = error.response.data as ApiErrorPayload | undefined;

  if (status >= 500) {
    return "Сервис временно недоступен. Попробуйте ещё раз через несколько секунд.";
  }

  if (safeServerKinds.has(String(payload?.kind)) && isSafeMessage(payload?.message)) {
    return payload.message.trim();
  }

  if (status >= 400 && status < 500 && isSafeMessage(payload?.error)) {
    return payload.error.trim();
  }

  if (status === 401) {
    return "Сессия истекла. Войдите в систему заново.";
  }

  if (status === 403) {
    return "Недостаточно прав для этого действия.";
  }

  if (status === 404) {
    return "Запрошенный объект не найден. Обновите страницу и попробуйте ещё раз.";
  }

  if (status === 409) {
    return "Такой объект уже существует или был изменён. Обновите страницу.";
  }

  if (status === 413 || status === 429) {
    return "Запрос слишком большой или выполняется слишком часто. Попробуйте позже.";
  }

  if (status >= 400 && status < 500) {
    return "Проверьте введённые данные и попробуйте ещё раз.";
  }

  return fallback;
}
