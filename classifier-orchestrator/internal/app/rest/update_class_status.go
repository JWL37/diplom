package rest

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"classifier-orchestrator/internal/usecase/update_class_status"
	"classifier-orchestrator/internal/usecase/update_class_status/dto"

	"github.com/go-chi/chi/v5"
)

type UpdateClassStatusHandler struct {
	logger  *slog.Logger
	usecase update_class_status.UseCase
}

func NewUpdateClassStatusHandler(logger *slog.Logger, usecase update_class_status.UseCase) http.Handler {
	return &UpdateClassStatusHandler{
		logger:  logger,
		usecase: usecase,
	}
}

func (h *UpdateClassStatusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	classIDRaw := chi.URLParam(r, "classId")
	classID, err := strconv.Atoi(classIDRaw)
	if err != nil || classID <= 0 {
		http.Error(w, "Invalid classId", http.StatusBadRequest)
		return
	}

	var req dto.UpdateClassStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request body", slog.String("error", err.Error()))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.Status != "DRAFT" && req.Status != "ACTIVE" {
		http.Error(w, "Invalid status", http.StatusBadRequest)
		return
	}

	req.ID = int32(classID)

	resp, err := h.usecase.Execute(context.WithoutCancel(r.Context()), req)
	if err != nil {
		if errors.Is(err, update_class_status.ErrClassNotFound) {
			http.Error(w, "Class not found", http.StatusNotFound)
			return
		}
		h.logger.Error("Failed to execute usecase", slog.String("error", err.Error()))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
