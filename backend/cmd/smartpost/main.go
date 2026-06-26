package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/redis/go-redis/v9"
	"github.com/smartpost/backend/internal/api"
	"github.com/smartpost/backend/internal/bot"
	"github.com/smartpost/backend/internal/config"
	"github.com/smartpost/backend/internal/database"
	"github.com/smartpost/backend/internal/scheduler"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("🚀 SmartPost starting...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Config error: %v", err)
	}
	log.Println("✅ Config loaded")

	dbPool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("❌ Database error: %v", err)
	}
	defer dbPool.Close()
	log.Println("✅ PostgreSQL connected")

	migrationSQL, err := os.ReadFile("internal/database/migrations/001_init.sql")
	if err != nil {
		log.Printf("⚠️ Could not read migration file: %v (skipping)", err)
	} else {
		if err := database.RunMigrations(ctx, dbPool, string(migrationSQL)); err != nil {
			log.Fatalf("❌ Migration error: %v", err)
		}
		log.Println("✅ Migrations applied")
	}

	rdbOpts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatalf("❌ Redis URL parse error: %v", err)
	}
	rdb := redis.NewClient(rdbOpts)
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("❌ Redis connection error: %v", err)
	}
	defer rdb.Close()
	log.Println("✅ Redis connected")

	publisher := scheduler.NewPublisher(cfg.RedisAddr())
	defer publisher.Close()

	tgBot, err := bot.New(cfg, dbPool, rdb, publisher)
	if err != nil {
		log.Fatalf("❌ Bot creation error: %v", err)
	}

	if err := tgBot.SetWebhook(ctx); err != nil {
		log.Printf("⚠️ Webhook error (will retry): %v", err)
	}

	worker := scheduler.NewWorker(dbPool, tgBot.API())
	go func() {
		if err := worker.Start(cfg.RedisAddr()); err != nil {
			log.Printf("❌ Asynq worker error: %v", err)
		}
	}()
	log.Println("✅ Asynq worker started")

	var webhookHandler http.Handler
	if cfg.WebhookURL != "" {
		webhookHandler = tgBot.API().WebhookHandler()
	}

	apiServer := api.NewServer(cfg, dbPool, publisher, webhookHandler)
	go func() {
		if err := apiServer.Start(); err != nil {
			log.Fatalf("❌ API server error: %v", err)
		}
	}()
	log.Println("✅ API server started")

	go tgBot.Start(ctx)
	log.Println("✅ Bot processor started")
	log.Println("🟢 SmartPost is running!")
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh
	log.Printf("⚠️ Received signal: %v, shutting down...", sig)

	cancel()
	log.Println("👋 SmartPost stopped")
}