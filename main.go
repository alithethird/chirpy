package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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
	w.Header().Set("Content-Type", "text/html") // normal header
	s := fmt.Sprintf(`<html>
	  <body>
	    <h1>Welcome, Chirpy Admin</h1>
	    <p>Chirpy has been visited %d times!</p>
	  </body>
	</html>`, cfg.fileserverHits.Load())
	w.Write([]byte(s))
}

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, req *http.Request) {
	cfg.fileserverHits.Swap(0)
	s := fmt.Sprintf("Hit counter reset: %d", cfg.fileserverHits.Load())
	w.Write([]byte(s))
}

func validationHandler(w http.ResponseWriter, req *http.Request) {
	type chirpData struct {
		Body string `json:"body"`
	}
	type errorJson struct {
		Error string `json:"error"`
	}
	type validJson struct {
		CleanedBody string `json:"cleaned_body"`
	}
	decoder := json.NewDecoder(req.Body)
	chirp := chirpData{}
	w.Header().Set("Content-Type", "application/json")
	err := decoder.Decode(&chirp)
	if err != nil {
		errVal := errorJson{Error: err.Error()}
		dat, err := json.Marshal(errVal)
		w.WriteHeader(400)
		if err != nil {
			w.Write([]byte("could not marshal error json"))
			return
		}
		w.Write(dat)
		return
	}
	if len(chirp.Body) > 140 {
		errVal := errorJson{Error: "Chirp is too long"}
		dat, err := json.Marshal(errVal)
		w.WriteHeader(400)
		if err != nil {
			w.Write([]byte("could not marshal error json"))
			return
		}
		w.Write(dat)
		return
	}

	strList := strings.Split(chirp.Body, " ")
	var newList []string
	for i := 0; i < len(strList); i++ {
		switch strings.ToLower(strList[i]) {
		case "kerfuffle":
			newList = append(newList, "****")
		case "sharbert":
			newList = append(newList, "****")
		case "fornax":
			newList = append(newList, "****")
		default:
			newList = append(newList, strList[i])
		}
	}

	valid := validJson{CleanedBody: strings.Join(newList, " ")}
	dat, err := json.Marshal(valid)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte("could not marshal error json"))
		return
	}
	w.WriteHeader(200)
	w.Write(dat)
}
func main() {
	piCfg := apiConfig{}
	mux := http.NewServeMux()
	srv := http.Server{Addr: ":8080", Handler: mux}
	mux.Handle("/app/", piCfg.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8") // normal header
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("GET /admin/metrics", piCfg.metricsHandler)
	mux.HandleFunc("POST /admin/reset", piCfg.resetHandler)
	mux.HandleFunc("POST /api/validate_chirp", validationHandler)
	for {
		srv.ListenAndServe()
	}
}
