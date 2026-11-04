package domain

import (
	"strings"
	"unicode"
)

type Kind string

const (
	KindValidation   Kind = "validation"
	KindNotFound     Kind = "not_found"
	KindConflict     Kind = "conflict"
	KindUnauthorized Kind = "unauthorized"
	KindForbidden    Kind = "forbidden"
	KindInternal     Kind = "internal"
	KindTimeOut      Kind = "time_out"
	KindLimit        Kind = "limit_exceeded"
)

type ErrorCode string

type AppError struct {
	Code    ErrorCode      `json:"code"`
	Message string         `json:"message"`
	Kind    Kind           `json:"kind"`
	Details map[string]any `json:"details,omitempty"`
}

func (e *AppError) Error() string {
	if e == nil {
		return ""
	}
	if e.Message != "" {
		return e.Message
	}
	return string(e.Code)
}

func NewAppError(code ErrorCode, message string, kind Kind, details map[string]any) *AppError {
	if kind == KindInternal {
		message = "Не удалось выполнить операцию. Попробуйте ещё раз позже."
	}
	if kind == KindTimeOut {
		message = "Сервис отвечает слишком долго. Попробуйте ещё раз."
	}

	return &AppError{
		Code:    code,
		Message: userMessage(code, message),
		Kind:    kind,
		Details: details,
	}
}

func NewNotFound(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: userMessage(code, message),
		Kind:    KindNotFound,
	}
}

func NewConflict(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: userMessage(code, message),
		Kind:    KindConflict,
	}
}

func NewValidationError(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: safeValidationMessage(code, message),
		Kind:    KindValidation,
	}
}

func NewForbidden(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: userMessage(code, message),
		Kind:    KindForbidden,
	}
}

func NewUnauthorized(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: userMessage(code, message),
		Kind:    KindUnauthorized,
	}
}

func NewLimitExceeded(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: userMessage(code, message),
		Kind:    KindLimit,
	}
}

