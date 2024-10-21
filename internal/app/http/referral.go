package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

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
