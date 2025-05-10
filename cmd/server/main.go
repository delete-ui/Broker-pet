package main

import (
	"Brocker-pet-project/internal/config"
	"Brocker-pet-project/internal/logger"
	"Brocker-pet-project/pkg/database"
	"fmt"
	"log"
)

func main() {
	fmt.Println("STARTED")

	cfg := config.MustLoad("local.yml")
	zaplog, err := logger.InitLogger(cfg)
	if err != nil {
		log.Fatalf("Error initializing logger: %v", err)
	}

	database.InitDB(cfg)

	DB := database.ReturnDB()
	defer DB.Close()

	zaplog.Info("Program started")

}
