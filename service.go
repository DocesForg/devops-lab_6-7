package main

import (
	"context"
	"strings"
)

type EmployeeService interface {
	Create(context.Context, Employee) (string, error)
	GetByID(context.Context, string) (Employee, error)
	ChangePosition(context.Context, string, string) error
}

type employeeService struct {
	repo EmployeeRepository
}

func NewEmployeeService(repo EmployeeRepository) EmployeeService {
	return &employeeService{repo: repo}
}

func (s *employeeService) Create(_ context.Context, e Employee) (string, error) {
	e.FullName = strings.TrimSpace(e.FullName)
	e.Position = strings.TrimSpace(e.Position)
	return s.repo.Create(e)
}

func (s *employeeService) GetByID(_ context.Context, id string) (Employee, error) {
	if strings.TrimSpace(id) == "" { return Employee{}, ErrInvalid }
	return s.repo.GetByID(id)
}

func (s *employeeService) ChangePosition(_ context.Context, id, position string) error {
	if strings.TrimSpace(id) == "" || strings.TrimSpace(position) == "" { return ErrInvalid }
	return s.repo.ChangePosition(id, strings.TrimSpace(position))
}
