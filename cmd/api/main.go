package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/theo-guerra/simple-shop/internal/database"
	"github.com/theo-guerra/simple-shop/internal/handlers"
)

func main() {
	// Lembre de usar a sua senha do PostgreSQL
	connStr := "user=postgres password=admin dbname=simple_shop host=localhost sslmode=disable"

	db, err := database.Conectar(connStr)
	if err != nil {
		log.Fatal("Falha crítica no banco:", err)
	}
	defer db.Close()

	h := &handlers.ProdutoHandler{DB: db}

	// 1. ROTAS DA API (O motor do backend)
	http.HandleFunc("/produtos", h.ListarProdutos)
	http.HandleFunc("/produtos/vender", h.Vender)
	http.HandleFunc("/produtos/deletar", h.Deletar)

	// 2. ROTA DO FRONT-END (A carroceria visual)
	// Isso diz ao Go para servir os arquivos da pasta "static" quando acessarmos a raiz do site "/"
	http.Handle("/", http.FileServer(http.Dir("./static")))

	fmt.Println("🚀 Servidor online!")
	fmt.Println("👉 Acesse a interface visual em: http://localhost:8080")
	fmt.Println("👉 API de dados em: http://localhost:8080/produtos")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Servidor parou:", err)
	}
}
