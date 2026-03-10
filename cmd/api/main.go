package main

import (
	"fmt"
	"net/http"

	"github.com/theo-guerra/simple-shop/internal/database"
	"github.com/theo-guerra/simple-shop/internal/handlers"
)

func main() {
	db, _ := database.Conectar("user=postgres password=admin dbname=simple_shop host=localhost sslmode=disable")

	hAuth := &handlers.AuthHandler{DB: db}
	hProd := &handlers.ProdutoHandler{DB: db}
	hCaixa := &handlers.CaixaHandler{DB: db}
	hLoja := &handlers.LojaHandler{DB: db}
	hRelatorios := &handlers.RelatorioHandler{DB: db} // O NOVO MOTOR

	// Rotas Abertas
	http.HandleFunc("/auth/login", hAuth.Login)

	// Rotas Protegidas (JWT)
	http.HandleFunc("/api/produtos", handlers.Autenticador(hProd.ListarProdutos))
	http.HandleFunc("/api/produtos/novo", handlers.Autenticador(hProd.Criar))
	http.HandleFunc("/api/produtos/editar", handlers.Autenticador(hProd.Editar))
	http.HandleFunc("/api/produtos/deletar", handlers.Autenticador(hProd.Deletar))
	http.HandleFunc("/api/vender", handlers.Autenticador(hProd.Vender))

	http.HandleFunc("/api/dashboard", handlers.Autenticador(hCaixa.DashboardMobile))
	http.HandleFunc("/api/movimentos", handlers.Autenticador(hCaixa.ListarMovimentosHoje))
	http.HandleFunc("/api/caixa/movimento", handlers.Autenticador(hCaixa.RegistrarMovimento))

	http.HandleFunc("/api/loja", handlers.Autenticador(hLoja.Config))

	// NOVAS ROTAS DE INTELIGÊNCIA DE NEGÓCIOS
	http.HandleFunc("/api/relatorios/top", handlers.Autenticador(hRelatorios.TopProdutos))
	http.HandleFunc("/api/relatorios/extrato", handlers.Autenticador(hRelatorios.ExtratoCompleto))

	http.Handle("/", http.FileServer(http.Dir("./static")))

	fmt.Println("🚀 ERP com Módulo de Relatórios Online!")
	http.ListenAndServe(":8080", nil)
}
