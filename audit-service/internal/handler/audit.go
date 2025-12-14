package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	// REMOVE: "github.com/gorilla/mux" - not used
	"audit-service/internal/model"
	"audit-service/internal/service"
)

type AuditHandler struct {
	service service.AuditService
}

func NewAuditHandler(s service.AuditService) *AuditHandler {
	return &AuditHandler{service: s}
}

func (h *AuditHandler) StoreEvent(w http.ResponseWriter, r *http.Request) {
	var event model.AuditEvent

	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON format")
		return
	}

	// Валидация обязательных полей
	if event.User == "" {
		respondWithError(w, http.StatusBadRequest, "Field 'user' is required")
		return
	}
	if event.Operation == "" {
		respondWithError(w, http.StatusBadRequest, "Field 'op' is required")
		return
	}

	storedEvent, err := h.service.StoreEvent(r.Context(), &event)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to store event")
		return
	}

	respondWithJSON(w, http.StatusCreated, storedEvent)
}

func (h *AuditHandler) FindEvents(w http.ResponseWriter, r *http.Request) {
	filters := parseQueryFilters(r.URL.Query())

	events, err := h.service.FindEvents(r.Context(), filters)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve events")
		return
	}

	respondWithJSON(w, http.StatusOK, events)
}

func parseQueryFilters(query map[string][]string) model.EventFilters {
	var filters model.EventFilters

	// Helper function to get first value from query parameter
	getFirst := func(key string) string {
		if values, ok := query[key]; ok && len(values) > 0 {
			return values[0]
		}
		return ""
	}

	// Парсинг временных меток
	parseTime := func(value string) *time.Time {
		if value == "" {
			return nil
		}
		formats := []string{
			time.RFC3339,
			"2006-01-02T15:04:05",
			"2006-01-02T15:04:05.999999",
		}

		for _, format := range formats {
			if t, err := time.Parse(format, value); err == nil {
				return &t
			}
		}
		return nil
	}

	// FIXED: Use getFirst instead of query.Get()
	if ts := getFirst("ev_ts"); ts != "" {
		filters.Timestamp = parseTime(ts)
	}
	if tsStart := getFirst("ev_ts_start"); tsStart != "" {
		filters.TimestampStart = parseTime(tsStart)
	}
	if tsEnd := getFirst("ev_ts_end"); tsEnd != "" {
		filters.TimestampEnd = parseTime(tsEnd)
	}

	// Парсинг списков
	getList := func(key string) []string {
		if values, ok := query[key]; ok && len(values) > 0 {
			return strings.Split(values[0], ",")
		}
		return nil
	}

	filters.Users = getList("ev_user")
	filters.Components = getList("ev_component")
	filters.Operations = getList("ev_op")

	// Парсинг числовых списков
	parseInt64List := func(key string) []int64 {
		strList := getList(key)
		if strList == nil {
			return nil
		}

		var intList []int64
		for _, str := range strList {
			if val, err := strconv.ParseInt(str, 10, 64); err == nil {
				intList = append(intList, val)
			}
		}
		return intList
	}

	filters.SessionIDs = parseInt64List("ev_session_id")
	filters.RequestIDs = parseInt64List("ev_req_id")

	// Парсинг пользовательских атрибутов
	filters.Attributes = make(map[string][]string)
	for key, values := range query {
		if !strings.HasPrefix(key, "ev_") && key != "ev_ts" && key != "ev_ts_start" && key != "ev_ts_end" {
			if len(values) > 0 {
				filters.Attributes[key] = strings.Split(values[0], ",")
			}
		}
	}

	return filters
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}
