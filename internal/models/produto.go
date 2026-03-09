package models

// Produto representa a estrutura de um item no estoque da loja.
type Produto struct {
	ID         int     `json:"id"`
	Nome       string  `json:"nome"`
	Preco      float64 `json:"preco"`
	Quantidade int     `json:"quantidade"`
}

// VendaRequest captura dados para processar baixas de estoque.
type VendaRequest struct {
	ID         int `json:"id"`
	Quantidade int `json:"quantidade"`
}

// DeleteRequest captura o ID do produto que será removido do sistema.
type DeleteRequest struct {
	ID int `json:"id"`
}
