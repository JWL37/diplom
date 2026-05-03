package rest

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"classifier-orchestrator/internal/usecase/list_classes"
)

type ListClassesHandler struct {
	logger  *slog.Logger
	usecase list_classes.UseCase
}

func NewListClassesHandler(logger *slog.Logger, usecase list_classes.UseCase) http.Handler {
	return &ListClassesHandler{
		logger:  logger,
		usecase: usecase,
	}
}

func (h *ListClassesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resp, err := h.usecase.Execute(context.WithoutCancel(r.Context()))
	if err != nil {
		h.logger.Error("Failed to execute usecase", slog.String("error", err.Error()))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
