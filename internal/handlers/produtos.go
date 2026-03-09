package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/theo-guerra/simple-shop/internal/models"
)

type ProdutoHandler struct {
	DB *sql.DB
}

// ListarProdutos (GET) - Traz o estoque atual
func (h *ProdutoHandler) ListarProdutos(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query("SELECT id, nome, preco, quantidade FROM produtos ORDER BY id ASC")
	if err != nil {
		http.Error(w, "Erro no banco", 500)
		return
	}
	defer rows.Close()

	var produtos []models.Produto
	for rows.Next() {
		var p models.Produto
		rows.Scan(&p.ID, &p.Nome, &p.Preco, &p.Quantidade)
		produtos = append(produtos, p)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(produtos)
}

// Criar (POST) - ESSA É A FUNÇÃO QUE ESTAVA FALTANDO! Cadastra novos itens.
func (h *ProdutoHandler) Criar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", 405)
		return
	}

	var p models.Produto
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "Dados inválidos", 400)
		return
	}

	_, err := h.DB.Exec("INSERT INTO produtos (nome, preco, quantidade) VALUES ($1, $2, $3)", p.Nome, p.Preco, p.Quantidade)
	if err != nil {
		http.Error(w, "Erro ao salvar", 500)
		return
	}

	log.Printf("🆕 Produto cadastrado: %s", p.Nome)
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintln(w, "✅ Produto cadastrado!")
}

// Vender (POST) - Reduz o estoque com segurança (Transação ACID)
func (h *ProdutoHandler) Vender(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", 405)
		return
	}

	var req models.VendaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Dados de venda inválidos", 400)
		return
	}

	tx, err := h.DB.Begin()
	if err != nil {
		http.Error(w, "Erro interno", 500)
		return
	}

	var estoqueAtual int
	err = tx.QueryRow("SELECT quantidade FROM produtos WHERE id = $1", req.ID).Scan(&estoqueAtual)
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

	_, err = tx.Exec("UPDATE produtos SET quantidade = quantidade - $1 WHERE id = $2", req.Quantidade, req.ID)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Falha na atualização", 500)
		return
	}

	tx.Commit()
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "✅ Venda processada!")
}

// Deletar (POST) - Remove o item do estoque permanentemente
func (h *ProdutoHandler) Deletar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", 405)
		return
	}

	var req models.DeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "ID inválido", 400)
		return
	}

	res, err := h.DB.Exec("DELETE FROM produtos WHERE id = $1", req.ID)
	if err != nil {
		http.Error(w, "Erro ao deletar", 500)
		return
	}

	count, _ := res.RowsAffected()
	if count == 0 {
		http.Error(w, "Produto não encontrado", 404)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "✅ Produto removido!")
}
