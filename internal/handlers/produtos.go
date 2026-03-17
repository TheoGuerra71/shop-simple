package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/theo-guerra/simple-shop/internal/models"
)

type ProdutoHandler struct {
	DB *sql.DB
}

// 🛡️ scanUrlImagemJSONB: Lê do banco de dados (JSONB) e transforma na lista de fotos pro Front
func scanUrlImagemJSONB(raw interface{}) []string {
	if raw == nil {
		return []string{}
	}
	var b []byte
	switch v := raw.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return []string{}
	}

	// Proteção contra dados velhos salvos antes da atualização
	if len(b) > 0 && b[0] != '[' {
		return []string{string(b)}
	}

	var arr []string
	err := json.Unmarshal(b, &arr)
	if err != nil || arr == nil {
		return []string{}
	}
	return arr
}

// 📦 LER PRODUTOS (Para o Painel - Protegido pela Senha)
func (h *ProdutoHandler) ListarProdutos(w http.ResponseWriter, r *http.Request) {
	usuarioID, ok := UsuarioIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Não autorizado", http.StatusUnauthorized)
		return
	}

	rows, err := h.DB.Query(
		`SELECT id, nome, COALESCE(categoria,'Geral'), preco, custo, quantidade, estoque_minimo,
         COALESCE(url_imagem::text,'[]'), COALESCE(visivel_catalogo, true)
         FROM produtos WHERE usuario_id = $1 ORDER BY id DESC`,
		usuarioID,
	)
	if err != nil {
		http.Error(w, "Erro ao buscar produtos: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var produtos []models.ProdutoApp
	for rows.Next() {
		var p models.ProdutoApp
		var urlImagemRaw interface{}
		var visivel bool

		// AQUI: Usando p.PrecoVenda para bater com o models.go
		err := rows.Scan(&p.ID, &p.Nome, &p.Categoria, &p.PrecoVenda, &p.Custo, &p.Quantidade, &p.EstoqueMinimo, &urlImagemRaw, &visivel)
		if err != nil {
			continue 
		}
		p.UsuarioID = usuarioID
		p.UrlImagem = scanUrlImagemJSONB(urlImagemRaw)
		p.VisivelCatalogo = visivel

		produtos = append(produtos, p)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(produtos)
}

// 🌐 LER PRODUTOS PÚBLICO (Para a Vitrine do Cliente Final)
func (h *ProdutoHandler) ListarProdutosPublico(w http.ResponseWriter, r *http.Request) {
	usuarioIDStr := r.URL.Query().Get("usuario_id")
	if usuarioIDStr == "" {
		http.Error(w, "Parâmetro usuario_id é obrigatório para a vitrine", http.StatusBadRequest)
		return
	}
	
	usuarioID, err := strconv.Atoi(usuarioIDStr)
	if err != nil || usuarioID <= 0 {
		http.Error(w, "usuario_id inválido", http.StatusBadRequest)
		return
	}

	rows, err := h.DB.Query(
		`SELECT id, nome, COALESCE(categoria,'Geral'), preco, quantidade,
         COALESCE(url_imagem::text,'[]')
         FROM produtos WHERE usuario_id = $1 AND visivel_catalogo = true AND quantidade > 0 ORDER BY id DESC`,
		usuarioID,
	)
	if err != nil {
		http.Error(w, "Erro ao buscar vitrine: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var produtos []models.ProdutoApp
	for rows.Next() {
		var p models.ProdutoApp
		var urlImagemRaw interface{}

		// AQUI: p.PrecoVenda
		err := rows.Scan(&p.ID, &p.Nome, &p.Categoria, &p.PrecoVenda, &p.Quantidade, &urlImagemRaw)
		if err != nil {
			continue
		}

		p.UsuarioID = usuarioID
		p.UrlImagem = scanUrlImagemJSONB(urlImagemRaw)
		p.VisivelCatalogo = true

		produtos = append(produtos, p)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(produtos)
}

// 🆕 CRIAR PRODUTO 
func (h *ProdutoHandler) Criar(w http.ResponseWriter, r *http.Request) {
	usuarioID, ok := UsuarioIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Não autorizado", http.StatusUnauthorized)
		return
	}

	var p models.ProdutoApp
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "Dados inválidos", http.StatusBadRequest)
		return
	}

	if p.UrlImagem == nil {
		p.UrlImagem = []string{}
	}
	urlImagemJSON, _ := json.Marshal(p.UrlImagem)

	visivel := true
	if !p.VisivelCatalogo {
		visivel = false
	}

	query := `INSERT INTO produtos (usuario_id, nome, categoria, preco, custo, quantidade, estoque_minimo, url_imagem, visivel_catalogo)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8::jsonb, $9)`

	// AQUI: p.PrecoVenda
	_, err := h.DB.Exec(query, usuarioID, p.Nome, p.Categoria, p.PrecoVenda, p.Custo, p.Quantidade, p.EstoqueMinimo, string(urlImagemJSON), visivel)
	if err != nil {
		http.Error(w, "Erro ao salvar no banco", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// ✏️ EDITAR PRODUTO
func (h *ProdutoHandler) Editar(w http.ResponseWriter, r *http.Request) {
	usuarioID, ok := UsuarioIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Não autorizado", http.StatusUnauthorized)
		return
	}

	var p models.ProdutoApp
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "Dados inválidos", http.StatusBadRequest)
		return
	}

	if p.UrlImagem == nil {
		p.UrlImagem = []string{}
	}
	urlImagemJSON, _ := json.Marshal(p.UrlImagem)

	_, err := h.DB.Exec(
		`UPDATE produtos SET nome=$1, categoria=$2, preco=$3, custo=$4, quantidade=$5, estoque_minimo=$6, url_imagem=$7::jsonb, visivel_catalogo=$8 WHERE id=$9 AND usuario_id=$10`,
		// AQUI: p.PrecoVenda
		p.Nome, p.Categoria, p.PrecoVenda, p.Custo, p.Quantidade, p.EstoqueMinimo, string(urlImagemJSON), p.VisivelCatalogo, p.ID, usuarioID,
	)
	if err != nil {
		http.Error(w, "Erro ao atualizar", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// 🗑️ DELETAR PRODUTO
func (h *ProdutoHandler) Deletar(w http.ResponseWriter, r *http.Request) {
	usuarioID, ok := UsuarioIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Não autorizado", http.StatusUnauthorized)
		return
	}

	var req struct {
		ID int `json:"id"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	_, err := h.DB.Exec("DELETE FROM produtos WHERE id = $1 AND usuario_id = $2", req.ID, usuarioID)
	if err != nil {
		http.Error(w, "Erro ao excluir", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// 🚚 REPOR ESTOQUE 
func (h *ProdutoHandler) Repor(w http.ResponseWriter, r *http.Request) {
	usuarioID, ok := UsuarioIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Não autorizado", http.StatusUnauthorized)
		return
	}

	var req struct {
		ID         int     `json:"id"`
		Quantidade int     `json:"quantidade"`
		CustoTotal float64 `json:"custo_total"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	if req.Quantidade <= 0 {
		http.Error(w, "Quantidade deve ser maior que zero", http.StatusBadRequest)
		return
	}

	var qtdAtual int
	var custoAtual float64
	
	err := h.DB.QueryRow("SELECT quantidade, custo FROM produtos WHERE id = $1 AND usuario_id = $2", req.ID, usuarioID).Scan(&qtdAtual, &custoAtual)
	if err != nil {
		http.Error(w, "Produto não encontrado", http.StatusNotFound)
		return
	}

	novaQtd := qtdAtual + req.Quantidade
	custoMedio := (custoAtual*float64(qtdAtual) + req.CustoTotal) / float64(novaQtd)

	_, err = h.DB.Exec("UPDATE produtos SET quantidade = $1, custo = $2 WHERE id = $3 AND usuario_id = $4", novaQtd, custoMedio, req.ID, usuarioID)
	if err != nil {
		http.Error(w, "Erro ao repor estoque", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
// 🛍️ LANÇAR VENDA (Abate Estoque e Calcula Total)
func (h *ProdutoHandler) Vender(w http.ResponseWriter, r *http.Request) {
	usuarioID, ok := UsuarioIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Não autorizado", http.StatusUnauthorized)
		return
	}

	var req struct {
		ID         int `json:"id"`
		Quantidade int `json:"quantidade"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	var qtdAtual int
	var preco float64
	var nome string
	
	// Confere se tem estoque
	err := h.DB.QueryRow("SELECT nome, quantidade, preco FROM produtos WHERE id=$1 AND usuario_id=$2", req.ID, usuarioID).Scan(&nome, &qtdAtual, &preco)
	if err != nil || qtdAtual < req.Quantidade {
		http.Error(w, "Estoque insuficiente", http.StatusBadRequest)
		return
	}

	// Abate o estoque na hora da venda
	h.DB.Exec("UPDATE produtos SET quantidade = quantidade - $1 WHERE id=$2 AND usuario_id=$3", req.Quantidade, req.ID, usuarioID)

	total := preco * float64(req.Quantidade)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"produto": nome,
		"total":   total,
	})
}