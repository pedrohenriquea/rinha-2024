package handlers

import (
	"api/models"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// RealizarTransacao é um handler para realizar transações
func RealizarTransacao(w http.ResponseWriter, r *http.Request) {
	// Extrair o ID do cliente a partir dos parâmetros da URL
	idCliente, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "[id] (na URL) deve ser um número inteiro representando a identificação do cliente", http.StatusBadRequest)
		return
	}

	// Decodificar o corpo da requisição em uma estrutura de Transacao
	var transacaoRequest models.Transacao
	if err := json.NewDecoder(r.Body).Decode(&transacaoRequest); err != nil {
		http.Error(w, "Erro ao decodificar o corpo da requisição", http.StatusUnprocessableEntity)
		return
	}

	// Validar request
	if transacaoRequest.Valor <= 0 {
		http.Error(w, "'valor' deve ser um número inteiro positivo que representa centavos", http.StatusUnprocessableEntity)
		return
	}

	if transacaoRequest.Tipo != "c" && transacaoRequest.Tipo != "d" {
		http.Error(w, "'tipo' deve ser apenas 'c' para crédito ou 'd' para débito.", http.StatusUnprocessableEntity)
		return
	}

	if len(transacaoRequest.Descricao) < 1 || len(transacaoRequest.Descricao) > 10 {
		http.Error(w, "'descricao' deve ser uma string de 1 a 10 caracteres.", http.StatusUnprocessableEntity)
		return
	}

	// Realizar a transação
	resp, err := models.InsertTransacaoSelectForUpdate(idCliente, transacaoRequest)
	if err != nil {
		switch {
		case err.Error() == "BUSCA_CLIENTE_EXCEPTION":
			http.Error(w, "Cliente não encontrado", http.StatusNotFound)
		case err.Error() == "LIMITE_EXCEPTION":
			http.Error(w, "Transação de débito ultrapassa o limite disponível do cliente", http.StatusUnprocessableEntity)
		default:
			log.Printf("Erro inesperado: %v", err)
			http.Error(w, "Erro inesperado", http.StatusInternalServerError)
		}
		return
	}

	// Codificar a resposta como JSON e enviá-la
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func BuscarExtrato(w http.ResponseWriter, r *http.Request) {
	idCliente, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "[id] (na URL) deve ser um número inteiro representando a identificação do cliente", http.StatusUnprocessableEntity)
		return
	}

	resp, err := models.GetExtrato(idCliente)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Cliente não encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "Erro interno do servidor", http.StatusInternalServerError)
		log.Printf("Erro ao buscar extrato: %v", err)
		return
	}

	// Configurar cabeçalho de resposta
	w.Header().Set("Content-Type", "application/json")

	// Escrever resposta no corpo da resposta
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Erro ao codificar resposta", http.StatusInternalServerError)
		log.Printf("Erro ao codificar resposta JSON: %v", err)
		return
	}
}
