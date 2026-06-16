# HachiSocial

Backend-каркас закрытой социальной платформы на Go и PostgreSQL.

## Что реализовано

- единый Go-процесс;
- конфигурация через переменные окружения;
- JSON-логи в `out/logs`;
- PostgreSQL connection pool;
- UUIDv4;
- users domain, application, PostgreSQL repository и HTTP transport;
- health endpoints;
- миграции пользователей и audit log;
- Docker Compose для приложения и PostgreSQL.
- встроенная стартовая web-страница на `http://localhost:8080`.

Auth и Telegram-активация будут следующими feature. Поэтому users endpoints
в production защищены `DenyPrincipalProvider`. В локальном Compose создаётся
тестовый пользователь `@hachi`, а `APP_DEV_USER_ID` включает development
identity без небезопасного HTTP-заголовка.

## Локальный запуск

Требования:

- Go 1.25+;
- Docker с Docker Compose.

Запуск всего окружения:

```bash
docker compose -f deployments/compose.yaml up --build
```

Проверка:

```bash
curl http://localhost:8080/live
curl http://localhost:8080/ready
```

Запуск Go-приложения вне контейнера:

```bash
cp .env.example .env
go run ./cmd/app
```

Приложение читает переменные среды напрямую, поэтому `.env` нужно загрузить
средствами shell или IDE.

## Проверки

```bash
make generate
go test ./...
go vet ./...
```
