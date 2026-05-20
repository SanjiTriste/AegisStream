package config

import (
	"encoding/hex"
	"log"
	"os"
)

type AppConfig struct {
	BatchSize     int
	WorkerCount   int
	ChaveCifraAES []byte
}

func CarregarConfiguracoes() AppConfig {
	chave := carregarChaveAES()

	return AppConfig{
		BatchSize:     500,
		WorkerCount:   3,
		ChaveCifraAES: chave,
	}
}

func carregarChaveAES() []byte {
	chaveHex := os.Getenv("AEGIS_AES_KEY")

	if chaveHex != "" {
		chave, err := hex.DecodeString(chaveHex)
		if err != nil {
			log.Fatalf("AEGIS_AES_KEY inválida (deve ser hex de 32 bytes): %v", err)
		}
		if len(chave) != 32 {
			log.Fatalf("AEGIS_AES_KEY deve ter exatamente 32 bytes (64 chars hex), tem %d", len(chave))
		}
		return chave
	}

	// Fallback para desenvolvimento local — NUNCA usar em produção
	log.Println("[AVISO] AEGIS_AES_KEY não definida. Usando chave de desenvolvimento.")
	log.Println("[AVISO] Para produção: export AEGIS_AES_KEY=$(openssl rand -hex 32)")
	return []byte("super_secreta_chave_de_32_bytes!")
}