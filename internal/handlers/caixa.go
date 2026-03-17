package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
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

	// 🕵️ LOG DE INVESTIGAÇÃO: Vai aparecer no terminal!
	fmt.Printf("\n🔍 Buscando dinheiro da gaveta do Lojista ID: %d\n", usuarioID)

	// 🛡️ A BARREIRA DE SEGURANÇA NA QUERY: "WHERE usuario_id = $1"
	rows, err := h.DB.Query("SELECT id, tipo, descricao, valor, data_mov FROM movimentos WHERE usuario_id = $1 ORDER BY id DESC", usuarioID)
	if err != nil {
		fmt.Println("❌ Erro no banco ao buscar movimentos:", err)
		http.Error(w, "Erro ao buscar movimentos", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var movimentos []models.Movimento
	var totalEncontrado int

	for rows.Next() {
		var m models.Movimento
		var data time.Time
		if err := rows.Scan(&m.ID, &m.Tipo, &m.Descricao, &m.Valor, &data); err == nil {
			m.DataMov = data.Format(time.RFC3339)
			movimentos = append(movimentos, m)
			totalEncontrado++
		}
	}

	fmt.Printf("✅ Lojista ID %d encontrou %d transações na gaveta dele.\n", usuarioID, totalEncontrado)

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

	fmt.Printf("💰 Lojista ID %d registrando novo movimento de R$ %.2f\n", usuarioID, m.Valor)

	_, err := h.DB.Exec("INSERT INTO movimentos (usuario_id, tipo, descricao, valor) VALUES ($1, $2, $3, $4)", usuarioID, m.Tipo, m.Descricao, m.Valor)
	if err != nil {
		fmt.Println("❌ Erro ao salvar movimento no banco:", err)
		http.Error(w, "Erro ao registrar o dinheiro", http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusCreated)
}