package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/pprof"
	"github.com/jaytnw/bms-service/internal/config"
	"github.com/jaytnw/bms-service/internal/handlers"
	"github.com/jaytnw/bms-service/internal/mqtt"
	"github.com/jaytnw/bms-service/internal/repository"
	"github.com/jaytnw/bms-service/internal/routes"
	"github.com/jaytnw/bms-service/internal/services"
	redisPkg "github.com/jaytnw/bms-service/pkg/redisclient"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env not loaded (using system env)")
	}

	// Load config
	cfg := config.LoadConfig()

	dsn := cfg.PostgresConfig.BuildDSN()
	if dsn == "" {
		log.Fatal("POSTGRES_DSN not set")
	}

	// Connect DB
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// Redis
	redisPkg.Init(cfg.RedisConfig)

	// ExternalAPI
	externalAPI := services.NewExternalAPIService("https://washeasy.me")

	// Wire DI
	statusRepo := repository.NewStatusRepo(db)
	statusService := services.NewStatusService(statusRepo, externalAPI, redisPkg.Client)
	statusHandler := handlers.NewStatusHandler(statusService)

	// Create Fiber app
	app := fiber.New()
	app.Use(cors.New())
	app.Use(logger.New())
	app.Use(pprof.New())

	baseClientID := cfg.MQTTConfig.ClientID


	clientID := fmt.Sprintf("%s-%d", baseClientID, time.Now().UnixNano())

	// MQTT Connect
	mqttClient := mqtt.NewClient(
		cfg.MQTTConfig.BrokerURL,
		clientID,
		cfg.MQTTConfig.Username,
		cfg.MQTTConfig.Password,
	)

	// MQTT Subscribe
	err = mqttClient.Subscribe("washingMachine/+/+/status", func(topic string, payload []byte, retained bool) {
		log.Printf("📥 Topic: %s | Payload: %s | Retained: %v", topic, string(payload), retained)
		statusService.HandleMQTTStatusUpdate(topic, payload)
	})
	

	if err != nil {
		log.Fatalf("❌ MQTT subscribe failed: %v", err)
	}

	// Setup routes
	routes.Setup(app, statusHandler)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("Server is running at http://localhost:%s", port)
	log.Fatal(app.Listen(":" + port))

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt) // listen for Ctrl+C
	<-quit                            // wait for signal

	log.Println("🛑 Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Fatalf("❌ Shutdown error: %v", err)
	}

	log.Println("✅ Server gracefully stopped.")
}
