package models

import (
	"database/sql"
	"time"
)

type ClienteTransacao struct {
	id          int64          `db:"id"`
	Limite      sql.NullInt64  `db:"limite"`
	Saldo       sql.NullInt64  `db:"saldo"`
	IdPai       sql.NullInt64  `db:"id_pai"`
	Valor       sql.NullInt64  `db:"valor"`
	Tipo        sql.NullString `db:"tipo"`
	Descricao   sql.NullString `db:"descricao"`
	RealizadaEm time.Time      `db:"realizada_em"`
}

type Transacao struct {
	Valor     int64  `json:"valor"`
	Tipo      string `json:"tipo"`
	Descricao string `json:"descricao"`
}

type Cliente struct {
	Limite int64 `json:"limite"`
	Saldo  int64 `json:"saldo"`
}

type TransacaoExtrato struct {
	Valor       int64     `json:"valor"`
	Tipo        string    `json:"tipo"`
	Descricao   string    `json:"descricao"`
	RealizadaEm time.Time `json:"realizada_em"`
}

type Extrato struct {
	Saldo             SaldoExtrato       `json:"saldo"`
	UltimasTransacoes []TransacaoExtrato `json:"ultimas_transacoes"`
}

type SaldoExtrato struct {
	Total       int64     `json:"total"`
	DataExtrato time.Time `json:"data_extrato"`
	Limite      int64     `json:"limite"`
}
