package models

type Transacao struct {
	Valor     int64  `json:"valor"`
	Tipo      string `json:"tipo"`
	Descricao string `json:"descricao"`
}

type Cliente struct {
	Limite int64 `json:"limite"`
	Saldo  int64 `json:"saldo"`
}
