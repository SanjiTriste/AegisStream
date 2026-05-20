package extract

import (
	"context"
	"encoding/csv"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/SanjiTriste/AegisStream/internal/domain"
)

// Variações aceitas para cada campo.
// Campos marcados com * são opcionais — o pipeline não rejeita o CSV se não existirem.
var mapeamentoColunas = map[string]string{
	// id da transação (obrigatório)
	"id":             "id_transacao",
	"id_transacao":   "id_transacao",
	"transacao_id":   "id_transacao",
	"transaction_id": "id_transacao",
	"txn_id":         "id_transacao",
	"order_id":       "id_transacao",

	// cpf / documento (opcional)
	"cpf":            "cpf",
	"cpf_cliente":    "cpf",
	"cpf_mascarado":  "cpf",
	"documento":      "cpf",
	"document":       "cpf",
	"national_id":    "cpf",
	"tax_id":         "cpf",

	// idade (opcional)
	"idade":          "idade",
	"idade_cliente":  "idade",
	"age":            "idade",
	"customer_age":   "idade",

	// status / situação (opcional)
	"status":              "status",
	"status_pagto":        "status",
	"status_pagamento":    "status",
	"payment_status":      "status",
	"transaction_status":  "status",
	"resultado":           "status",
	"situacao":            "status",
	"situação":            "status",
}

func detectarSchema(headers []string, logger *slog.Logger) domain.SchemaCSV {
	schema := domain.SchemaCSV{
		IDTransacao: -1,
		CPF:         -1,
		Idade:       -1,
		Status:      -1,
	}

	for i, h := range headers {
		normalizado := strings.ToLower(strings.TrimSpace(h))
		campo, encontrado := mapeamentoColunas[normalizado]
		if !encontrado {
			continue
		}
		switch campo {
		case "id_transacao":
			schema.IDTransacao = i
		case "cpf":
			schema.CPF = i
		case "idade":
			schema.Idade = i
		case "status":
			schema.Status = i
		}
	}

	logger.Info("Schema CSV detectado",
		"id_transacao_col", schema.IDTransacao,
		"cpf_col", schema.CPF,
		"idade_col", schema.Idade,
		"status_col", schema.Status,
	)

	return schema
}

func LeitorCSV(ctx context.Context, caminho string, jobs chan<- domain.RegistroBruto, logger *slog.Logger) {
	defer close(jobs)

	file, err := os.Open(caminho)
	if err != nil {
		logger.Error("Arquivo não encontrado", "arquivo", caminho, "erro", err)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)

	headers, err := reader.Read()
	if err != nil {
		logger.Error("Falha ao ler header do CSV", "arquivo", caminho, "erro", err)
		return
	}

	schema := detectarSchema(headers, logger)

	// Só o id_transacao é obrigatório — os outros são opcionais
	if schema.IDTransacao < 0 {
		logger.Error("CSV rejeitado: coluna de ID não encontrada",
			"arquivo", caminho,
			"nomes_aceitos", "id, id_transacao, transaction_id, order_id, txn_id",
		)
		return
	}

	linhaCount := 1
	for {
		select {
		case <-ctx.Done():
			logger.Info("Leitura interrompida pelo sistema (Shutdown)")
			return
		default:
		}

		linha, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			logger.Warn("Erro de parse na linha", "linha", linhaCount, "erro", err)
			continue
		}

		linhaCount++
		jobs <- domain.RegistroBruto{
			Linha:  linhaCount,
			Dados:  linha,
			Schema: schema,
		}
	}

	logger.Info("Streaming concluído", "linhas_brutas_lidas", linhaCount)
}