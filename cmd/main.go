package cmd

import (
	"context"
	"data_processor/internal/repo"
	data_processor "data_processor/internal/transport"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
)

func main() {
	// Инициализация подключения к БД
	pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Инициализация репозиториев
	repositories := repo.NewPgxRepository(pool)

	// Создание gRPC сервера
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
			log.Printf("gRPC method: %s", info.FullMethod)
			return handler(ctx, req)
		}),
	)

	server := data_processor.NewServer(repositories)

	// Регистрация сервисов
	data_processor.RegisterUserServiceServer(grpcServer, server)
	data_processor.RegisterOrganizationServiceServer(grpcServer, server)
	data_processor.RegisterTeamServiceServer(grpcServer, server)
	data_processor.RegisterApplicationServiceServer(grpcServer, server)
	data_processor.RegisterVersionServiceServer(grpcServer, server)
	data_processor.RegisterScanServiceServer(grpcServer, server)
	data_processor.RegisterScanInfoServiceServer(grpcServer, server)
	data_processor.RegisterScanRuleServiceServer(grpcServer, server)
	data_processor.RegisterPermissionServiceServer(grpcServer, server)
	data_processor.RegisterRoleServiceServer(grpcServer, server)

	// Запуск сервера
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Println("Server started on :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
