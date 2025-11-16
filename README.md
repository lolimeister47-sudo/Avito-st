

---

# PR Reviewer Assignment Service

Тестовое задание на стажёра Backend (осенняя волна 2025)

Этот сервис автоматически назначает ревьюеров на Pull Request'ы, позволяет управлять командами и пользователями, поддерживает переназначение ревьюеров и идемпотентный merge. Проект реализован в стиле Clean Architecture и полностью соответствует OpenAPI-спецификации из задания.

---

## Возможности сервиса

### Команды

* Создание команды с участниками
* Получение команды
* Изменение активности пользователя

### Pull Request'ы

* Создание PR с автоматическим назначением до двух активных ревьюеров
* Переназначение ревьювера на другого активного участника команды
* Merge PR (идемпотентный)
* Получение PR’ов, где пользователь является ревьювером

### Бизнес-правила

* Ревьюеры выбираются только из команды автора
* Автор не может быть ревьювером
* Неактивные пользователи не назначаются
* Если из команды доступен только один кандидат — назначается один. Если ноль — никто.
* После MERGED изменять ревьюверов нельзя
* Переназначение возможно только если старый ревьювер действительно назначен
* Переназначение выбирает случайного активного кандидата из той же команды

---

## Архитектура проекта

```
/cmd/pr-service         – входная точка приложения
/internal
    /app                – запуск приложения, wiring зависимостей
    /config             – работа с переменными окружения
    /domain             – сущности и доменные ошибки
    /usecase            – бизнес-логика
    /adapter
        /repo/postgres  – репозитории для PostgreSQL
        /http           – HTTP сервер, роутер, OpenAPI-обработчики
    /db/migrations      – SQL миграции
```

Построено по принципам Clean Architecture:

* Domain — «чистые» сущности и доменные ошибки
* Usecase — бизнес-логика без зависимостей от инфраструктуры
* Adapter — Postgres и HTTP реализация
* App — объединяет все слои и запускает HTTP-сервер

---

## Запуск через Docker

Убедитесь, что файл `docker-compose.yml` и миграция `internal/db/migrations/001_init.sql` находятся на месте.

Для запуска:

```
make up
```

или вручную:

```
docker-compose up --build
```

Миграции применяются автоматически через `docker-entrypoint-initdb.d`.

Сервис будет доступен по адресу:

```
http://localhost:8080
```

---

## Переменные окружения

Файл `.env`:

```
DB_DSN=postgres://pr_service:pr_service@db:5432/pr_service?sslmode=disable
HTTP_ADDR=:8080
```

Если `.env` отсутствует — используется конфигурация по умолчанию.

---

## Запуск локально (без Docker)

```
make build
./bin/pr-service
```

---

## Примеры API запросов

### Healthcheck

```
GET /health
```

### Создание команды

```
POST /team/add
{
  "team_name": "backend",
  "members": [
    {"user_id":"u1","username":"Alice","is_active":true},
    {"user_id":"u2","username":"Bob","is_active":true},
    {"user_id":"u3","username":"Charlie","is_active":true}
  ]
}
```

### Получение команды

```
GET /team/get?team_name=backend
```

### Деактивация пользователя

```
POST /users/setIsActive
{
  "user_id": "u3",
  "is_active": false
}
```

### Создание PR

```
POST /pullRequest/create
{
  "pull_request_id": "pr-1",
  "pull_request_name": "Add login",
  "author_id": "u1"
}
```

### Merge PR (идемпотентно)

```
POST /pullRequest/merge
{
  "pull_request_id": "pr-1"
}
```

### Переназначение ревьювера

```
POST /pullRequest/reassign
{
  "pull_request_id": "pr-1",
  "old_user_id": "u2"
}
```

### Получение PR’ов пользователя

```
GET /users/getReview?user_id=u2
```

---

## Postman Collection

Файл:
`prservice-postman-collection.json`

Содержит полный набор запросов.

---


Просто скажи.
