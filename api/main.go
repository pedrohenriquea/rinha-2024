package main

import (
	"api/configs"
	"api/handlers"
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v4/pgxpool"
)

func main() {
	err := configs.Load()
	if err != nil {
		panic(err)
	}

	databaseUrl := "postgres://admin:123@localhost:5432/rinha"
	dbPool, err := pgxpool.Connect(context.Background(), databaseUrl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	r := chi.NewRouter()
	r.Post("/clientes/{id}/transacoes", func(w http.ResponseWriter, r *http.Request) {
		handlers.RealizarTransacao(w, r, dbPool)
	})
	r.Get("/clientes/{id}/extrato", func(w http.ResponseWriter, r *http.Request) {
		handlers.BuscarExtrato(w, r, dbPool)
	})

	http.ListenAndServe(fmt.Sprintf(":%s", configs.GetServerPort()), r)
}
