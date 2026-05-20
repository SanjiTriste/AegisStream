package logger

import (
	"log/slog"
	"os"
)

// New retorna um logger padrão do Go configurado para logs em formato JSON.
func New() *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	return slog.New(handler)
}
