// The Call of Cthulhu character sheet API: a random investigator sheet and the lists the front shows. Small on
// purpose: the standard library only, so the container wakes in milliseconds.
package main

import (
	"encoding/json"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
	"time"

	"github.com/lucasmagalhaees/cthulhu-char-golang/sheet"
)

const base = "/char-generator"

func send(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

// routes builds the API; newRand gives each request its own random source.
func routes(newRand func() *rand.Rand) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET "+base+"/health", func(w http.ResponseWriter, _ *http.Request) {
		send(w, map[string]string{"status": "UP"})
	})
	mux.HandleFunc("GET "+base+"/random-char", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		send(w, sheet.Generate(newRand(), q.Get("stereotype"), q.Get("nativeLanguageRegion"), q.Get("foreignLanguageRegion")))
	})
	mux.HandleFunc("GET "+base+"/random-char/region/native", func(w http.ResponseWriter, _ *http.Request) { send(w, sheet.RegionOptions(true)) })
	mux.HandleFunc("GET "+base+"/random-char/region/foreign", func(w http.ResponseWriter, _ *http.Request) { send(w, sheet.RegionOptions(false)) })
	mux.HandleFunc("GET "+base+"/random-char/stereotype", func(w http.ResponseWriter, _ *http.Request) { send(w, sheet.StereotypeOptions()) })
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, base+"/health", http.StatusFound)
	})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		mux.ServeHTTP(w, r)
	})
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}
	h := routes(func() *rand.Rand { return rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), rand.Uint64())) })
	log.Printf("listening on :%s", port)
	log.Fatal((&http.Server{Addr: ":" + port, Handler: h, ReadHeaderTimeout: 5 * time.Second}).ListenAndServe())
}
