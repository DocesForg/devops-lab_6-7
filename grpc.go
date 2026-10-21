package main

import (
	"context"
	"errors"
	"net"

	"github.com/go-kit/kit/endpoint"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

const employeeGRPCServiceName = "employee.v1.EmployeeService"

type employeeGRPCServer struct {
	create        *kitgrpc.Server
	getByID       *kitgrpc.Server
	changePosition *kitgrpc.Server
}

func newEmployeeGRPCServer(service EmployeeService) *employeeGRPCServer {
	return &employeeGRPCServer{
		create: kitgrpc.NewServer(
			makeCreateEndpoint(service),
			decodeCreateGRPCRequest,
			encodeGRPCResponse,
		),
		getByID: kitgrpc.NewServer(
			makeGetByIDEndpoint(service),
			decodeGetByIDGRPCRequest,
			encodeGRPCResponse,
		),
		changePosition: kitgrpc.NewServer(
			makeChangePositionEndpoint(service),
			decodeChangePositionGRPCRequest,
			encodeGRPCResponse,
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
		return map[string]interface{}{
			"id": e.ID, "full_name": e.FullName, "position": e.Position,
			"salary": e.Salary, "hired_on": float64(e.HiredOn),
		}, nil
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
	values, err := structRequest(request)
	if err != nil {
		return nil, err
	}
	e := Employee{}
	e.FullName, _ = values["full_name"].(string)
	e.Position, _ = values["position"].(string)
	e.Salary, _ = values["salary"].(float64)
	if hiredOn, ok := values["hired_on"].(float64); ok {
		e.HiredOn = int64(hiredOn)
	}
	return e, nil
}

func decodeGetByIDGRPCRequest(_ context.Context, request interface{}) (interface{}, error) {
	values, err := structRequest(request)
	if err != nil {
		return nil, err
	}
	id, _ := values["id"].(string)
	return map[string]interface{}{"id": id}, nil
}

func decodeChangePositionGRPCRequest(_ context.Context, request interface{}) (interface{}, error) {
	values, err := structRequest(request)
	if err != nil {
		return nil, err
	}
	id, _ := values["id"].(string)
	position, _ := values["position"].(string)
	return map[string]interface{}{"id": id, "position": position}, nil
}

func structRequest(request interface{}) (map[string]interface{}, error) {
	message, ok := request.(*structpb.Struct)
	if !ok || message == nil {
		return nil, status.Error(codes.InvalidArgument, "request must be a protobuf Struct")
	}
	return message.AsMap(), nil
}

func encodeGRPCResponse(_ context.Context, response interface{}) (interface{}, error) {
	values, ok := response.(map[string]interface{})
	if !ok {
		return nil, status.Error(codes.Internal, "invalid service response")
	}
	message, err := structpb.NewStruct(values)
	if err != nil {
		return nil, status.Error(codes.Internal, "could not encode response")
	}
	return message, nil
}

func (s *employeeGRPCServer) Create(ctx context.Context, request *structpb.Struct) (*structpb.Struct, error) {
	return s.serve(ctx, s.create, request)
}

func (s *employeeGRPCServer) GetByID(ctx context.Context, request *structpb.Struct) (*structpb.Struct, error) {
	return s.serve(ctx, s.getByID, request)
}

func (s *employeeGRPCServer) ChangePosition(ctx context.Context, request *structpb.Struct) (*structpb.Struct, error) {
	return s.serve(ctx, s.changePosition, request)
}

func (s *employeeGRPCServer) serve(ctx context.Context, handler *kitgrpc.Server, request *structpb.Struct) (*structpb.Struct, error) {
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
	message, ok := response.(*structpb.Struct)
	if !ok {
		return nil, status.Error(codes.Internal, "invalid response type")
	}
	return message, nil
}

func registerEmployeeGRPCServer(server *grpc.Server, implementation *employeeGRPCServer) {
	server.RegisterService(&grpc.ServiceDesc{
		ServiceName: employeeGRPCServiceName,
		HandlerType: (*employeeGRPCService)(nil),
		Methods: []grpc.MethodDesc{
			{MethodName: "Create", Handler: employeeCreateHandler},
			{MethodName: "GetByID", Handler: employeeGetByIDHandler},
			{MethodName: "ChangePosition", Handler: employeeChangePositionHandler},
		},
	}, implementation)
}

type employeeGRPCService interface {
	Create(context.Context, *structpb.Struct) (*structpb.Struct, error)
	GetByID(context.Context, *structpb.Struct) (*structpb.Struct, error)
	ChangePosition(context.Context, *structpb.Struct) (*structpb.Struct, error)
}

func employeeCreateHandler(srv interface{}, ctx context.Context, decode func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	return invokeEmployeeMethod(srv, ctx, decode, interceptor, "/"+employeeGRPCServiceName+"/Create", (*employeeGRPCServer).Create)
}

func employeeGetByIDHandler(srv interface{}, ctx context.Context, decode func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	return invokeEmployeeMethod(srv, ctx, decode, interceptor, "/"+employeeGRPCServiceName+"/GetByID", (*employeeGRPCServer).GetByID)
}

func employeeChangePositionHandler(srv interface{}, ctx context.Context, decode func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	return invokeEmployeeMethod(srv, ctx, decode, interceptor, "/"+employeeGRPCServiceName+"/ChangePosition", (*employeeGRPCServer).ChangePosition)
}

type employeeMethod func(*employeeGRPCServer, context.Context, *structpb.Struct) (*structpb.Struct, error)

func invokeEmployeeMethod(srv interface{}, ctx context.Context, decode func(interface{}) error, interceptor grpc.UnaryServerInterceptor, fullMethod string, method employeeMethod) (interface{}, error) {
	request := new(structpb.Struct)
	if err := decode(request); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return method(srv.(*employeeGRPCServer), ctx, request)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: fullMethod}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return method(srv.(*employeeGRPCServer), ctx, req.(*structpb.Struct))
	}
	return interceptor(ctx, request, info, handler)
}

func listenGRPC(address string) (net.Listener, error) {
	return net.Listen("tcp", address)
}
