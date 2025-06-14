package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"ports-and-adapters-architecture/api/rest"
	"ports-and-adapters-architecture/internal/adapters/cache"
	"ports-and-adapters-architecture/internal/adapters/messaging"
	"ports-and-adapters-architecture/internal/adapters/payment"
	"ports-and-adapters-architecture/internal/adapters/persistence"
	"ports-and-adapters-architecture/internal/domain"
	"ports-and-adapters-architecture/internal/usecase"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
)

func main() {
	// Load configuration
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database
	db, err := initDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize Redis cache
	redisCache := cache.NewRedisCache(
		cfg.GetString("redis.addr"),
		cfg.GetString("redis.password"),
		cfg.GetInt("redis.db"),
	)
	defer redisCache.Close()

	// Test Redis connection
	if err := redisCache.Ping(context.Background()); err != nil {
		log.Printf("Warning: Redis connection failed: %v", err)
	}

	// Initialize Kafka
	kafkaPublisher := messaging.NewKafkaEventPublisher(
		cfg.GetStringSlice("kafka.brokers"),
	)
	defer kafkaPublisher.Close()

	// Initialize repositories
	userRepo := persistence.NewPostgresUserRepository(db)
	walletRepo := persistence.NewPostgresWalletRepository(db)
	transactionRepo := persistence.NewPostgresTransactionRepository(db)
	paymentRepo := persistence.NewPostgresPaymentRepository(db)

	// Initialize payment gateways
	midtransGateway := payment.NewMidtransGateway(
		cfg.GetString("payment.midtrans.server_key"),
		cfg.GetString("payment.midtrans.client_key"),
		cfg.GetBool("payment.midtrans.is_production"),
	)

	stripeGateway := payment.NewStripeGateway(
		cfg.GetString("payment.stripe.api_key"),
		cfg.GetString("payment.stripe.webhook_secret"),
		cfg.GetBool("payment.stripe.is_test"),
	)

	// Initialize services
	userService := usecase.NewUserService(userRepo, kafkaPublisher, redisCache)
	walletService := usecase.NewWalletService(
		walletRepo,
		userRepo,
		transactionRepo,
		kafkaPublisher,
		redisCache,
	)
	transactionService := usecase.NewTransactionService(
		transactionRepo,
		walletRepo,
		kafkaPublisher,
		redisCache,
	)
	paymentService := usecase.NewPaymentService(
		paymentRepo,
		walletRepo,
		transactionRepo,
		kafkaPublisher,
		redisCache,
	)

	// Register payment gateways
	paymentService.RegisterGateway(domain.PaymentProviderMidtrans, midtransGateway)
	paymentService.RegisterGateway(domain.PaymentProviderStripe, stripeGateway)

	// Initialize Echo
	e := echo.New()

	// Setup routes
	rest.SetupRoutes(e, walletService, paymentService)

	// Start server
	go func() {
		port := cfg.GetString("server.port")
		if port == "" {
			port = "8080"
		}

		log.Printf("Starting server on port %s", port)
		if err := e.Start(":" + port); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func loadConfig() (*viper.Viper, error) {
	v := viper.New()

	// Set config name and paths
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./config")
	v.AddConfigPath(".")

	// Set defaults
	setDefaults(v)

	// Read base config
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Override with environment-specific config
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "local"
	}

	v.SetConfigName(fmt.Sprintf("config.%s", env))
	if err := v.MergeInConfig(); err != nil {
		log.Printf("No environment-specific config found for %s: %v", env, err)
	}

	// Allow environment variables to override
	v.AutomaticEnv()

	return v, nil
}

func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.port", "8080")
	v.SetDefault("server.timeout", "30s")

	// Database defaults
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", "5432")
	v.SetDefault("database.name", "mini_ewallet")
	v.SetDefault("database.user", "postgres")
	v.SetDefault("database.password", "postgres")
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 25)
	v.SetDefault("database.conn_max_lifetime", "5m")

	// Redis defaults
	v.SetDefault("redis.addr", "localhost:6379")
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)

	// Kafka defaults
	v.SetDefault("kafka.brokers", []string{"localhost:9092"})
	v.SetDefault("kafka.consumer_group", "mini-ewallet")

	// Payment gateway defaults
	v.SetDefault("payment.midtrans.is_production", false)
	v.SetDefault("payment.stripe.is_test", true)
}

func initDatabase(cfg *viper.Viper) (*sql.DB, error) {
	db, err := persistence.NewPostgresConnection(
		cfg.GetString("database.host"),
		cfg.GetString("database.port"),
		cfg.GetString("database.user"),
		cfg.GetString("database.password"),
		cfg.GetString("database.name"),
	)
	if err != nil {
		return nil, err
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.GetInt("database.max_open_conns"))
	db.SetMaxIdleConns(cfg.GetInt("database.max_idle_conns"))
	db.SetConnMaxLifetime(cfg.GetDuration("database.conn_max_lifetime"))

	return db, nil
}
