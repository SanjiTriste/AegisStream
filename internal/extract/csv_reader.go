package extract

import (
	"context"
	"encoding/csv"
	"io"
	"log/slog"
	"os"

	"github.com/SanjiTriste/AegisStream/internal/domain"
)

func LeitorCSV(ctx context.Context, caminho string, jobs chan<- domain.RegistroBruto, logger *slog.Logger) {
	defer close(jobs)

	file, err := os.Open(caminho)
	if err != nil {
		logger.Error("Arquivo nao encontrado", "arquivo", caminho, "erro", err)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Read()
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
		jobs <- domain.RegistroBruto{Linha: linhaCount, Dados: linha}
	}
	logger.Info("Streaming concluido", "linhas_brutas_lidas", linhaCount)
}