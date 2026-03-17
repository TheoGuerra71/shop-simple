package models

// 👤 USUÁRIO: O dono da loja (Lojista)
type Usuario struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
	Senha string `json:"senha"`
}

// ⚙️ CONFIGURAÇÃO DA LOJA: Identidade visual e contato do catálogo
type LojaConfig struct {
	UsuarioID  int    `json:"usuario_id"`
	NomeLoja   string `json:"nome_loja"`
	Whatsapp   string `json:"whatsapp"`
	Instagram  string `json:"instagram"`
	CorHex     string `json:"cor_hex"`
	MsgSuporte string `json:"msg_suporte"`
}

// 📦 PRODUTO: O item físico que vai para a vitrine
type ProdutoApp struct {
	ID              int      `json:"id"`
	UsuarioID       int      `json:"usuario_id"`
	Nome            string   `json:"nome"`
	Categoria       string   `json:"categoria"`
	Preco           float64  `json:"preco"` 
	PrecoVenda      float64  `json:"preco_venda"`
	Custo           float64  `json:"custo"`
	Quantidade      int      `json:"quantidade"`
	EstoqueMinimo   int      `json:"estoque_minimo"`
	UrlImagem       []string `json:"url_imagem"`
	VisivelCatalogo bool     `json:"visivel_catalogo"`
}

// 💰 MOVIMENTO: A "gaveta" do caixa
type Movimento struct {
	ID        int     `json:"id"`
	UsuarioID int     `json:"usuario_id"` 
	Tipo      string  `json:"tipo"`       
	Descricao string  `json:"descricao"`
	Valor     float64 `json:"valor"`
	DataMov   string  `json:"data_mov"`
}

// ==========================================
// 🤝 SISTEMA DE FIADO E CLIENTES
// ==========================================

// 🧑‍🤝‍🧑 CLIENTE: A pessoa que compra na loja
type Cliente struct {
	ID        int    `json:"id"`
	UsuarioID int    `json:"usuario_id"` 
	Nome      string `json:"nome"`
	Telefone  string `json:"telefone"`
}

// 📝 FIADO: A conta pendente do cliente
type Fiado struct {
	ID          int     `json:"id"`
	UsuarioID   int     `json:"usuario_id"`
	ClienteID   int     `json:"cliente_id"`   // 🚀 AQUI ESTÁ A PEÇA QUE FALTAVA!
	NomeCliente string  `json:"nome_cliente"` 
	Valor       float64 `json:"valor"`
	DataDivida  string  `json:"data_divida"`  
	Pago        bool    `json:"pago"`         
	Descricao   string  `json:"descricao"`
}

// 💸 BAIXA FIADO: O molde para receber o pagamento de uma conta
type BaixaFiadoRequest struct {
	ID    int     `json:"id"`    
	Valor float64 `json:"valor"` 
}