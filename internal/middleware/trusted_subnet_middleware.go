package middleware

import (
	"net"
	"net/http"
)

// TrustedSubnetMiddleware отклоняет запросы, у которых X-Real-IP не входит
// в указанную доверенную подсеть. При пустой подсети блокирует всё.
func TrustedSubnetMiddleware(trustedSubnet *net.IPNet) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if trustedSubnet == nil {
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}

			ip := net.ParseIP(r.Header.Get("X-Real-IP"))
			if ip == nil || !trustedSubnet.Contains(ip) {
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
