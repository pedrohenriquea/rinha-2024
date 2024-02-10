package models

import "api/db"

func UpdateSaldoCliente(valorSaldo int64, idCliente int) (int64, error) {
	conn, err := db.OpenConnection()
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	res, err := conn.Exec(`UPDATE clientes SET saldo=$1 WHERE id=$2`, valorSaldo, idCliente)
	if err != nil {
		return 0, nil
	}

	return res.RowsAffected()
}
