package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"rest-refs/internal/app/api"
	"rest-refs/internal/app/models"
)

func (h *Handler) RegisterUserHandler(w http.ResponseWriter, r *http.Request) {
	var input models.RegisterRequest

	// Decode request body into input struct
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Неправильный формат данных", http.StatusBadRequest)
		return
	}

	user := models.User{
		Email:    input.Email,
		Password: input.Password,
	}

	// Attempt to create user using service
	err := h.service.Authorization.CreateUser(user)
	if err != nil {
		if errors.Is(err, api.ErrUserAlreadyExists) {
			http.Error(w, "Такой пользователь уже существует", http.StatusConflict)
			return
		}

		http.Error(w, "Проблема на сервере", http.StatusInternalServerError)
		return
	}

	// Respond with created user
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}

// LoginUserHandler handles user login requests
// It parses request body to get email and password, generates token
// and responds with token if successful
func (h *Handler) LoginUserHandler(w http.ResponseWriter, r *http.Request) {
	var input models.LoginRequest

	// Decode request body into input struct
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Неправильный формат данных", http.StatusBadRequest)
		return
	}

	user := models.User{
		Email:    input.Email,
		Password: input.Password,
	}

	// Attempt to generate token using service
	token, err := h.service.Authorization.GenerateToken(user)
	if err != nil {
		http.Error(w, "Проблема на сервере", http.StatusInternalServerError)
		return
	}
	// Respond with created user
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(token)
}

func (h *Handler) RegisterWithReferralHandler(w http.ResponseWriter, r *http.Request) {

}