func userMessage(code ErrorCode, fallback string) string {
	messages := map[ErrorCode]string{
		"invalid_uuid":                           "Укажите корректный идентификатор объекта.",
		"time.invalid":                           "Проверьте даты и время.",
		"project.invalid":                        "Проверьте название, описание, дедлайн и приоритет проекта.",
		"project.name_invalid":                   "Укажите название проекта длиной от 1 до 40 символов.",
		"project.description_invalid":            "Описание проекта не должно превышать допустимую длину.",
		"project.priority_invalid":               "Выберите допустимый приоритет проекта.",
		"project.deadline_invalid":               "Дедлайн проекта должен быть корректной датой и не может быть раньше даты создания проекта.",
		"project.status_invalid":                 "Выберите допустимый статус проекта.",
		"project.metrics_invalid":                "Показатели проекта должны быть числами от 0 до 100.",
		"project.risk_invalid":                   "Укажите допустимый уровень риска проекта.",
		"project.not_found":                      "Проект не найден. Обновите страницу.",
		"stage.invalid":                          "Проверьте название, цель, дедлайн и порядок этапа.",
		"stage.name_invalid":                     "Укажите название этапа.",
		"stage.goal_invalid":                     "Укажите цель этапа.",
		"stage.order_invalid":                    "Порядок этапа должен быть положительным целым числом.",
		"stage.deadline_invalid":                 "Укажите корректный дедлайн этапа.",
		"stage.status_invalid":                   "Выберите допустимый статус этапа.",
		"stage.metrics_invalid":                  "Показатели этапа должны быть числами от 0 до 100.",
		"stage.risk_invalid":                     "Укажите допустимый уровень риска этапа.",
		"stage.not_found":                        "Этап не найден. Обновите страницу.",
		"task.invalid":                           "Проверьте название, результат, дедлайн и приоритет задачи.",
		"task.not_found":                         "Задача не найдена. Обновите страницу.",
		"task.outcome.invalid":                   "Опишите ожидаемый результат задачи.",
		"task.difficulty.invalid":                "Сложность задачи должна быть числом от 1 до 10.",
		"task.priority.invalid":                  "Выберите допустимый приоритет задачи.",
		"task.deadline.invalid":                  "Укажите корректный дедлайн задачи.",
		"task.status.invalid":                    "Выберите допустимый статус задачи.",
		"task.metrics.invalid":                   "Показатели задачи должны быть числами от 0 до 100.",
		"task.blocks.invalid":                    "Добавьте хотя бы одну часть задачи.",
		"task.block.invalid":                     "Проверьте данные части задачи.",
		"task.block.type.invalid":                "Выберите допустимый тип части задачи.",
		"task.block.weight.invalid":              "Вес блока должен быть неотрицательным числом.",
		"task.block.mastery.invalid":             "Освоение части задачи должно быть числом от 0 до 100.",
		"assessment_snapshot.invalid":            "Проверьте данные оценки.",
		"assessment_snapshot.score.invalid":      "Оценка должна быть числом от 0 до 100.",
		"assessment_snapshot.source.invalid":     "Укажите источник оценки.",
		"assessment_snapshot.confidence.invalid": "Уверенность должна быть числом от 0 до 100.",
		"readiness.invalid":                      "Проверьте настройки готовности.",
		"weight_distribution.invalid":            "Сумма весов должна быть корректной.",
		"decay_profile.invalid":                  "Проверьте настройки затухания актуальности.",
		"recommendation.invalid":                 "Проверьте настройки рекомендаций.",
		"calibration.invalid":                    "Проверьте настройки калибровки.",
		"task.policy.invalid":                    "Проверьте правила задач.",
		"aggregation.policy.invalid":             "Проверьте правила агрегации.",
		"risk.policy.invalid":                    "Проверьте правила расчёта риска.",
		"recommendation.policy.invalid":          "Проверьте правила рекомендаций.",
		"calibration.policy.invalid":             "Проверьте правила калибровки.",
		"library.name_invalid":                   "Укажите название библиотеки без запрещённых символов.",
		"library.description_invalid":            "Описание библиотеки не должно превышать допустимую длину.",
		"library.access_invalid":                 "Проверьте пользователей и их права доступа.",
		"library.not_found":                      "Библиотека не найдена. Обновите страницу.",
		"folder.name_invalid":                    "Укажите корректное название папки.",
		"folder.access_invalid":                  "Проверьте пользователей и их права доступа к папке.",
		"folder.not_found":                       "Папка не найдена. Обновите страницу.",
		"file.type_invalid":                      "Файл этого типа нельзя загрузить.",
		"file.size_invalid":                      "Размер файла превышает допустимый лимит.",
		"file.not_found":                         "Файл не найден. Обновите страницу.",
		"api_key.invalid":                        "Проверьте название и параметры API-ключа.",
		"api_key.not_found":                      "API-ключ не найден. Обновите страницу.",
		"api_key.forbidden":                      "Операция с API-ключами недоступна.",
		"api_key.expired":                        "Срок действия API-ключа истёк.",
		"api_key.revoked":                        "API-ключ уже отозван.",
		"api_key.limit_exceeded":                 "Достигнут лимит API-ключей.",
		"api_key.scope_denied":                   "У API-ключа нет нужной области доступа.",
	}

	if localized, ok := messages[code]; ok {
		return localized
	}

	return strings.TrimSpace(fallback)
}

func safeValidationMessage(code ErrorCode, message string) string {
	message = strings.TrimSpace(message)
	if localized := userMessage(code, ""); localized != "" {
		return localized
	}
	if message == "" || len(message) > 240 || containsUnsafeValidationContent(message) {
		return "Проверьте введённые данные и попробуйте ещё раз."
	}

	return message
}

func containsUnsafeValidationContent(message string) bool {
	lowerMessage := strings.ToLower(message)
	if strings.Contains(lowerMessage, "http://") ||
		strings.Contains(lowerMessage, "https://") ||
		strings.Contains(lowerMessage, "stack trace") ||
		strings.Contains(lowerMessage, "sql") ||
		strings.Contains(lowerMessage, "/var/") ||
		strings.Contains(lowerMessage, "/home/") {
		return true
	}

	for _, char := range message {
		if unicode.IsControl(char) {
			return true
		}
	}

	return false
}

func (k Kind) HTTPCode() int {
	switch k {
	case KindValidation:
		return 400
	case KindNotFound:
		return 404
	case KindConflict:
		return 409
	case KindUnauthorized:
		return 401
	case KindForbidden:
		return 403
	case KindInternal:
		return 500
	case KindTimeOut:
		return 504
	case KindLimit:
		return 402
	default:
		return 500
	}
}
