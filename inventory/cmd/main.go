package main

import (
	"log/slog"
	"net"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"

	inventoryService "github.com/PianyCoder/ship-trader/inventory/pkg/service"
	inventoryv1 "github.com/PianyCoder/ship-trader/shared/pkg/proto/inventory/v1"
)

const (
	grpcAddress = ":50051"

	// keepalive параметры gRPC сервера
	grpcMaxConnectionIdle     = 15 * time.Minute // Закрыть idle-соединение (нет активных RPC)
	grpcMaxConnectionAge      = 30 * time.Minute // Принудительная ротация для балансировки
	grpcMaxConnectionAgeGrace = 5 * time.Second  // Время на завершение активных RPC
	grpcKeepaliveTime         = 5 * time.Minute  // Интервал ping`ов для обнаружения мертвых соединений
	grpcKeepaliveTimeout      = 5 * time.Second  // Время ожидания pong (1с как-то маловато, возможен скачок сети)
	grpcMinPingInterval       = 5 * time.Minute  // Минимальный интервал ping`ов от клиента (защита от DoS)
)

func main() {
	//nolint:noctx // Контекст здесь не нужен: GracefulStop() сам закроет listener и прервёт Accept()
	lis, err := net.Listen("tcp", grpcAddress)
	if err != nil {
		slog.Error("не удалось создать listener", "error", err)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     grpcMaxConnectionIdle,
			MaxConnectionAge:      grpcMaxConnectionAge,
			MaxConnectionAgeGrace: grpcMaxConnectionAgeGrace,
			Time:                  grpcKeepaliveTime,
			Timeout:               grpcKeepaliveTimeout,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             grpcMinPingInterval,
			PermitWithoutStream: true, // разрешаем "теплые соединения", без активных RPC
		}),
	)

	// Регистрация inventory сервиса на gRPC-сервере
	inventoryv1.RegisterInventoryServiceServer(grpcServer, inventoryService.NewServer())

	// Включаем reflection для postman/grpcurl
	reflection.Register(grpcServer)

	slog.Info("запуск InventoryService", "адрес", grpcAddress)

	// TODO: Реализовать graceful shutdown
	// При получении сигнала SIGINT/SIGTERM сервер должен:
	// 1. Перестать принимать новые соединения
	// 2. Дождаться завершения текущих запросов
	// 3. Корректно завершить работу
	// Подсказка: используйте signal.NotifyContext и grpcServer.GracefulStop()

	err = grpcServer.Serve(lis)
	if err != nil {
		slog.Error("ошибка запуска сервера", "error", err)
		os.Exit(1)
	}
}
