package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/agomesd/chirpy/internal/database"
	"github.com/agomesd/chirpy/internal/handlers"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const PORT = "8080"
const FILE_PATH_ROOT = "/"

type apiConfig struct {
	fileServerHits atomic.Int32
	queries        *database.Queries
	platform       string
	secret         string
	polkaKey       string
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

func (cfg *apiConfig) resetServer(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	cfg.queries.DeleteUsers(r.Context())
	cfg.fileServerHits.Store(0)
}

func main() {
	godotenv.Load()

	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	secret := os.Getenv("SECRET")
	polkaKey := os.Getenv("POLKA_KEY")

	db, err := sql.Open("postgres", dbURL)

	if err != nil {
		log.Fatalf("Failed to open database: %s", err)
	}

	dbQueries := database.New(db)

	mux := http.NewServeMux()

	apiCfg := apiConfig{}

	apiCfg.queries = dbQueries
	apiCfg.platform = platform
	apiCfg.secret = secret
	apiCfg.polkaKey = polkaKey

	userService := handlers.UserService{}
	chirpService := handlers.ChirpService{}
	webhookService := handlers.WebhookService{}

	userService.DB = apiCfg.queries
	userService.JWTSecret = apiCfg.secret

	chirpService.DB = apiCfg.queries
	chirpService.JWTSecret = apiCfg.secret

	webhookService.DB = apiCfg.queries
	webhookService.APIKey = apiCfg.polkaKey

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /api/healthz", healthCheck)
	mux.HandleFunc("GET /admin/metrics", apiCfg.getFileServerHits)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetServer)

	mux.HandleFunc("POST /api/users", userService.HandleCreateUser)
	mux.HandleFunc("PUT /api/users", userService.HandleUpdateUserEmailPassword)
	mux.HandleFunc("POST /api/login", userService.HandleLoginUser)
	mux.HandleFunc("POST /api/refresh", userService.HandleRefreshToken)
	mux.HandleFunc("POST /api/revoke", userService.HandleRevokeToken)

	mux.HandleFunc("POST /api/chirps", chirpService.HandleCreateChirp)
	mux.HandleFunc("GET /api/chirps", chirpService.HandleGetChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", chirpService.HandleGetChirpByID)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", chirpService.HandleDeleteChirp)

	mux.HandleFunc("POST /api/polka/webhooks", webhookService.HandlePolkaWebhook)

	server := http.Server{
		Handler: mux,
		Addr:    ":" + PORT,
	}

	log.Printf("Serving files from %s on port: %s\n", FILE_PATH_ROOT, PORT)
	log.Fatal(server.ListenAndServe())

}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("charset", "utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK\n"))
}
