package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	recurrence "example.com/taskservice/internal/domain/recurrence"
	taskrepo "example.com/taskservice/internal/usecase/task"
)

// RecurrenceHandler обрабатывает HTTP запросы для периодических задач и правил
type RecurrenceHandler struct {
	logger  *slog.Logger
	usecase *taskrepo.Service
}

// NewRecurrenceHandler создает новый handler
func NewRecurrenceHandler(logger *slog.Logger, usecase *taskrepo.Service) *RecurrenceHandler {
	return &RecurrenceHandler{
		logger:  logger,
		usecase: usecase,
	}
}

// CreateRecurrenceRule обрабатывает POST /api/recurrence-rules
func (h *RecurrenceHandler) CreateRecurrenceRule(w http.ResponseWriter, r *http.Request) {
	var req CreateRecurrenceRuleDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Преобразуем DTO в domain model
	recType, err := recurrence.FromString(req.RecurrenceType)
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid recurrence type: %v", err), http.StatusBadRequest)
		return
	}

	params := recurrence.RecurrenceParams{
		IntervalDays:  req.Params.IntervalDays,
		MonthDays:     req.Params.MonthDays,
		SpecificDates: req.Params.SpecificDates,
		StartDate:     req.Params.StartDate,
		EndDate:       req.Params.EndDate,
	}

	rule := &recurrence.Rule{
		Name:           req.Name,
		RecurrenceType: recType,
		Params:         params,
	}

	created, err := h.usecase.CreateRecurrenceRule(r.Context(), rule)
	if err != nil {
		h.logger.Error("create recurrence rule failed", "error", err)
		http.Error(w, fmt.Sprintf("failed to create recurrence rule: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newRecurrenceRuleDTO(created))
}

// GetRecurrenceRule обрабатывает GET /api/recurrence-rules/{id}
func (h *RecurrenceHandler) GetRecurrenceRule(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	if idStr == "" {
		http.Error(w, "missing recurrence rule id", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid recurrence rule id", http.StatusBadRequest)
		return
	}

	rule, err := h.usecase.GetRecurrenceRule(r.Context(), id)
	if err != nil {
		h.logger.Error("get recurrence rule failed", "id", id, "error", err)
		http.Error(w, fmt.Sprintf("recurrence rule not found: %v", err), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newRecurrenceRuleDTO(rule))
}

// ListRecurrenceRules обрабатывает GET /api/recurrence-rules
func (h *RecurrenceHandler) ListRecurrenceRules(w http.ResponseWriter, r *http.Request) {
	// Опциональный параметр для фильтрации по enabled статусу
	var enabled *bool
	enabledParam := r.URL.Query().Get("enabled")
	if enabledParam != "" {
		e := enabledParam == "true"
		enabled = &e
	}

	rules, err := h.usecase.ListRecurrenceRules(r.Context(), enabled)
	if err != nil {
		h.logger.Error("list recurrence rules failed", "error", err)
		http.Error(w, fmt.Sprintf("failed to list recurrence rules: %v", err), http.StatusInternalServerError)
		return
	}

	response := make([]RecurrenceRuleDTO, len(rules))
	for i, rule := range rules {
		response[i] = newRecurrenceRuleDTO(&rule)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateRecurrenceRule обрабатывает PUT /api/recurrence-rules/{id}
func (h *RecurrenceHandler) UpdateRecurrenceRule(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	if idStr == "" {
		http.Error(w, "missing recurrence rule id", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid recurrence rule id", http.StatusBadRequest)
		return
	}

	var req CreateRecurrenceRuleDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid request: %v", err), http.StatusBadRequest)
		return
	}

	recType, err := recurrence.FromString(req.RecurrenceType)
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid recurrence type: %v", err), http.StatusBadRequest)
		return
	}

	params := recurrence.RecurrenceParams{
		IntervalDays:  req.Params.IntervalDays,
		MonthDays:     req.Params.MonthDays,
		SpecificDates: req.Params.SpecificDates,
		StartDate:     req.Params.StartDate,
		EndDate:       req.Params.EndDate,
	}

	rule := &recurrence.Rule{
		ID:             id,
		Name:           req.Name,
		RecurrenceType: recType,
		Params:         params,
		IsEnabled:      true,
	}

	updated, err := h.usecase.UpdateRecurrenceRule(r.Context(), rule)
	if err != nil {
		h.logger.Error("update recurrence rule failed", "id", id, "error", err)
		http.Error(w, fmt.Sprintf("failed to update recurrence rule: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newRecurrenceRuleDTO(updated))
}

// DeleteRecurrenceRule обрабатывает DELETE /api/recurrence-rules/{id}
func (h *RecurrenceHandler) DeleteRecurrenceRule(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	if idStr == "" {
		http.Error(w, "missing recurrence rule id", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid recurrence rule id", http.StatusBadRequest)
		return
	}

	if err := h.usecase.DeleteRecurrenceRule(r.Context(), id); err != nil {
		h.logger.Error("delete recurrence rule failed", "id", id, "error", err)
		http.Error(w, fmt.Sprintf("failed to delete recurrence rule: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// CreateTaskWithRecurrence обрабатывает POST /api/tasks/with-recurrence
func (h *RecurrenceHandler) CreateTaskWithRecurrence(w http.ResponseWriter, r *http.Request) {
	var req CreateTaskWithRecurrenceDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Получаем правило (либо по ID, либо создаем новое)
	var rule *recurrence.Rule

	if req.RecurrenceRuleID != nil {
		// Получаем существующее правило
		var err error
		rule, err = h.usecase.GetRecurrenceRule(r.Context(), *req.RecurrenceRuleID)
		if err != nil {
			h.logger.Error("get recurrence rule failed", "id", *req.RecurrenceRuleID, "error", err)
			http.Error(w, fmt.Sprintf("recurrence rule not found: %v", err), http.StatusNotFound)
			return
		}
	} else if req.RecurrenceRuleData != nil {
		// Создаем новое правило
		recType, err := recurrence.FromString(req.RecurrenceRuleData.RecurrenceType)
		if err != nil {
			http.Error(w, fmt.Sprintf("invalid recurrence type: %v", err), http.StatusBadRequest)
			return
		}

		params := recurrence.RecurrenceParams{
			IntervalDays:  req.RecurrenceRuleData.Params.IntervalDays,
			MonthDays:     req.RecurrenceRuleData.Params.MonthDays,
			SpecificDates: req.RecurrenceRuleData.Params.SpecificDates,
			StartDate:     req.RecurrenceRuleData.Params.StartDate,
			EndDate:       req.RecurrenceRuleData.Params.EndDate,
		}

		rule = &recurrence.Rule{
			Name:           req.RecurrenceRuleData.Name,
			RecurrenceType: recType,
			Params:         params,
		}

		createdRule, err := h.usecase.CreateRecurrenceRule(r.Context(), rule)
		if err != nil {
			h.logger.Error("create recurrence rule failed", "error", err)
			http.Error(w, fmt.Sprintf("failed to create recurrence rule: %v", err), http.StatusInternalServerError)
			return
		}
		rule = createdRule
	} else {
		http.Error(w, "either recurrence_rule_id or recurrence_rule_data must be provided", http.StatusBadRequest)
		return
	}

	// Создаем задачу с периодичностью
	input := taskrepo.CreateWithRecurrenceInput{
		Title:         req.Title,
		Description:   req.Description,
		RecurrenceRule: rule,
	}

	task, err := h.usecase.CreateWithRecurrence(r.Context(), input)
	if err != nil {
		h.logger.Error("create task with recurrence failed", "error", err)
		http.Error(w, fmt.Sprintf("failed to create task with recurrence: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newTaskDTO(task))
}
