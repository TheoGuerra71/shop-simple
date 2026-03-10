package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

type RelatorioHandler struct {
	DB *sql.DB
}

// TopProdutos (GET /api/relatorios/top) - Retorna o ranking de mais vendidos
func (h *RelatorioHandler) TopProdutos(w http.ResponseWriter, r *http.Request) {
	// Puxa os 10 produtos que mais saíram da loja
	rows, err := h.DB.Query("SELECT nome, preco, vendas_qtd FROM produtos WHERE vendas_qtd > 0 ORDER BY vendas_qtd DESC LIMIT 10")
	if err != nil {
		http.Error(w, "Erro ao gerar ranking", 500)
		return
	}
	defer rows.Close()

	var ranking []map[string]interface{}
	for rows.Next() {
		var nome string
		var preco float64
		var vendas int
		rows.Scan(&nome, &preco, &vendas)

		ranking = append(ranking, map[string]interface{}{
			"nome":             nome,
			"preco":            preco,
			"vendas":           vendas,
			"total_arrecadado": preco * float64(vendas),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ranking)
}

// ExtratoCompleto (GET /api/relatorios/extrato) - Traz o Livro Caixa detalhado
func (h *RelatorioHandler) ExtratoCompleto(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query("SELECT id, tipo, descricao, valor, data_mov FROM caixa_movimentos ORDER BY data_mov DESC LIMIT 100")
	if err != nil {
		http.Error(w, "Erro ao buscar extrato", 500)
		return
	}
	defer rows.Close()

	var extrato []map[string]interface{}
	for rows.Next() {
		var id int
		var tipo, desc, data string
		var valor float64
		rows.Scan(&id, &tipo, &desc, &valor, &data)

		extrato = append(extrato, map[string]interface{}{
			"id": id, "tipo": tipo, "descricao": desc, "valor": valor, "data": data,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(extrato)
}
