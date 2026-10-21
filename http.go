package main

import (
	"encoding/json"
	"bytes"
	"context"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-kit/kit/endpoint"
)

type API struct {
	createEndpoint        endpoint.Endpoint
	getByIDEndpoint       endpoint.Endpoint
	changePositionEndpoint endpoint.Endpoint
	logger                *log.Logger
}

func NewHTTPHandler(service EmployeeService, logger *log.Logger) http.Handler {
	api := &API{
		createEndpoint: makeCreateEndpoint(service),
		getByIDEndpoint: makeGetByIDEndpoint(service),
		changePositionEndpoint: makeChangePositionEndpoint(service),
		logger: logger,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.Trim(r.URL.Path, "/")
		switch {
		case path == "healthz" && r.Method == http.MethodGet:
			writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		case path == "employees" && r.Method == http.MethodPost:
			api.create(w, r)
		case strings.HasPrefix(path, "employees/") && r.Method == http.MethodGet:
			r = withEmployeeID(r, strings.TrimPrefix(path, "employees/"))
			api.getByID(w, r)
		case strings.HasPrefix(path, "employees/") && strings.HasSuffix(path, "/position") && r.Method == http.MethodPatch:
			id := strings.TrimSuffix(strings.TrimPrefix(path, "employees/"), "/position")
			r = withEmployeeID(r, id)
			api.changePosition(w, r)
		default:
			writeError(w, http.StatusNotFound, "route not found")
		}
	})
	return requestLogging(logger, mux)
}

func (a *API) create(w http.ResponseWriter, r *http.Request) {
	var e Employee
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&e); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	response, err := a.createEndpoint(r.Context(), e)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, response)
}

func (a *API) getByID(w http.ResponseWriter, r *http.Request) {
	response, err := a.getByIDEndpoint(r.Context(), map[string]interface{}{"id": employeeID(r)})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (a *API) changePosition(w http.ResponseWriter, r *http.Request) {
	var body struct { Position string `json:"position"` }
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	response, err := a.changePositionEndpoint(r.Context(), map[string]interface{}{
		"id": employeeID(r), "position": body.Position,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, response)
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
	var payload bytes.Buffer
	if err := json.NewEncoder(&payload).Encode(value); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if _, err := w.Write(payload.Bytes()); err != nil {
		log.Printf("failed to write HTTP response: %v", err)
	}
}

type employeeIDKey struct{}

func withEmployeeID(r *http.Request, id string) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), employeeIDKey{}, id))
}

func employeeID(r *http.Request) string {
	id, _ := r.Context().Value(employeeIDKey{}).(string)
	return id
}

func requestLogging(logger *log.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		if logger != nil { logger.Printf("method=%s path=%s duration=%s", r.Method, strings.TrimSpace(r.URL.Path), time.Since(start)) }
	})
}

