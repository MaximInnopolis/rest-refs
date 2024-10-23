package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"rest-refs/internal/app/api"
	"rest-refs/internal/app/models"
	"rest-refs/internal/app/repository/postgresql"
)

// CreateReferralCodeHandler creates a new referral code
// @Summary Create a new referral code
// @Description Creates a referral code for the authenticated user
// @Tags referral_code
// @Accept  json
// @Produce  json
// @Param ReferralCodeCreateRequest body models.ReferralCodeCreateRequest true "Referral code request"
// @Success 201 {object} models.ReferralCodeResponse "Referral code created"
// @Failure 400 {string} string "Invalid data format or date"
// @Failure 401 {string} string "Authentication error"
// @Failure 409 {string} string "Referral code already exists"
// @Failure 500 {string} string "Server error"
// @Router /referral_code [post]
func (h *Handler) CreateReferralCodeHandler(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("CreateReferralCodeHandler[http]: Создание реферального кода")

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

	if expirationDate.Before(time.Now()) {
		http.Error(w, "Срок годности реферального кода не может быть в прошлом", http.StatusBadRequest)
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
			http.Error(w, "Активный реферальный код уже существует", http.StatusConflict)
			return
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

	h.logger.Debugf("CreateReferralCodeHandler[http]: Реферальный код успешно создан")
}

// DeleteReferralCodeHandler deletes the active referral code
// @Summary Delete an active referral code
// @Description Deletes the referral code of the authenticated user
// @Tags referral_code
// @Success 204 "Referral code deleted"
// @Failure 401 {string} string "Authentication error"
// @Failure 404 {string} string "Referral code not found"
// @Failure 500 {string} string "Server error"
// @Router /referral_code [delete]
func (h *Handler) DeleteReferralCodeHandler(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("DeleteReferralCodeHandler[http]: Удаление реферального кода")

	userID, ok := r.Context().Value("UserID").(int)
	if !ok {
		http.Error(w, "Ошибка аутентификации", http.StatusUnauthorized)
		return
	}

	// Call the service to delete the referral code
	err := h.service.DeleteReferralCode(userID)
	if err != nil {
		if errors.Is(err, postgresql.ErrReferralCodeNotFound) {
			http.Error(w, "Активный реферальный код не найден", http.StatusNotFound)
			return
		}
		http.Error(w, "Проблема на сервере", http.StatusInternalServerError)
		return
	}

	// Send a successful response
	w.WriteHeader(http.StatusNoContent)

	h.logger.Debugf("DeleteReferralCodeHandler[http]: Реферальный код успешно удален")
}

// GetReferralCodeByEmailHandler retrieves referral code by referrer email
// @Summary Get referral code by referrer email
// @Description Retrieves the referral code by the email of the referrer
// @Tags referral_code
// @Produce  json
// @Param email path string true "Referrer email"
// @Success 200 {object} models.ReferralCodeResponse "Referral code found"
// @Failure 400 {string} string "Email cannot be empty"
// @Failure 404 {string} string "Referral code not found"
// @Failure 500 {string} string "Server error"
// @Router /referral_code/email/{email} [get]
func (h *Handler) GetReferralCodeByEmailHandler(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("GetReferralCodeByEmailHandler[http]: Получение реферального кода по email реферера")

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
	json.NewEncoder(w).Encode(
		models.ReferralCodeResponse{
			Code:       referralCode.Code,
			Expiration: referralCode.Expiration,
		})

	h.logger.Debugf("GetReferralCodeByEmailHandler[http]: Реферальный код успешно получен по email реферера")
}
