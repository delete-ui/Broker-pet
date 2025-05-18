package handlers

import (
	"Brocker-pet-project/internal/models"
	"Brocker-pet-project/internal/repository"
	"Brocker-pet-project/pkg/jwt"
	"encoding/json"
	"go.uber.org/zap"
	"net/http"
)

type UserHandler struct {
	repo *repository.UserRepository
	log  *zap.Logger
}

func NewUserHandler(repo *repository.UserRepository, log *zap.Logger) *UserHandler {
	return &UserHandler{repo: repo, log: log}
}

func (h *UserHandler) NewUserPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.log.Error("Invalid request method", zap.String("excepted: ", http.MethodPost), zap.String("got: ", r.Method))
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	if r.Header.Get("content-type") != "application/json" {
		h.log.Error("Invalid content type", zap.String("excepted: ", "application/json"), zap.String("got: ", r.Header.Get("content-type")))
		http.Error(w, "Invalid content type", http.StatusUnsupportedMediaType)
		return
	}

	var user models.User

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		h.log.Error("Error decoding user", zap.Error(err))
		http.Error(w, "Invalid server error", http.StatusInternalServerError)
		return
	}

	userResponse := h.repo.NewUser(user.Username, user.Password)

	if userResponse == nil {
		h.log.Error("Error creating new user")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	if err := json.NewEncoder(w).Encode(userResponse); err != nil {
		h.log.Error("Error encoding response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.log.Debug("User post request successfully handled", zap.Int64("id: ", user.Id), zap.String(" username: ", user.Username))

}

func (h *UserHandler) LoginIn(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		h.log.Error("Invalid request method", zap.String("excepted: ", http.MethodGet), zap.String("got: ", r.Method))
		http.Error(w, "Invalid request method ", http.StatusMethodNotAllowed)
		return
	}

	var user models.User

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		h.log.Error("Error decoding user", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	userResponse := h.repo.GetUserByUsername(user.Username, user.Password)
	if userResponse == nil {
		h.log.Error("Error getting user by username", zap.String("username: ", user.Username))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	token, err := jwt.GenerateToken(userResponse.Id)
	if err != nil {
		h.log.Error("Failed to generate token", zap.Error(err))
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})

	h.log.Debug("User get request successfully handled", zap.String("username: ", user.Username))

}
