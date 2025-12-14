-- Создание таблицы audit_events
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

-- Индексы для ускорения фильтрации
CREATE INDEX IF NOT EXISTS idx_audit_events_timestamp ON audit_events(timestamp);
CREATE INDEX IF NOT EXISTS idx_audit_events_user_id ON audit_events(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_events_component ON audit_events(component);
CREATE INDEX IF NOT EXISTS idx_audit_events_operation ON audit_events(operation);
CREATE INDEX IF NOT EXISTS idx_audit_events_session_id ON audit_events(session_id);
CREATE INDEX IF NOT EXISTS idx_audit_events_request_id ON audit_events(request_id);
CREATE INDEX IF NOT EXISTS idx_audit_events_created_at ON audit_events(created_at);

-- GIN индекс для эффективного поиска по JSONB полям
CREATE INDEX IF NOT EXISTS idx_audit_events_attributes ON audit_events USING GIN (attributes);
CREATE INDEX IF NOT EXISTS idx_audit_events_response ON audit_events USING GIN (response);

-- Индекс для комбинированных запросов
CREATE INDEX IF NOT EXISTS idx_audit_events_user_timestamp 
ON audit_events(user_id, timestamp DESC);