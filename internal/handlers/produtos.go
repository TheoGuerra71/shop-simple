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

// ListarProdutos (GET /produtos) exibe o estoque atualizado.
func (h *ProdutoHandler) ListarProdutos(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query("SELECT id, nome, preco, quantidade FROM produtos ORDER BY id ASC")
	if err != nil {
		http.Error(w, "Erro ao buscar dados", 500)
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

// Vender (POST /produtos/vender) realiza baixa segura no estoque.
func (h *ProdutoHandler) Vender(w http.ResponseWriter, r *http.Request) {
	var req models.VendaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Dados de venda inválidos", 400)
		return
	}

	tx, err := h.DB.Begin()
	if err != nil {
		http.Error(w, "Erro na transação", 500)
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
		http.Error(w, "Erro no update", 500)
		return
	}

	tx.Commit()
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "✅ Venda de %d unidade(s) do ID %d processada!", req.Quantidade, req.ID)
}

// Deletar (POST /produtos/deletar) remove um item permanentemente.
func (h *ProdutoHandler) Deletar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", 405)
		return
	}

	var req models.DeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "ID inválido para exclusão", 400)
		return
	}

	// Executa o comando SQL para remover o registro baseado no ID.
	res, err := h.DB.Exec("DELETE FROM produtos WHERE id = $1", req.ID)
	if err != nil {
		http.Error(w, "Erro ao deletar item", 500)
		return
	}

	// Verifica se alguma linha foi realmente afetada (se o ID existia).
	count, _ := res.RowsAffected()
	if count == 0 {
		http.Error(w, "Produto não encontrado para remoção", 404)
		return
	}

	log.Printf("🗑️ Produto ID %d removido do sistema.", req.ID)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "✅ Produto ID %d removido com sucesso!", req.ID)
}
