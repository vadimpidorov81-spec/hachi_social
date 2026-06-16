# HachiSocial Backend: ТЗ на фичу пользователей

Статус: **согласовано, первая итерация реализована**.

Принятые решения имеют приоритет над предварительными вариантами ниже:

- UUIDv4;
- username после нормализации соответствует `^[a-z]{1,20}$`;
- смена username не входит в MVP;
- `internal/domain/users` содержит domain, а
  `internal/feature/users/{transport,application,repository}` — три слоя feature;
- единый процесс запускается из `cmd/app`;
- логирование через `log/slog` в `out/logs`;
- OpenTelemetry и агрегация метрик не входят в MVP;
- auth и Telegram-активация реализуются следующими feature.

Документ описывает первую backend-фичу HachiSocial: доменную модель пользователя,
хранение профиля и HTTP API для работы с ним. Реализация начинается только после
согласования раздела «Решения для подтверждения».

## 1. Цель

Реализовать изолированный модуль `users`, который:

- создаёт пользователя через внутренний service-метод;
- хранит основную информацию и публичный профиль;
- возвращает текущего пользователя;
- возвращает публичный профиль другого пользователя;
- позволяет владельцу редактировать свой профиль;
- позволяет администратору менять статус пользователя;
- соблюдает принятые границы `transport -> service -> repository`;
- не зависит от HTTP, PostgreSQL и конкретного логгера в domain-слое.

## 2. Границы первой итерации

### Входит

- domain-сущность `User`;
- value objects и доменная валидация;
- статусы пользователя;
- роли платформы;
- создание пользователя внутренним use case;
- получение пользователя по ID и username;
- получение текущего профиля;
- редактирование собственного профиля;
- публичное чтение профиля;
- административная блокировка и разблокировка;
- PostgreSQL repository;
- миграции `golang-migrate`;
- HTTP transport;
- unit- и integration-тесты;
- логирование через `log/slog`;
- общие HTTP middleware, необходимые для запуска API.

### Не входит

- покупка доступа;
- Telegram-бот;
- коды активации;
- пароль и credentials;
- вход, выход и сессии;
- восстановление доступа;
- аватар и загрузка файлов;
- подписки на пользователей;
- блокировка одного пользователя другим;
- сообщества;
- посты и комментарии;
- дневник и привычки;
- полноценная административная панель.

Auth будет отдельной фичей. Он вызовет внутренний use case `users.Create` после
успешной активации и будет передавать идентификатор пользователя в request
context. До появления auth user-модуль тестируется через подставной
`PrincipalProvider`; небезопасный production-механизм авторизации через
произвольный HTTP-заголовок не реализуется.

## 3. Архитектурные правила

Фича располагается в `internal/domain/users` и `internal/feature/users`:

```text
internal/
├── domain/
│   └── users/
└── feature/
    └── users/
        ├── transport/
        │   └── http/
        ├── application/
        └── repository/
            └── postgres/
```

Направление зависимостей:

```text
transport -> service -> domain
                 |
                 v
       repository interfaces
                 ^
                 |
       postgres repository
```

Правила:

- domain использует только стандартную библиотеку;
- domain не логирует и не знает о JSON, SQL или HTTP;
- service работает с domain-типами и интерфейсами зависимостей;
- transport не обращается к repository напрямую;
- PostgreSQL repository не содержит бизнес-правила;
- типы `pgx` и `sqlc` не выходят из repository;
- dependency injection выполняется вручную в `cmd/app/main.go`;
- глобальные изменяемые зависимости не используются.

## 4. Domain-модель

### User

```go
type User struct {
    id          ID
    username    Username
    displayName DisplayName
    bio         Bio
    timezone    Timezone
    role        Role
    status      Status
    createdAt   time.Time
    updatedAt   time.Time
}
```

Поля изменяются только через конструктор и domain-методы.

Предполагаемые методы:

