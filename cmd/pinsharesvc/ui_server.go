package main

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/sys/windows/svc/debug"
)

type UIServer struct {
	config   *ServiceConfig
	eventLog debug.Log
	server   *http.Server
}

func NewUIServer(config *ServiceConfig, eventLog debug.Log) *UIServer {
	return &UIServer{
		config:   config,
		eventLog: eventLog,
	}
}

// Start starts the UI server
func (us *UIServer) Start(ctx context.Context) error {
	mux := http.NewServeMux()

	// Get UI directory path
	uiDir := filepath.Join(us.config.InstallDirectory, "ui")

	// Check if UI directory exists
	if _, err := os.Stat(uiDir); os.IsNotExist(err) {
		us.logError(fmt.Sprintf("UI directory not found at %s", uiDir), nil)
		return fmt.Errorf("UI directory not found: %w", err)
	}

	us.logInfo(fmt.Sprintf("Serving UI from: %s", uiDir))

	// Create file server for static files
	fileSystem := http.Dir(uiDir)
	fileServer := http.FileServer(fileSystem)

	// Proxy handler for API requests
	pinshareProxy := us.createReverseProxy(
		fmt.Sprintf("http://localhost:%d", us.config.PinShareAPIPort),
		"/api",
	)

	ipfsProxy := us.createReverseProxy(
		fmt.Sprintf("http://localhost:%d", us.config.IPFSAPIPort),
		"/ipfs-api",
	)

	// Route handlers
	mux.Handle("/api/", pinshareProxy)
	mux.Handle("/ipfs-api/", ipfsProxy)
	mux.HandleFunc("/", us.createSPAHandler(fileServer, fileSystem))

	// Create server
	us.server = &http.Server{
		Addr:           fmt.Sprintf(":%d", us.config.UIPort),
		Handler:        us.loggingMiddleware(mux),
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	us.logInfo(fmt.Sprintf("Starting UI server on http://localhost:%d", us.config.UIPort))

	// Start server in goroutine
	go func() {
		if err := us.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			us.logError("UI server error", err)
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()
	return nil
}

// Stop gracefully stops the UI server
func (us *UIServer) Stop() {
	if us.server == nil {
		return
	}

	us.logInfo("Stopping UI server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := us.server.Shutdown(ctx); err != nil {
		us.logError("Error shutting down UI server", err)
	} else {
		us.logInfo("UI server stopped")
	}
}

// createReverseProxy creates a reverse proxy for API requests
func (us *UIServer) createReverseProxy(targetURL, pathPrefix string) http.Handler {
	target, _ := url.Parse(targetURL)

	proxy := httputil.NewSingleHostReverseProxy(target)

	// Customize the director to strip the path prefix
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = target.Host

		// Strip the path prefix
		if pathPrefix != "" && strings.HasPrefix(req.URL.Path, pathPrefix) {
			req.URL.Path = strings.TrimPrefix(req.URL.Path, pathPrefix)
			if !strings.HasPrefix(req.URL.Path, "/") {
				req.URL.Path = "/" + req.URL.Path
			}
		}
	}

	// Error handler
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		us.logError(fmt.Sprintf("Proxy error for %s", r.URL.Path), err)
		http.Error(w, "Service temporarily unavailable", http.StatusServiceUnavailable)
	}

	return proxy
}

// createSPAHandler creates a handler for Single Page Application routing
func (us *UIServer) createSPAHandler(fileServer http.Handler, fileSystem http.FileSystem) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Try to serve the requested file
		path := r.URL.Path

		// Check if file exists
		_, err := fileSystem.Open(path)
		if err != nil {
			// File doesn't exist, check if it's a directory with index.html
			if !strings.HasSuffix(path, "/") {
				path = path + "/"
			}
			indexPath := path + "index.html"
			_, err = fileSystem.Open(indexPath)

			if err != nil {
				// Neither file nor directory with index exists
				// Serve index.html for client-side routing (SPA)
				indexFile, err := fileSystem.Open("/index.html")
				if err != nil {
					http.Error(w, "Not found", http.StatusNotFound)
					return
				}
				defer indexFile.Close()

				stat, err := indexFile.Stat()
				if err != nil {
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					return
				}

				// Serve index.html
				if f, ok := indexFile.(fs.File); ok {
					http.ServeContent(w, r, "index.html", stat.ModTime(), f.(interface {
						fs.File
						io.Seeker
					}))
				}
				return
			}
		}

		// File or directory exists, serve it
		fileServer.ServeHTTP(w, r)
	}
}

// loggingMiddleware logs HTTP requests
func (us *UIServer) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code
		ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Call the next handler
		next.ServeHTTP(ww, r)

		// Log the request (only log non-200 or slow requests to reduce noise)
		duration := time.Since(start)
		if ww.statusCode != http.StatusOK || duration > 1*time.Second {
			us.logInfo(fmt.Sprintf("%s %s - %d (%v)", r.Method, r.URL.Path, ww.statusCode, duration))
		}
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Logging helpers
func (us *UIServer) logInfo(msg string) {
	if us.eventLog != nil {
		us.eventLog.Info(1, msg)
	}
}

func (us *UIServer) logError(msg string, err error) {
	errMsg := msg
	if err != nil {
		errMsg = fmt.Sprintf("%s: %v", msg, err)
	}
	if us.eventLog != nil {
		us.eventLog.Error(1, errMsg)
	}
}
