package handlers

import (
	"encoding/json"
	"go.uber.org/zap"
	"log"
	"net/http"
)

type EmailRequest struct {
	Email string `json:"email"`
}

// AddEmail обрабатывает запрос на запись электронной почты.
func (h *Handler) AddEmail(w http.ResponseWriter, r *http.Request) {
	// Декодирование JSON-запроса
	var emailReq EmailRequest
	if err := json.NewDecoder(r.Body).Decode(&emailReq); err != nil {
		log.Printf("Failed to decode JSON: %v", err)
		http.Error(w, "Failed to decode JSON", http.StatusBadRequest)
		return
	}

	if err := h.Services.AddEmail(emailReq.Email); err != nil {
		h.logger.Error("Ошибка при регистрации пользователя", zap.Error(err))
		http.Error(w, "Ошибка при регистрации пользователя", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
