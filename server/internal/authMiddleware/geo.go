package authMiddleware

import (
	"go.uber.org/zap"
	"net"
	"net/http"
	"strings"
)

type MiddlewareRobotTaskDenied struct {
	logger *zap.Logger
}

func NewMiddlewareRobotTaskDenied(logger *zap.Logger) *MiddlewareRobotTaskDenied {
	return &MiddlewareRobotTaskDenied{logger: logger}
}

func (m *MiddlewareRobotTaskDenied) RequestToIP(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if isIPAddress(r.Host) {
				m.logger.Warn("Access denied", zap.String("Host", r.Host))
				http.Error(w, "Access denied", http.StatusForbidden)
				return
			}
			// Proceed to the next middleware/handler
			next.ServeHTTP(w, r)
		},
	)
}

func isIPAddress(host string) bool {
	// Remove port if present
	host = strings.Split(host, ":")[0]
	return net.ParseIP(host) != nil
}
