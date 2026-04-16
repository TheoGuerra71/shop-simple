package handlers

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
)

type RelatorioHandler struct {
	DB *sql.DB
}

// TopProdutos (GET /api/relatorios/top) - Ranking dos mais vendidos do lojista
func (h *RelatorioHandler) TopProdutos(w http.ResponseWriter, r *http.Request) {
	usuarioID, ok := UsuarioIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Não autorizado", http.StatusUnauthorized)
		return
	}

	rows, err := h.DB.Query(
		"SELECT nome, preco, vendas_qtd FROM produtos WHERE usuario_id = $1 AND vendas_qtd > 0 ORDER BY vendas_qtd DESC LIMIT 10",
		usuarioID,
	)
	if err != nil {
		slog.Error("Erro ao gerar ranking", "usuario_id", usuarioID, "erro", err)
		http.Error(w, "Erro ao gerar ranking", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var ranking []map[string]interface{}
	for rows.Next() {
		var nome string
		var preco float64
		var vendas int
		if err := rows.Scan(&nome, &preco, &vendas); err != nil {
			continue
		}
		ranking = append(ranking, map[string]interface{}{
			"nome":             nome,
			"preco":            preco,
			"vendas":           vendas,
			"total_arrecadado": preco * float64(vendas),
		})
	}

	if ranking == nil {
		ranking = []map[string]interface{}{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ranking)
}

// ExtratoCompleto (GET /api/relatorios/extrato) - Livro Caixa do lojista
func (h *RelatorioHandler) ExtratoCompleto(w http.ResponseWriter, r *http.Request) {
	usuarioID, ok := UsuarioIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Não autorizado", http.StatusUnauthorized)
		return
	}

	rows, err := h.DB.Query(
		"SELECT id, tipo, descricao, valor, data_mov FROM movimentos WHERE usuario_id = $1 ORDER BY data_mov DESC LIMIT 100",
		usuarioID,
	)
	if err != nil {
		slog.Error("Erro ao buscar extrato", "usuario_id", usuarioID, "erro", err)
		http.Error(w, "Erro ao buscar extrato", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var extrato []map[string]interface{}
	for rows.Next() {
		var id int
		var tipo, desc, data string
		var valor float64
		if err := rows.Scan(&id, &tipo, &desc, &valor, &data); err != nil {
			continue
		}
		extrato = append(extrato, map[string]interface{}{
			"id": id, "tipo": tipo, "descricao": desc, "valor": valor, "data": data,
		})
	}

	if extrato == nil {
		extrato = []map[string]interface{}{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(extrato)
}
