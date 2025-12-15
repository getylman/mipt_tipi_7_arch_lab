-- Инициализация ТЕСТОВОЙ БД (отдельная от продовой)
CREATE TABLE IF NOT EXISTS audit_events (
    id BIGSERIAL PRIMARY KEY,
    timestamp TIMESTAMP NOT NULL,
    user_id TEXT NOT NULL,
    component TEXT,
    operation TEXT NOT NULL,
    session_id BIGINT,
    request_id BIGINT,
    response JSONB,
    attributes JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Дополнительные таблицы для тестов (если нужно)
CREATE TABLE IF NOT EXISTS test_metadata (
    test_name TEXT PRIMARY KEY,
    last_run TIMESTAMP,
    events_generated INT DEFAULT 0
);

-- Индексы для тестовой БД
CREATE INDEX IF NOT EXISTS idx_test_user_ts ON audit_events(user_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_test_timestamp ON audit_events(timestamp);
