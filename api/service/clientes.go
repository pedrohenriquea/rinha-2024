package service

import (
	"api/models"
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

func InsertTransacao(idCliente int, transacaoRequest models.Transacao, dbPool *pgxpool.Pool) (*models.Cliente, error) {
	ctx := context.Background()

	// Iniciando a transação
	tx, err := dbPool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
			return
		}
		if err := tx.Commit(ctx); err != nil {
			return
		}
	}()

	// Busca os dados do cliente
	var cliente models.Cliente
	var ultimasTransacoes []models.TransacaoExtrato
	if err := tx.QueryRow(ctx, `SELECT limite, saldo, ultimas_transacoes FROM cliente WHERE id=$1 FOR UPDATE`, idCliente).Scan(&cliente.Limite, &cliente.Saldo, &ultimasTransacoes); err != nil {
		return nil, errors.New("BUSCA_CLIENTE_EXCEPTION")
	}

	// Verificar se a transação de débito não ultrapassa o limite disponível
	if strings.EqualFold(transacaoRequest.Tipo, "d") {
		cliente.Saldo -= transacaoRequest.Valor
		if cliente.Saldo < -cliente.Limite {
			return nil, errors.New("LIMITE_EXCEPTION")
		}
	} else {
		cliente.Saldo += transacaoRequest.Valor
	}

	novaTransacao := models.TransacaoExtrato{
		Valor:       transacaoRequest.Valor,
		Tipo:        transacaoRequest.Tipo,
		Descricao:   transacaoRequest.Descricao,
		RealizadaEm: time.Now(),
	}

	ultimasTransacoes = adicionarNovaTransacao(ultimasTransacoes, novaTransacao)

	// Atualizar o saldo do cliente no banco de dados
	if _, err := tx.Exec(ctx, `UPDATE cliente SET saldo=$1, ultimas_transacoes=$2 WHERE id=$3`, cliente.Saldo, ultimasTransacoes, idCliente); err != nil {
		return nil, errors.New("ATUALIZA_SALDO_EXCEPTION")
	}

	// Preparar a resposta
	resp := &models.Cliente{
		Limite: cliente.Limite,
		Saldo:  cliente.Saldo,
	}

	return resp, nil
}

func adicionarNovaTransacao(ultimasTransacoes []models.TransacaoExtrato, novaTransacao models.TransacaoExtrato) []models.TransacaoExtrato {
	// Adicionando o novo registro no início da lista
	ultimasTransacoes = append([]models.TransacaoExtrato{novaTransacao}, ultimasTransacoes...)

	// Garantindo que a lista tenha no máximo 10 registros
	if len(ultimasTransacoes) > 10 {
		ultimasTransacoes = ultimasTransacoes[:10]
	}

	// Ordenando a lista
	sort.Slice(ultimasTransacoes, func(i, j int) bool {
		return ultimasTransacoes[i].Valor > ultimasTransacoes[j].Valor
	})

	return ultimasTransacoes
}

func GetExtrato(idCliente int, dbPool *pgxpool.Pool) (extrato models.Extrato, err error) {

	var cliente models.Cliente
	var ultimasTransacoes []models.TransacaoExtrato
	if err := dbPool.QueryRow(context.Background(), `SELECT limite, saldo, ultimas_transacoes FROM cliente WHERE id=$1 FOR UPDATE`, idCliente).Scan(&cliente.Limite, &cliente.Saldo, &ultimasTransacoes); err != nil {
		return models.Extrato{}, errors.New("BUSCA_CLIENTE_EXCEPTION")
	}
	// Criar extrato
	extrato = models.Extrato{
		Saldo: models.SaldoExtrato{
			Total:       cliente.Saldo,
			DataExtrato: time.Now(),
			Limite:      cliente.Limite,
		},
		UltimasTransacoes: ultimasTransacoes,
	}

	return extrato, nil
}
