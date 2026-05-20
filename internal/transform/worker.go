package transform

import (
	"log/slog"
	"strconv"
	"strings"
	"sync"

	"github.com/SanjiTriste/AegisStream/internal/config"
	"github.com/SanjiTriste/AegisStream/internal/domain"
)

func WorkerPool(id int, cfg config.AppConfig, jobs <-chan domain.RegistroBruto, cleanJobs chan<- domain.RegistroVarejo, logger *slog.Logger, wg *sync.WaitGroup) {
	defer wg.Done()

	for raw := range jobs {
		s := raw.Schema

		// Só o ID é obrigatório
		if len(raw.Dados) <= s.IDTransacao {
			logger.Warn("Linha ignorada: sem coluna de ID",
				"worker_id", id, "linha", raw.Linha)
			continue
		}

		idTransacao := strings.TrimSpace(raw.Dados[s.IDTransacao])
		if idTransacao == "" {
			continue
		}

		// Campos opcionais — usa valor default se a coluna não existir
		cpfBruto := campoOpcional(raw.Dados, s.CPF)
		status := strings.ToUpper(campoOpcional(raw.Dados, s.Status))
		if status == "" {
			status = "DESCONHECIDO"
		}

		idade := 0
		if s.Idade >= 0 && s.Idade < len(raw.Dados) {
			idadeStr := strings.TrimSpace(raw.Dados[s.Idade])
			if v, err := strconv.Atoi(idadeStr); err == nil {
				idade = v
			}
		}

		cleanJobs <- domain.RegistroVarejo{
			IDTransacao:      idTransacao,
			IdadeCliente:     idade,
			StatusPagto:      status,
			CPFMascarado:     MascararCPF(cpfBruto),
			CPFCriptografado: CriptografarAES(cpfBruto, cfg.ChaveCifraAES),
		}
	}
}

// campoOpcional retorna o valor da coluna ou "" se o índice não existir.
func campoOpcional(dados []string, idx int) string {
	if idx < 0 || idx >= len(dados) {
		return ""
	}
	return strings.TrimSpace(dados[idx])
}