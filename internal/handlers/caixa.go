package handlers

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/theo-guerra/simple-shop/internal/models"
)

type CaixaHandler struct {
	DB *sql.DB
}

func (h *CaixaHandler) DashboardMobile(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// 💰 LISTAR MOVIMENTOS (Com Barreira e Radar)
func (h *CaixaHandler) ListarMovimentosHoje(w http.ResponseWriter, r *http.Request) {
	usuarioID, ok := UsuarioIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Não autorizado", http.StatusUnauthorized)
		return
	}

	// Paginação: ?limite=50&pagina=1 (padrão: 50 por página)
	limite := 50
	pagina := 1
	if l, err := strconv.Atoi(r.URL.Query().Get("limite")); err == nil && l > 0 && l <= 200 {
		limite = l
	}
	if p, err := strconv.Atoi(r.URL.Query().Get("pagina")); err == nil && p > 0 {
		pagina = p
	}
	offset := (pagina - 1) * limite

	rows, err := h.DB.Query(
		"SELECT id, tipo, descricao, valor, data_mov FROM movimentos WHERE usuario_id = $1 ORDER BY id DESC LIMIT $2 OFFSET $3",
		usuarioID, limite, offset,
	)
	if err != nil {
		slog.Error("Erro ao buscar movimentos", "usuario_id", usuarioID, "erro", err)
		http.Error(w, "Erro ao buscar movimentos", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var movimentos []models.Movimento
	for rows.Next() {
		var m models.Movimento
		var data time.Time
		if err := rows.Scan(&m.ID, &m.Tipo, &m.Descricao, &m.Valor, &data); err == nil {
			m.DataMov = data.Format(time.RFC3339)
			movimentos = append(movimentos, m)
		}
	}

	if movimentos == nil {
		movimentos = []models.Movimento{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(movimentos)
}

// 💸 REGISTRAR MOVIMENTO (Com Radar)
func (h *CaixaHandler) RegistrarMovimento(w http.ResponseWriter, r *http.Request) {
	usuarioID, ok := UsuarioIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Não autorizado", http.StatusUnauthorized)
		return
	}

	var m models.Movimento
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, "Dados inválidos", http.StatusBadRequest)
		return
	}

	if m.Tipo != "entrada" && m.Tipo != "saida" {
		http.Error(w, "Tipo deve ser 'entrada' ou 'saida'", http.StatusBadRequest)
		return
	}
	if m.Valor <= 0 {
		http.Error(w, "Valor deve ser maior que zero", http.StatusBadRequest)
		return
	}

	_, err := h.DB.Exec("INSERT INTO movimentos (usuario_id, tipo, descricao, valor) VALUES ($1, $2, $3, $4)", usuarioID, m.Tipo, m.Descricao, m.Valor)
	if err != nil {
		slog.Error("Erro ao registrar movimento", "usuario_id", usuarioID, "erro", err)
		http.Error(w, "Erro ao registrar o dinheiro", http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusCreated)
}