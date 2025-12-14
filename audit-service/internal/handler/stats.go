package handler

import (
    "encoding/json"
    "net/http"
    "runtime"
    "sync/atomic"
    "time"
)

type StatsHandler struct {
    startTime    time.Time
    version      string
    totalRequests uint64
    totalErrors   uint64
    dbConnected   atomic.Bool
}

func NewStatsHandler(version string) *StatsHandler {
    return &StatsHandler{
        startTime: time.Now(),
        version:   version,
    }
}

type StatsResponse struct {
    Version      string    `json:"version"`
    Uptime       string    `json:"uptime"`
    Goroutines   int       `json:"goroutines"`
    TotalRequests uint64   `json:"total_requests"`
    TotalErrors   uint64   `json:"total_errors"`
    DBConnected   bool     `json:"db_connected"`
    Timestamp    time.Time `json:"timestamp"`
}

func (h *StatsHandler) Stats(w http.ResponseWriter, r *http.Request) {
    stats := StatsResponse{
        Version:      h.version,
        Uptime:       time.Since(h.startTime).String(),
        Goroutines:   runtime.NumGoroutine(),
        TotalRequests: atomic.LoadUint64(&h.totalRequests),
        TotalErrors:   atomic.LoadUint64(&h.totalErrors),
        DBConnected:   h.dbConnected.Load(),
        Timestamp:    time.Now().UTC(),
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(stats)
}

func (h *StatsHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
    if !h.dbConnected.Load() {
        http.Error(w, "Database not connected", http.StatusServiceUnavailable)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func (h *StatsHandler) IncrementRequests() {
    atomic.AddUint64(&h.totalRequests, 1)
}

func (h *StatsHandler) IncrementErrors() {
    atomic.AddUint64(&h.totalErrors, 1)
}

func (h *StatsHandler) SetDBConnected(connected bool) {
    h.dbConnected.Store(connected)
}

func (h *StatsHandler) Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        h.IncrementRequests()
        next.ServeHTTP(w, r)
    })
}