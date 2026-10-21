package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEmployeeLifecycle(t *testing.T) {
	svc := NewEmployeeService(NewMemoryRepository())
	id, err := svc.Create(context.Background(), Employee{FullName: "Иван Иванов", Position: "Инженер", Salary: 80000})
	if err != nil { t.Fatal(err) }
	e, err := svc.GetByID(context.Background(), id)
	if err != nil { t.Fatal(err) }
	if e.FullName != "Иван Иванов" || e.Position != "Инженер" { t.Fatalf("unexpected employee: %+v", e) }
	if err := svc.ChangePosition(context.Background(), id, "Ведущий инженер"); err != nil { t.Fatal(err) }
	e, err = svc.GetByID(context.Background(), id)
	if err != nil { t.Fatal(err) }
	if e.Position != "Ведущий инженер" { t.Fatalf("position was not changed: %q", e.Position) }
}

func TestGetMissingEmployee(t *testing.T) {
	svc := NewEmployeeService(NewMemoryRepository())
	_, err := svc.GetByID(context.Background(), "missing")
	if err != ErrNotFound { t.Fatalf("expected ErrNotFound, got %v", err) }
}

func TestHTTPCreateAndGet(t *testing.T) {
	handler := NewHTTPHandler(NewEmployeeService(NewMemoryRepository()), nil)
	body := []byte(`{"full_name":"Анна Смирнова","position":"Аналитик","salary":65000}`)
	create := httptest.NewRecorder()
	handler.ServeHTTP(create, httptest.NewRequest(http.MethodPost, "/employees", bytes.NewReader(body)))
	if create.Code != http.StatusCreated { t.Fatalf("create status=%d body=%s", create.Code, create.Body.String()) }
	var result map[string]string
	if err := json.Unmarshal(create.Body.Bytes(), &result); err != nil { t.Fatal(err) }
	get := httptest.NewRecorder()
	handler.ServeHTTP(get, httptest.NewRequest(http.MethodGet, "/employees/"+result["id"], nil))
	if get.Code != http.StatusOK { t.Fatalf("get status=%d body=%s", get.Code, get.Body.String()) }
}
