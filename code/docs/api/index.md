---
title: Использование REST API
description: Практическое руководство по REST API Learning Copilot
---

# Использование REST API

REST API предназначен для интеграций с Learning Copilot: автоматического
создания учебных планов, обновления задач, записи оценок и работы с
материалами.

## Адрес и формат запросов

В локальном окружении API доступен по адресу:

```text
http://app.lvh.me/v1
```

В другом окружении замените домен на адрес приложения. Все пути ниже указаны
относительно `/v1`. Для JSON-запросов передавайте заголовок:

```http
Content-Type: application/json
Accept: application/json
```

Идентификаторы проектов, этапов, задач, библиотек, папок и файлов имеют формат
UUID. Даты передаются в ISO 8601, например `2026-12-31T23:59:59Z`.

Полная машинная спецификация API хранится в файле
`code/backend/docs/swagger.yaml` репозитория. Эта страница описывает
рекомендуемый порядок использования API и основные сценарии.

## Авторизация

### Bearer-токен

Защищённые методы принимают заголовок:

```http
Authorization: Bearer <JWT>
```

### API-ключ

Для внешней интеграции создайте ключ в разделе **API-ключи** интерфейса или
через методы `/access/api-keys`. Секрет возвращается только при создании.
Передавайте его тем же способом:

```http
Authorization: Bearer flk_<secret>
```

Ключ должен иметь область доступа, соответствующую операции. Методы `GET`,
`HEAD` и `OPTIONS` требуют `:read`, остальные методы — `:write`.

Доступные домены:

| Домен | Назначение |
|---|---|
| `projects` | проекты, этапы, задачи, оценки и политики |
| `storage` | библиотеки, папки и файлы |
| `chat` | зарезервирован для функций чата |

Для создания API-ключей используется JWT-авторизация. Пример запроса:

```bash
curl -X POST "http://app.lvh.me/v1/access/api-keys" \
  -H "Authorization: Bearer <JWT>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Синхронизация учебного плана",
    "scopes": ["projects:read", "projects:write"],
    "expires_at": "2027-01-01T00:00:00Z"
  }'
```

Сохраните поле `secret` из ответа. В списке ключей секрет не возвращается.

## Рекомендуемый сценарий работы

1. Создайте проект через `POST /projects`.
2. Получите его идентификатор и добавьте этап через
   `POST /projects/{project_id}/stages`.
3. Добавьте задачи через
   `POST /projects/{project_id}/stages/{stage_id}/tasks`.
4. Записывайте результаты обучения через
   `POST /projects/{project_id}/stages/{stage_id}/tasks/{task_id}/assessments`.
5. Получайте готовность задачи и дашборд проекта.
6. После изменения данных при необходимости запускайте пересчёт готовности.

### Проект, этап и задача

```bash
BASE_URL="http://app.lvh.me/v1"
AUTH="Authorization: Bearer flk_<secret>"

curl -X POST "$BASE_URL/projects" \
  -H "$AUTH" -H "Content-Type: application/json" \
  -d '{
    "name": "Подготовка к экзамену",
    "description": "План подготовки по неделям",
    "priority": "normal",
    "deadline_at": "2026-12-31T23:59:59Z"
  }'
```

Для создания этапа используйте `name`, `goal`, `order` и необязательный
`deadline_at`:

```bash
curl -X POST "$BASE_URL/projects/<project_id>/stages" \
  -H "$AUTH" -H "Content-Type: application/json" \
  -d '{
    "name": "Клеточная биология",
    "goal": "Разобраться в строении клетки",
    "order": 1
  }'
```

Задача требует `outcome`, `difficulty` от 1 до 10, `priority` и непустой
массив `blocks`. Типы блоков: `comprehension`, `memorization`, `practice`.

```bash
curl -X POST "$BASE_URL/projects/<project_id>/stages/<stage_id>/tasks" \
  -H "$AUTH" -H "Content-Type: application/json" \
  -d '{
    "outcome": "Объяснить механизмы клеточного дыхания",
    "difficulty": 6,
    "priority": "normal",
    "blocks": [
      {"type": "comprehension", "weight": 40},
      {"type": "memorization", "weight": 30},
      {"type": "practice", "weight": 30}
    ]
  }'
```

