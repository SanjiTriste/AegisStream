package domain

type RegistroBruto struct {
	Linha int
	Dados []string
}

type RegistroVarejo struct {
	IDTransacao      string
	IdadeCliente     int
	StatusPagto      string
	CPFMascarado     string
	CPFCriptografado string
}