package main

import (
	"log"
	"os"
	"time"
	"transfers-api/internal/config"
	"transfers-api/internal/consumer"
	"transfers-api/internal/handlers"
	"transfers-api/internal/logging"
	"transfers-api/internal/repositories"
	"transfers-api/internal/repository/ccache"
	"transfers-api/internal/services"
	"transfers-api/internal/transport"
	"transfers-api/internal/version"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	// init logger
	logger := logging.Logger
	logger.Info("logger started")

	// init config
	cfg := config.ParseFromEnv()
	logger.Infof("config loaded: %v", cfg.String())

	rabbitMQURL := os.Getenv("RABBITMQ_URL")
	if rabbitMQURL == "" {
		rabbitMQURL = "amqp://guest:guest@localhost:5672/"
	}

	conn, err := amqp.Dial(rabbitMQURL)
	if err != nil {
		logger.Fatalf("rabbitmq dial error: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		logger.Fatalf("rabbitmq channel error: %v", err)
	}
	defer ch.Close()

	if _, err := ch.QueueDeclare("transfers", true, false, false, false, nil); err != nil {
		logger.Fatalf("rabbitmq queue declare error: %v", err)
	}

	consumer := consumer.NewTransferConsumer(ch, "transfers")
	go func() {
		if err := consumer.Start(); err != nil {
			log.Printf("[TransferConsumer] stopped: %v", err)
		}
	}()
	logger.Infof("rabbitmq consumer started on queue: %s", "transfers")

	// init repositories
	transfersDB := repositories.NewTransfersMongoDBRepository(cfg.MongoDBConfig)
	localCache := ccache.NewTransferCcacheRepo(300 * time.Second)
	logger.Info("repositories created")

	// init services
	transfersService := services.NewTransfersService(cfg.Business, transfersDB, localCache)
	logger.Infof("services created")

	// init handlers
	transfersHandler := handlers.NewTransfersHandler(transfersService)
	logger.Infof("handlers created")

	// init server
	server := transport.NewHTTPServer(transfersHandler)
	server.MapRoutes()
	logger.Infof("server created, running %s@%s", version.AppName, version.Version)

	// run server
	server.Run(":8080")
}
