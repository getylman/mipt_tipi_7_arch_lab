-- +goose Up
-- CREATE TABLE audit_events (
--     id BIGSERIAL PRIMARY KEY,
--     timestamp TIMESTAMP NOT NULL,
--     user_id TEXT NOT NULL,
--     component TEXT,
--     operation TEXT NOT NULL,
--     session_id BIGINT,
--     request_id BIGINT,
--     response JSONB,
--     attributes JSONB,
--     created_at TIMESTAMP DEFAULT NOW()
-- );

-- -- Индексы для тестов
-- CREATE INDEX idx_test_timestamp ON audit_events(timestamp);
-- CREATE INDEX idx_test_user ON audit_events(user_id);

-- +goose Down