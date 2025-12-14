package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"audit-service/internal/model"

	"github.com/lib/pq"
)

type AuditRepository interface {
	StoreEvent(ctx context.Context, event *model.AuditEvent) (*model.AuditEvent, error)
	FindEvents(ctx context.Context, filters model.EventFilters) ([]*model.AuditEvent, error)
}

type postgresRepository struct {
	db *sql.DB
}

func NewAuditRepository(db *sql.DB) AuditRepository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) StoreEvent(ctx context.Context, event *model.AuditEvent) (*model.AuditEvent, error) {
	query := `
        INSERT INTO audit_events 
        (timestamp, user_id, component, operation, session_id, request_id, response, attributes)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id, created_at
    `

	err := r.db.QueryRowContext(ctx, query,
		event.Timestamp,
		event.User,
		event.Component,
		event.Operation,
		event.SessionID,
		event.RequestID,
		event.Response,
		event.Attributes,
	).Scan(&event.ID, &event.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to store audit event: %w", err)
	}

	return event, nil
}

func (r *postgresRepository) FindEvents(ctx context.Context, filters model.EventFilters) ([]*model.AuditEvent, error) {
	var conditions []string
	var args []interface{}
	argCounter := 1

	// Обработка временных фильтров
	if filters.Timestamp != nil {
		conditions = append(conditions, fmt.Sprintf("timestamp = $%d", argCounter))
		args = append(args, *filters.Timestamp)
		argCounter++
	} else {
		if filters.TimestampStart != nil {
			conditions = append(conditions, fmt.Sprintf("timestamp >= $%d", argCounter))
			args = append(args, *filters.TimestampStart)
			argCounter++
		}
		if filters.TimestampEnd != nil {
			conditions = append(conditions, fmt.Sprintf("timestamp <= $%d", argCounter))
			args = append(args, *filters.TimestampEnd)
			argCounter++
		}
	}

	// Обработка фильтров по спискам
	addListFilter := func(values []string, column string) {
		if len(values) > 0 {
			conditions = append(conditions, fmt.Sprintf("%s = ANY($%d)", column, argCounter))
			args = append(args, pq.Array(values))
			argCounter++
		}
	}

	addIntListFilter := func(values []int64, column string) {
		if len(values) > 0 {
			conditions = append(conditions, fmt.Sprintf("%s = ANY($%d)", column, argCounter))
			args = append(args, pq.Array(values))
			argCounter++
		}
	}

	addListFilter(filters.Users, "user_id")
	addListFilter(filters.Components, "component")
	addListFilter(filters.Operations, "operation")
	addIntListFilter(filters.SessionIDs, "session_id")
	addIntListFilter(filters.RequestIDs, "request_id")

	// Обработка фильтров по атрибутам JSONB
	for key, values := range filters.Attributes {
		if len(values) > 0 {
			jsonPath := fmt.Sprintf("attributes->>'%s'", key)
			conditions = append(conditions, fmt.Sprintf("%s = ANY($%d)", jsonPath, argCounter))
			args = append(args, pq.Array(values))
			argCounter++
		}
	}

	// Сборка запроса
	query := "SELECT id, timestamp, user_id, component, operation, session_id, request_id, response, attributes, created_at FROM audit_events"
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY timestamp DESC LIMIT 1000"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	var events []*model.AuditEvent
	for rows.Next() {
		var event model.AuditEvent
		err := rows.Scan(
			&event.ID,
			&event.Timestamp,
			&event.User,
			&event.Component,
			&event.Operation,
			&event.SessionID,
			&event.RequestID,
			&event.Response,
			&event.Attributes,
			&event.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, &event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return events, nil
}
