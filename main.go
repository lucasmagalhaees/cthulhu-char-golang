// The Call of Cthulhu character sheet API: a random investigator sheet and the lists the front shows. Small on
// purpose: the standard library only, so the container wakes in milliseconds.
package main

import (
	"encoding/json"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/lucasmagalhaees/cthulhu-char-golang/sheet"
)

const base = "/char-generator"

// maxBatch is the most sheets one call gives.
const maxBatch = 30

// version is the commit this binary was built from: set by the build (-ldflags "-X main.version=..."), or read from
// Vercel's environment, or "dev".
var version = ""

func currentVersion() string {
	if version != "" {
		return version
	}
	if v := os.Getenv("VERCEL_GIT_COMMIT_SHA"); len(v) >= 7 {
		return v[:7]
	}
	return "dev"
}

func send(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

// fresh marks a random answer as never cached; listing marks a fixed list as cached for a day.
func fresh(w http.ResponseWriter)   { w.Header().Set("Cache-Control", "no-store") }
func listing(w http.ResponseWriter) { w.Header().Set("Cache-Control", "public, max-age=86400") }

// seeded gives the request's own source: from ?seed= when given (the same seed, the same sheets), else newRand.
func seeded(r *http.Request, newRand func() *rand.Rand) (*rand.Rand, bool) {
	s := r.URL.Query().Get("seed")
	if s == "" {
		return newRand(), true
	}
	n, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return nil, false
	}
	return rand.New(rand.NewPCG(n, n^0x9e3779b97f4a7c15)), true
}

// routes builds the API; newRand gives each request its own random source.
func routes(newRand func() *rand.Rand) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET "+base+"/health", func(w http.ResponseWriter, _ *http.Request) {
		fresh(w)
		send(w, map[string]string{"status": "UP", "version": currentVersion()})
	})
	mux.HandleFunc("GET "+base+"/random-char", func(w http.ResponseWriter, r *http.Request) {
		fresh(w)
		src, ok := seeded(r, newRand)
		if !ok {
			http.Error(w, "seed must be a number", http.StatusBadRequest)
			return
		}
		q := r.URL.Query()
		send(w, sheet.Generate(src, q.Get("stereotype"), q.Get("nativeLanguageRegion"), q.Get("foreignLanguageRegion")))
	})
	// n sheets in one call (1 to maxBatch), with the same options as /random-char
	mux.HandleFunc("GET "+base+"/random-chars", func(w http.ResponseWriter, r *http.Request) {
		fresh(w)
		q := r.URL.Query()
		n := 1
		if s := q.Get("n"); s != "" {
			v, err := strconv.Atoi(s)
			if err != nil || v < 1 || v > maxBatch {
				http.Error(w, "n must be 1 to "+strconv.Itoa(maxBatch), http.StatusBadRequest)
				return
			}
			n = v
		}
		src, ok := seeded(r, newRand)
		if !ok {
			http.Error(w, "seed must be a number", http.StatusBadRequest)
			return
		}
		out := make([]sheet.Record, n)
		for i := range out {
			out[i] = sheet.Generate(src, q.Get("stereotype"), q.Get("nativeLanguageRegion"), q.Get("foreignLanguageRegion"))
		}
		send(w, out)
	})
	mux.HandleFunc("GET "+base+"/random-char/region/native", func(w http.ResponseWriter, _ *http.Request) { listing(w); send(w, sheet.RegionOptions(true)) })
	mux.HandleFunc("GET "+base+"/random-char/region/foreign", func(w http.ResponseWriter, _ *http.Request) { listing(w); send(w, sheet.RegionOptions(false)) })
	mux.HandleFunc("GET "+base+"/random-char/stereotype", func(w http.ResponseWriter, _ *http.Request) { listing(w); send(w, sheet.StereotypeOptions()) })
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
