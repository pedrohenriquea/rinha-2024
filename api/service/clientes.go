package service

import (
	"api/models"
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

func BuscarClienteTx(idCliente int, tx pgx.Tx) (clienteTransacoes models.ClienteTransacoes, err error) {
	ctx := context.Background()

	if _, err := tx.Prepare(ctx, "busca-cliente-tx", "SELECT limite, saldo, ultimas_transacoes FROM cliente WHERE id=$1 FOR UPDATE"); err != nil {
		panic(err)
	}

	// Busca os dados do cliente
	if err := tx.QueryRow(ctx, "busca-cliente-tx", idCliente).Scan(&clienteTransacoes.Cliente.Limite, &clienteTransacoes.Cliente.Saldo, &clienteTransacoes.UltimasTransacoes); err != nil {
		return models.ClienteTransacoes{}, errors.New("BUSCA_CLIENTE_EXCEPTION")
	}

	return clienteTransacoes, nil
}

func BuscarCliente(idCliente int, dbPool *pgxpool.Pool) (clienteTransacoes models.ClienteTransacoes, err error) {
	ctx := context.Background()

	// Busca os dados do cliente
	if err := dbPool.QueryRow(ctx, "SELECT limite, saldo, ultimas_transacoes FROM cliente WHERE id=$1", idCliente).Scan(&clienteTransacoes.Cliente.Limite, &clienteTransacoes.Cliente.Saldo, &clienteTransacoes.UltimasTransacoes); err != nil {
		return models.ClienteTransacoes{}, errors.New("BUSCA_CLIENTE_EXCEPTION")
	}

	return clienteTransacoes, nil
}

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
	clienteDB, err := BuscarClienteTx(idCliente, tx)
	if err != nil {
		return nil, err
	}
	cliente := clienteDB.Cliente
	ultimasTransacoes := clienteDB.UltimasTransacoes

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

	// Busca os dados do cliente
	clienteDB, err := BuscarCliente(idCliente, dbPool)
	if err != nil {
		return models.Extrato{}, err
	}
	cliente := clienteDB.Cliente
	ultimasTransacoes := clienteDB.UltimasTransacoes

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
