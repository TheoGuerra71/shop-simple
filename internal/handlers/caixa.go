package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/theo-guerra/simple-shop/internal/models"
)

// CaixaHandler gerencia as rotas do painel financeiro do aplicativo.
type CaixaHandler struct {
	DB *sql.DB
}

// DashboardMobile (GET /caixa/dashboard) - Calcula a inteligência financeira do App
func (h *CaixaHandler) DashboardMobile(w http.ResponseWriter, r *http.Request) {
	var resumo models.ResumoApp

	// 1. Vendas Hoje (Soma das Entradas com a data exata de hoje)
	h.DB.QueryRow("SELECT COALESCE(SUM(valor), 0) FROM caixa_movimentos WHERE tipo = 'ENTRADA' AND DATE(data_mov) = CURRENT_DATE").Scan(&resumo.VendasHoje)

	// 2. Vendas no Mês (Soma das Entradas no mês e ano atuais)
	h.DB.QueryRow("SELECT COALESCE(SUM(valor), 0) FROM caixa_movimentos WHERE tipo = 'ENTRADA' AND EXTRACT(MONTH FROM data_mov) = EXTRACT(MONTH FROM CURRENT_DATE) AND EXTRACT(YEAR FROM data_mov) = EXTRACT(YEAR FROM CURRENT_DATE)").Scan(&resumo.VendasMes)

	// 3. Ticket Médio (Total de Vendas Hoje dividido pela Quantidade de Vendas Hoje)
	var qtdVendasHoje int
	h.DB.QueryRow("SELECT COUNT(*) FROM caixa_movimentos WHERE tipo = 'ENTRADA' AND DATE(data_mov) = CURRENT_DATE").Scan(&qtdVendasHoje)
	if qtdVendasHoje > 0 {
		resumo.TicketMedio = resumo.VendasHoje / float64(qtdVendasHoje)
	}

	// 4. Itens em Estoque (Total de produtos cadastrados no catálogo)
	h.DB.QueryRow("SELECT COUNT(*) FROM produtos").Scan(&resumo.ItensEstoque)

	// 5. Total de Saídas (Soma das Saídas com a data exata de hoje)
	h.DB.QueryRow("SELECT COALESCE(SUM(valor), 0) FROM caixa_movimentos WHERE tipo = 'SAIDA' AND DATE(data_mov) = CURRENT_DATE").Scan(&resumo.TotalSaidas)

	// 6. Fechamento de Caixa e Lucro (Entradas - Saídas do dia)
	resumo.TotalEntradas = resumo.VendasHoje
	resumo.LucroLiquido = resumo.TotalEntradas - resumo.TotalSaidas

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resumo)
}

// RegistrarMovimento (POST /caixa/movimento) - Lança Entradas ou Saídas manuais
func (h *CaixaHandler) RegistrarMovimento(w http.ResponseWriter, r *http.Request) {
	var mov models.Movimentacao
	if err := json.NewDecoder(r.Body).Decode(&mov); err != nil {
		http.Error(w, "Dados inválidos", 400)
		return
	}

	_, err := h.DB.Exec("INSERT INTO caixa_movimentos (tipo, descricao, valor) VALUES ($1, $2, $3)", mov.Tipo, mov.Descricao, mov.Valor)
	if err != nil {
		http.Error(w, "Erro ao registrar no caixa", 500)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// ListarMovimentosHoje (GET /caixa/movimentos/hoje) - Alimenta a lista da tela inicial do App
func (h *CaixaHandler) ListarMovimentosHoje(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query("SELECT id, tipo, descricao, valor, data_mov FROM caixa_movimentos WHERE DATE(data_mov) = CURRENT_DATE ORDER BY id DESC")
	if err != nil {
		http.Error(w, "Erro ao buscar movimentos", 500)
		return
	}
	defer rows.Close()

	var movimentos []models.Movimentacao
	for rows.Next() {
		var m models.Movimentacao
		rows.Scan(&m.ID, &m.Tipo, &m.Descricao, &m.Valor, &m.DataMov)
		movimentos = append(movimentos, m)
	}

	if movimentos == nil {
		movimentos = []models.Movimentacao{} // Evita retornar 'null' para o front-end
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(movimentos)
}
