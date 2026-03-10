package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/theo-guerra/simple-shop/internal/models"
)

type ProdutoHandler struct {
	DB *sql.DB
}

// ListarProdutos (GET) - Agora traz custo e estoque mínimo
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

// Criar (POST) - Salva com os novos campos financeiros
func (h *ProdutoHandler) Criar(w http.ResponseWriter, r *http.Request) {
	var p models.ProdutoApp
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "Dados inválidos", 400)
		return
	}

	query := `INSERT INTO produtos (nome, preco, custo, quantidade, estoque_minimo, url_imagem) 
              VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := h.DB.Exec(query, p.Nome, p.PrecoVenda, p.Custo, p.Quantidade, p.EstoqueMinimo, p.UrlImagem)
	if err != nil {
		http.Error(w, "Erro ao salvar produto", 500)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// Vender (POST) - A função mais importante: Baixa estoque + Gera Entrada no Caixa
func (h *ProdutoHandler) Vender(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID         int `json:"id"`
		Quantidade int `json:"quantidade"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Erro nos dados", 400)
		return
	}

	tx, err := h.DB.Begin()
	if err != nil {
		http.Error(w, "Erro na transação", 500)
		return
	}

	var estoqueAtual int
	var preco float64
	var nome string

	// Busca dados do produto
	err = tx.QueryRow("SELECT quantidade, preco, nome FROM produtos WHERE id = $1", req.ID).Scan(&estoqueAtual, &preco, &nome)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Produto não encontrado", 404)
		return
	}

	if estoqueAtual < req.Quantidade {
		tx.Rollback()
		http.Error(w, "Estoque insuficiente", 409)
		return
	}

	// 1. Reduz o Estoque
	_, err = tx.Exec("UPDATE produtos SET quantidade = quantidade - $1 WHERE id = $2", req.Quantidade, req.ID)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Erro ao atualizar estoque", 500)
		return
	}

	// 2. Registra a ENTRADA no Caixa (Para o gráfico de lucro e vendas hoje)
	valorTotal := preco * float64(req.Quantidade)
	desc := fmt.Sprintf("Venda: %dx %s", req.Quantidade, nome)
	_, err = tx.Exec("INSERT INTO caixa_movimentos (tipo, descricao, valor) VALUES ('ENTRADA', $1, $2)", desc, valorTotal)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Erro ao registrar no caixa", 500)
		return
	}

	tx.Commit()
	w.WriteHeader(http.StatusOK)
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
