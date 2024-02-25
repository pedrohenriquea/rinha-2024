package models

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

func GetExtrato(idCliente int, dbPool *pgxpool.Pool) (extrato Extrato, err error) {
	ctx := context.Background()

	// Iniciar transação
	tx, err := dbPool.Begin(ctx)
	if err != nil {
		return extrato, err
	}
	defer tx.Rollback(ctx)

	// Obter informações do cliente
	cliente, err := GetClienteByID(ctx, idCliente, tx)
	if err != nil {
		return extrato, err
	}

	// Obter últimas transações do extrato
	ultimasTransacoes, err := GetUltimasTransacoesExtrato(ctx, idCliente, tx)
	if err != nil {
		return extrato, err
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
		return extrato, err
	}

	return extrato, nil
}

func GetClienteByID(ctx context.Context, idCliente int, tx pgx.Tx) (cliente Cliente, err error) {
	// Obter os resultados
	err = tx.QueryRow(ctx, `SELECT limite, saldo FROM clientes WHERE id=$1`, idCliente).Scan(&cliente.Limite, &cliente.Saldo)
	if err != nil {
		return cliente, err
	}

	return cliente, nil
}

func GetUltimasTransacoesExtrato(ctx context.Context, idCliente int, tx pgx.Tx) (transacoes []TransacaoExtrato, err error) {
	// Executar consulta
	rows, err := tx.Query(ctx, `SELECT valor, tipo, descricao, realizada_em FROM transacoes WHERE cliente_id=$1 ORDER BY realizada_em DESC LIMIT 10`, idCliente)
	if err != nil {
		return
	}
	defer rows.Close()

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
