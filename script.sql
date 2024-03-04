CREATE TABLE cliente (
	id SERIAL PRIMARY KEY,
	nome VARCHAR(50) NOT NULL,
	limite INTEGER NOT NULL,
	saldo INTEGER NOT NULL,
	ultimas_transacoes json,
	versao INTEGER NOT NULL
);

DO $$
BEGIN
        INSERT INTO cliente (nome, limite, saldo, ultimas_transacoes, versao)
		VALUES
			('o barato sai caro', 1000 * 100, 0, '[]', 1),
			('zan corp ltda', 800 * 100, 0, '[]', 1),
			('les cruders', 10000 * 100, 0, '[]', 1),
			('padaria joia de cocaia', 100000 * 100, 0, '[]', 1),
			('kid mais', 5000 * 100, 0, '[]', 1);
END;
$$;