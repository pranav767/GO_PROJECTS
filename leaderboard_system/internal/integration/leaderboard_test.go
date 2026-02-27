//go:build integration

package integration

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
	redismod "github.com/testcontainers/testcontainers-go/modules/redis"

	"leaderboard_system/internal/repository"
	"leaderboard_system/internal/service"
)

type noopBroadcaster struct{}

func (n *noopBroadcaster) Broadcast(_ []byte) {}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestFullFlow_Integration(t *testing.T) {
	ctx := context.Background()

	// Start MySQL container
	mysqlC, err := mysql.RunContainer(ctx,
		testcontainers.WithImage("mysql:8.0"),
		mysql.WithDatabase("leaderboard_system"),
		mysql.WithUsername("root"),
		mysql.WithPassword("testpass"),
		mysql.WithScripts("../../internal/db/db.sql"),
	)
	if err != nil {
		t.Fatalf("failed to start mysql: %v", err)
	}
	t.Cleanup(func() {
		if err := mysqlC.Terminate(ctx); err != nil {
			t.Logf("failed to terminate mysql: %v", err)
		}
	})

	// Start Redis container
	redisC, err := redismod.RunContainer(ctx,
		testcontainers.WithImage("redis:7.2-alpine"),
	)
	if err != nil {
		t.Fatalf("failed to start redis: %v", err)
	}
	t.Cleanup(func() {
		if err := redisC.Terminate(ctx); err != nil {
			t.Logf("failed to terminate redis: %v", err)
		}
	})

	// Get connection strings
	mysqlDSN, err := mysqlC.ConnectionString(ctx, "parseTime=true")
	if err != nil {
		t.Fatalf("failed to get mysql connection string: %v", err)
	}

	redisEndpoint, err := redisC.Endpoint(ctx, "")
	if err != nil {
		t.Fatalf("failed to get redis endpoint: %v", err)
	}

	// Connect via repository layer
	mysqlDB, err := repository.NewMySQLFromDSN(mysqlDSN)
	if err != nil {
		t.Fatalf("failed to connect mysql: %v", err)
	}
	defer mysqlDB.Close()

	redisClient, err := repository.NewRedis(repository.RedisConfig{Addr: redisEndpoint})
	if err != nil {
		t.Fatalf("failed to connect redis: %v", err)
	}
	defer redisClient.Close()

	// Build repositories
	userRepo := repository.NewUserRepo(mysqlDB)
	gameRepo := repository.NewGameRepo(mysqlDB)
	lbRepo := repository.NewLeaderboardRepo(redisClient)
	historyRepo := repository.NewScoreHistoryRepo(mysqlDB)

	// Build services
	logger := testLogger()
	authSvc := service.NewAuthService(userRepo, []byte("test-secret"), "", logger)
	gameSvc := service.NewGameService(gameRepo, logger)
	lbSvc := service.NewLeaderboardService(lbRepo, gameRepo, userRepo, historyRepo, &noopBroadcaster{})

	// Register users
	if err = authSvc.Register(ctx, "alice", "password123"); err != nil {
		t.Fatalf("register alice failed: %v", err)
	}
	if err = authSvc.Register(ctx, "bob", "password456"); err != nil {
		t.Fatalf("register bob failed: %v", err)
	}

	// Create game
	if _, err = gameSvc.CreateGame(ctx, "chess", "a board game"); err != nil {
		t.Fatalf("create game failed: %v", err)
	}

	// Submit scores
	alice, _ := userRepo.GetUserByUsername(ctx, "alice")
	bob, _ := userRepo.GetUserByUsername(ctx, "bob")

	if err = lbSvc.SubmitScore(ctx, alice.ID, "chess", 1500); err != nil {
		t.Fatalf("submit score alice failed: %v", err)
	}
	if err = lbSvc.SubmitScore(ctx, bob.ID, "chess", 2000); err != nil {
		t.Fatalf("submit score bob failed: %v", err)
	}

	// Verify leaderboard order
	entries, err := lbSvc.GetLeaderboard(ctx, "chess", 10)
	if err != nil {
		t.Fatalf("get leaderboard failed: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Score != 2000 || entries[0].Username != "bob" {
		t.Fatalf("expected bob at top with 2000, got %s with %f", entries[0].Username, entries[0].Score)
	}

	// Verify MAX score policy
	if err = lbSvc.SubmitScore(ctx, alice.ID, "chess", 1000); err != nil {
		t.Fatalf("submit lower score failed: %v", err)
	}
	aliceEntry, err := lbSvc.GetUserRank(ctx, "chess", alice.ID)
	if err != nil {
		t.Fatalf("get user rank failed: %v", err)
	}
	if aliceEntry.Score != 1500 {
		t.Fatalf("MAX policy violated: expected 1500, got %f", aliceEntry.Score)
	}
}
