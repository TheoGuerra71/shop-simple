package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/theo-guerra/simple-shop/internal/models"
)

type CatalogoHandler struct {
	DB *sql.DB
}

// ServirCatalogo (GET /api/catalogo) - Retorna apenas produtos com estoque > 0
func (h *CatalogoHandler) ServirCatalogo(w http.ResponseWriter, r *http.Request) {
	// O cliente não precisa ver o custo ou o ID interno, apenas nome, preço e se tem estoque
	rows, err := h.DB.Query("SELECT nome, preco, quantidade, url_imagem FROM produtos WHERE quantidade > 0 ORDER BY nome ASC")
	if err != nil {
		http.Error(w, "Erro ao carregar catálogo", 500)
		return
	}
	defer rows.Close()

	var catalogo []models.ProdutoApp
	for rows.Next() {
		var p models.ProdutoApp
		rows.Scan(&p.Nome, &p.PrecoVenda, &p.Quantidade, &p.UrlImagem)
		catalogo = append(catalogo, p)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(catalogo)
}
