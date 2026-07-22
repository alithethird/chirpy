package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiHandler struct{}

func (apiHandler) ServeHTTP(http.ResponseWriter, *http.Request) {}

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, req *http.Request) {
	s := fmt.Sprintf("Hits: %d", cfg.fileserverHits.Load())
	w.Write([]byte(s))
}

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, req *http.Request) {
	cfg.fileserverHits.Swap(0)
	s := fmt.Sprintf("Hit counter reset: %d", cfg.fileserverHits.Load())
	w.Write([]byte(s))
}

func main() {
	piCfg := apiConfig{}
	mux := http.NewServeMux()
	srv := http.Server{Addr: ":8080", Handler: mux}
	mux.Handle("/app/", piCfg.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8") // normal header
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("/metrics", piCfg.metricsHandler)
	mux.HandleFunc("/reset", piCfg.resetHandler)
	for {
		srv.ListenAndServe()
	}
}
