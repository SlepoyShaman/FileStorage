package http

import (
	"context"
	"crypto/tls"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"text/template"
	"time"

	"github.com/SlepoyShaman/FileStorage/common/settings"
	"github.com/SlepoyShaman/FileStorage/database/storage/bolt"
)

// Embed the files in the frontend/dist directory
//
//go:embed embed/*
var assets embed.FS

// Custom dirFS to handle both embedded and non-embedded file systems
type dirFS struct {
	http.Dir
}

// Implement the Open method for dirFS, which wraps http.Dir
func (d dirFS) Open(name string) (fs.File, error) {
	return d.Dir.Open(name)
}

var (
	store   *bolt.BoltStore
	config  *settings.Settings
	assetFs fs.FS
)

func StartHttp(ctx context.Context, storage *bolt.BoltStore, shutdownComplete chan struct{}) {
	store = storage
	config = &settings.Config
	var err error
	// Determine filesystem mode and set asset paths
	if settings.Env.EmbeddedFs {
		// Embedded mode: Serve files from the embedded assets
		assetFs, err = fs.Sub(assets, "embed")
		if err != nil {
			log.Fatal("fs.Sub failed: %v", err)
		}
		entries, err := fs.ReadDir(assetFs, ".")
		if err != nil || len(entries) == 0 {
			log.Fatal("Could not embed frontend. Does dist exist? %v", err)
		}
	} else {
		// Dev mode: Serve files from http/dist directory
		assetFs = dirFS{Dir: http.Dir("http/dist")}
	}

	// In development mode, we want to reload the templates on each request.
	// In production (embedded), we parse them once.
	templates := template.New("").Funcs(template.FuncMap{
		"marshal": func(v interface{}) (string, error) {
			a, err := json.Marshal(v)
			return string(a), err
		},
	})
	if !settings.Env.IsDevMode {
		templates = template.Must(templates.ParseFS(assetFs, "public/index.html"))
	}

	router := http.NewServeMux()
	// API group routing
	api := http.NewServeMux()
	// Public group routing (new structure)
	publicRoutes := http.NewServeMux()

	// Resources routes
	api.HandleFunc("GET /resources", withUser(resourceGetHandler))
	api.HandleFunc("POST /resources", withUser(resourcePostHandler))

	// Mount the route groups
	apiPath := config.Server.BaseURL + "api"
	publicPath := config.Server.BaseURL + "public"
	router.Handle(apiPath+"/", http.StripPrefix(apiPath, api))
	router.Handle(publicPath+"/", http.StripPrefix(publicPath, publicRoutes))

	//

	// redirect to baseUrl if not root
	if config.Server.BaseURL != "/" {
		router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, config.Server.BaseURL, http.StatusMovedPermanently)
		})
	}

	var scheme string
	port := ""
	srv := &http.Server{
		Addr:    fmt.Sprintf("%v:%v", config.Server.ListenAddress, config.Server.Port),
		Handler: muxWithMiddleware(router),
	}
	listenAddress := config.Server.ListenAddress
	if listenAddress == "0.0.0.0" {
		listenAddress = "localhost"
	}
	go func() {
		// Determine whether to use HTTPS (TLS) or HTTP
		if config.Server.TLSCert != "" && config.Server.TLSKey != "" {
			// Load the TLS certificate and key
			cer, err := tls.LoadX509KeyPair(config.Server.TLSCert, config.Server.TLSKey)
			if err != nil {
				log.Fatal("Could not load certificate: %v", err)
			}

			// Create a custom TLS configuration
			tlsConfig := &tls.Config{
				MinVersion:   tls.VersionTLS12,
				Certificates: []tls.Certificate{cer},
			}

			// Set HTTPS scheme and default port for TLS
			scheme = "https"
			if config.Server.Port != 443 {
				port = fmt.Sprintf(":%d", config.Server.Port)
			}

			// Build the full URL with host and port
			fullURL := fmt.Sprintf("%s://%s%s%s", scheme, listenAddress, port, config.Server.BaseURL)
			slog.Info("Running at               : %s", fullURL)

			// Create a TLS listener and serve
			listener, err := tls.Listen("tcp", srv.Addr, tlsConfig)
			if err != nil {
				log.Fatal("Could not start TLS server: %v", err)
			}
			if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
				log.Fatal("Server error: %v", err)
			}
		} else {
			// Set HTTP scheme and the default port for HTTP
			scheme = "http"
			if config.Server.Port != 80 {
				port = fmt.Sprintf(":%d", config.Server.Port)
			}

			// Build the full URL with host and port
			fullURL := fmt.Sprintf("%s://%s%s%s", scheme, listenAddress, port, config.Server.BaseURL)
			slog.Info("Running at               : %s", fullURL)

			// Start HTTP server
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatal("Server error: %v", err)
			}
		}
	}()

	// Wait for context cancellation to shut down the server
	<-ctx.Done()
	slog.Info("Shutting down HTTP server...")

	// Close all SSE sessions

	// Persist in-memory state before shutting down the HTTP server
	if store != nil {
		if store.Share != nil {
			if err := store.Share.Flush(); err != nil {
				slog.Error("Failed to flush share storage: %v", err)
			}
		}
		if store.Access != nil {
			if err := store.Access.Flush(); err != nil {
				slog.Error("Failed to flush access storage: %v", err)
			}
		}
	}

	// Graceful shutdown with a timeout - 30 seconds, in case downloads are happening
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("HTTP server forced to shut down: %v", err)
	} else {
		slog.Info("HTTP server shut down gracefully.")
	}

	// Signal that shutdown is complete
	close(shutdownComplete)
}
