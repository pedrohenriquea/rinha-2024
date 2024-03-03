package models

import (
	"database/sql"
	"sort"
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

func (item ClienteTransacao) IsSaldoExtrato() bool {
	return !item.IdPai.Valid
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

type TransacaoExtratoRoutine struct {
	Valor       int64         `json:"valor"`
	Tipo        string        `json:"tipo"`
	Descricao   string        `json:"descricao"`
	RealizadaEm time.Time     `json:"realizada_em"`
	Saldo       *SaldoExtrato `json:"saldo"`
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

type ItemExtratoCanal struct {
	SaldoExtrato     SaldoExtrato
	TransacaoExtrato TransacaoExtrato
	IdPai            int64
	RealizadaEm      time.Time
}

func (item ItemExtratoCanal) IsSaldoExtrato() bool {
	// Verifica se o campo SaldoExtrato está preenchido
	return item.SaldoExtrato != SaldoExtrato{}
}

type ByRealizadaEmDesc []TransacaoExtrato

func (a ByRealizadaEmDesc) Len() int           { return len(a) }
func (a ByRealizadaEmDesc) Less(i, j int) bool { return a[i].RealizadaEm.After(a[j].RealizadaEm) }
func (a ByRealizadaEmDesc) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

// Função para ordenar a lista de ItemExtratoCanal por RealizadaEm em ordem decrescente
func OrdenarPorRealizadaEmDesc(lista []TransacaoExtrato) {
	sort.Sort(ByRealizadaEmDesc(lista))
}
