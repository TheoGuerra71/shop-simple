package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/theo-guerra/simple-shop/internal/models"
)

type CaixaHandler struct {
	DB *sql.DB
}

type DashboardResponse struct {
	VendasHoje     float64 `json:"vendas_hoje"`
	VendasOntem    float64 `json:"vendas_ontem"`
	VendasMes      float64 `json:"vendas_mes"`
	TicketMedio    float64 `json:"ticket_medio"`
	TotalEntradas  float64 `json:"total_entradas"`
	TotalSaidas    float64 `json:"total_saidas"`
	LucroLiquido   float64 `json:"lucro_liquido"`
	PixHoje        float64 `json:"pix_hoje"`
	CartaoHoje     float64 `json:"cartao_hoje"`
	DinheiroHoje   float64 `json:"dinheiro_hoje"`
	AlertasEstoque int     `json:"alertas_estoque"`
}

func (h *CaixaHandler) DashboardMobile(w http.ResponseWriter, r *http.Request) {
	var d DashboardResponse
	h.DB.QueryRow("SELECT COALESCE(SUM(valor), 0) FROM caixa_movimentos WHERE tipo = 'ENTRADA' AND DATE(data_mov) = CURRENT_DATE").Scan(&d.VendasHoje)
	h.DB.QueryRow("SELECT COALESCE(SUM(valor), 0) FROM caixa_movimentos WHERE tipo = 'ENTRADA' AND DATE(data_mov) = CURRENT_DATE - INTERVAL '1 day'").Scan(&d.VendasOntem)
	h.DB.QueryRow("SELECT COALESCE(SUM(valor), 0) FROM caixa_movimentos WHERE tipo = 'ENTRADA' AND EXTRACT(MONTH FROM data_mov) = EXTRACT(MONTH FROM CURRENT_DATE) AND EXTRACT(YEAR FROM data_mov) = EXTRACT(YEAR FROM CURRENT_DATE)").Scan(&d.VendasMes)
	h.DB.QueryRow("SELECT COALESCE(SUM(valor), 0) FROM caixa_movimentos WHERE tipo = 'ENTRADA' AND DATE(data_mov) = CURRENT_DATE AND descricao LIKE '%[Pix]%'").Scan(&d.PixHoje)
	h.DB.QueryRow("SELECT COALESCE(SUM(valor), 0) FROM caixa_movimentos WHERE tipo = 'ENTRADA' AND DATE(data_mov) = CURRENT_DATE AND (descricao LIKE '%[Cartão]%' OR descricao LIKE '%[Crédito]%' OR descricao LIKE '%[Débito]%')").Scan(&d.CartaoHoje)
	h.DB.QueryRow("SELECT COALESCE(SUM(valor), 0) FROM caixa_movimentos WHERE tipo = 'ENTRADA' AND DATE(data_mov) = CURRENT_DATE AND descricao LIKE '%[Dinheiro]%'").Scan(&d.DinheiroHoje)
	h.DB.QueryRow("SELECT COALESCE(SUM(valor), 0) FROM caixa_movimentos WHERE tipo = 'SAIDA' AND DATE(data_mov) = CURRENT_DATE").Scan(&d.TotalSaidas)

	d.TotalEntradas = d.VendasHoje
	d.LucroLiquido = d.TotalEntradas - d.TotalSaidas

	var qtdVendas int
	h.DB.QueryRow("SELECT COUNT(*) FROM caixa_movimentos WHERE tipo = 'ENTRADA' AND DATE(data_mov) = CURRENT_DATE").Scan(&qtdVendas)
	if qtdVendas > 0 {
		d.TicketMedio = d.VendasHoje / float64(qtdVendas)
	}
	h.DB.QueryRow("SELECT COUNT(*) FROM produtos WHERE quantidade <= estoque_minimo").Scan(&d.AlertasEstoque)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(d)
}

// 🚀 O SEGREDO DO BEZOS: Puxar os últimos 30 dias ao invés de apenas hoje!
func (h *CaixaHandler) ListarMovimentosHoje(w http.ResponseWriter, r *http.Request) {
	// Puxa histórico de 30 dias para o JS processar os gráficos e extratos granulares
	rows, err := h.DB.Query("SELECT id, tipo, descricao, valor, data_mov FROM caixa_movimentos WHERE data_mov >= CURRENT_DATE - INTERVAL '30 days' ORDER BY data_mov DESC")
	if err != nil {
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
		movimentos = []models.Movimentacao{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(movimentos)
}

func (h *CaixaHandler) RegistrarMovimento(w http.ResponseWriter, r *http.Request) {
	var mov models.Movimentacao
	json.NewDecoder(r.Body).Decode(&mov)
	h.DB.Exec("INSERT INTO caixa_movimentos (tipo, descricao, valor) VALUES ($1, $2, $3)", mov.Tipo, mov.Descricao, mov.Valor)
	w.WriteHeader(http.StatusCreated)
}
