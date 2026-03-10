package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/theo-guerra/simple-shop/internal/models"
)

type ProdutoHandler struct {
	DB *sql.DB
}

// ListarProdutos (GET)
func (h *ProdutoHandler) ListarProdutos(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query("SELECT id, nome, preco, custo, quantidade, estoque_minimo, url_imagem FROM produtos ORDER BY id ASC")
	if err != nil {
		http.Error(w, "Erro ao buscar produtos", 500)
		return
	}
	defer rows.Close()

	var produtos []models.ProdutoApp
	for rows.Next() {
		var p models.ProdutoApp
		rows.Scan(&p.ID, &p.Nome, &p.PrecoVenda, &p.Custo, &p.Quantidade, &p.EstoqueMinimo, &p.UrlImagem)
		produtos = append(produtos, p)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(produtos)
}

// Criar (POST)
func (h *ProdutoHandler) Criar(w http.ResponseWriter, r *http.Request) {
	var p models.ProdutoApp
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "Dados inválidos", 400)
		return
	}

	query := `INSERT INTO produtos (nome, preco, custo, quantidade, estoque_minimo, url_imagem) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := h.DB.Exec(query, p.Nome, p.PrecoVenda, p.Custo, p.Quantidade, p.EstoqueMinimo, p.UrlImagem)
	if err != nil {
		http.Error(w, "Erro ao salvar produto", 500)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// Editar (POST)
func (h *ProdutoHandler) Editar(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID     int     `json:"id"`
		Nome   string  `json:"nome"`
		Preco  float64 `json:"preco"`
		Custo  float64 `json:"custo"`
		Minimo int     `json:"estoque_minimo"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	_, err := h.DB.Exec("UPDATE produtos SET nome=$1, preco=$2, custo=$3, estoque_minimo=$4 WHERE id=$5", req.Nome, req.Preco, req.Custo, req.Minimo, req.ID)
	if err != nil {
		http.Error(w, "Erro ao atualizar", 500)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// Vender (POST) - A MÁGICA ACONTECE AQUI
func (h *ProdutoHandler) Vender(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID         int `json:"id"`
		Quantidade int `json:"quantidade"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	tx, err := h.DB.Begin()
	if err != nil {
		return
	}

	var estoqueAtual int
	var preco float64
	var nome string

	err = tx.QueryRow("SELECT quantidade, preco, nome FROM produtos WHERE id = $1", req.ID).Scan(&estoqueAtual, &preco, &nome)
	if err != nil || estoqueAtual < req.Quantidade {
		tx.Rollback()
		http.Error(w, "Estoque insuficiente", 409)
		return
	}

	// 1. Reduz estoque E aumenta o ranking de "vendas_qtd" (Para o relatório de Mais Vendidos)
	_, err = tx.Exec("UPDATE produtos SET quantidade = quantidade - $1, vendas_qtd = vendas_qtd + $1 WHERE id = $2", req.Quantidade, req.ID)
	if err != nil {
		tx.Rollback()
		return
	}

	// 2. Não precisamos mais lançar a entrada no caixa aqui, porque o Front-End vai mandar a forma de pagamento (Dinheiro, Pix) direto para o CaixaHandler.

	tx.Commit()
	// Retornamos os dados do produto para o Front-End montar o Recibo do WhatsApp!
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"produto": nome,
		"preco":   preco,
		"total":   preco * float64(req.Quantidade),
	})
}

// Deletar (POST)
func (h *ProdutoHandler) Deletar(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID int `json:"id"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	h.DB.Exec("DELETE FROM produtos WHERE id = $1", req.ID)
	w.WriteHeader(http.StatusOK)
}
