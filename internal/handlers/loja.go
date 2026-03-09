package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/theo-guerra/simple-shop/internal/models"
)

type LojaHandler struct {
	DB *sql.DB
}

// ObterConfig puxa os dados de WhatsApp, Instagram e Nome da Loja
func (h *LojaHandler) ObterConfig(w http.ResponseWriter, r *http.Request) {
	var config models.LojaConfig
	err := h.DB.QueryRow("SELECT nome_loja, whatsapp, instagram FROM loja_config WHERE id = 1").
		Scan(&config.NomeLoja, &config.Whatsapp, &config.Instagram)

	if err != nil {
		http.Error(w, "Erro ao carregar configurações", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

// AtualizarConfig salva as edições feitas na configuração da loja
func (h *LojaHandler) AtualizarConfig(w http.ResponseWriter, r *http.Request) {
	var config models.LojaConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, "JSON inválido", 400)
		return
	}

	_, err := h.DB.Exec("UPDATE loja_config SET nome_loja = $1, whatsapp = $2, instagram = $3 WHERE id = 1",
		config.NomeLoja, config.Whatsapp, config.Instagram)

	if err != nil {
		http.Error(w, "Erro ao salvar", 500)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "✅ Configurações atualizadas!")
}

// ListarTarefas mostra a To-Do List do lojista
func (h *LojaHandler) ListarTarefas(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query("SELECT id, descricao, concluida FROM tarefas ORDER BY id ASC")
	if err != nil {
		http.Error(w, "Erro no banco", 500)
		return
	}
	defer rows.Close()

	var tarefas []models.Tarefa
	for rows.Next() {
		var t models.Tarefa
		rows.Scan(&t.ID, &t.Descricao, &t.Concluida)
		tarefas = append(tarefas, t)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tarefas)
}

// CriarTarefa adiciona um novo item na To-Do List
func (h *LojaHandler) CriarTarefa(w http.ResponseWriter, r *http.Request) {
	var t models.Tarefa
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		http.Error(w, "Dados inválidos", 400)
		return
	}

	_, err := h.DB.Exec("INSERT INTO tarefas (descricao) VALUES ($1)", t.Descricao)
	if err != nil {
		http.Error(w, "Erro ao salvar", 500)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintln(w, "✅ Tarefa adicionada!")
}
