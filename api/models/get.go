package models

import "api/db"

func GetUltimasTransacoes(idCliente int) (transacoes []Transacao, err error) {
	conn, err := db.OpenConnection()
	if err != nil {
		return
	}
	defer conn.Close()

	rows, err := conn.Query(`SELECT id, cliente_id, valor, tipo, descricao, realizada_em FROM transacoes WHERE cliente_id=$1`, idCliente)
	if err != nil {
		return
	}

	for rows.Next() {
		var transacao Transacao

		err = rows.Scan(&transacao.Valor, &transacao.Tipo, &transacao.Descricao)
		if err != nil {
			continue
		}

		transacoes = append(transacoes, transacao)
	}

	return
}

func GetClienteByID(idCliente int) (cliente Cliente, err error) {
	conn, err := db.OpenConnection()
	if err != nil {
		return
	}
	defer conn.Close()

	row := conn.QueryRow(`SELECT limite, saldo FROM clientes WHERE id=$1`, idCliente)

	err = row.Scan(&cliente.Limite, &cliente.Saldo)

	return

}
