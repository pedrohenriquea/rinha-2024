package models

import (
	"api/db"
	"log"
	"time"
)

func GetExtrato(idCliente int) (extrato Extrato, err error) {
	// Obter informações do cliente
	cliente, err := GetClienteByID(idCliente)
	if err != nil {
		return
	}

	// Obter últimas transações do extrato
	ultimasTransacoes, err := GetUltimasTransacoesExtrato(idCliente)
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

	return extrato, nil
}

func GetClienteByID(idCliente int) (cliente Cliente, err error) {
	conn, err := db.OpenConnection()
	if err != nil {
		return cliente, err
	}
	defer conn.Close()

	// Preparar a consulta
	stmt, err := conn.Prepare(`SELECT limite, saldo FROM clientes WHERE id=$1`)
	if err != nil {
		return cliente, err
	}
	defer stmt.Close()

	// Executar a consulta
	row := stmt.QueryRow(idCliente)

	// Obter os resultados
	err = row.Scan(&cliente.Limite, &cliente.Saldo)
	if err != nil {
		return cliente, err
	}

	return cliente, nil
}

func GetUltimasTransacoesExtrato(idCliente int) (transacoes []TransacaoExtrato, err error) {
	conn, err := db.OpenConnection()
	if err != nil {
		return
	}
	defer conn.Close()

	// Preparar a consulta
	stmt, err := conn.Prepare(`SELECT valor, tipo, descricao, realizada_em FROM transacoes WHERE cliente_id=$1 ORDER BY realizada_em DESC LIMIT 10`)
	if err != nil {
		return
	}
	defer stmt.Close()

	// Executar a consulta
	rows, err := stmt.Query(idCliente)
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
