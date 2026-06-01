package handler

import (
	"crypto/tls"
	"os"
	"path/filepath"

	"golang.org/x/crypto/acme/autocert"
)

// newAutocertManager создаёт autocert.Manager, который автоматически получает
// и обновляет TLS-сертификат Let's Encrypt для указанного домена.
// Сертификаты кэшируются в поддиректории внутри os.TempDir().
func newAutocertManager(domain string) *autocert.Manager {
	return &autocert.Manager{
		Cache:      autocert.DirCache(filepath.Join(os.TempDir(), "autocert-cache")),
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(domain),
	}
}

// tlsConfigFor возвращает *tls.Config из менеджера autocert.
func tlsConfigFor(domain string) *tls.Config {
	return newAutocertManager(domain).TLSConfig()
}
