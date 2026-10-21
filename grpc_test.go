package main

import (
	"context"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/structpb"
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

	createRequest, _ := structpb.NewStruct(map[string]interface{}{
		"full_name": "Иван Иванов", "position": "Инженер", "salary": 80000.0,
	})
	createResponse := new(structpb.Struct)
	if err := conn.Invoke(ctx, "/"+employeeGRPCServiceName+"/Create", createRequest, createResponse); err != nil {
		t.Fatalf("Create: %v", err)
	}
	id, ok := createResponse.AsMap()["id"].(string)
	if !ok || id == "" {
		t.Fatalf("Create response has no id: %v", createResponse.AsMap())
	}

	getRequest, _ := structpb.NewStruct(map[string]interface{}{"id": id})
	getResponse := new(structpb.Struct)
	if err := conn.Invoke(ctx, "/"+employeeGRPCServiceName+"/GetByID", getRequest, getResponse); err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got := getResponse.AsMap()["full_name"]; got != "Иван Иванов" {
		t.Fatalf("unexpected full_name: %v", got)
	}

	changeRequest, _ := structpb.NewStruct(map[string]interface{}{"id": id, "position": "Ведущий инженер"})
	changeResponse := new(structpb.Struct)
	if err := conn.Invoke(ctx, "/"+employeeGRPCServiceName+"/ChangePosition", changeRequest, changeResponse); err != nil {
		t.Fatalf("ChangePosition: %v", err)
	}

	getResponse = new(structpb.Struct)
	if err := conn.Invoke(ctx, "/"+employeeGRPCServiceName+"/GetByID", getRequest, getResponse); err != nil {
		t.Fatalf("GetByID after change: %v", err)
	}
	if got := getResponse.AsMap()["position"]; got != "Ведущий инженер" {
		t.Fatalf("unexpected position after change: %v", got)
	}
}
