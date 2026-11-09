package main

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

func TestPostgresRepositoryCRUD(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to run PostgreSQL integration test")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		t.Fatalf("connect to PostgreSQL: %v", err)
	}

	repo, err := NewPostgresRepository(db)
	if err != nil {
		t.Fatalf("initialize repository schema: %v", err)
	}
	id, err := repo.Create(Employee{
		FullName: "Integration Test",
		Position: "Engineer",
		Salary:   42000,
	})
	if err != nil {
		t.Fatalf("create employee: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM employees WHERE id = $1", id)
	})

	got, err := repo.GetByID(id)
	if err != nil {
		t.Fatalf("read employee: %v", err)
	}
	if got.FullName != "Integration Test" || got.Position != "Engineer" || got.Salary != 42000 {
		t.Fatalf("unexpected employee: %+v", got)
	}

	if err := repo.ChangePosition(id, "Senior Engineer"); err != nil {
		t.Fatalf("update employee position: %v", err)
	}
	got, err = repo.GetByID(id)
	if err != nil {
		t.Fatalf("read updated employee: %v", err)
	}
	if got.Position != "Senior Engineer" {
		t.Fatalf("position = %q, want Senior Engineer", got.Position)
	}
}
