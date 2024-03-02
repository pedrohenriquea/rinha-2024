package handlers

import (
	"api/models"
	"api/service"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

func RealizarTransacao(c *fiber.Ctx, dbPool *pgxpool.Pool) {
	// Extrair o ID do cliente a partir dos parâmetros da URL
	idCliente, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		c.Status(fiber.StatusBadRequest).SendString("[id] (na URL) deve ser um número inteiro representando a identificação do cliente")
		return
	}

	// Decodificar o corpo da requisição em uma estrutura de Transacao
	var transacaoRequest models.Transacao
	if err := c.BodyParser(&transacaoRequest); err != nil {
		c.Status(fiber.StatusUnprocessableEntity).SendString("Erro ao decodificar o corpo da requisição")
		return
	}

	// Validar request
	if transacaoRequest.Valor < 0 {
		c.Status(fiber.StatusUnprocessableEntity).SendString("'valor' deve ser um número inteiro positivo que representa centavos")
		return
	}

	if transacaoRequest.Tipo != "c" && transacaoRequest.Tipo != "d" {
		c.Status(fiber.StatusUnprocessableEntity).SendString("'tipo' deve ser apenas 'c' para crédito ou 'd' para débito.")
		return
	}

	if len(transacaoRequest.Descricao) < 1 || len(transacaoRequest.Descricao) > 10 {
		c.Status(fiber.StatusUnprocessableEntity).SendString("'descricao' deve ser uma string de 1 a 10 caracteres.")
		return
	}

	// Realizar a transação
	resp, err := service.InsertTransacao(idCliente, transacaoRequest, dbPool)
	if err != nil {
		switch {
		case err.Error() == "BUSCA_CLIENTE_EXCEPTION":
			c.Status(fiber.StatusNotFound).SendString("Cliente não encontrado")
		case err.Error() == "LIMITE_EXCEPTION":
			c.Status(fiber.StatusUnprocessableEntity).SendString("Transação de débito ultrapassa o limite disponível do cliente")
		default:
			log.Printf("Erro inesperado: %v", err)
			c.Status(fiber.StatusInternalServerError).SendString("Erro inesperado")
		}
		return
	}

	c.Status(fiber.StatusOK).JSON(resp)
}

func BuscarExtrato(c *fiber.Ctx, dbPool *pgxpool.Pool) {
	idCliente, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		c.Status(fiber.StatusUnprocessableEntity).SendString("[id] (na URL) deve ser um número inteiro representando a identificação do cliente")
		return
	}
	resp, err := service.GetExtrato(idCliente, dbPool)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.Status(fiber.StatusNotFound).SendString("Cliente não encontrado")
			return
		}
		if err.Error() == "BUSCA_CLIENTE_EXCEPTION" {
			c.Status(fiber.StatusNotFound).SendString("Cliente não encontrado")
			return
		}
		c.Status(fiber.StatusInternalServerError).SendString("Erro inesperado")
		log.Printf("Erro ao buscar extrato: %v", err)
		return
	}

	c.Status(fiber.StatusOK).JSON(resp)
}
