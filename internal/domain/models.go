package domain

// SchemaCSV mapeia os índices das colunas encontradas no header do CSV.
// Cada campo começa em -1 (não encontrado).
type SchemaCSV struct {
	IDTransacao int
	CPF         int
	Idade       int
	Status      int
}


// Valido verifica se todos os campos obrigatórios foram encontrados no header.
func (s SchemaCSV) Valido() bool {
	return s.IDTransacao >= 0 && s.CPF >= 0 && s.Idade >= 0 && s.Status >= 0
}

// RegistroBruto carrega a linha crua do CSV junto com o schema detectado.
type RegistroBruto struct {
	Linha  int
	Dados  []string
	Schema SchemaCSV
}

type RegistroVarejo struct {
	IDTransacao      string
	IdadeCliente     int
	StatusPagto      string
	CPFMascarado     string
	CPFCriptografado string
}
