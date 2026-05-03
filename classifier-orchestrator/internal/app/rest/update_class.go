package rest

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"classifier-orchestrator/internal/usecase/update_class"
	"classifier-orchestrator/internal/usecase/update_class/dto"

	"github.com/go-chi/chi/v5"
)

type UpdateClassHandler struct {
	logger  *slog.Logger
	usecase update_class.UseCase
}

func NewUpdateClassHandler(logger *slog.Logger, usecase update_class.UseCase) http.Handler {
	return &UpdateClassHandler{
		logger:  logger,
		usecase: usecase,
	}
}

func (h *UpdateClassHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	classIDRaw := chi.URLParam(r, "classId")
	classID, err := strconv.Atoi(classIDRaw)
	if err != nil || classID <= 0 {
		http.Error(w, "Invalid classId", http.StatusBadRequest)
		return
	}

	var req dto.UpdateClassRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request body", slog.String("error", err.Error()))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if len(req.PositiveGoldens) == 0 {
		http.Error(w, "Positive goldens are required", http.StatusBadRequest)
		return
	}

	req.ID = int32(classID)

	resp, err := h.usecase.Execute(context.WithoutCancel(r.Context()), req)
	if err != nil {
		if errors.Is(err, update_class.ErrClassNotFound) {
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
