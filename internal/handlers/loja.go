package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/theo-guerra/simple-shop/internal/models"
)

type LojaHandler struct {
	DB *sql.DB
}

func (h *LojaHandler) Config(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		var c models.LojaConfig
		err := h.DB.QueryRow("SELECT nome_loja, whatsapp, instagram, cor_hex, msg_suporte FROM loja_config WHERE id = 1").
			Scan(&c.NomeLoja, &c.Whatsapp, &c.Instagram, &c.CorHex, &c.MsgSuporte)
		if err != nil {
			c = models.LojaConfig{NomeLoja: "Minha Loja", CorHex: "#10b981", MsgSuporte: "Olá!"}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(c)
	} else if r.Method == http.MethodPost {
		var req models.LojaConfig
		json.NewDecoder(r.Body).Decode(&req)

		// Se a cor vier vazia, garante um padrão
		if req.CorHex == "" {
			req.CorHex = "#10b981"
		}

		_, err := h.DB.Exec("UPDATE loja_config SET nome_loja=$1, whatsapp=$2, instagram=$3, cor_hex=$4, msg_suporte=$5 WHERE id=1",
			req.NomeLoja, req.Whatsapp, req.Instagram, req.CorHex, req.MsgSuporte)
		if err != nil {
			http.Error(w, "Erro ao salvar", 500)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
