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

	dealRepository := repository.NewDealRepository(database.ReturnDB())
	userRepository := repository.NewUserRepository(database.ReturnDB())
	redisClient := redis.NewRedisClient(cfg)

	dealHandler := handlers.NewDealHandler(dealRepository, redisClient, zaplog)
	userHandler := handlers.NewUserHandler(userRepository, zaplog)

	r.Post("/api/new_deal", dealHandler.NewDealPost)
	r.Post("/api/registration", userHandler.NewUserPost)
	r.Get("/api/login", userHandler.LoginIn)

	protectedRoute := r.Group(func(r chi.Router) {
		r.Get("/api/all_deals", dealHandler.AllDealsGet)
		r.Get("/api/all_processed_deals", dealHandler.AllProcessedDealsGet)
		r.Get("/api/all_not_processed_deals", dealHandler.AllNotProcessedDealsGet)
	})

	profitRepository := repository.NewProfitRepository(database.ReturnDB())

	dealWorker := worker2.NewDealWorker(zaplog, dealRepository, profitRepository)

	go func() {
		var res bool

		for res {
			time.Sleep(10 * time.Second)
			dealWorker.MarkAsProcessed()
		}
	}()

	dealWorker.MarkAsProcessed()

	r.Handle("/", middleware.AuthMiddleware(protectedRoute))

	zaplog.Info("Program started")

	if err := http.ListenAndServe(cfg.Server.Port, r); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}

}

//TODO: CREATING PROTECTED ROUTES, REDIS DEL IN MARKASPROCESSED FUNC, CREATE PROTECTED ROUTES FOR PROFIT
