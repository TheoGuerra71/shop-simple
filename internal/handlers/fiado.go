package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/theo-guerra/simple-shop/internal/models"
)

type FiadoHandler struct {
	DB *sql.DB
}

// CadastrarCliente (POST /clientes/novo) - Salva o cliente e retorna o ID gerado.
func (h *FiadoHandler) CadastrarCliente(w http.ResponseWriter, r *http.Request) {
	var c models.Cliente
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, "Dados inválidos", 400)
		return
	}

	// O RETURNING id devolve o número gerado pelo PostgreSQL na hora da criação
	err := h.DB.QueryRow("INSERT INTO clientes (nome, telefone) VALUES ($1, $2) RETURNING id", c.Nome, c.Telefone).Scan(&c.ID)
	if err != nil {
		http.Error(w, "Erro ao salvar cliente", 500)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(c)
}

// ListarFiados (GET /fiados) - Mostra todas as dívidas pendentes e pagas.
func (h *FiadoHandler) ListarFiados(w http.ResponseWriter, r *http.Request) {
	// JOIN une a tabela de fiados com a de clientes para pegarmos o nome da pessoa
	query := `
		SELECT f.id, f.cliente_id, c.nome, f.valor, f.descricao, f.data_divida, f.pago
		FROM fiados f
		JOIN clientes c ON f.cliente_id = c.id
		ORDER BY f.pago ASC, f.data_divida DESC
	`
	rows, err := h.DB.Query(query)
	if err != nil {
		http.Error(w, "Erro ao buscar fiados", 500)
		return
	}
	defer rows.Close()

	var fiados []models.Fiado
	for rows.Next() {
		var f models.Fiado
		rows.Scan(&f.ID, &f.ClienteID, &f.NomeCliente, &f.Valor, &f.Descricao, &f.DataDivida, &f.Pago)
		fiados = append(fiados, f)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fiados)
}

// NovoFiado (POST /fiados/novo) - Adiciona uma dívida ao caderno.
func (h *FiadoHandler) NovoFiado(w http.ResponseWriter, r *http.Request) {
	var f models.Fiado
	if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
		http.Error(w, "Dados inválidos", 400)
		return
	}

	_, err := h.DB.Exec("INSERT INTO fiados (cliente_id, valor, descricao) VALUES ($1, $2, $3)",
		f.ClienteID, f.Valor, f.Descricao)
	if err != nil {
		http.Error(w, "Erro ao registrar fiado", 500)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintln(w, "✅ Fiado registrado com sucesso!")
}

// DarBaixa (POST /fiados/pagar) - Muda o status da dívida para paga (TRUE).
func (h *FiadoHandler) DarBaixa(w http.ResponseWriter, r *http.Request) {
	var req models.BaixaFiadoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "ID inválido", 400)
		return
	}

	_, err := h.DB.Exec("UPDATE fiados SET pago = TRUE WHERE id = $1", req.ID)
	if err != nil {
		http.Error(w, "Erro ao dar baixa", 500)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "✅ Dívida %d marcada como paga!", req.ID)
}
