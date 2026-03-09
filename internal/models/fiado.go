package models

// Cliente representa uma pessoa cadastrada para comprar fiado.
type Cliente struct {
	ID       int    `json:"id"`
	Nome     string `json:"nome"`
	Telefone string `json:"telefone"`
}

// Fiado representa o registro de uma dívida no caderno.
type Fiado struct {
	ID          int     `json:"id"`
	ClienteID   int     `json:"cliente_id"`
	NomeCliente string  `json:"nome_cliente"` // Este campo é preenchido via JOIN no banco
	Valor       float64 `json:"valor"`
	Descricao   string  `json:"descricao"`
	DataDivida  string  `json:"data_divida"`
	Pago        bool    `json:"pago"`
}

// BaixaFiadoRequest captura o ID da dívida para dar baixa quando o cliente pagar.
type BaixaFiadoRequest struct {
	ID int `json:"id"`
}
