package models

import (
	"time"
)

type Transacao struct {
	Valor     int64  `json:"valor"`
	Tipo      string `json:"tipo"`
	Descricao string `json:"descricao"`
}

type Cliente struct {
	Limite int64 `json:"limite"`
	Saldo  int64 `json:"saldo"`
}

type ClienteTransacoes struct {
	Cliente           Cliente
	UltimasTransacoes []TransacaoExtrato
	Versao            int64
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
