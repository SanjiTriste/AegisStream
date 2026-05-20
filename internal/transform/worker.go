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
		idadeStr := strings.TrimSpace(raw.Dados[7])
		idade, err := strconv.Atoi(idadeStr)
		if err != nil {
			logger.Warn("Dado invalido descartado", "worker_id", id, "linha", raw.Linha, "campo", "idade")
			continue
		}

		status := strings.ToUpper(strings.TrimSpace(raw.Dados[9]))
		if status == "" {
			continue
		}

		cpfBruto := raw.Dados[6]
		cleanJobs <- domain.RegistroVarejo{
			IDTransacao:      strings.TrimSpace(raw.Dados[0]),
			IdadeCliente:     idade,
			StatusPagto:      status,
			CPFMascarado:     MascararCPF(cpfBruto),
			CPFCriptografado: CriptografarAES(cpfBruto, cfg.ChaveCifraAES),
		}
	}
}