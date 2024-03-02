package service

import (
	"api/models"
	"context"
	"errors"
	"log"
	"sort"
	"strings"
	"sync"
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
	if err := tx.QueryRow(ctx, `SELECT limite, saldo FROM cliente_transacao WHERE id=$1 and id_pai is null FOR UPDATE`, idCliente).Scan(&cliente.Limite, &cliente.Saldo); err != nil {
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

	// Inserir a transação no banco de dados
	var transacaoID int
	if err := tx.QueryRow(ctx, `INSERT INTO cliente_transacao (id_pai, valor, tipo, descricao, realizada_em) VALUES($1, $2, $3, $4, now()) RETURNING id`, idCliente, transacaoRequest.Valor, transacaoRequest.Tipo, transacaoRequest.Descricao).Scan(&transacaoID); err != nil {
		return nil, errors.New("INSERE_TRANSACAO_EXCEPTION")
	}

	// Atualizar o saldo do cliente no banco de dados
	if _, err := tx.Exec(ctx, `UPDATE cliente_transacao SET saldo=$1 WHERE id=$2`, cliente.Saldo, idCliente); err != nil {
		return nil, errors.New("ATUALIZA_SALDO_EXCEPTION")
	}

	// Preparar a resposta
	resp := &models.Cliente{
		Limite: cliente.Limite,
		Saldo:  cliente.Saldo,
	}

	return resp, nil
}

func GetExtrato(idCliente int, dbPool *pgxpool.Pool) (extrato models.Extrato, err error) {

	query := `
	SELECT limite, saldo, id_pai, valor, tipo, descricao, realizada_em
	FROM cliente_transacao
	WHERE id = $1 or id_pai = $1
	ORDER BY realizada_em DESC
	LIMIT 11 OFFSET 0
	`

	rows, err := dbPool.Query(context.Background(), query, idCliente)
	if err != nil {
		return extrato, err
	}

	// Iterar sobre os resultados
	var saldoExtrato models.SaldoExtrato
	var transacoes []models.TransacaoExtrato
	var primeiroRegistro bool = true

	for rows.Next() {
		var transacao models.ClienteTransacao
		if err := rows.Scan(&transacao.Limite, &transacao.Saldo, &transacao.IdPai, &transacao.Valor, &transacao.Tipo, &transacao.Descricao, &transacao.RealizadaEm); err != nil {
			log.Println(err)
			continue
		}

		if primeiroRegistro && transacao.IdPai.Valid {
			return models.Extrato{}, errors.New("BUSCA_CLIENTE_EXCEPTION")
		}

		primeiroRegistro = false

		if !transacao.IdPai.Valid {
			saldoExtrato = models.SaldoExtrato{
				Total:       transacao.Saldo.Int64,
				DataExtrato: time.Now(),
				Limite:      transacao.Limite.Int64,
			}
		} else {
			transExtrato := models.TransacaoExtrato{
				Valor:       transacao.Valor.Int64,
				Tipo:        transacao.Tipo.String,
				Descricao:   transacao.Descricao.String,
				RealizadaEm: transacao.RealizadaEm,
			}
			transacoes = append(transacoes, transExtrato)
		}
	}
	if err := rows.Err(); err != nil {
		log.Println(err)
	}

	// Criar extrato
	extrato = models.Extrato{
		Saldo:             saldoExtrato,
		UltimasTransacoes: transacoes,
	}

	return extrato, nil
}

func GetExtratoGoRoutine(idCliente int, dbPool *pgxpool.Pool) (extrato models.Extrato, err error) {

	// Canal para enviar transações processadas
	transacoesCh := make(chan []models.TransacaoExtratoRoutine)

	// Canal para sinalizar quando todas as goroutines terminaram
	doneCh := make(chan struct{})
	defer close(doneCh)

	// Consulta SQL
	query := `
        SELECT limite, saldo, id_pai, valor, tipo, descricao, realizada_em
        FROM cliente_transacao
        WHERE id = $1 OR id_pai = $1
        ORDER BY realizada_em DESC
    `

	// Função para processar um lote de resultados
	processBatch := func(offset int, batchSize int) {

		rows, err := dbPool.Query(context.Background(), query+" LIMIT $2 OFFSET $3", idCliente, batchSize, offset)

		if err != nil {
			log.Println(err)
			return
		}
		defer rows.Close()

		var transacoes []models.TransacaoExtratoRoutine

		for rows.Next() {
			var transacao models.ClienteTransacao
			if err := rows.Scan(&transacao.Limite, &transacao.Saldo, &transacao.IdPai, &transacao.Valor, &transacao.Tipo, &transacao.Descricao, &transacao.RealizadaEm); err != nil {
				log.Println(err)
				continue
			}

			// Processar transação...
			if !transacao.IdPai.Valid {
				transSaldo := models.TransacaoExtratoRoutine{
					Saldo: &models.SaldoExtrato{
						Total:       transacao.Saldo.Int64,
						DataExtrato: time.Now(),
						Limite:      transacao.Limite.Int64,
					},
					RealizadaEm: transacao.RealizadaEm,
				}
				transacoes = append(transacoes, transSaldo)
			} else {
				transExtrato := models.TransacaoExtratoRoutine{
					Valor:       transacao.Valor.Int64,
					Tipo:        transacao.Tipo.String,
					Descricao:   transacao.Descricao.String,
					RealizadaEm: transacao.RealizadaEm,
				}
				transacoes = append(transacoes, transExtrato)
			}

		}
		if err := rows.Err(); err != nil {
			log.Println(err)
		}

		// Enviar transações processadas para o canal
		transacoesCh <- transacoes
	}

	// Inicie as goroutines para processar os lotes
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			switch i {
			case 0:
				processBatch(0, 6)
			case 1:
				processBatch(6, 5)
			}

		}(i)
	}

	// Aguarde o término de todas as goroutines
	go func() {
		wg.Wait()
		close(transacoesCh)
	}()

	// Receba as transações processadas dos canais e agregue-as
	var transacoes []models.TransacaoExtratoRoutine
	for trans := range transacoesCh {
		transacoes = append(transacoes, trans...)
	}

	OrdenarPorRealizadaEm(transacoes)
	if len(transacoes) == 0 {
		return models.Extrato{}, errors.New("BUSCA_CLIENTE_EXCEPTION")
	}

	var saldo models.SaldoExtrato
	if transacoes[0].Saldo != nil {
		saldo = *transacoes[0].Saldo
	} else {
		return models.Extrato{}, errors.New("BUSCA_CLIENTE_EXCEPTION")
	}

	extrato = models.Extrato{
		Saldo:             saldo,
		UltimasTransacoes: ConverterParaTransacaoExtrato(transacoes),
	}

	return extrato, nil
}

type ByRealizadaEm []models.TransacaoExtratoRoutine

func (a ByRealizadaEm) Len() int           { return len(a) }
func (a ByRealizadaEm) Less(i, j int) bool { return a[i].RealizadaEm.After(a[j].RealizadaEm) }
func (a ByRealizadaEm) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func OrdenarPorRealizadaEm(transacoes []models.TransacaoExtratoRoutine) {
	sort.Sort(ByRealizadaEm(transacoes))
}

func ConverterParaTransacaoExtrato(transacoes []models.TransacaoExtratoRoutine) []models.TransacaoExtrato {
	var transacoesExtrato []models.TransacaoExtrato
	for _, transacao := range transacoes {
		if transacao.Saldo == nil {
			transacaoExtrato := models.TransacaoExtrato{
				Valor:       transacao.Valor,
				Tipo:        transacao.Tipo,
				Descricao:   transacao.Descricao,
				RealizadaEm: transacao.RealizadaEm,
			}
			transacoesExtrato = append(transacoesExtrato, transacaoExtrato)
		}
	}
	return transacoesExtrato
}
