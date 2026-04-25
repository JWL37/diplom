package rest

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"classifier-orchestrator/internal/usecase/create_class"
	"classifier-orchestrator/internal/usecase/create_class/dto"
)

type CreateClassHandler struct {
	logger  *slog.Logger
	usecase create_class.UseCase
}

func NewCreateClassHandler(logger *slog.Logger, usecase create_class.UseCase) http.Handler {
	return &CreateClassHandler{
		logger:  logger,
		usecase: usecase,
	}
}

func (h *CreateClassHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateClassRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request body", slog.String("error", err.Error()))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.usecase.Execute(context.WithoutCancel(r.Context()), req)
	if err != nil {
		h.logger.Error("Failed to execute usecase", slog.String("error", err.Error()))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}