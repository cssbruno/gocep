package main

import (
	"log"
	"net/http"

	"github.com/cssbruno/gocep/pkg/cep"
	"github.com/cssbruno/gocep/pkg/util"

	"github.com/rs/cors"
)

var (
	Port = "0.0.0.0:8080"
)

func main() {

	mux := http.NewServeMux()
	mux.HandleFunc("/cep/{cep...}", HandlerCEP)
	mux.HandleFunc("/cep", NotFound)
	mux.HandleFunc("/", NotFound)
	muxcors := cors.Default().Handler(mux)
	server := &http.Server{
		Addr:    Port,
		Handler: muxcors,
	}

	log.Println("Run My Server ", Port)
	log.Fatal(server.ListenAndServe())
}

func HandlerCEP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	cepstr := r.PathValue("cep")
	if err := util.CheckCEP(cepstr); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	result, address, err := cep.Search(cepstr)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !cep.ValidCEP(address) {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var b []byte
	if len(result) > 0 {
		b = []byte(result)
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(b)
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusFound)
	return
}
