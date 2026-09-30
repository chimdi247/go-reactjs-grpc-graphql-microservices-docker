//go:generate protoc ./account.proto --go_out=plugins=grpc:./pb
package account

import (
	"context"
	"fmt"
	"net"

	"github.com/akhilsharma90/go-graphql-microservice/account/pb"
	"github.com/akhilsharma90/go-graphql-microservice/pkg/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/metric"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type grpcServer struct {
	pb.UnimplementedAccountServiceServer
	service Service
}

// ListenGRPC starts the gRPC server with OpenTelemetry auto-instrumentation:
// otelgrpc's stats handler traces every RPC with no per-method code, and a
// unary interceptor records request-count/latency metrics (see
// pkg/telemetry) so the Grafana dashboard has real per-service numbers.
// It also registers the standard gRPC health service so Docker/Compose
// healthchecks can use it.
func ListenGRPC(s Service, meter metric.Meter, port int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	serv := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.UnaryInterceptor(telemetry.GRPCUnaryMetricsInterceptor(meter)),
	)
	pb.RegisterAccountServiceServer(serv, &grpcServer{service: s})

	healthServer := health.NewServer()
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(serv, healthServer)

	reflection.Register(serv)
	return serv.Serve(lis)
}

func (s *grpcServer) PostAccount(ctx context.Context, r *pb.PostAccountRequest) (*pb.PostAccountResponse, error) {
	a, err := s.service.PostAccount(ctx, r.Name, r.Email, r.Password)
	if err != nil {
		return nil, err
	}
	return &pb.PostAccountResponse{Account: &pb.Account{
		Id:    a.ID,
		Name:  a.Name,
		Email: a.Email,
	}}, nil
}

func (s *grpcServer) GetAccount(ctx context.Context, r *pb.GetAccountRequest) (*pb.GetAccountResponse, error) {
	a, err := s.service.GetAccount(ctx, r.Id)
	if err != nil {
		return nil, err
	}
	return &pb.GetAccountResponse{
		Account: &pb.Account{
			Id:    a.ID,
			Name:  a.Name,
			Email: a.Email,
		},
	}, nil
}

func (s *grpcServer) GetAccounts(ctx context.Context, r *pb.GetAccountsRequest) (*pb.GetAccountsResponse, error) {
	res, err := s.service.GetAccounts(ctx, r.Skip, r.Take)
	if err != nil {
		return nil, err
	}
	accounts := []*pb.Account{}
	for _, p := range res {
		accounts = append(
			accounts,
			&pb.Account{
				Id:    p.ID,
				Name:  p.Name,
				Email: p.Email,
			},
		)
	}
	return &pb.GetAccountsResponse{Accounts: accounts}, nil
}

// Authenticate verifies email+password (bcrypt) and returns the account on
// success. The password hash never leaves this service — the response
// only ever carries the public Account fields.
func (s *grpcServer) Authenticate(ctx context.Context, r *pb.AuthenticateRequest) (*pb.AuthenticateResponse, error) {
	a, err := s.service.Authenticate(ctx, r.Email, r.Password)
	if err != nil {
		if err == ErrInvalidCredentials {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.AuthenticateResponse{
		Account: &pb.Account{
			Id:    a.ID,
			Name:  a.Name,
			Email: a.Email,
		},
	}, nil
}

// GetAccountsCount backs the "total users" business-KPI metric/dashboard
// panel and an admin-facing GraphQL field.
func (s *grpcServer) GetAccountsCount(ctx context.Context, r *pb.GetAccountsCountRequest) (*pb.GetAccountsCountResponse, error) {
	count, err := s.service.GetAccountsCount(ctx)
	if err != nil {
		return nil, err
	}
	return &pb.GetAccountsCountResponse{Count: count}, nil
}
