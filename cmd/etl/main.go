package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/SanjiTriste/AegisStream/internal/config"
	"github.com/SanjiTriste/AegisStream/internal/domain"
	"github.com/SanjiTriste/AegisStream/internal/extract"
	"github.com/SanjiTriste/AegisStream/internal/load"
	"github.com/SanjiTriste/AegisStream/internal/transform"
	"github.com/SanjiTriste/AegisStream/pkg/logger"
	"github.com/fsnotify/fsnotify"
)

func main() {
	log := logger.New()
	cfg := config.CarregarConfiguracoes()

	if err := os.MkdirAll("inbox", 0755); err != nil {
		log.Error("Falha ao criar pasta inbox", "erro", err)
		os.Exit(1)
	}
	if err := os.MkdirAll("outbox", 0755); err != nil {
		log.Error("Falha ao criar pasta outbox", "erro", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Error("Falha ao iniciar watcher", "erro", err)
		os.Exit(1)
	}
	defer watcher.Close()

	// Debounce — evita processar o mesmo arquivo duas vezes
	// (o fsnotify dispara múltiplos eventos no Windows)
	var (
		debounce   = make(map[string]time.Time)
		debounceMu sync.Mutex
	)

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&(fsnotify.Create|fsnotify.Write) != 0 {
					if strings.ToLower(filepath.Ext(event.Name)) != ".csv" {
						continue
					}

					// Ignora se o mesmo arquivo foi processado nos últimos 2 segundos
					debounceMu.Lock()
					ultima, visto := debounce[event.Name]
					if visto && time.Since(ultima) < 2*time.Second {
						debounceMu.Unlock()
						continue
					}
					debounce[event.Name] = time.Now()
					debounceMu.Unlock()

					log.Info("CSV detetado, a iniciar processamento...", "arquivo", event.Name)
					executarETL(ctx, event.Name, cfg, log)
				}

			case <-ctx.Done():
				return

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Error("Erro no monitor", "erro", err)
			}
		}
	}()

	if err := watcher.Add("./inbox"); err != nil {
		log.Error("Falha ao vigiar pasta inbox", "erro", err)
		os.Exit(1)
	}

	fmt.Println("AegisStream Ativo! A aguardar CSVs na pasta ./inbox...")
	fmt.Println("Pressione Ctrl+C para encerrar.")

	<-ctx.Done()
	log.Info("Sinal recebido. AegisStream a encerrar...")
}

func executarETL(ctx context.Context, caminhoCSV string, cfg config.AppConfig, log *slog.Logger) {
	start := time.Now()

	db, err := load.InicializarBanco()
	if err != nil {
		log.Error("Falha critica ao iniciar banco", "erro", err)
		return
	}
	defer db.Close()

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

	go extract.LeitorCSV(ctx, caminhoCSV, jobs, log)

	wgWorkers.Wait()
	close(cleanJobs)
	wgLoader.Wait()

	nomeArquivo := filepath.Base(caminhoCSV)
	caminhoDestino := filepath.Join("outbox", nomeArquivo+".processado")

	if err := os.Rename(caminhoCSV, caminhoDestino); err == nil {
		log.Info("Ficheiro processado e movido para outbox", "destino", caminhoDestino)
	} else {
		log.Error("Falha ao mover ficheiro", "erro", err)
	}

	log.Info("Processamento Finalizado", "tempo_ms", time.Since(start).Milliseconds())
}
