package authMiddleware

import (
	"context"
	"github.com/dgrijalva/jwt-go"
	"go.uber.org/zap"
	"net"
	"net/http"
	"strings"
	"time"
)

type contextKey string

const UserCtxKey contextKey = "user"

const (
	salt       = "demo"
	SigningKey = "demo#4#%demo#demo"
	TokenTTL   = 12 * time.Hour
)

type TokenClaims struct {
	jwt.StandardClaims
	UserID int    `json:"user_id"`
	IP     string `json:"ip"`
}

type MiddlewareJWT struct {
	logger     *zap.Logger
	signingKey []byte
}

func NewMiddlewareJWT(logger *zap.Logger, signingKey string) *MiddlewareJWT {
	return &MiddlewareJWT{
		logger:     logger,
		signingKey: []byte(signingKey),
	}
}

func (m *MiddlewareJWT) JWTAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					m.logger.Error("panic recovered in JWTAuth middleware", zap.Any("error", err))
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}
			}()

			cookie, err := r.Cookie("jwt")
			if err != nil {
				m.logger.Info("JWT cookie not found", zap.Error(err))
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			tokenString := cookie.Value

			token, err := jwt.ParseWithClaims(
				tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
					if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, http.ErrAbortHandler
					}
					return m.signingKey, nil
				},
			)

			if err != nil {
				m.logger.Error("Failed to parse token", zap.Error(err))
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			if claims, ok := token.Claims.(*TokenClaims); ok && token.Valid {
				if claims.IP != GetIP(r) {
					m.logger.Warn(
						"Token IP does not match request IP", zap.String("tokenIP", claims.IP),
						zap.String("requestIP", GetIP(r)),
					)
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
				ctx := context.WithValue(r.Context(), UserCtxKey, claims.UserID)
				next.ServeHTTP(w, r.WithContext(ctx))
			} else {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
			}
		},
	)
}

func GetIP(r *http.Request) string {
	// Try to get the IP from the X-Forwarded-For header if it exists
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ips := strings.Split(forwarded, ",")
		// Return the first IP address in the X-Forwarded-For header
		return strings.TrimSpace(ips[0])
	}
	// Fallback to using the remote address
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// If there's an error, just return the remote address as is
		return r.RemoteAddr
	}
	return ip
}
