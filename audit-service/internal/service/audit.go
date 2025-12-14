package service

import (
    "context"
    "fmt"
    "time"

    "audit-service/internal/model"
    "audit-service/internal/repository"
)

type AuditService interface {
    StoreEvent(ctx context.Context, event *model.AuditEvent) (*model.AuditEvent, error)
    FindEvents(ctx context.Context, filters model.EventFilters) ([]*model.AuditEvent, error)
}

type auditService struct {
    repo repository.AuditRepository
}

func NewAuditService(repo repository.AuditRepository) AuditService {
    return &auditService{repo: repo}
}

func (s *auditService) StoreEvent(ctx context.Context, event *model.AuditEvent) (*model.AuditEvent, error) {
    // Валидация временной метки
    if event.Timestamp.IsZero() {
        event.Timestamp = time.Now().UTC()
    }
    
    // Ограничение на будущие даты
    if event.Timestamp.After(time.Now().Add(5 * time.Minute)) {
        return nil, fmt.Errorf("timestamp cannot be more than 5 minutes in the future")
    }
    
    // Базовая валидация
    if len(event.User) > 255 {
        return nil, fmt.Errorf("user field too long")
    }
    if len(event.Operation) > 100 {
        return nil, fmt.Errorf("operation field too long")
    }
    
    return s.repo.StoreEvent(ctx, event)
}

func (s *auditService) FindEvents(ctx context.Context, filters model.EventFilters) ([]*model.AuditEvent, error) {
    // Валидация временных диапазонов
    if filters.TimestampStart != nil && filters.TimestampEnd != nil {
        if filters.TimestampStart.After(*filters.TimestampEnd) {
            return nil, fmt.Errorf("timestamp_start cannot be after timestamp_end")
        }
        
        // Ограничение диапазона 30 дней для производительности
        if filters.TimestampEnd.Sub(*filters.TimestampStart) > 30*24*time.Hour {
            return nil, fmt.Errorf("date range cannot exceed 30 days")
        }
    }
    
    return s.repo.FindEvents(ctx, filters)
}