```go
func NewUser(
    id ID,
    username Username,
    displayName DisplayName,
    timezone Timezone,
    now time.Time,
) (*User, error)

func (u *User) UpdateProfile(
    displayName DisplayName,
    bio Bio,
    timezone Timezone,
    now time.Time,
) error

func (u *User) ChangeUsername(username Username, now time.Time) error
func (u *User) Block(now time.Time) error
func (u *User) Activate(now time.Time) error
func (u *User) ChangeRole(role Role, now time.Time) error
```

Repository получает отдельный способ восстановить валидную сущность из
сохранённого состояния. Он не должен использовать публичный конструктор нового
пользователя так, будто пользователь создаётся заново.

### ID

- внутренний ID: UUID;
- генерируется через интерфейс `IDGenerator`;
- transport принимает и возвращает строковое UUID-представление;
- используется UUIDv4.

### Username

Предварительные правила:

- длина от 1 до 20 символов;
- разрешены только латинские буквы;
- регистр при сравнении не учитывается;
- каноническое хранение в lowercase;
- username уникален;
- зарезервированные имена запрещены;
- username присутствует в URL публичного профиля.

Примеры:

- допустимо: `alex`, `a`;
- недопустимо: `_alex`, `alex99`, `алекс`, `alex-name`.

Зарезервированный минимум:

```text
admin
administrator
api
auth
bot
help
moderator
root
settings
support
system
users
```

### DisplayName

- после `strings.TrimSpace` длина от 1 до 50 Unicode-символов;
- переносы строк запрещены;
- управляющие символы запрещены;
- имя не обязано быть уникальным.

### Bio

- необязательное поле;
- максимум 500 Unicode-символов;
- пробелы по краям удаляются;
- обычный текст без HTML;
- сохранение переносов строк разрешено.

### Timezone

- IANA timezone, например `Europe/Moscow`;
- значение проверяется через `time.LoadLocation`;
- значение по умолчанию передаётся явно при создании пользователя;
- timezone необходим для будущего расчёта дневника и стриков.

### Role

В первой версии:

```text
user
moderator
admin
```

Новый пользователь получает роль `user`. Изменять роль может только
администратор. Роли платформы не заменяют будущие роли внутри сообщества.

### Status

В первой версии:

```text
active
blocked
```

- новый пользователь создаётся со статусом `active`;
- `blocked` запрещает создание авторизованной сессии и выполнение защищённых
  действий;
- причина блокировки и администратор хранятся в moderation/audit, а не в `User`;
- физическое удаление пользователя этой итерацией не реализуется.

## 5. Валидация по слоям

### Transport

- корректный JSON;
- отсутствие неизвестных полей;
- лимит размера request body;
- корректный UUID в URL;
- наличие обязательных полей;
- строковый тип входных значений.

### Domain

- формат и длина username;
- зарезервированный username;
- длина display name и bio;
- корректность timezone;
- допустимые role и status;
- допустимые переходы состояния.

### Service

- уникальность username;
- существование пользователя;
- права на действие;
- запрет изменения заблокированного пользователя, где это требуется;
- защита от блокировки собственного admin-аккаунта.

### PostgreSQL

- `NOT NULL`;
- `UNIQUE` для нормализованного username;
- `CHECK` для role и status;
- ограничения длины как дополнительная защита;
- индексы для рабочих запросов.

## 6. Service use cases

### CreateUser

Внутренний use case, не публикуется как открытый HTTP endpoint.

Вход:

```go
type CreateUserCommand struct {
    Username    string
    DisplayName string
    Timezone    string
}
```

Алгоритм:

1. Создать domain value objects.
2. Проверить, что username не занят.
3. Сгенерировать ID.
4. Создать `User` с ролью `user` и статусом `active`.
5. Сохранить пользователя.
6. Вернуть безопасное представление пользователя.

Конфликт уникальности должен корректно обрабатываться и при race condition:
решающей гарантией остаётся уникальный индекс PostgreSQL.

