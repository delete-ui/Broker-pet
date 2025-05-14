package main

import (
	"Brocker-pet-project/internal/config"
	"Brocker-pet-project/internal/handlers"
	"Brocker-pet-project/internal/logger"
	"Brocker-pet-project/internal/repository"
	worker2 "Brocker-pet-project/internal/worker"
	"Brocker-pet-project/pkg/database"
	"Brocker-pet-project/pkg/middleware"
	"Brocker-pet-project/pkg/redis"
	"fmt"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"time"
)

func main() {
	fmt.Println("STARTED")

	cfg := config.MustLoad("local.yml")
	zaplog, err := logger.InitLogger(cfg)
	if err != nil {
		log.Fatalf("Error initializing logger: %v", err)
	}

	database.InitDB(cfg)

	r := chi.NewRouter()

	redisClient := redis.NewRedisClient(cfg)
	dealRepository := repository.NewDealRepository(database.ReturnDB(), redisClient)
	userRepository := repository.NewUserRepository(database.ReturnDB())
	profitRepository := repository.NewProfitRepository(database.ReturnDB())

	profitHandler := handlers.NewProfitHandler(profitRepository, zaplog)
	dealHandler := handlers.NewDealHandler(dealRepository, redisClient, zaplog)
	userHandler := handlers.NewUserHandler(userRepository, zaplog)

	r.Post("/api/new_deal", dealHandler.NewDealPost)
	r.Post("/api/registration", userHandler.NewUserPost)
	r.Get("/api/login", userHandler.LoginIn)

	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware)

		r.Get("/api/all_deals", dealHandler.AllDealsGet)
		r.Get("/api/all_processed_deals", dealHandler.AllProcessedDealsGet)
		r.Get("/api/all_not_processed_deals", dealHandler.AllNotProcessedDealsGet)
		r.Get("/api/all_clear_profit", profitHandler.AllClearProfitGET)
	})

	dealWorker := worker2.NewDealWorker(zaplog, dealRepository, profitRepository)

	go Worker(dealWorker)

	zaplog.Info("Program started")

	if err := http.ListenAndServe(cfg.Server.Port, r); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}

}

func Worker(dealWorker *worker2.DealWorker) {
	res := true

	for res == true {
		time.Sleep(3 * time.Second)
		dealWorker.MarkAsProcessed()
	}
}

//TODO: TESTS,DOCKER,CI/CD
