package handler

import (
	"crypto/tls"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/crypto/acme/autocert"
)

func newAutocertManager(domain, cacheDir string) (*autocert.Manager, error) {
	if cacheDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		cacheDir = filepath.Join(homeDir, ".autocert-cache")
	}

	if err := os.MkdirAll(cacheDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create TLS cache directory %s: %w", cacheDir, err)
	}

	return &autocert.Manager{
		Cache:      autocert.DirCache(cacheDir),
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(domain),
	}, nil
}

func tlsConfigFor(domain, cacheDir string) (*tls.Config, error) {
	m, err := newAutocertManager(domain, cacheDir)
	if err != nil {
		return nil, err
	}
	return m.TLSConfig(), nil
}
