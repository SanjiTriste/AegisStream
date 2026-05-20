package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath" // NOVO: Pacote para lidar com caminhos de pastas
	"sync"
	"syscall"
	"time"

	"github.com/SanjiTriste/AegisStream/internal/config"
	"github.com/SanjiTriste/AegisStream/internal/domain"
	"github.com/SanjiTriste/AegisStream/internal/extract"
	"github.com/SanjiTriste/AegisStream/internal/load"
	"github.com/SanjiTriste/AegisStream/internal/transform"
	"github.com/SanjiTriste/AegisStream/pkg/logger"
)

func main() {
	caminhoCSV := flag.String("csv", "", "Caminho do arquivo CSV de entrada")
	flag.Parse()

	log := logger.New()
	cfg := config.CarregarConfiguracoes()

	if *caminhoCSV == "" {
		fmt.Println("Erro: Caminho do arquivo nao informado. Uso: go run cmd/etl/main.go -csv=vendas_reais.csv")
		os.Exit(1)
	}

	log.Info("Iniciando AegisStream ETL", "arquivo", *caminhoCSV, "workers", cfg.WorkerCount)
	start := time.Now()

	db, err := load.InicializarBanco()
	if err != nil {
		log.Error("Falha critica ao iniciar banco", "erro", err)
		os.Exit(1)
	}
	defer db.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	jobs := make(chan domain.RegistroBruto, 1000)
	cleanJobs := make(chan domain.RegistroVarejo, 1000)
	var wgWorkers sync.WaitGroup
	var wgLoader sync.WaitGroup

	wgLoader.Add(1)
	go load.Loader(ctx, db, cfg, cleanJobs, log, &wgLoader)

	for w := 1; w <= cfg.WorkerCount; w++ {
		wgWorkers.Add(1)
		go transform.WorkerPool(w, cfg, jobs, cleanJobs, log, &wgWorkers)
	}

	go extract.LeitorCSV(ctx, *caminhoCSV, jobs, log)

	wgWorkers.Wait()
	close(cleanJobs)
	wgLoader.Wait()

	// ==========================================
	// NOVO: Movendo o arquivo para a pasta outbox
	// ==========================================

	// Pega apenas o nome do arquivo (ex: extrai "vendas_reais.csv" de "C:/pasta/vendas_reais.csv")
	nomeArquivo := filepath.Base(*caminhoCSV)

	// Monta o caminho de destino apontando para a pasta outbox
	caminhoDestino := filepath.Join("outbox", nomeArquivo+".processado")

	if err := os.Rename(*caminhoCSV, caminhoDestino); err == nil {
		log.Info("Arquivo fonte movido para outbox com sucesso", "caminho_destino", caminhoDestino)
	} else {
		log.Error("Falha ao mover arquivo para outbox", "erro", err)
	}

	log.Info("AegisStream Finalizado", "tempo_execucao_ms", time.Since(start).Milliseconds())
}
