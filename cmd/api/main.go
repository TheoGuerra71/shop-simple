package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux" // Supondo que você usa o Gorilla Mux

	// Ajuste os imports abaixo para o caminho real do seu projeto
	"github.com/theo-guerra/simple-shop/internal/database"
	"github.com/theo-guerra/simple-shop/internal/handlers"
)

func main() {
	// 1. Conecta no Banco de Dados
	db, err := database.Conectar("user=postgres password=admin dbname=simple_shop host=localhost sslmode=disable")
	if err != nil {
		log.Fatal("Erro ao conectar no banco:", err)
	}
	defer db.Close()

	// 2. Inicializa os Handlers
	authHandler := &handlers.AuthHandler{DB: db}
	caixaHandler := &handlers.CaixaHandler{DB: db}
	// produtosHandler := &handlers.ProdutosHandler{DB: db} (Se você tiver um separado)

	// 3. Inicializa o Roteador
	r := mux.NewRouter()

	// ==========================================
	// 🔓 ROTAS PÚBLICAS (Não precisam de Cookie)
	// ==========================================
	r.HandleFunc("/auth/login", authHandler.Login).Methods("POST", "OPTIONS")
	r.HandleFunc("/auth/logout", authHandler.Logout).Methods("POST", "OPTIONS") // A NOVA ROTA AQUI!

	// Servir os arquivos estáticos (HTML, CSS, JS) - Login e Catálogo são públicos
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static")))

	// ==========================================
	// 🔒 ROTAS PROTEGIDAS (Existem o Cookie HttpOnly)
	// ==========================================
	api := r.PathPrefix("/api").Subrouter()
	api.Use(handlers.AuthMiddleware)

	// Rotas do Caixa/BI
	api.HandleFunc("/dashboard", caixaHandler.DashboardMobile).Methods("GET")
	api.HandleFunc("/movimentos", caixaHandler.ListarMovimentosHoje).Methods("GET")
	api.HandleFunc("/caixa/movimento", caixaHandler.RegistrarMovimento).Methods("POST")

	// SE ESSAS ESTIVEREM COMENTADAS, O PAINEL TRAVA:
	api.HandleFunc("/produtos", produtosHandler.ListarProdutos).Methods("GET")
	api.HandleFunc("/produtos/novo", produtosHandler.NovoProduto).Methods("POST")
	api.HandleFunc("/produtos/editar", produtosHandler.EditarProduto).Methods("POST")
	api.HandleFunc("/produtos/deletar", produtosHandler.DeletarProduto).Methods("POST")
	api.HandleFunc("/produtos/repor", produtosHandler.ReporEstoque).Methods("POST")
	api.HandleFunc("/vender", caixaHandler.Vender).Methods("POST") // (ou o handler correspondente)

	api.HandleFunc("/loja", lojaHandler.ObterConfig).Methods("GET")
	api.HandleFunc("/loja", lojaHandler.SalvarConfig).Methods("POST")
	// 4. Liga o Servidor
	log.Println("🚀 ERP com Motor CFO rodando na porta 8080...")
	log.Fatal(http.ListenAndServe(":8080", r))
}
