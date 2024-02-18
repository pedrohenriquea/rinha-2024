package main

import (
	"api/configs"
	"api/handlers"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {
	err := configs.Load()
	if err != nil {
		panic(err)
	}

	r := chi.NewRouter()
	r.Post("/clientes/{id}/transacoes", handlers.RealizarTransacao)
	r.Get("/clientes/{id}/extrato", handlers.BuscarExtrato)

	http.ListenAndServe(fmt.Sprintf(":%s", configs.GetServerPort()), r)
}
