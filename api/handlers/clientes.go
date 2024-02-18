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
	err = json.NewDecoder(r.Body).Decode(&transacaoRequest)
	if err != nil {
		http.Error(w, "Erro ao decodificar o corpo da requisição", http.StatusBadRequest)
		return
	}

	// Validar request
	if transacaoRequest.Valor <= 0 {
		http.Error(w, "'valor' deve ser um número inteiro positivo que representa centavos", http.StatusBadRequest)
		return
	}

	if transacaoRequest.Tipo != "c" && transacaoRequest.Tipo != "d" {
		http.Error(w, "'tipo' deve ser apenas c para crédito ou d para débito.", http.StatusBadRequest)
		return
	}

	if len(transacaoRequest.Descricao) < 1 || len(transacaoRequest.Descricao) > 10 {
		http.Error(w, "'descricao' deve ser uma string de 1 a 10 caracteres.", http.StatusBadRequest)
		return
	}

	resp, err := models.InsertTransacaoSelectForUpdate(idCliente, transacaoRequest)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Cliente não encontrado", http.StatusNotFound)
		} else if err.Error() == "LIMITE_EXCEPTION" {
			http.Error(w, "Transação de débito ultrapassa o limite disponível do cliente", http.StatusUnprocessableEntity)
		} else if err.Error() == "BUSCA_CLIENTE_EXCEPTION" {
			http.Error(w, "Cliente não encontrado", http.StatusNotFound)
		} else {
			log.Printf("Erro inesperado: %v", err)
			http.Error(w, "Erro inesperado", http.StatusInternalServerError)
		}
	}

	// Codificar a resposta como JSON e enviá-la
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func BuscarExtrato(w http.ResponseWriter, r *http.Request) {
	idCliente, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "[id] (na URL) deve ser um número inteiro representando a identificação do cliente", http.StatusBadRequest)
		return
	}

	resp, err := models.GetExtrato(idCliente)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Cliente não encontrado", http.StatusNotFound)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
