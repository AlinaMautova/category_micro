package grpc

import (
	"context"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	authp "github.com/lyerzhanl/contracts/gen/go/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"log"
	"log/slog"
	"strings"
)

func MiddlewareLogger(l *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}

func AuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if info.FullMethod != "/catalogue.CatalogueService/CreateItem" {
		return handler(ctx, req)
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Printf("failed to get metadata from context")
	}
	tkn, found := md["authorization"]
	if !found && len(tkn) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "authentication is required")
	}

	getUserInfoRequest := &authp.GetUserInfoRequest{Token: tkn[0]}
	getUserInfoResponse, err := UserInfoServiceClient.GetUserInfo(ctx, getUserInfoRequest)
	if err != nil {
		log.Printf("failed to get user info %v", err)
		switch {
		case strings.Contains(err.Error(), "unknown user"):
			return nil, status.Errorf(codes.InvalidArgument, "invalid input")
		default:
			return nil, status.Errorf(codes.Internal, "get user info failed")
		}
	}
	isAdminRequest := &authp.IsAdminRequest{UserId: int64(getUserInfoResponse.User.Id)}
	isAdminResponse, err := AuthServiceClient.IsAdmin(ctx, isAdminRequest)
	if err != nil {
		log.Printf("permissions fail %v", err)
		return nil, status.Errorf(codes.PermissionDenied, "permission failed")
	}

	if !isAdminResponse.IsAdmin {
		return nil, status.Errorf(codes.PermissionDenied, "permission failed")
	}

	return handler(ctx, req)
}
