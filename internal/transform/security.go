package transform

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"io"
	"strings"
)

func CriptografarAES(texto string, chave []byte) string {
	if texto == "" {
		return ""
	}
	bloco, err := aes.NewCipher(chave)
	if err != nil {
		return ""
	}
	gcm, err := cipher.NewGCM(bloco)
	if err != nil {
		return ""
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return ""
	}
	textoCifrado := gcm.Seal(nonce, nonce, []byte(texto), nil)
	return hex.EncodeToString(textoCifrado)
}

func MascararCPF(cpf string) string {
	cpf = strings.TrimSpace(cpf)
	if len(cpf) >= 14 {
		return "***.***." + cpf[8:]
	}
	return "CPF_INVALIDO"
}