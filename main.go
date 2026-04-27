package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"

	"github.com/agomesd/chirpy/internal/handlers"
)

const PORT = "8080"
const FILE_PATH_ROOT = "/"

type apiConfig struct {
	fileServerHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileServerHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) getFileServerHits(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	hits := fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", cfg.fileServerHits.Load())
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(hits))
}

func (cfg *apiConfig) resetFileServerHits(_ http.ResponseWriter, _ *http.Request) {
	cfg.fileServerHits.Store(0)
}

func main() {
	mux := http.NewServeMux()

	apiCfg := apiConfig{}

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /api/healthz", handlers.Healthz)
	mux.HandleFunc("GET /admin/metrics", apiCfg.getFileServerHits)
	mux.HandleFunc("POST /api/validate_chirp", handlers.ValidateChirp)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetFileServerHits)

	server := http.Server{
		Handler: mux,
		Addr:    ":" + PORT,
	}

	log.Printf("Serving files from %s on port: %s\n", FILE_PATH_ROOT, PORT)
	log.Fatal(server.ListenAndServe())

}
