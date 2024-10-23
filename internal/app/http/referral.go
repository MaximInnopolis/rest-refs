package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// GetReferralsByReferrerIDHandler retrieves referrals based on referrer ID
// @Summary Get referrals by referrer ID
// @Description Retrieves a list of referrals based on the referrer's ID
// @Tags referral
// @Accept  json
// @Produce  json
// @Param referrer_id path int true "Referrer ID"
// @Success 200 {array} models.ReferralInfoResponse "List of referrals"
// @Failure 400 {string} string "Invalid ID format"
// @Failure 404 {string} string "Referrals not found"
// @Failure 500 {string} string "Internal server error"
// @Router /referral/id/{referrer_id} [get]
func (h *Handler) GetReferralsByReferrerIDHandler(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("GetReferralsByReferrerIDHandler[http]: Получение рефералов по id реферера")

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

	h.logger.Debugf("GetReferralsByReferrerIDHandler[http]: Рефералы успешно получены по id реферера")
}
