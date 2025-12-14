package model

import (
    "database/sql/driver"
    "encoding/json"
    "errors"
    "time"
)

type AuditEvent struct {
    ID         int64           `json:"id" db:"id"`
    Timestamp  time.Time       `json:"timestamp" db:"timestamp"`
    User       string          `json:"user" db:"user_id"`
    Component  *string         `json:"component,omitempty" db:"component"`
    Operation  string          `json:"op" db:"operation"`
    SessionID  *int64          `json:"session_id,omitempty" db:"session_id"`
    RequestID  *int64          `json:"req_id,omitempty" db:"request_id"`
    Response   *JSONB          `json:"res,omitempty" db:"response"`
    Attributes *JSONB          `json:"attributes,omitempty" db:"attributes"`
    CreatedAt  time.Time       `json:"created_at" db:"created_at"`
}

type EventFilters struct {
    Timestamp     *time.Time         `json:"ev_ts,omitempty"`
    TimestampStart *time.Time        `json:"ev_ts_start,omitempty"`
    TimestampEnd  *time.Time         `json:"ev_ts_end,omitempty"`
    Users         []string           `json:"ev_user,omitempty"`
    Components    []string           `json:"ev_component,omitempty"`
    Operations    []string           `json:"ev_op,omitempty"`
    SessionIDs    []int64            `json:"ev_session_id,omitempty"`
    RequestIDs    []int64            `json:"ev_req_id,omitempty"`
    Attributes    map[string][]string `json:"-"`
}

type JSONB map[string]interface{}

func (j *JSONB) Value() (driver.Value, error) {
    if j == nil {
        return nil, nil
    }
    return json.Marshal(j)
}

func (j *JSONB) Scan(value interface{}) error {
    if value == nil {
        *j = nil
        return nil
    }
    
    b, ok := value.([]byte)
    if !ok {
        return errors.New("type assertion to []byte failed")
    }
    
    if len(b) == 0 {
        *j = nil
        return nil
    }
    
    return json.Unmarshal(b, j)
}