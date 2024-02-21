package models

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

func GetExtrato(idCliente int, dbPool *pgxpool.Pool) (extrato Extrato, err error) {
 
    // Obter informações do cliente
    cliente, err := GetClienteByID(idCliente, dbPool)
    if err != nil {
        return
    }
 
    // Obter saldo
    saldo, err := GetSaldo(idCliente, dbPool)
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
            Total:       saldo,
            DataExtrato: time.Now(),
            Limite:      cliente.Limite,
        },
        UltimasTransacoes: ultimasTransacoes,
    }
 
    return extrato, nil
}

func GetSaldo(idCliente int, dbPool *pgxpool.Pool) (saldo int64, err error) {
	ctx := context.Background()

	// Preparar a consulta
	row := dbPool.QueryRow(ctx, `SELECT COALESCE(SUM(valor), 0) FROM transacoes WHERE cliente_id=$1`, idCliente)
	if err != nil {
		return saldo, err
	}
	// Obter os resultados
	err = row.Scan(&saldo)
	if err != nil {
		return saldo, err
	}

	return saldo, nil
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
