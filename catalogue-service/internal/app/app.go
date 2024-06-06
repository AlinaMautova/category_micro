package app

import (
	grpc "catalogue-service/internal/app/grpc"
	"catalogue-service/internal/services/catalog"
	"catalogue-service/internal/storage"
	"log/slog"
	"time"
)

type App struct {
	GRPCServer *grpc.App
}

func New(log *slog.Logger, grpcPort int, dsn string, tokenTTL time.Duration) *App {
	// TODO: database setup
	itemRepo, err := storage.New(dsn)
	if err != nil {
		panic(err)
	}

	// TODO: catalogue service setup in services/catalogue
	catalogueService := catalog.New(log, itemRepo, tokenTTL)

	// TODO: grpc app setup
	grpcApp := grpc.New(log, catalogueService, grpcPort)

	return &App{GRPCServer: grpcApp}
}
