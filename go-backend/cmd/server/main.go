package main

import (
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rdruid-talent-analyzer/go-backend/internal/wcl"
	"github.com/rdruid-talent-analyzer/go-backend/internal/web"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	clientID := os.Getenv("WCL_CLIENT_ID")
	clientSecret := os.Getenv("WCL_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		log.Fatal().Msg("WCL_CLIENT_ID and WCL_CLIENT_SECRET must be set")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	inner := wcl.NewClient(clientID, clientSecret)
	cacheDir := "data/cache"
	client := wcl.NewCachedClient(inner, cacheDir)
	resultCacheDir := "data/results"

	apiRouter := web.NewRouter(client, resultCacheDir)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Mount API routes
	r.Mount("/", apiRouter)

	// Serve frontend static files
	frontendDir := findFrontendDir()
	if frontendDir != "" {
		assetsDir := filepath.Join(frontendDir, "assets")
		if info, err := os.Stat(assetsDir); err == nil && info.IsDir() {
			fileServer := http.FileServer(http.Dir(assetsDir))
			r.Handle("/assets/*", http.StripPrefix("/assets/", fileServer))
		}

		// SPA fallback: serve index.html for all unmatched routes
		indexPath := filepath.Join(frontendDir, "index.html")
		if _, err := os.Stat(indexPath); err == nil {
			r.NotFound(func(w http.ResponseWriter, r *http.Request) {
				// Don't serve index.html for API routes
				if strings.HasPrefix(r.URL.Path, "/api/") {
					http.Error(w, `{"detail":"Not found"}`, 404)
					return
				}
				http.ServeFile(w, r, indexPath)
			})
		}
	}

	log.Info().Str("port", port).Msg("Starting server")
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal().Err(err).Msg("Server failed")
	}
}

func findFrontendDir() string {
	candidates := []string{
		"../frontend/dist",
		"frontend/dist",
	}
	for _, dir := range candidates {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			// Check it has at least an index.html
			if _, err := fs.Stat(os.DirFS(dir), "index.html"); err == nil {
				return dir
			}
		}
	}
	return ""
}
