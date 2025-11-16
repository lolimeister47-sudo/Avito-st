-- Создание таблицы команд
CREATE TABLE IF NOT EXISTS teams (
    team_name TEXT PRIMARY KEY
);

-- Создание таблицы пользователей
CREATE TABLE IF NOT EXISTS users (
    user_id   TEXT PRIMARY KEY,
    username  TEXT NOT NULL,
    team_name TEXT NOT NULL REFERENCES teams(team_name) ON DELETE CASCADE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE
);

-- Статус PR хранится как текст, но ограничим допустимые значения
CREATE TABLE IF NOT EXISTS pull_requests (
    pull_request_id   TEXT PRIMARY KEY,
    pull_request_name TEXT NOT NULL,
    author_id         TEXT NOT NULL REFERENCES users(user_id),
    status            TEXT NOT NULL CHECK (status IN ('OPEN', 'MERGED')),
    created_at        TIMESTAMPTZ,
    merged_at         TIMESTAMPTZ
);

-- Назначенные ревьюверы (0..2 на PR)
CREATE TABLE IF NOT EXISTS pull_request_reviewers (
    pull_request_id TEXT NOT NULL REFERENCES pull_requests(pull_request_id) ON DELETE CASCADE,
    reviewer_id     TEXT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    PRIMARY KEY (pull_request_id, reviewer_id)
);
