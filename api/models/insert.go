package models

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v4/pgxpool"
)

func InsertTransacaoSelectForUpdate(idCliente int, transacaoRequest Transacao, dbPool *pgxpool.Pool) (_ *Cliente, err error) {
	ctx := context.Background()

	// Iniciando a transação
	tx, err := dbPool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Busca os dados do cliente
	var cliente Cliente
	rowCliente := dbPool.QueryRow(ctx, `SELECT limite, saldo FROM clientes WHERE id=$1 FOR UPDATE`, idCliente)
	err = rowCliente.Scan(&cliente.Limite, &cliente.Saldo)
	if err != nil {
		return nil, errors.New("BUSCA_CLIENTE_EXCEPTION")
	}

	// Saldo
	var saldoTransient = cliente.Saldo

	// Verificar se a transação de débito não ultrapassa o limite disponível
	if strings.EqualFold(transacaoRequest.Tipo, "d") {
		saldoTransient = cliente.Saldo - transacaoRequest.Valor
		if saldoTransient < -cliente.Limite {
			return nil, errors.New("LIMITE_EXCEPTION")
		}
	} else {
		saldoTransient = cliente.Saldo + transacaoRequest.Valor
	}

	// Inserir a transação no banco de dados
	sqlInsertTransacao := `INSERT INTO transacoes (cliente_id, valor, tipo, descricao, realizada_em) VALUES($1, $2, $3, $4, now()) RETURNING id`
	_, err = dbPool.Exec(ctx, sqlInsertTransacao, idCliente, transacaoRequest.Valor, transacaoRequest.Tipo, transacaoRequest.Descricao)
	if err != nil {
		return nil, errors.New("INSERE_TRANSACAO_EXCEPTION")
	}

	// Atualizar o saldo do cliente no banco de dados
	_, err = dbPool.Exec(ctx, `UPDATE clientes SET saldo=$1 WHERE id=$2`, saldoTransient, idCliente)
	if err != nil {
		return nil, errors.New("ATUALIZA_SALDO_EXCEPTION")
	}

	// Preparar a resposta
	resp := Cliente{
		Limite: cliente.Limite,
		Saldo:  saldoTransient,
	}

	// Commit da transação
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &resp, nil
}
