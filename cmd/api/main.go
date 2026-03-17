package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"

	"github.com/theo-guerra/simple-shop/internal/database"
	"github.com/theo-guerra/simple-shop/internal/handlers"
)

func main() {
	db, err := database.Conectar("user=postgres password=admin dbname=simple_shop host=localhost sslmode=disable")
	if err != nil {
		log.Fatal("Erro ao conectar no banco:", err)
	}
	defer db.Close()

	authHandler := &handlers.AuthHandler{DB: db}
	caixaHandler := &handlers.CaixaHandler{DB: db}
	produtosHandler := &handlers.ProdutoHandler{DB: db}
	lojaHandler := &handlers.LojaHandler{DB: db}
	publicoHandler := &handlers.PublicoHandler{DB: db}

	r := mux.NewRouter()

	// ==========================================
	// 🔓 ROTAS PÚBLICAS (Acesso sem Senha)
	// ==========================================
	r.HandleFunc("/auth/login", authHandler.Login).Methods("POST", "OPTIONS")
	r.HandleFunc("/auth/cadastro", authHandler.Cadastro).Methods("POST", "OPTIONS")
	r.HandleFunc("/auth/recuperar/solicitar", authHandler.SolicitarRecuperacao).Methods("POST", "OPTIONS")
	r.HandleFunc("/auth/recuperar/validar", authHandler.ValidarRecuperacao).Methods("POST", "OPTIONS")
	r.HandleFunc("/auth/logout", authHandler.Logout).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/loja", publicoHandler.GetLojaByUrl).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/produtos/vitrine", publicoHandler.GetProdutosVitrine).Methods("GET", "OPTIONS")

	// ==========================================
	// 🔒 ROTAS PROTEGIDAS (Apenas com Login)
	// ==========================================
	api := r.PathPrefix("/api").Subrouter()
	api.Use(handlers.AuthMiddleware)

	api.HandleFunc("/dashboard", caixaHandler.DashboardMobile).Methods("GET")
	api.HandleFunc("/movimentos", caixaHandler.ListarMovimentosHoje).Methods("GET")
	api.HandleFunc("/caixa/movimento", caixaHandler.RegistrarMovimento).Methods("POST")

	api.HandleFunc("/produtos/painel", produtosHandler.ListarProdutos).Methods("GET")
	api.HandleFunc("/produtos/novo", produtosHandler.Criar).Methods("POST")
	api.HandleFunc("/produtos/editar", produtosHandler.Editar).Methods("POST")
	api.HandleFunc("/produtos/deletar", produtosHandler.Deletar).Methods("POST")
	api.HandleFunc("/produtos/repor", produtosHandler.Repor).Methods("POST")
	api.HandleFunc("/vender", produtosHandler.Vender).Methods("POST")

	api.HandleFunc("/loja", lojaHandler.Config).Methods("GET", "POST")

	// ==========================================
	// 💎 ROTEADOR DE VANITY URL
	// ==========================================
	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/" {
			http.ServeFile(w, r, "./static/index.html")
			return
		}
		fullPath := filepath.Join("./static", path)
		if _, err := os.Stat(fullPath); err == nil {
			http.ServeFile(w, r, fullPath)
			return
		}
		http.ServeFile(w, r, "./static/catalogo.html")
	})

	log.Println("🚀 ERP Operacional na porta 7000...")
	log.Fatal(http.ListenAndServe(":7000", r))
}