package models

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

func GetExtrato(idCliente int, dbPool *pgxpool.Pool) (extrato Extrato, err error) {
	ctx := context.Background()

	// Iniciando a transação
	tx, err := dbPool.Begin(ctx)
	if err != nil {
		return Extrato{}, err
	}
	defer tx.Rollback(ctx)

	// Obter informações do cliente
	cliente, err := GetClienteByID(idCliente, dbPool)
	if err != nil {
		return
	}

	// Obter últimas transações do extrato
	ultimasTransacoes, err := GetUltimasTransacoesExtrato(idCliente, dbPool)
	if err != nil {
		return
	}

	// Criar extrato
	extrato = Extrato{
		Saldo: SaldoExtrato{
			Total:       cliente.Saldo,
			DataExtrato: time.Now(),
			Limite:      cliente.Limite,
		},
		UltimasTransacoes: ultimasTransacoes,
	}

	// Commit da transação
	if err := tx.Commit(ctx); err != nil {
		return Extrato{}, err
	}

	return extrato, nil
}

func GetClienteByID(idCliente int, dbPool *pgxpool.Pool) (cliente Cliente, err error) {
	ctx := context.Background()

	// Preparar a consulta
	row := dbPool.QueryRow(ctx, `SELECT limite, saldo FROM clientes WHERE id=$1`, idCliente)
	if err != nil {
		return cliente, err
	}
	// Obter os resultados
	err = row.Scan(&cliente.Limite, &cliente.Saldo)
	if err != nil {
		return cliente, err
	}

	return cliente, nil
}

func GetUltimasTransacoesExtrato(idCliente int, dbPool *pgxpool.Pool) (transacoes []TransacaoExtrato, err error) {
	ctx := context.Background()

	// Preparar a consulta
	rows, err := dbPool.Query(ctx, `SELECT valor, tipo, descricao, realizada_em FROM transacoes WHERE cliente_id=$1 ORDER BY realizada_em DESC LIMIT 10`, idCliente)
	if err != nil {
		return
	}

	// Iterar sobre os resultados
	for rows.Next() {
		var transacao TransacaoExtrato
		if err := rows.Scan(&transacao.Valor, &transacao.Tipo, &transacao.Descricao, &transacao.RealizadaEm); err != nil {
			log.Println(err)
			continue
		}
		transacoes = append(transacoes, transacao)
	}
	if err := rows.Err(); err != nil {
		log.Println(err)
	}

	return
}
