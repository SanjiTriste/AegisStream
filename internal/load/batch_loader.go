package load

import (
	"context"
	"database/sql"
	"log/slog"
	"sync"
	"time"

	"github.com/SanjiTriste/AegisStream/internal/config"
	"github.com/SanjiTriste/AegisStream/internal/domain"
	_ "modernc.org/sqlite" // NOVO: Driver 100% Go (Substitui o CGO)
)

func InicializarBanco() (*sql.DB, error) {
	// NOVO: Aqui agora é "sqlite" em vez de "sqlite3"
	db, err := sql.Open("sqlite", "./banco_varejo.db")
	if err != nil {
		return nil, err
	}

	query := `
	CREATE TABLE IF NOT EXISTS transacoes_varejo (
		id_transacao TEXT PRIMARY KEY,
		idade_cliente INTEGER,
		status_pagto TEXT,
		cpf_mascarado TEXT,
		cpf_criptografado TEXT,
		atualizado_em DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = db.Exec(query)
	return db, err
}

func Loader(ctx context.Context, db *sql.DB, cfg config.AppConfig, cleanJobs <-chan domain.RegistroVarejo, logger *slog.Logger, wg *sync.WaitGroup) {
	defer wg.Done()
	lote := make([]domain.RegistroVarejo, 0, cfg.BatchSize)

	for registro := range cleanJobs {
		lote = append(lote, registro)
		if len(lote) >= cfg.BatchSize {
			gravarLoteComRetry(ctx, db, lote, logger)
			lote = lote[:0]
		}
	}

	if len(lote) > 0 {
		gravarLoteComRetry(ctx, db, lote, logger)
	}
}

func gravarLoteComRetry(ctx context.Context, db *sql.DB, lote []domain.RegistroVarejo, logger *slog.Logger) {
	maxTentativas := 3
	for tentativa := 1; tentativa <= maxTentativas; tentativa++ {
		err := executarTransacaoBD(ctx, db, lote)
		if err == nil {
			logger.Info("Lote salvo", "registros", len(lote), "tentativa", tentativa)
			return
		}

		logger.Warn("Falha ao gravar lote", "tentativa", tentativa, "erro", err.Error())
		tempoEspera := time.Duration(1<<tentativa) * time.Second

		select {
		case <-ctx.Done():
			logger.Error("Gravacao cancelada (Shutdown)")
			return
		case <-time.After(tempoEspera):
		}
	}
	logger.Error("FALHA CRITICA: Lote perdido apos 3 tentativas.")
}

func executarTransacaoBD(ctx context.Context, db *sql.DB, lote []domain.RegistroVarejo) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO transacoes_varejo (id_transacao, idade_cliente, status_pagto, cpf_mascarado, cpf_criptografado)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(id_transacao) DO UPDATE SET
			idade_cliente = excluded.idade_cliente,
			status_pagto = excluded.status_pagto,
			cpf_mascarado = excluded.cpf_mascarado,
			cpf_criptografado = excluded.cpf_criptografado,
			atualizado_em = CURRENT_TIMESTAMP;
	`
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, reg := range lote {
		_, err := stmt.ExecContext(ctx, reg.IDTransacao, reg.IdadeCliente, reg.StatusPagto, reg.CPFMascarado, reg.CPFCriptografado)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}
