package config

type AppConfig struct {
	BatchSize     int
	WorkerCount   int
	ChaveCifraAES []byte
}

func CarregarConfiguracoes() AppConfig {
	return AppConfig{
		BatchSize:     500,
		WorkerCount:   3,
		ChaveCifraAES: []byte("super_secreta_chave_de_32_bytes!"),
	}
}