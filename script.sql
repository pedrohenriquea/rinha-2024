CREATE TABLE cliente_transacao (
	id SERIAL PRIMARY KEY,
	nome VARCHAR(50) NULL,
	limite INTEGER NULL,
	saldo INTEGER  NULL,
	id_pai INTEGER NULL,
	valor INTEGER NULL,
	tipo CHAR(1) NULL,
	descricao VARCHAR(10) NULL,
	realizada_em TIMESTAMP NULL DEFAULT NOW(),
	CONSTRAINT fk_cliente_transacao_id_pai
		FOREIGN KEY (id_pai) REFERENCES cliente_transacao(id)
);
CREATE INDEX idx_id_pai ON cliente_transacao(id_pai);
CREATE INDEX idx_realizada_em ON cliente_transacao(realizada_em);

DO $$
BEGIN
        INSERT INTO cliente_transacao (nome, limite, saldo, realizada_em)
		VALUES
			('o barato sai caro', 1000 * 100, 0, '9999-12-30'),
			('zan corp ltda', 800 * 100, 0, '9999-12-30'),
			('les cruders', 10000 * 100, 0, '9999-12-30'),
			('padaria joia de cocaia', 100000 * 100, 0, '9999-12-30'),
			('kid mais', 5000 * 100, 0, '9999-12-30');

END;
$$;