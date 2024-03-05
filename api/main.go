package main

import (
	"api/configs"
	"api/handlers"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v4/pgxpool"
)

func main() {
	err := configs.Load()
	if err != nil {
		log.Fatalf("Erro ao carregar as configurações: %v", err)
	}

	dbPool, err := connectToDatabase(configs.GetDB())
	if err != nil {
		log.Fatalf("Erro ao conectar ao banco de dados: %v", err)
	}
	defer dbPool.Close()

	app := setupApp(dbPool)

	serverPort := os.Getenv("HTTP_PORT")
	//serverPort := configs.GetServerPort()
	log.Printf("Servidor HTTP iniciado na porta %s...\n", serverPort)
	if err := app.Listen(":" + serverPort); err != nil {
		log.Fatalf("Erro ao iniciar o servidor: %v", err)
	}
}

func connectToDatabase(conf configs.DBConfig) (*pgxpool.Pool, error) {
	connectionString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		conf.Host, conf.Port, conf.User, conf.Pass, conf.Database)

	return pgxpool.Connect(context.Background(), connectionString)
}

func setupApp(dbPool *pgxpool.Pool) *fiber.App {
	app := fiber.New()

	app.Post("/clientes/:id/transacoes", func(c *fiber.Ctx) error {
		handlers.RealizarTransacao(c, dbPool)
		return nil
	})

	app.Get("/clientes/:id/extrato", func(c *fiber.Ctx) error {
		handlers.BuscarExtrato(c, dbPool)
		return nil
	})

	return app
}
