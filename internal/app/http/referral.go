package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"rest-refs/internal/app/api"
	"rest-refs/internal/app/models"
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

}

func (h *Handler) GetReferralCodeByEmailHandler(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) GetReferralsByReferrerIDHandler(w http.ResponseWriter, r *http.Request) {

}
