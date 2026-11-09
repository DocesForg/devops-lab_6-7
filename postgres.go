package main

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// PostgresRepository stores employees in PostgreSQL.
type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) (*PostgresRepository, error) {
	if db == nil {
		return nil, errors.New("postgres database connection is nil")
	}
	repo := &PostgresRepository{db: db}
	if err := repo.createSchema(); err != nil {
		return nil, fmt.Errorf("create employees table: %w", err)
	}
	return repo, nil
}

func (r *PostgresRepository) createSchema() error {
	const query = `
CREATE TABLE IF NOT EXISTS employees (
    id BIGSERIAL PRIMARY KEY,
    full_name TEXT NOT NULL,
    position TEXT NOT NULL,
    salary DOUBLE PRECISION NOT NULL CHECK (salary >= 0),
    hired_on BIGINT NOT NULL
)`
	_, err := r.db.Exec(query)
	return err
}

func (r *PostgresRepository) Create(e Employee) (string, error) {
	if strings.TrimSpace(e.FullName) == "" || strings.TrimSpace(e.Position) == "" || e.Salary < 0 {
		return "", ErrInvalid
	}
	if e.HiredOn == 0 {
		e.HiredOn = time.Now().Unix()
	}
	var id int64
	err := r.db.QueryRow(
		`INSERT INTO employees (full_name, position, salary, hired_on)
		 VALUES ($1, $2, $3, $4) RETURNING id`,
		e.FullName, e.Position, e.Salary, e.HiredOn,
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("insert employee: %w", err)
	}
	return fmt.Sprint(id), nil
}

func (r *PostgresRepository) GetByID(id string) (Employee, error) {
	var e Employee
	err := r.db.QueryRow(
		`SELECT id, full_name, position, salary, hired_on FROM employees WHERE id = $1`, id,
	).Scan(&e.ID, &e.FullName, &e.Position, &e.Salary, &e.HiredOn)
	if errors.Is(err, sql.ErrNoRows) {
		return Employee{}, ErrNotFound
	}
	if err != nil {
		return Employee{}, fmt.Errorf("select employee: %w", err)
	}
	return e, nil
}

func (r *PostgresRepository) ChangePosition(id, position string) error {
	result, err := r.db.Exec(`UPDATE employees SET position = $1 WHERE id = $2`, position, id)
	if err != nil {
		return fmt.Errorf("update employee position: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read update result: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}
