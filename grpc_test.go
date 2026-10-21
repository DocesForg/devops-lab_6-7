package main

import (
	"context"
	"net"
	"testing"
	"time"

	employeev1 "github.com/DocesForg/devops-lab_6-7/gen/employee/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestGRPCEmployeeLifecycle(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := grpc.NewServer()
	registerEmployeeGRPCServer(server, newEmployeeGRPCServer(NewEmployeeService(NewMemoryRepository())))
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(func() {
		server.Stop()
		_ = listener.Close()
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := grpc.NewClient(listener.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	client := employeev1.NewEmployeeServiceClient(conn)

	created, err := client.Create(ctx, &employeev1.CreateEmployeeRequest{
		FullName: "Иван Иванов", Position: "Инженер", Salary: 80000,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.GetId() == "" {
		t.Fatal("Create response has no employee ID")
	}

	got, err := client.GetByID(ctx, &employeev1.EmployeeIDRequest{Id: created.GetId()})
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.GetEmployee().GetFullName() != "Иван Иванов" {
		t.Fatalf("unexpected full_name: %q", got.GetEmployee().GetFullName())
	}

	_, err = client.ChangePosition(ctx, &employeev1.ChangePositionRequest{
		Id: created.GetId(), Position: "Ведущий инженер",
	})
	if err != nil {
		t.Fatalf("ChangePosition: %v", err)
	}

	got, err = client.GetByID(ctx, &employeev1.EmployeeIDRequest{Id: created.GetId()})
	if err != nil {
		t.Fatalf("GetByID after change: %v", err)
	}
	if got.GetEmployee().GetPosition() != "Ведущий инженер" {
		t.Fatalf("unexpected position after change: %q", got.GetEmployee().GetPosition())
	}
}
