package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"audit-service/config"
	"audit-service/db"
	"audit-service/internal/handler"
	"audit-service/internal/repository"
	"audit-service/internal/service"
	"audit-service/pkg/postgres"

	"github.com/gorilla/mux"
)

func main() {
	// 1. Загрузка конфигурации
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Подключение к БД
	dbConn, err := postgres.NewConnection(
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
	)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbConn.Close()

	// 3. Применение миграций
	if err := db.RunMigrations(dbConn); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// 4. Инициализация слоев
	auditRepo := repository.NewAuditRepository(dbConn)
	auditService := service.NewAuditService(auditRepo)
	auditHandler := handler.NewAuditHandler(auditService)
	statsHandler := handler.NewStatsHandler(cfg.AppVersion)

	// 5. Настройка health-check для БД
	go monitorDBConnection(dbConn, statsHandler)

	// 6. Настройка маршрутизатора
	router := mux.NewRouter()

	// API эндпоинты
	apiRouter := router.PathPrefix("/audit").Subrouter()
	apiRouter.HandleFunc("/events/", auditHandler.StoreEvent).Methods("POST")
	apiRouter.HandleFunc("/events/query", auditHandler.FindEvents).Methods("GET")

	// Сервисные эндпоинты
	router.HandleFunc("/stats", statsHandler.Stats).Methods("GET")
	router.HandleFunc("/health", statsHandler.HealthCheck).Methods("GET")

	// Middleware для сбора статистики
	router.Use(statsHandler.Middleware)

	// 7. Graceful shutdown
	srv := &http.Server{
		Addr:         ":" + string(cfg.ServerPort),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Starting audit service on port %d", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}

func monitorDBConnection(db *sql.DB, statsHandler *handler.StatsHandler) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err := db.PingContext(ctx)
		cancel()

		statsHandler.SetDBConnected(err == nil)
		if err != nil {
			log.Printf("Database connection check failed: %v", err)
		}
	}
}