### GetCurrentUser

- получает user ID из параметра service-метода;
- возвращает полное безопасное представление текущего пользователя;
- не возвращает credentials, Telegram ID и внутренние audit-данные.

### GetPublicProfile

- ищет пользователя по username;
- возвращает только публичные поля;
- заблокированный пользователь не доступен через публичный endpoint;
- отсутствие и блокировка внешне возвращаются одинаковым `404`, чтобы не
  раскрывать состояние аккаунта.

### UpdateProfile

Разрешено изменять:

- `display_name`;
- `bio`;
- `timezone`.

Требования:

- пользователь меняет только собственный профиль;
- частичное обновление выполняется через `PATCH`;
- переданное поле обновляется даже при пустом значении, если пустое значение
  допустимо;
- отсутствие поля означает «не изменять»;
- `updated_at` меняется при успешном обновлении.

### ChangeUsername

Условный use case, который реализуется только при подтверждении смены username:

- проверяется новый username;
- проверяется уникальность;
- обновление атомарно;
- смена username ограничивается отдельной политикой.

На первой итерации предлагается разрешить смену не чаще одного раза в 30 дней.
Для этого понадобится поле `username_changed_at`.

### SetUserStatus

Административный use case:

- принимает actor ID, target ID и новый status;
- actor должен иметь роль `admin`;
- нельзя заблокировать самого себя;
- повторная установка текущего статуса идемпотентна;
- операция должна создать запись в audit log.

### SetUserRole

Предлагается определить интерфейс и domain-метод, но не публиковать endpoint в
первой итерации. Первого администратора безопаснее назначать миграцией или
отдельной CLI-командой.

## 7. Service ports

Предварительные интерфейсы:

```go
type Repository interface {
    Create(ctx context.Context, user *domain.User) error
    GetByID(ctx context.Context, id domain.ID) (*domain.User, error)
    GetByUsername(ctx context.Context, username domain.Username) (*domain.User, error)
    UpdateProfile(ctx context.Context, user *domain.User) error
    UpdateUsername(ctx context.Context, user *domain.User) error
    UpdateStatus(ctx context.Context, user *domain.User) error
}

type IDGenerator interface {
    New() domain.ID
}

type Clock interface {
    Now() time.Time
}

type AuditWriter interface {
    UserStatusChanged(
        ctx context.Context,
        actorID domain.ID,
        targetID domain.ID,
        from domain.Status,
        to domain.Status,
    ) error
}
```

Интерфейсы принадлежат service-слою как потребителю. Перед реализацией методы
repository можно уточнить, но универсального generic CRUD repository не будет.

## 8. PostgreSQL

### Таблица `users`

Предварительная миграция:

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY,
    username VARCHAR(20) NOT NULL,
    display_name VARCHAR(50) NOT NULL,
    bio VARCHAR(500) NOT NULL DEFAULT '',
    timezone VARCHAR(64) NOT NULL,
    role VARCHAR(16) NOT NULL DEFAULT 'user',
    status VARCHAR(16) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,

    CONSTRAINT users_username_format_chk
        CHECK (username ~ '^[a-z]{1,20}$'),
    CONSTRAINT users_role_chk
        CHECK (role IN ('user', 'moderator', 'admin')),
    CONSTRAINT users_status_chk
        CHECK (status IN ('active', 'blocked'))
);

CREATE UNIQUE INDEX users_username_unique_idx
    ON users (username);

CREATE INDEX users_status_idx
    ON users (status);
