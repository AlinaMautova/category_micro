package grpc

import (
	catalogueGrpc "catalogue-service/internal/transport/grpc/catalog"
	"fmt"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	authp "github.com/lyerzhanl/contracts/gen/go/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"log"
	"log/slog"
	"net"
)

type App struct {
	log        *slog.Logger
	grpcServer *grpc.Server
	port       int
}

var (
	AuthServiceClient     authp.AuthClient
	UserInfoServiceClient authp.UserServiceClient
)

func New(
	log *slog.Logger,
	catalogueService catalogueGrpc.Catalogue,
	port int,
) *App {
	loggingOpts := []logging.Option{
		logging.WithLogOnEvents(
			//logging.StartCall, logging.FinishCall,
			logging.PayloadReceived, logging.PayloadSent,
		),
		// Add any other option (check functions starting with logging.With).
	}

	recoveryOpts := []recovery.Option{
		recovery.WithRecoveryHandler(func(p interface{}) (err error) {
			log.Error("Recovered from panic", slog.Any("panic", p))

			return status.Errorf(codes.Internal, "internal error")
		}),
	}

	gRPCServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
		recovery.UnaryServerInterceptor(recoveryOpts...),
		logging.UnaryServerInterceptor(MiddlewareLogger(log), loggingOpts...),
	), grpc.UnaryInterceptor(AuthInterceptor))

	catalogueGrpc.Register(gRPCServer, catalogueService)

	ConnectToAuthService()
	return &App{
		log:        log,
		port:       port,
		grpcServer: gRPCServer,
	}
}

// MustRun runs gRPC server and panics if any error occurs.
func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

// Run runs gRPC server.
func (a *App) Run() error {
	const op = "grpc.Run"

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	a.log.Info("grpc server started", slog.String("addr", l.Addr().String()))

	if err := a.grpcServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func ConnectToAuthService() {
	conn, err := grpc.NewClient("0.0.0.0:"+
		"44044", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to auth service: %v", err)
	}
	AuthServiceClient = authp.NewAuthClient(conn)
	UserInfoServiceClient = authp.NewUserServiceClient(conn)
}

func (a *App) Stop() {
	const op = "grpc.Stop"

	a.log.With(slog.String("op", op)).
		Info("stopping gRPC server", slog.Int("port", a.port))

	a.grpcServer.GracefulStop()
}
