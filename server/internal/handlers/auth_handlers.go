package handlers

import (
	"github.com/egosha7/site-go/internal/authMiddleware"
	"github.com/egosha7/site-go/internal/domain"
	"go.uber.org/zap"
	"net/http"
	"time"
)

// AuthUser обрабатывает запрос аутентификации пользователя.
func (h *Handler) AuthUser(w http.ResponseWriter, r *http.Request) {
	login := r.FormValue("login")
	password := r.FormValue("password")

	user := &domain.User{
		Login:    login,
		Password: password,
	}

	ip := authMiddleware.GetIP(r)

	token, err := h.Services.GenerateToken(user, ip)
	if err != nil {
		http.Error(w, "Неверная пара логин/пароль", http.StatusUnauthorized)
		h.logger.Error("Failed to check user validity", zap.Error(err))
		return
	}

	h.logger.Info("User authenticated", zap.String("login", user.Login), zap.String("token", token))

	// Устанавливаем токен в cookie
	http.SetCookie(
		w, &http.Cookie{
			Name:     "jwt",
			Value:    token,
			Expires:  time.Now().Add(12 * time.Hour),
			HttpOnly: true,
		},
	)

	// Редирект на страницу /admin/puppy
	http.Redirect(w, r, "/admin/puppies", http.StatusSeeOther)
}
