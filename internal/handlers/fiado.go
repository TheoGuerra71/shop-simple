package handlers

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/theo-guerra/simple-shop/internal/models"
)

type FiadoHandler struct {
	DB *sql.DB
}

// CadastrarCliente (POST /api/clientes/novo)
func (h *FiadoHandler) CadastrarCliente(w http.ResponseWriter, r *http.Request) {
	usuarioID, ok := UsuarioIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Não autorizado", http.StatusUnauthorized)
		return
	}

	var c models.Cliente
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, "Dados inválidos", http.StatusBadRequest)
		return
	}

	if msg := validarCampoObrigatorio(c.Nome, "nome"); msg != "" {
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	err := h.DB.QueryRow(
		"INSERT INTO clientes (usuario_id, nome, telefone) VALUES ($1, $2, $3) RETURNING id",
		usuarioID, c.Nome, c.Telefone,
	).Scan(&c.ID)
	if err != nil {
		slog.Error("Erro ao salvar cliente", "usuario_id", usuarioID, "erro", err)
		http.Error(w, "Erro ao salvar cliente", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(c)
}

// ListarClientes (GET /api/clientes)
func (h *FiadoHandler) ListarClientes(w http.ResponseWriter, r *http.Request) {
	usuarioID, ok := UsuarioIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Não autorizado", http.StatusUnauthorized)
		return
	}

	rows, err := h.DB.Query("SELECT id, nome, telefone FROM clientes WHERE usuario_id = $1 ORDER BY nome ASC", usuarioID)
	if err != nil {
		slog.Error("Erro ao buscar clientes", "usuario_id", usuarioID, "erro", err)
		http.Error(w, "Erro ao buscar clientes", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var clientes []models.Cliente
	for rows.Next() {
		var c models.Cliente
		if err := rows.Scan(&c.ID, &c.Nome, &c.Telefone); err != nil {
			continue
		}
		c.UsuarioID = usuarioID
		clientes = append(clientes, c)
	}

	if clientes == nil {
		clientes = []models.Cliente{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(clientes)
}

// ListarFiados (GET /api/fiados)
func (h *FiadoHandler) ListarFiados(w http.ResponseWriter, r *http.Request) {
	usuarioID, ok := UsuarioIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Não autorizado", http.StatusUnauthorized)
		return
	}

	query := `
		SELECT f.id, f.cliente_id, c.nome, f.valor, f.descricao, f.data_divida, f.pago
		FROM fiados f
		JOIN clientes c ON f.cliente_id = c.id
		WHERE f.usuario_id = $1
		ORDER BY f.pago ASC, f.data_divida DESC
	`
	rows, err := h.DB.Query(query, usuarioID)
	if err != nil {
		slog.Error("Erro ao buscar fiados", "usuario_id", usuarioID, "erro", err)
		http.Error(w, "Erro ao buscar fiados", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var fiados []models.Fiado
	for rows.Next() {
		var f models.Fiado
		if err := rows.Scan(&f.ID, &f.ClienteID, &f.NomeCliente, &f.Valor, &f.Descricao, &f.DataDivida, &f.Pago); err != nil {
			continue
		}
		fiados = append(fiados, f)
	}

	if fiados == nil {
		fiados = []models.Fiado{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fiados)
}

// NovoFiado (POST /api/fiados/novo)
func (h *FiadoHandler) NovoFiado(w http.ResponseWriter, r *http.Request) {
	usuarioID, ok := UsuarioIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Não autorizado", http.StatusUnauthorized)
		return
	}

	var f models.Fiado
	if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
		http.Error(w, "Dados inválidos", http.StatusBadRequest)
		return
	}

	if f.ClienteID <= 0 {
		http.Error(w, "Cliente é obrigatório", http.StatusBadRequest)
		return
	}
	if f.Valor <= 0 {
		http.Error(w, "Valor deve ser maior que zero", http.StatusBadRequest)
		return
	}

	_, err := h.DB.Exec(
		"INSERT INTO fiados (usuario_id, cliente_id, valor, descricao) VALUES ($1, $2, $3, $4)",
		usuarioID, f.ClienteID, f.Valor, f.Descricao,
	)
	if err != nil {
		slog.Error("Erro ao registrar fiado", "usuario_id", usuarioID, "erro", err)
		http.Error(w, "Erro ao registrar fiado", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// DarBaixa (POST /api/fiados/pagar)
func (h *FiadoHandler) DarBaixa(w http.ResponseWriter, r *http.Request) {
	usuarioID, ok := UsuarioIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Não autorizado", http.StatusUnauthorized)
		return
	}

	var req models.BaixaFiadoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Dados inválidos", http.StatusBadRequest)
		return
	}

	result, err := h.DB.Exec("UPDATE fiados SET pago = TRUE WHERE id = $1 AND usuario_id = $2", req.ID, usuarioID)
	if err != nil {
		slog.Error("Erro ao dar baixa", "usuario_id", usuarioID, "fiado_id", req.ID, "erro", err)
		http.Error(w, "Erro ao dar baixa", http.StatusInternalServerError)
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		http.Error(w, "Fiado não encontrado", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}
