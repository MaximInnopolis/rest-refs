package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"rest-refs/internal/app/api"
	"rest-refs/internal/app/models"
	"rest-refs/internal/app/repository/postgresql"
)

func (h *Handler) CreateReferralCodeHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("UserID").(int)
	if !ok {
		http.Error(w, "Ошибка аутентификации", http.StatusUnauthorized)
		return
	}

	var input models.ReferralCodeCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Неправильный формат данных", http.StatusBadRequest)
		return
	}

	// Parse expiration date from string to time
	expirationDate, err := time.Parse("02.01.2006", input.ExpirationDate)
	if err != nil {
		http.Error(w, "Неправильный формат даты", http.StatusBadRequest)
		return
	}

	// Create the referral code model
	referralCode := models.ReferralCode{
		ReferrerID: userID,
		Expiration: expirationDate,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	createdCode, err := h.service.CreateReferralCode(referralCode)
	if err != nil {
		if errors.Is(err, api.ErrReferralCodeAlreadyExists) {
			http.Error(w, "Реферальный код уже существует", http.StatusConflict)
		}

		http.Error(w, "Проблема на сервере", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(
		models.ReferralCodeResponse{
			Code:       createdCode.Code,
			Expiration: createdCode.Expiration,
		})
}

func (h *Handler) DeleteReferralCodeHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("UserID").(int)
	if !ok {
		http.Error(w, "Ошибка аутентификации", http.StatusUnauthorized)
		return
	}

	// Call the service to delete the referral code
	err := h.service.DeleteReferralCode(userID)
	if err != nil {
		if errors.Is(err, postgresql.ErrReferralCodeNotFound) {
			http.Error(w, "Реферальный код не найден", http.StatusNotFound)
			return
		}
		http.Error(w, "Проблема на сервере", http.StatusInternalServerError)
		return
	}

	// Send a successful response
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetReferralCodeByEmailHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	email := vars["email"]

	if email == "" {
		http.Error(w, "Email не может быть пустым", http.StatusBadRequest)
		return
	}

	referralCode, err := h.service.GetReferralCodeByReferrerEmail(email)
	if err != nil {
		if errors.Is(err, postgresql.ErrReferralCodeNotFound) {
			http.Error(w, "Реферальный код не найден", http.StatusNotFound)
			return
		}
		http.Error(w, "Проблема на сервере", http.StatusInternalServerError)
		return
	}

	// Отправляем успешный ответ с реферальным кодом
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(referralCode)

}

func (h *Handler) GetReferralsByReferrerIDHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["referrer_id"]

	// Convert ID string to integer
	referrerID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Неправильный формат ID", http.StatusBadRequest)
		return
	}

	referrals, err := h.service.GetReferralsByReferrerID(referrerID)
	if err != nil {
		http.Error(w, "Ошибка получения рефералов", http.StatusInternalServerError)
		return
	}

	if len(referrals) == 0 {
		http.Error(w, "Рефералы не найдены", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(referrals); err != nil {
		http.Error(w, "Ошибка кодирования ответа", http.StatusInternalServerError)
		return
	}

}
