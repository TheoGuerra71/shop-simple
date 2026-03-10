package main

import (
	"fmt"
	"net/http"

	"github.com/theo-guerra/simple-shop/internal/database"
	"github.com/theo-guerra/simple-shop/internal/handlers"
)

func main() {
	// 1. Conexão (Ajuste sua senha aqui)
	db, _ := database.Conectar("user=postgres password=admin dbname=simple_shop host=localhost sslmode=disable")

	hAuth := &handlers.AuthHandler{DB: db}
	hProd := &handlers.ProdutoHandler{DB: db}
	hCaixa := &handlers.CaixaHandler{DB: db}
	hLoja := &handlers.LojaHandler{DB: db}

	// 2. Rotas Públicas
	http.HandleFunc("/auth/login", hAuth.Login)
	http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("./static/public"))))

	// 3. Rotas Protegidas (JWT)
	http.HandleFunc("/api/produtos", handlers.Autenticador(hProd.ListarProdutos))
	http.HandleFunc("/api/vender", handlers.Autenticador(hProd.Vender))
	http.HandleFunc("/api/dashboard", handlers.Autenticador(hCaixa.DashboardMobile))
	http.HandleFunc("/api/movimentos", handlers.Autenticador(hCaixa.ListarMovimentosHoje))
	http.HandleFunc("/api/loja", handlers.Autenticador(hLoja.ObterConfig))

	// 4. Servir Frontend Admin
	http.Handle("/", http.FileServer(http.Dir("./static")))

	fmt.Println("🚀 ERP BLINDADO NIVEL 1000: http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
