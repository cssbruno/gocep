package main

import (
	"log"
	"net/http"
	"runtime"

	"github.com/cssbruno/gocep/config"
	"github.com/cssbruno/gocep/handlers"

	"github.com/jeffotoni/gcolor"
	"github.com/rs/cors"
)

func main() {
	runtime.GOMAXPROCS(config.NumCPU)

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/cep/{cep...}", handlers.SearchCep)
	mux.HandleFunc("/v1/cep", handlers.NotFound)
	mux.HandleFunc("/", handlers.NotFound)
	muxcors := cors.Default().Handler(mux)
	server := &http.Server{
		Addr:    config.Port,
		Handler: muxcors,
	}
	log.Println(gcolor.YellowCor("Server Run Port"), config.Port)
	log.Println(gcolor.YellowCor("/v1/cep/{cep}"))
	log.Fatal(server.ListenAndServe())
}