```

Username хранится в lowercase, поэтому обычного уникального индекса достаточно.
Нормализация выполняется domain-объектом до записи.

### Миграции

```text
db/migrations/
├── 000001_create_users.up.sql
└── 000001_create_users.down.sql
```

Требования:

- миграции выполняются через `golang-migrate`;
- production-приложение не запускает миграции автоматически;
- up/down проверяются в CI на чистой PostgreSQL;
- миграции после публикации не редактируются, создаётся новая версия.

## 9. HTTP API

Базовый prefix: `/api/v1`.

### `GET /users/me`

Требует авторизованный контекст.

Ответ `200`:

```json
{
  "data": {
    "id": "019...",
    "username": "alex",
    "display_name": "Alex",
    "bio": "Backend developer",
    "timezone": "Europe/Moscow",
    "role": "user",
    "status": "active",
    "created_at": "2026-06-13T10:00:00Z",
    "updated_at": "2026-06-13T10:00:00Z"
  }
}
```

### `PATCH /users/me`

Тело:

```json
{
  "display_name": "Alex",
  "bio": "Go developer",
  "timezone": "Europe/Moscow"
}
```

Все поля опциональны. Пустой объект возвращает `400`.

Ответ: `200` с обновлённым представлением.

### `PUT /users/me/username`

Условный endpoint, который реализуется только при подтверждении смены username.

Тело:

```json
{
  "username": "new_username"
}
```

Ответ: `200` с обновлённым представлением.

### `GET /users/{username}`

Публичный endpoint внутри платформы.

Ответ `200`:

```json
{
  "data": {
    "id": "019...",
    "username": "alex",
    "display_name": "Alex",
    "bio": "Backend developer"
  }
}
```

Timezone, role, status и технические даты публично не возвращаются.

### `PUT /admin/users/{id}/status`

Требует роль `admin`.

Тело:

```json
{
  "status": "blocked"
}
```

Ответ: `204 No Content`.

## 10. Формат ошибок

```json
{
  "error": {
    "code": "username_already_taken",
    "message": "Username is already taken",
    "request_id": "019..."
  }
}
```

Минимальные коды:

| Code | HTTP |
|---|---:|
| `invalid_request` | 400 |
| `validation_failed` | 400 |
| `unauthorized` | 401 |
| `forbidden` | 403 |
| `user_not_found` | 404 |
| `username_already_taken` | 409 |
| `username_change_too_soon` | 409 |
| `invalid_user_status` | 422 |
| `internal_error` | 500 |

Domain и service возвращают типизированные ошибки. HTTP transport переводит их
в статус и публичный код. Внутренний текст ошибки и SQL не возвращаются клиенту.

## 11. Middleware первой итерации

Предлагаемый порядок:

```text
RequestID
-> Recoverer
-> AccessLog
-> SecurityHeaders
-> CORS
-> BodyLimit
-> Authentication
-> Authorization
-> Handler
```

Требования:

- `RequestID` принимает корректный входной ID или генерирует новый;
- `Recoverer` перехватывает panic, пишет stack trace и возвращает `500`;
- access log содержит method, route, status, duration и request ID;
- для известного пользователя добавляется user ID;
- body, cookies и authorization headers не логируются;
- health endpoints не создают шум в обычных access logs;
- middleware передают request-scoped logger через context.

`Authentication` предоставляет transport-слою согласованный `Principal` с user
ID и ролью. В unit- и HTTP-тестах используется подставная реализация. Реальный
разбор session cookie появится в auth-фиче.

Rate limiting можно добавить в auth-фиче, когда появятся чувствительные
публичные endpoints.

## 12. Logger

- библиотека: стандартный `log/slog`;
- encoder: JSON;
- вывод: stdout и файлы в `out/logs`;
- уровни настраиваются через config;
- базовые поля: `service`, `environment`, `version`;
- request-поля: `request_id`, `user_id`;
- service логирует начало только значимых операций и их ошибки;
- repository не логирует каждую успешную SQL-операцию;
- одна ошибка не должна логироваться на каждом слое;
- domain не зависит от logger;
- файл закрывается при завершении процесса.

## 13. Конфигурация и запуск

Минимальные переменные:

```text
APP_ENV
APP_VERSION
HTTP_ADDR
DATABASE_URL
LOG_LEVEL
LOG_DIR
```

API должен:

- проверять config при запуске;
- подключаться к PostgreSQL;
- настраивать pool `pgxpool`;
- собирать зависимости вручную;
- запускать HTTP server;
- корректно завершаться по сигналу;
- прекращать принимать новые запросы;
- ждать завершения активных запросов с timeout;
- закрывать PostgreSQL pool и logger.

## 14. Health endpoints

### `GET /live`

- подтверждает, что процесс работает;
- не обращается к PostgreSQL;
- возвращает `200`.

### `GET /ready`

- выполняет быстрый ping PostgreSQL;
- возвращает `200`, если API готов принимать запросы;
- возвращает `503`, если обязательная зависимость недоступна.

## 15. Тестирование

### Domain unit tests

- валидные и невалидные username;
- lowercase-нормализация;
- зарезервированные username;
- границы display name и bio;
- timezone;
- создание пользователя;
- обновление профиля;
- переходы `active <-> blocked`;
- повторная установка статуса;

### Service unit tests

- успешное создание;
- конфликт username;
- race conflict, возвращённый repository;
- получение текущего пользователя;
- публичный профиль active-пользователя;
- сокрытие blocked-пользователя;
- частичное обновление профиля;
- ограничение смены username;
- admin блокирует пользователя;
- non-admin получает forbidden;
- admin не может заблокировать себя;

### Repository integration tests

Запускаются с настоящей PostgreSQL:

- create/get/update;
- уникальность username;
- mapping database/domain;
- корректная работа ограничений;
- конкурентное создание одинакового username;
- миграция up/down.

### HTTP tests

- успешные ответы и JSON schema;
- неизвестные JSON-поля;
- некорректный body;
- отсутствие авторизации;
- role authorization;
- mapping service errors;
- request ID;
- panic recovery.

## 16. Нефункциональные требования

- все публичные методы принимают `context.Context`;
- SQL-запросы используют context и параметры;
- HTTP server имеет read/write/idle timeouts;
- request body ограничен;
- ошибки оборачиваются с контекстом через `%w`;
- sensitive data не попадает в логи;
- код проходит `gofmt`, `go vet`, `staticcheck` и настроенный linter;
- package-level test coverage не является самоцелью, но все бизнес-ветки
  должны быть проверены;
- документация API обновляется вместе с endpoint-ами.

## 17. Критерии приёмки

Фича считается готовой, когда:

1. Проект запускается локально с PostgreSQL через Docker Compose.
2. Миграции создают и откатывают таблицу `users`.
3. Внутренний service use case создаёт валидного пользователя.
4. Username нормализуется и защищён уникальным индексом.
5. Реализованы согласованные HTTP endpoint-ы.
6. Пользователь может читать и менять только собственный полный профиль.
7. Публичный endpoint не раскрывает закрытые поля и blocked-аккаунты.
8. Только admin может менять статус другого пользователя.
9. Ошибки имеют единый JSON-формат.
10. `slog`, request ID, panic recovery и access logging работают.
11. Domain, service и repository покрыты необходимыми тестами.
12. Integration-тесты проходят на настоящей PostgreSQL.
13. `go test ./...`, линтеры и `git diff --check` проходят без ошибок.

## 18. Принятые решения

1. Auth, пароли и Telegram-активация реализуются следующими feature.
2. Внутренние идентификаторы используют UUIDv4.
3. Username соответствует `^[a-z]{1,20}$` и хранится в lowercase.
4. Публичный профиль доступен только авторизованным участникам платформы.
5. Смена username не входит в MVP.
6. Роли: `user`, `moderator`, `admin`.
7. Статусы: `active`, `blocked`.
8. Профиль хранится в таблице `users`.
9. Endpoint административной смены статуса входит в первую итерацию.
10. Создание первого администратора будет отдельным эксплуатационным сценарием.
11. Изменение статуса записывается в таблицу `audit_log`.
12. OpenTelemetry и агрегация метрик не входят в MVP.