### Оценка и готовность

Оценка передаётся в диапазоне от 0 до 100. Обязательные поля —
`block_type`, `source`, `raw_score`, `normalized_score`,
`confidence_before` и `confidence_after`.

```bash
curl -X POST "$BASE_URL/projects/<project_id>/stages/<stage_id>/tasks/<task_id>/assessments" \
  -H "$AUTH" -H "Content-Type: application/json" \
  -d '{
    "block_type": "practice",
    "source": "manual_confirmation",
    "raw_score": 78,
    "normalized_score": 75,
    "confidence_before": 62,
    "confidence_after": 71
  }'
```

Для чтения данных используйте:

- `GET /projects/{id}/dashboard` — показатели проекта;
- `GET /projects/{project_id}/stages/{stage_id}/board` — доска этапа;
- `GET /projects/{project_id}/stages/{stage_id}/tasks/{task_id}/readiness` —
  готовность задачи.

Пересчёт запускается отдельными `POST`-методами:

- `/projects/{id}/readiness/recalculate`;
- `/projects/{project_id}/stages/{stage_id}/readiness/recalculate`;
- `/projects/{project_id}/stages/{stage_id}/tasks/{task_id}/readiness/recalculate`.

### Библиотеки и файлы

Основные методы хранилища:

| Действие | Метод |
|---|---|
| Создать библиотеку | `POST /storage/libraries` |
| Список библиотек | `GET /storage/libraries` |
| Дерево библиотеки | `GET /storage/libraries/{id}/tree` |
| Создать папку | `POST /storage/folders` |
| Загрузить небольшой файл | `POST /storage/files/upload` |
| Начать multipart-загрузку | `POST /storage/files/init` |
| Получить файл | `GET /storage/files/{id}/download` |
| Удалить файл | `DELETE /storage/files/{id}` |

Для multipart-загрузки порядок такой: `init` → получение URL частей через
`/storage/files/{id}/parts` → загрузка частей в объектное хранилище →
`complete`. Незавершённую загрузку отменяйте через `abort`.

Для файла сохраняются имя, размер и MIME-тип как метаданные. API не требует
заранее заданный список MIME-типов: если клиент не передал тип, используется
`application/octet-stream`. Имя файла может иметь до 255 символов и не должно
содержать управляющие символы или разделители пути `/` и `\`; полные имена `.`
и `..` запрещены.
Максимальный размер файла — 500 МБ.

### Управление ключами

| Действие | Метод |
|---|---|
| Создать ключ | `POST /access/api-keys` |
| Список ключей без секретов | `GET /access/api-keys` |
| Отозвать ключ | `DELETE /access/api-keys/{id}` |
| Удалить отозванный ключ | `DELETE /access/api-keys/{id}/permanent` |
| Ротировать ключ | `POST /access/api-keys/{id}/rotate` |

Ротация возвращает новый секрет. Сохраните его до окончания переходного
периода и замените старый секрет в интеграции.

## Ответы и ошибки

Успешные ответы используют стандартные коды HTTP: `200` для чтения и
изменения, `201` для создания, `204` для удаления без тела. Ошибка обычно
имеет вид:

```json
{
  "error": {
    "kind": "validation",
    "code": "…",
    "message": "…"
  }
}
```

Основные коды:

| HTTP | Значение | Действие |
|---:|---|---|
| `400` | Некорректный запрос | Проверьте JSON, обязательные поля и диапазоны |
| `401` | Нет авторизации или ключ недействителен | Обновите JWT или замените ключ |
| `403` | Недостаточно области доступа | Добавьте нужный `domain:read` или `domain:write` |
| `404` | Объект не найден | Проверьте UUID и владельца объекта |
| `409` | Конфликт состояния | Получите объект заново и повторите действие |
| `413` | Превышен лимит | Уменьшите запрос или используйте multipart-загрузку |
| `500` | Ошибка сервиса | Повторите запрос позже |

Не повторяйте автоматически необратимые операции без проверки результата.
Для повторной отправки оценки или создания объекта сначала убедитесь, что
предыдущий запрос не был принят.
