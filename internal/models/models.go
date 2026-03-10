package models

// --- SEGURANÇA ---
type Usuario struct {
	Email string `json:"email"`
	Senha string `json:"senha"`
}

// --- PRODUTOS E ESTOQUE ---
type ProdutoApp struct {
	ID            int     `json:"id"`
	Nome          string  `json:"nome"`
	Categoria     string  `json:"categoria"`
	PrecoVenda    float64 `json:"preco"`
	Custo         float64 `json:"custo"`
	Quantidade    int     `json:"quantidade"`
	EstoqueMinimo int     `json:"estoque_minimo"`
	UrlImagem     string  `json:"url_imagem"`
}

// --- FINANCEIRO E CAIXA ---
type Movimentacao struct {
	ID        int     `json:"id"`
	Tipo      string  `json:"tipo"`
	Descricao string  `json:"descricao"`
	Valor     float64 `json:"valor"`
	DataMov   string  `json:"data_mov"`
}

type SessaoCaixa struct {
	ID              int     `json:"id"`
	Status          string  `json:"status"`
	FundoTroco      float64 `json:"fundo_troco"`
	TotalFechamento float64 `json:"total_fechamento"`
}

// --- CONFIGURAÇÕES DA LOJA (CATÁLOGO FARM) ---
type LojaConfig struct {
	NomeLoja   string `json:"nome_loja"`
	Whatsapp   string `json:"whatsapp"`
	Instagram  string `json:"instagram"`
	CorHex     string `json:"cor_hex"`
	MsgSuporte string `json:"msg_suporte"`
}

// --- LOGÍSTICA ---
type ChecklistItem struct {
	ID        int    `json:"id"`
	Tarefa    string `json:"tarefa"`
	Concluido bool   `json:"concluido"`
}

// --- SISTEMA DE FIADO E CLIENTES (AJUSTADO AOS NOMES EXATOS DO SEU CÓDIGO) ---
type Cliente struct {
	ID       int    `json:"id"`
	Nome     string `json:"nome"`
	Telefone string `json:"telefone"`
}

type Fiado struct {
	ID          int     `json:"id"`
	ClienteID   int     `json:"cliente_id"`
	NomeCliente string  `json:"nome_cliente"`
	Descricao   string  `json:"descricao"`
	Valor       float64 `json:"valor"`
	DataDivida  string  `json:"data_divida"` // Consertado: Era 'Data'
	Pago        bool    `json:"pago"`        // Consertado: Era 'Status string'
}

type BaixaFiadoRequest struct {
	ID    int     `json:"id"` // Consertado: Era 'FiadoID'
	Valor float64 `json:"valor"`
}
