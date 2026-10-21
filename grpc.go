package main

import (
	"context"
	"errors"
	"net"

	"github.com/go-kit/kit/endpoint"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	employeev1 "github.com/DocesForg/devops-lab_6-7/gen/employee/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type employeeGRPCServer struct {
	employeev1.UnimplementedEmployeeServiceServer
	create         *kitgrpc.Server
	getByID        *kitgrpc.Server
	changePosition *kitgrpc.Server
}

func newEmployeeGRPCServer(service EmployeeService) *employeeGRPCServer {
	return &employeeGRPCServer{
		create: kitgrpc.NewServer(
			makeCreateEndpoint(service),
			decodeCreateGRPCRequest,
			encodeCreateGRPCResponse,
		),
		getByID: kitgrpc.NewServer(
			makeGetByIDEndpoint(service),
			decodeGetByIDGRPCRequest,
			encodeGetByIDGRPCResponse,
		),
		changePosition: kitgrpc.NewServer(
			makeChangePositionEndpoint(service),
			decodeChangePositionGRPCRequest,
			encodeChangePositionGRPCResponse,
		),
	}
}

func makeCreateEndpoint(service EmployeeService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		e := request.(Employee)
		id, err := service.Create(ctx, e)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"id": id}, nil
	}
}

func makeGetByIDEndpoint(service EmployeeService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		id := request.(map[string]interface{})["id"].(string)
		e, err := service.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}
		return e, nil
	}
}

func makeChangePositionEndpoint(service EmployeeService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		values := request.(map[string]interface{})
		id, _ := values["id"].(string)
		position, _ := values["position"].(string)
		if err := service.ChangePosition(ctx, id, position); err != nil {
			return nil, err
		}
		return map[string]interface{}{"status": "updated"}, nil
	}
}

func decodeCreateGRPCRequest(_ context.Context, request interface{}) (interface{}, error) {
	req, ok := request.(*employeev1.CreateEmployeeRequest)
	if !ok || req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid create request")
	}
	return Employee{
		FullName: req.GetFullName(),
		Position: req.GetPosition(),
		Salary:   req.GetSalary(),
		HiredOn:  req.GetHiredOn(),
	}, nil
}

func decodeGetByIDGRPCRequest(_ context.Context, request interface{}) (interface{}, error) {
	req, ok := request.(*employeev1.EmployeeIDRequest)
	if !ok || req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid employee ID request")
	}
	return map[string]interface{}{"id": req.GetId()}, nil
}

func decodeChangePositionGRPCRequest(_ context.Context, request interface{}) (interface{}, error) {
	req, ok := request.(*employeev1.ChangePositionRequest)
	if !ok || req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid change-position request")
	}
	return map[string]interface{}{"id": req.GetId(), "position": req.GetPosition()}, nil
}

func encodeCreateGRPCResponse(_ context.Context, response interface{}) (interface{}, error) {
	values := response.(map[string]interface{})
	return &employeev1.CreateEmployeeResponse{Id: values["id"].(string)}, nil
}

func encodeGetByIDGRPCResponse(_ context.Context, response interface{}) (interface{}, error) {
	e := response.(Employee)
	return &employeev1.GetEmployeeResponse{Employee: &employeev1.Employee{
		Id: e.ID, FullName: e.FullName, Position: e.Position,
		Salary: e.Salary, HiredOn: e.HiredOn,
	}}, nil
}

func encodeChangePositionGRPCResponse(_ context.Context, response interface{}) (interface{}, error) {
	values := response.(map[string]interface{})
	return &employeev1.ChangePositionResponse{Status: values["status"].(string)}, nil
}

func (s *employeeGRPCServer) Create(ctx context.Context, request *employeev1.CreateEmployeeRequest) (*employeev1.CreateEmployeeResponse, error) {
	response, err := s.serve(ctx, s.create, request)
	if err != nil {
		return nil, err
	}
	result, ok := response.(*employeev1.CreateEmployeeResponse)
	if !ok {
		return nil, status.Error(codes.Internal, "invalid response type")
	}
	return result, nil
}

func (s *employeeGRPCServer) GetByID(ctx context.Context, request *employeev1.EmployeeIDRequest) (*employeev1.GetEmployeeResponse, error) {
	response, err := s.serve(ctx, s.getByID, request)
	if err != nil {
		return nil, err
	}
	result, ok := response.(*employeev1.GetEmployeeResponse)
	if !ok {
		return nil, status.Error(codes.Internal, "invalid response type")
	}
	return result, nil
}

func (s *employeeGRPCServer) ChangePosition(ctx context.Context, request *employeev1.ChangePositionRequest) (*employeev1.ChangePositionResponse, error) {
	response, err := s.serve(ctx, s.changePosition, request)
	if err != nil {
		return nil, err
	}
	result, ok := response.(*employeev1.ChangePositionResponse)
	if !ok {
		return nil, status.Error(codes.Internal, "invalid response type")
	}
	return result, nil
}

func (s *employeeGRPCServer) serve(ctx context.Context, handler *kitgrpc.Server, request interface{}) (interface{}, error) {
	_, response, err := handler.ServeGRPC(ctx, request)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalid):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case errors.Is(err, ErrNotFound):
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, "internal server error")
		}
	}
	return response, nil
}

func registerEmployeeGRPCServer(server *grpc.Server, implementation *employeeGRPCServer) {
	employeev1.RegisterEmployeeServiceServer(server, implementation)
}

func listenGRPC(address string) (net.Listener, error) {
	return net.Listen("tcp", address)
}
