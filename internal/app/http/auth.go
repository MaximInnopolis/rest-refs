package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"rest-refs/internal/app/api"
	"rest-refs/internal/app/models"
	"rest-refs/internal/app/repository/postgresql"
)

func (h *Handler) RegisterUserHandler(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("RegisterUserHandler[http]: Регистрация пользователя")

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
	err := h.service.Authorization.RegisterUser(user)
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
	h.logger.Infof("RegisterUserHandler[http]: Регистрация пользователя прошла успешно")
}

// LoginUserHandler handles user login requests
// It parses request body to get email and password, generates token
// and responds with token if successful
func (h *Handler) LoginUserHandler(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("LoginUserHandler[http]: Логин пользователя")

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

	h.logger.Debugf("LoginUserHandler[http]: Логин пользователя прошел успешно")
}

func (h *Handler) RegisterWithReferralHandler(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("RegisterWithReferralHandler[http]: Регистрация реферала")

	var input models.RegisterRequest

	// Decode request body into input struct
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Неправильный формат данных", http.StatusBadRequest)
		return
	}

	if input.ReferralCode == "" {
		http.Error(w, "Реферальный код не может быть пустым", http.StatusBadRequest)
		return
	}

	user := models.User{
		Email:    input.Email,
		Password: input.Password,
	}

	// Attempt to register referral using service
	err := h.service.Referral.RegisterWithReferralCode(input.ReferralCode, user)
	if err != nil {
		if errors.Is(err, postgresql.ErrReferralCodeNotFound) {
			http.Error(w, "Введенный реферальный код не существует", http.StatusNotFound)
			return
		}

		if errors.Is(err, api.ErrReferralCodeNotActive) {
			http.Error(w, "Введенный реферальный код неактивен", http.StatusBadRequest)
			return
		}

		if errors.Is(err, api.ErrUserAlreadyExists) {
			http.Error(w, "Пользователь с указанными данными уже зарегистрирован", http.StatusConflict)
			return
		}

		http.Error(w, "Проблема на сервере", http.StatusInternalServerError)
		return
	}

	// Respond with created user
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	h.logger.Infof("RegisterUserHandler[http]: Регистрация реферала прошла успешно")
}
