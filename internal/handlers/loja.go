package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/theo-guerra/simple-shop/internal/models"
)

type LojaHandler struct {
	DB *sql.DB
}

// defaultLojaConfig retorna um design padrão caso o lojista ainda não tenha salvo nada
func defaultLojaConfig() models.LojaConfig {
	return models.LojaConfig{NomeLoja: "Minha Loja", CorHex: "#1c1917", MsgSuporte: "Olá!"}
}

// ⚙️ Config (GET/POST protegido) — Lê e salva a configuração do painel do lojista logado.
func (h *LojaHandler) Config(w http.ResponseWriter, r *http.Request) {
	usuarioID, ok := UsuarioIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Não autorizado", http.StatusUnauthorized)
		return
	}

	// GET: Devolve os dados para preencher os campos na tela
	if r.Method == http.MethodGet {
		var c models.LojaConfig
		err := h.DB.QueryRow(
			"SELECT nome_loja, whatsapp, instagram, cor_hex, msg_suporte FROM loja_config WHERE usuario_id = $1",
			usuarioID,
		).Scan(&c.NomeLoja, &c.Whatsapp, &c.Instagram, &c.CorHex, &c.MsgSuporte)
		
		if err != nil {
			c = defaultLojaConfig()
			c.UsuarioID = usuarioID
		} else {
			c.UsuarioID = usuarioID
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(c)
		return
	}

	// POST: Salva as alterações feitas no formulário
	if r.Method == http.MethodPost {
		var req models.LojaConfig
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Dados inválidos", http.StatusBadRequest)
			return
		}
		if req.CorHex == "" {
			req.CorHex = "#1c1917"
		}
		
		// Insere a loja nova ou atualiza se já existir
		_, err := h.DB.Exec(
			`INSERT INTO loja_config (usuario_id, nome_loja, whatsapp, instagram, cor_hex, msg_suporte)
             VALUES ($1, $2, $3, $4, $5, $6)
             ON CONFLICT (usuario_id) DO UPDATE SET nome_loja=$2, whatsapp=$3, instagram=$4, cor_hex=$5, msg_suporte=$6`,
			usuarioID, req.NomeLoja, req.Whatsapp, req.Instagram, req.CorHex, req.MsgSuporte,
		)
		if err != nil {
			http.Error(w, "Erro ao salvar: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

// 🌐 ConfigPublico (GET público) — A MÁGICA DA VANITY URL!
// Agora ele busca a loja usando o "?slug=minhamonalisa" em vez do número do ID.
func (h *LojaHandler) ConfigPublico(w http.ResponseWriter, r *http.Request) {
	slug := r.URL.Query().Get("slug")
	if slug == "" {
		http.Error(w, "Slug inválido", http.StatusBadRequest)
		return
	}

	var c models.LojaConfig
	var uid int

	// O Motor Inteligente: Ele pega o nome no banco (ex: "Minha Monalisa"), 
	// tira os espaços, joga pra letra minúscula ("minhamonalisa") e compara com o link!
	err := h.DB.QueryRow(
		"SELECT usuario_id, nome_loja, whatsapp, instagram, cor_hex, msg_suporte FROM loja_config WHERE REPLACE(LOWER(nome_loja), ' ', '') = $1",
		strings.ToLower(slug),
	).Scan(&uid, &c.NomeLoja, &c.Whatsapp, &c.Instagram, &c.CorHex, &c.MsgSuporte)

	if err != nil {
		// Se digitar o nome da loja errado, retorna erro 404 para o front mostrar a mensagem de "Loja não encontrada"
		http.Error(w, "Loja não encontrada", http.StatusNotFound)
		return
	}

	// Devolvemos o ID real da loja de forma oculta para o JS poder buscar os produtos certos
	c.UsuarioID = uid
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(c)
}