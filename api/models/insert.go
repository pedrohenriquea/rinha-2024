package models

import (
	"api/db"
	"context"
	"errors"
	"log"
	"strings"
)

func InsertTransacao(idCliente int, transacao Transacao) (id int64, err error) {
	conn, err := db.OpenConnection()
	if err != nil {
		return
	}
	defer conn.Close()

	sql := `INSERT INTO transacoes (cliente_id, valor, tipo, descricao, realizada_em) VALUES($1, $2, $3, $4, now()) RETURNING id`

	err = conn.QueryRow(sql, idCliente, transacao.Valor, transacao.Tipo, transacao.Descricao).Scan(&id)

	return
}

func InsertTransacaoSelectForUpdate(idCliente int, transacaoRequest Transacao) (_ *Cliente, err error) {
	conn, err := db.OpenConnection()
	if err != nil {
		log.Println("1")
		return nil, err
	}
	defer conn.Close()

	// Iniciando a transação
	tx, err := conn.BeginTx(context.Background(), nil)
	if err != nil {
		log.Println("2")
		return nil, err
	}
	defer tx.Rollback()

	// Busca os dados do cliente
	var cliente Cliente
	rowCliente := conn.QueryRow(`SELECT limite, saldo FROM clientes WHERE id=$1 FOR UPDATE`, idCliente)
	err = rowCliente.Scan(&cliente.Limite, &cliente.Saldo)
	if err != nil {
		log.Println("3")
		return nil, errors.New("BUSCA_CLIENTE_EXCEPTION")
	}

	// Saldo
	var saldoTransient = cliente.Saldo

	// Verificar se a transação de débito não ultrapassa o limite disponível
	if strings.EqualFold(transacaoRequest.Tipo, "d") {
		saldoTransient = cliente.Saldo - transacaoRequest.Valor
		if saldoTransient < -cliente.Limite {
			log.Println("4")
			return nil, errors.New("LIMITE_EXCEPTION")
		}
	} else {
		saldoTransient = cliente.Saldo + transacaoRequest.Valor
	}

	// Inserir a transação no banco de dados
	sqlInsertTransacao := `INSERT INTO transacoes (cliente_id, valor, tipo, descricao, realizada_em) VALUES($1, $2, $3, $4, now()) RETURNING id`
	_, err = conn.Exec(sqlInsertTransacao, idCliente, transacaoRequest.Valor, transacaoRequest.Tipo, transacaoRequest.Descricao)
	if err != nil {
		log.Println("5")
		return nil, errors.New("INSERE_TRANSACAO_EXCEPTION")
	}

	// Atualizar o saldo do cliente no banco de dados
	_, err = conn.Exec(`UPDATE clientes SET saldo=$1 WHERE id=$2`, saldoTransient, idCliente)
	if err != nil {
		log.Println("6")
		return nil, errors.New("ATUALIZA_SALDO_EXCEPTION")
	}

	// Preparar a resposta
	resp := Cliente{
		Limite: cliente.Limite,
		Saldo:  saldoTransient,
	}

	// Commit da transação
	if err := tx.Commit(); err != nil {
		log.Println("7")
		return nil, err
	}

	return &resp, nil
}
