package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"
)

type API struct {
	service EmployeeService
	logger  *log.Logger
}

func NewHTTPHandler(service EmployeeService, logger *log.Logger) http.Handler {
	api := &API{service: service, logger: logger}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /employees", api.create)
	mux.HandleFunc("GET /employees/{id}", api.getByID)
	mux.HandleFunc("PATCH /employees/{id}/position", api.changePosition)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	return requestLogging(logger, mux)
}

func (a *API) create(w http.ResponseWriter, r *http.Request) {
	var e Employee
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&e); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	id, err := a.service.Create(r.Context(), e)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (a *API) getByID(w http.ResponseWriter, r *http.Request) {
	e, err := a.service.GetByID(r.Context(), r.PathValue("id"))
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, e)
}

func (a *API) changePosition(w http.ResponseWriter, r *http.Request) {
	var body struct { Position string `json:"position"` }
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if err := a.service.ChangePosition(r.Context(), r.PathValue("id"), body.Position); err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func writeDomainError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrInvalid):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func requestLogging(logger *log.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		if logger != nil { logger.Printf("method=%s path=%s duration=%s", r.Method, strings.TrimSpace(r.URL.Path), time.Since(start)) }
	})
}

var _ = context.Background
