// Package server runs the HTTP(S) listeners and stops them gracefully.
//
// TLS_MODE decides the shape:
//
//	off    one plain HTTP listener on APP_PORT (a proxy/platform adds HTTPS)
//	auto   HTTPS on TLS_PORT with Let's Encrypt certificates, obtained and
//	       renewed automatically, plus HTTP on TLS_HTTP_PORT that answers
//	       Let's Encrypt's checks and redirects everything else to HTTPS
//	files  HTTPS on TLS_PORT with your certificate files, plus the redirect
package server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"ecommerce-backend/internal/config"

	"golang.org/x/crypto/acme/autocert"
)

// ShutdownTimeout is how long running requests get to finish on stop.
const ShutdownTimeout = 10 * time.Second

// Run serves h until ctx is cancelled (Ctrl+C / SIGTERM) or a listener fails.
func Run(ctx context.Context, cfg config.Config, h http.Handler) error {
	api := newServer(h)
	var redirect *http.Server // the :80 listener when TLS is on

	switch cfg.TLS.Mode {
	case config.TLSAuto:
		if err := os.MkdirAll(cfg.TLS.CacheDir, 0o700); err != nil {
			return fmt.Errorf("TLS_CACHE_DIR: %w", err)
		}
		m := &autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(cfg.TLS.Domains...), // never request certificates for other names
			Cache:      autocert.DirCache(cfg.TLS.CacheDir),        // keep certificates across restarts (Let's Encrypt has rate limits)
			Email:      cfg.TLS.Email,
		}
		api.TLSConfig = m.TLSConfig()
		redirect = newServer(m.HTTPHandler(RedirectToHTTPS(cfg.TLS.HTTPSPort)))
	case config.TLSFiles:
		// Fail at start-up, not on the first visitor, when the files are wrong.
		cert, err := tls.LoadX509KeyPair(cfg.TLS.CertFile, cfg.TLS.KeyFile)
		if err != nil {
			return fmt.Errorf("load TLS_CERT_FILE / TLS_KEY_FILE: %w", err)
		}
		api.TLSConfig = &tls.Config{Certificates: []tls.Certificate{cert}}
		if cfg.TLS.HTTPPort != "off" {
			redirect = newServer(RedirectToHTTPS(cfg.TLS.HTTPSPort))
		}
	}
	if api.TLSConfig != nil {
		api.TLSConfig.MinVersion = tls.VersionTLS12
		api.Addr = ":" + cfg.TLS.HTTPSPort
	} else {
		api.Addr = ":" + cfg.Port
	}

	errs := make(chan error, 2)
	go func() {
		slog.Info("API listening", "addr", api.Addr, "https", api.TLSConfig != nil, "tls_mode", cfg.TLS.Mode)
		if api.TLSConfig != nil {
			errs <- api.ListenAndServeTLS("", "") // certificates come from TLSConfig
		} else {
			errs <- api.ListenAndServe()
		}
	}()
	if redirect != nil {
		redirect.Addr = ":" + cfg.TLS.HTTPPort
		go func() {
			slog.Info("redirecting HTTP to HTTPS", "addr", redirect.Addr)
			errs <- redirect.ListenAndServe()
		}()
	}

	var runErr error
	select {
	case err := <-errs:
		if !errors.Is(err, http.ErrServerClosed) {
			runErr = err
		}
	case <-ctx.Done():
		slog.Info("shutting down")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), ShutdownTimeout)
	defer cancel()
	err := api.Shutdown(shutdownCtx)
	if redirect != nil {
		err = errors.Join(err, redirect.Shutdown(shutdownCtx))
	}
	return errors.Join(runErr, err)
}

func newServer(h http.Handler) *http.Server {
	return &http.Server{
		Handler:           h,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       60 * time.Second, // whole request incl. body: room for a 5 MB photo on a slow phone connection
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}

// RedirectToHTTPS sends every plain-HTTP request to the same address over
// HTTPS. 308 keeps the method and body (a POST stays a POST).
func RedirectToHTTPS(httpsPort string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := r.Host
		if h, _, err := net.SplitHostPort(host); err == nil {
			host = h
		}
		if host == "" {
			http.Error(w, "use HTTPS", http.StatusBadRequest)
			return
		}
		if httpsPort != "" && httpsPort != "443" {
			host = net.JoinHostPort(host, httpsPort)
		}
		http.Redirect(w, r, "https://"+host+r.URL.RequestURI(), http.StatusPermanentRedirect)
	})
}
