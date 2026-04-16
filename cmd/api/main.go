package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gorilla/mux"

	"github.com/theo-guerra/simple-shop/internal/config"
	"github.com/theo-guerra/simple-shop/internal/database"
	"github.com/theo-guerra/simple-shop/internal/handlers"
)

func main() {
	cfg, err := config.Carregar()
	if err != nil {
		slog.Error("Erro ao carregar configuração", "erro", err)
		os.Exit(1)
	}

	handlers.SetJWTSecret(cfg.JWTSecret)
	handlers.SetProduction(cfg.IsProduction())

	db, err := database.Conectar(cfg.DSN())
	if err != nil {
		slog.Error("Erro ao conectar no banco", "erro", err)
		os.Exit(1)
	}
	defer db.Close()

	authHandler := &handlers.AuthHandler{DB: db}
	caixaHandler := &handlers.CaixaHandler{DB: db}
	produtosHandler := &handlers.ProdutoHandler{DB: db}
	lojaHandler := &handlers.LojaHandler{DB: db}
	publicoHandler := &handlers.PublicoHandler{DB: db}
	fiadoHandler := &handlers.FiadoHandler{DB: db}
	relatorioHandler := &handlers.RelatorioHandler{DB: db}

	r := mux.NewRouter()
	r.Use(handlers.SecurityHeaders)
	r.Use(handlers.CORS)
	r.Use(handlers.RequestLogger)

	// Rate limiter: 10 tentativas por minuto em endpoints de auth
	authLimiter := handlers.NewRateLimiter(10, time.Minute)

	// ==========================================
	// ROTAS PÚBLICAS (Acesso sem Senha)
	// ==========================================
	r.HandleFunc("/auth/login", authLimiter.LimitarPorIP(authHandler.Login)).Methods("POST", "OPTIONS")
	r.HandleFunc("/auth/cadastro", authLimiter.LimitarPorIP(authHandler.Cadastro)).Methods("POST", "OPTIONS")
	r.HandleFunc("/auth/recuperar/solicitar", authLimiter.LimitarPorIP(authHandler.SolicitarRecuperacao)).Methods("POST", "OPTIONS")
	r.HandleFunc("/auth/recuperar/validar", authLimiter.LimitarPorIP(authHandler.ValidarRecuperacao)).Methods("POST", "OPTIONS")
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

	api.HandleFunc("/clientes", fiadoHandler.ListarClientes).Methods("GET")
	api.HandleFunc("/clientes/novo", fiadoHandler.CadastrarCliente).Methods("POST")
	api.HandleFunc("/fiados", fiadoHandler.ListarFiados).Methods("GET")
	api.HandleFunc("/fiados/novo", fiadoHandler.NovoFiado).Methods("POST")
	api.HandleFunc("/fiados/pagar", fiadoHandler.DarBaixa).Methods("POST")

	api.HandleFunc("/relatorios/top", relatorioHandler.TopProdutos).Methods("GET")
	api.HandleFunc("/relatorios/extrato", relatorioHandler.ExtratoCompleto).Methods("GET")

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

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Inicia o servidor em goroutine separada
	go func() {
		slog.Info("ERP Operacional", "porta", cfg.Port, "ambiente", cfg.AppEnv)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Erro no servidor", "erro", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown: espera SIGINT/SIGTERM e drena requests em andamento
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Desligando servidor...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Erro ao desligar servidor", "erro", err)
		os.Exit(1)
	}
	slog.Info("Servidor encerrado com sucesso")
}