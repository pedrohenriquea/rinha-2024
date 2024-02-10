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
		http.Error(w, "ID do cliente inválido", http.StatusNotFound)
		return
	}

	// Decodificar o corpo da requisição em uma estrutura de Transacao
	var transacaoRequest models.Transacao
	err = json.NewDecoder(r.Body).Decode(&transacaoRequest)
	if err != nil {
		http.Error(w, "Erro ao decodificar o corpo da requisição", http.StatusBadRequest)
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
