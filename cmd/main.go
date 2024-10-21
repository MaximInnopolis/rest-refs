package main

import (
	"os"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"rest-refs/internal/app/api"
	"rest-refs/internal/app/config"
	httpHandler "rest-refs/internal/app/http"
	"rest-refs/internal/app/repository"
	"rest-refs/internal/app/repository/database"
)

func main() {
	// Initialize logger
	log := logrus.New()
	log.SetOutput(os.Stdout)
	log.SetLevel(logrus.DebugLevel)

	// Create config
	cfg, err := config.New()
	if err != nil {
		log.Errorf("Ошибка при чтении конфига: %v", err)
		os.Exit(1)
	}

	// Create a new connection pool to database
	pool, err := database.NewPool(cfg.DbUrl)
	if err != nil {
		log.Errorf("Ошибка при создании соединения к базе данных: %v", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Create a new Database with connection pool
	db := database.NewDatabase(pool)

	// Create a new repo with Database and logger
	repo := repository.New(*db, log)

	// Create a new service
	refService := api.New(repo, log)

	// Create Http handler
	handler := httpHandler.New(*refService, log)

	// Init Router
	r := mux.NewRouter()

	handler.RegisterRoutes(r)

	// Start server
	handler.StartServer(cfg.HttpPort)
}
