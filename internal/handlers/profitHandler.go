package handlers

import (
	"Brocker-pet-project/internal/repository"
	"encoding/json"
	"go.uber.org/zap"
	"net/http"
)

type ProfitHandler struct {
	repo *repository.ProfitRepository
	log  *zap.Logger
}

func NewProfitHandler(repo *repository.ProfitRepository, log *zap.Logger) *ProfitHandler {
	return &ProfitHandler{repo: repo, log: log}
}

func (h *ProfitHandler) AllClearProfitGET(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		h.log.Error("Invalid request method", zap.String("excepted: ", http.MethodGet), zap.String("got: ", r.Method))
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	if r.Header.Get("Content-Type") != "application/json" {
		h.log.Error("Invalid request Content-Type header", zap.String("excepted: ", "application/json"), zap.String("got: ", r.Header.Get("Content-Type")))
		http.Error(w, "Invalid request Content-Type header", http.StatusUnsupportedMediaType)
		return
	}

	profits := h.repo.GetAllProfitInfo(r.Context())
	if profits == nil {
		h.log.Error("Error reading sql response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	if err := json.NewEncoder(w).Encode(profits); err != nil {
		h.log.Error("Error encoding response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.log.Debug("All clear profit get request successfully handled")

}
