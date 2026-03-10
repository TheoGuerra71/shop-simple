package models

// Movimentacao registra tudo o que entra e sai do seu caixa no app
type Movimentacao struct {
	ID        int     `json:"id"`
	Tipo      string  `json:"tipo"` // "ENTRADA" ou "SAIDA"
	Descricao string  `json:"descricao"`
	Valor     float64 `json:"valor"`
	DataMov   string  `json:"data_mov"`
}

// ResumoApp contém todos os números para montar o Dashboard idêntico ao da foto
type ResumoApp struct {
	VendasHoje    float64 `json:"vendas_hoje"`
	VendasMes     float64 `json:"vendas_mes"`
	TicketMedio   float64 `json:"ticket_medio"`
	ItensEstoque  int     `json:"itens_estoque"`
	TotalEntradas float64 `json:"total_entradas"`
	TotalSaidas   float64 `json:"total_saidas"`
	LucroLiquido  float64 `json:"lucro_liquido"`
}

// ProdutoApp atualiza o nosso produto para ter Custo e Limite de Reposição
type ProdutoApp struct {
	ID            int     `json:"id"`
	Nome          string  `json:"nome"`
	PrecoVenda    float64 `json:"preco"`
	Custo         float64 `json:"custo"`
	Quantidade    int     `json:"quantidade"`
	EstoqueMinimo int     `json:"estoque_minimo"`
	UrlImagem     string  `json:"url_imagem"`
}
