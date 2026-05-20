package rest

import (
	"log/slog"
	"net/http"

	"classifier-orchestrator/internal/usecase/create_class"
	"classifier-orchestrator/internal/usecase/list_classes"
	"classifier-orchestrator/internal/usecase/update_class"
	"classifier-orchestrator/internal/usecase/update_class_status"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(logger *slog.Logger, createClassUseCase create_class.UseCase, listClassesUseCase list_classes.UseCase, updateClassUseCase update_class.UseCase, updateClassStatusUseCase update_class_status.UseCase) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/ping", HandlePing(logger))
	r.Get("/swagger", HandleSwaggerUI())
	r.Get("/swagger/", HandleSwaggerUI())
	r.Get("/swagger/openapi.yaml", HandleOpenAPI())

	r.Method(http.MethodPost, "/api/v1/classes", NewCreateClassHandler(logger, createClassUseCase))
	r.Method(http.MethodGet, "/api/v1/classes", NewListClassesHandler(logger, listClassesUseCase))
	r.Method(http.MethodPatch, "/api/v1/classes/{classId}", NewUpdateClassHandler(logger, updateClassUseCase))
	r.Method(http.MethodPatch, "/api/v1/classes/{classId}/status", NewUpdateClassStatusHandler(logger, updateClassStatusUseCase))

	return r
}
