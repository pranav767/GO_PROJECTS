package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"log/slog"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	pb "leaderboard_system/api/gen/leaderboard/v1"
	grpcdelivery "leaderboard_system/internal/delivery/grpc"
	"leaderboard_system/internal/delivery/grpc/interceptors"
	"leaderboard_system/internal/delivery/ws"
	"leaderboard_system/internal/repository"
	"leaderboard_system/internal/service"
)

func main() {
	// Load env
	_ = godotenv.Load()

	// Structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Init MySQL
	mysqlDB, err := repository.NewMySQL(repository.MySQLConfig{
		User:     envOrDefault("DB_USER", "root"),
		Password: envOrDefault("DB_PASS", "adminpass"),
		Host:     envOrDefault("DB_HOST", "localhost"),
		Port:     envOrDefault("DB_PORT", "3306"),
		DBName:   envOrDefault("DB_NAME", "leaderboard_system"),
	})
	if err != nil {
		logger.Error("mysql init failed", slog.Any("error", err))
		os.Exit(1)
	}

	// Init Redis with connection pooling
	redisClient, err := repository.NewRedis(repository.RedisConfig{
		Addr:         envOrDefault("REDIS_ADDR", "localhost:6379"),
		PoolSize:     20,
		MinIdleConns: 5,
	})
	if err != nil {
		logger.Error("redis init failed", slog.Any("error", err))
		os.Exit(1)
	}

	// Build repositories
	userRepo := repository.NewUserRepo(mysqlDB)
	gameRepo := repository.NewGameRepo(mysqlDB)
	lbRepo := repository.NewLeaderboardRepo(redisClient)
	historyRepo := repository.NewScoreHistoryRepo(mysqlDB)

	// Build WebSocket hub
	hub := ws.NewHub(logger)
	hub.Start()

	jwtSecret := []byte(os.Getenv("HMAC_SECRET"))
	adminUsername := os.Getenv("ADMIN_USERNAME")

	// Build services
	authSvc := service.NewAuthService(userRepo, jwtSecret, adminUsername, logger)
	gameSvc := service.NewGameService(gameRepo, logger)
	lbSvc := service.NewLeaderboardService(lbRepo, gameRepo, userRepo, historyRepo, hub)
	historySvc := service.NewScoreHistoryService(historyRepo)

	// Create gRPC server with chained interceptors
	validationInterceptor, err := interceptors.ValidationUnaryInterceptor()
	if err != nil {
		logger.Error("failed to create validation interceptor", slog.Any("error", err))
		os.Exit(1)
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptors.RecoveryUnaryInterceptor(logger),
			validationInterceptor,
			interceptors.AuthUnaryInterceptor(authSvc),
		),
	)

	// Register gRPC service implementations
	pb.RegisterAuthServiceServer(grpcServer, grpcdelivery.NewAuthServer(authSvc))
	pb.RegisterGameServiceServer(grpcServer, grpcdelivery.NewGameServer(gameSvc))
	pb.RegisterLeaderboardServiceServer(grpcServer, grpcdelivery.NewLeaderboardServer(lbSvc))
	pb.RegisterScoreHistoryServiceServer(grpcServer, grpcdelivery.NewScoreHistoryServer(historySvc))

	// Enable server reflection so grpcurl and other tools can discover services
	reflection.Register(grpcServer)

	// Start gRPC listener
	grpcPort := envOrDefault("GRPC_PORT", "9090")
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		logger.Error("failed to listen for gRPC", slog.Any("error", err))
		os.Exit(1)
	}

	go func() {
		logger.Info("gRPC server starting", slog.String("port", grpcPort))
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error("gRPC server failed", slog.Any("error", err))
		}
	}()

	// Create grpc-gateway mux for REST compatibility
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	gwMux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	grpcEndpoint := ":" + grpcPort

	if err := registerGatewayHandlers(ctx, gwMux, grpcEndpoint, opts); err != nil {
		logger.Error("failed to register gateway handlers", slog.Any("error", err))
		os.Exit(1)
	}

	// HTTP mux: WebSocket + healthz + grpc-gateway
	httpMux := http.NewServeMux()
	httpMux.Handle("/ws/leaderboard", hub)
	httpMux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		mysql, redis := "ok", "ok"
		code := http.StatusOK
		if mysqlDB.PingContext(r.Context()) != nil {
			mysql, code = "error", http.StatusServiceUnavailable
		}
		if redisClient.Ping(r.Context()).Err() != nil {
			redis, code = "error", http.StatusServiceUnavailable
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		fmt.Fprintf(w, `{"mysql":%q,"redis":%q}`, mysql, redis)
	})
	httpMux.Handle("/", gwMux)

	httpPort := envOrDefault("HTTP_PORT", "8080")
	httpServer := &http.Server{Addr: ":" + httpPort, Handler: httpMux}

	go func() {
		logger.Info("HTTP gateway starting", slog.String("port", httpPort))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server failed", slog.Any("error", err))
		}
	}()

	logger.Info("leaderboard system is running",
		slog.String("grpc_port", grpcPort),
		slog.String("http_port", httpPort),
	)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down servers")
	if err := httpServer.Shutdown(context.Background()); err != nil {
		logger.Error("HTTP shutdown error", slog.Any("error", err))
	}
	grpcServer.GracefulStop()
	hub.Shutdown()
	mysqlDB.Close()
	redisClient.Close()
	logger.Info("shutdown complete")
}

func registerGatewayHandlers(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error {
	if err := pb.RegisterAuthServiceHandlerFromEndpoint(ctx, mux, endpoint, opts); err != nil {
		return err
	}
	if err := pb.RegisterGameServiceHandlerFromEndpoint(ctx, mux, endpoint, opts); err != nil {
		return err
	}
	if err := pb.RegisterLeaderboardServiceHandlerFromEndpoint(ctx, mux, endpoint, opts); err != nil {
		return err
	}
	return pb.RegisterScoreHistoryServiceHandlerFromEndpoint(ctx, mux, endpoint, opts)
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
