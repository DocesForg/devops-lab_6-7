package main

import (
	"errors"
	"strings"
	"sync"
	"time"
)

var (
	ErrNotFound = errors.New("employee not found")
	ErrInvalid  = errors.New("invalid employee")
)

type Employee struct {
	ID       string  `json:"id"`
	FullName string  `json:"full_name"`
	Position string  `json:"position"`
	Salary   float64 `json:"salary"`
	HiredOn  int64   `json:"hired_on"`
}

type EmployeeRepository interface {
	Create(Employee) (string, error)
	GetByID(string) (Employee, error)
	ChangePosition(string, string) error
}

type MemoryRepository struct {
	mu        sync.RWMutex
	employees map[string]Employee
	nextID    uint64
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{employees: make(map[string]Employee)}
}

func (r *MemoryRepository) Create(e Employee) (string, error) {
	if strings.TrimSpace(e.FullName) == "" || strings.TrimSpace(e.Position) == "" || e.Salary < 0 {
		return "", ErrInvalid
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nextID++
	e.ID = formatID(r.nextID)
	if e.HiredOn == 0 {
		e.HiredOn = time.Now().Unix()
	}
	r.employees[e.ID] = e
	return e.ID, nil
}

func (r *MemoryRepository) GetByID(id string) (Employee, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	e, ok := r.employees[id]
	if !ok {
		return Employee{}, ErrNotFound
	}
	return e, nil
}

func (r *MemoryRepository) ChangePosition(id, position string) error {
	if strings.TrimSpace(position) == "" {
		return ErrInvalid
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	e, ok := r.employees[id]
	if !ok {
		return ErrNotFound
	}
	e.Position = position
	r.employees[id] = e
	return nil
}

func formatID(n uint64) string {
	const digits = "0123456789"
	if n == 0 { return "0" }
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = digits[n%10]
		n /= 10
	}
	return string(buf[i:])
}
