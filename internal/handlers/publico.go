package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/theo-guerra/simple-shop/internal/models"
)

type PublicoHandler struct {
	DB *sql.DB
}

// 🏪 1. Busca o Nome e o WhatsApp da Loja baseada na URL
func (h *PublicoHandler) GetLojaByUrl(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	
	var config models.LojaConfig
	err := h.DB.QueryRow(`
		SELECT usuario_id, nome_loja, whatsapp 
		FROM loja_config 
		WHERE REPLACE(LOWER(nome_loja), ' ', '') = $1
	`, strings.ToLower(url)).Scan(&config.UsuarioID, &config.NomeLoja, &config.Whatsapp)

	if err != nil {
		http.Error(w, "Loja não encontrada", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

// 👗 2. Busca apenas os produtos marcados para aparecer no catálogo
func (h *PublicoHandler) GetProdutosVitrine(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	
	var usuarioID int
	err := h.DB.QueryRow(`SELECT usuario_id FROM loja_config WHERE REPLACE(LOWER(nome_loja), ' ', '') = $1`, strings.ToLower(url)).Scan(&usuarioID)
	if err != nil {
		return 
	}

	// 🔥 A CORREÇÃO ESTÁ AQUI: Trocamos "preco_venda" por "preco" para bater com o seu Banco de Dados!
	rows, err := h.DB.Query(`
		SELECT id, nome, categoria, preco, url_imagem 
		FROM produtos 
		WHERE usuario_id = $1 AND visivel_catalogo = true 
		ORDER BY id DESC
	`, usuarioID)
	
	if err != nil {
		fmt.Printf("❌ Erro no banco: %v\n", err)
		http.Error(w, "Erro ao buscar produtos", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var produtos []models.ProdutoApp
	for rows.Next() {
		var p models.ProdutoApp
		var urlImagemJSON sql.NullString
		
		// O Go pega o "preco" do banco e guarda na caixinha "PrecoVenda" do nosso molde
		if err := rows.Scan(&p.ID, &p.Nome, &p.Categoria, &p.PrecoVenda, &urlImagemJSON); err == nil {
			if urlImagemJSON.Valid {
				json.Unmarshal([]byte(urlImagemJSON.String), &p.UrlImagem)
			}
			produtos = append(produtos, p)
		}
	}

	if len(produtos) == 0 {
		produtos = []models.ProdutoApp{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(produtos)
}