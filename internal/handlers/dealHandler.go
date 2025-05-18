package handlers

import (
	"Brocker-pet-project/internal/models"
	"Brocker-pet-project/internal/repository"
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type DealHandler struct {
	repo      *repository.DealRepository
	redisRepo *redis.Client
	log       *zap.Logger
}

func NewDealHandler(repo *repository.DealRepository, redisRepo *redis.Client, log *zap.Logger) *DealHandler {
	return &DealHandler{repo: repo, redisRepo: redisRepo, log: log}
}

func (h *DealHandler) NewDealPost(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		h.log.Error("Invalid request method", zap.String("excepted: ", http.MethodPost), zap.String("got: ", r.Method))
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if r.Header.Get("content-type") != "application/json" {
		h.log.Error("Invalid content type", zap.String("excepted: ", "application/json"), zap.String("got: ", r.Header.Get("content-type")))
		http.Error(w, "Invalid media type", http.StatusUnsupportedMediaType)
		return
	}

	var deal models.Deal

	if err := json.NewDecoder(r.Body).Decode(&deal); err != nil {
		h.log.Error("Error decoding deal", zap.Error(err))
		http.Error(w, "Invalid server error", http.StatusInternalServerError)
		return
	}

	dealResponse := *h.repo.CreateNewDeal(deal.Title, deal.Expenses, deal.Profit)

	ctx := context.Background()
	h.redisRepo.Del(ctx, "notProcessedDeals:all", "processedDeals:all", "allDeals:get")

	w.Header().Set("content-type", "application/json")
	if err := json.NewEncoder(w).Encode(dealResponse); err != nil {
		h.log.Error("Error encoding deal", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.log.Debug("New deal post request successfully handled ", zap.Int64("deal id: ", dealResponse.Id))
}

func (h *DealHandler) AllProcessedDealsGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.log.Error("Invalid request method", zap.String("excepted: ", http.MethodGet), zap.String("got: ", r.Method))
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := context.Background()
	cacheKey := "processedDeals:all"

	cachedDeals, err := h.redisRepo.Get(ctx, cacheKey).Bytes()
	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write(cachedDeals)
		h.log.Debug("Served by redis cache")
		return
	}

	var deals *[]models.Deal

	deals = h.repo.GetAllProcessedDeals(r.Context())

	tasksJSON, _ := json.Marshal(deals)
	h.redisRepo.Set(ctx, cacheKey, tasksJSON, 5*time.Minute)

	w.Header().Set("content-type", "application/json")
	if err := json.NewEncoder(w).Encode(deals); err != nil {
		h.log.Error("Error encoding deals", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.log.Debug("Get all processed deals GET request successfully handled")

}

func (h *DealHandler) AllNotProcessedDealsGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.log.Error("Invalid request method", zap.String("excepted: ", http.MethodGet), zap.String("got: ", r.Method))
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := context.Background()
	cacheKey := "notProcessedDeals:all"

	cachedDeals, err := h.redisRepo.Get(ctx, cacheKey).Bytes()
	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write(cachedDeals)
		h.log.Debug("Served by redis cache")
		return
	}

	var deals *[]models.Deal

	deals = h.repo.GetAllNotProcessedDeals(r.Context())

	tasksJSON, _ := json.Marshal(deals)
	h.redisRepo.Set(ctx, cacheKey, tasksJSON, 5*time.Minute)

	w.Header().Set("content-type", "application/json")
	if err := json.NewEncoder(w).Encode(deals); err != nil {
		h.log.Error("Error encoding deals", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.log.Debug("Get all not processed deals GET request successfully handled")

}

func (h *DealHandler) AllDealsGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.log.Error("Invalid request method", zap.String("excepted: ", http.MethodGet), zap.String("got: ", r.Method))
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := context.Background()
	cacheKey := "allDeals:get"

	cachedDeals, err := h.redisRepo.Get(ctx, cacheKey).Bytes()
	if err == nil {
		w.Header().Set("content-type", "application/json")
		w.Write(cachedDeals)
		h.log.Debug("Served by redis cache")
		return
	}

	var deals *[]models.Deal

	deals = h.repo.GetAllDeals(r.Context())

	tasksJSON, _ := json.Marshal(deals)
	h.redisRepo.Set(ctx, cacheKey, tasksJSON, 5*time.Minute)

	w.Header().Set("content-type", "application/json")
	if err := json.NewEncoder(w).Encode(deals); err != nil {
		h.log.Error("Error encoding deals", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.log.Debug("Get all deals GET request successfully handled")

}
