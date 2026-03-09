package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/theo-guerra/simple-shop/internal/database"
	"github.com/theo-guerra/simple-shop/internal/handlers"
)

func main() {
	// 1. Configuração do Banco
	connStr := "user=postgres password=admin dbname=simple_shop host=localhost sslmode=disable"
	db, err := database.Conectar(connStr)
	if err != nil {
		log.Fatal("Falha crítica no banco:", err)
	}
	defer db.Close()

	// 2. Inicialização dos Controladores (Handlers)
	hProdutos := &handlers.ProdutoHandler{DB: db}
	hLoja := &handlers.LojaHandler{DB: db}   // Requer o arquivo loja.go criado
	hFiado := &handlers.FiadoHandler{DB: db} // Requer o arquivo fiado.go criado

	// ==========================================
	// MAPEAMENTO DE ROTAS (O mapa do sistema)
	// ==========================================

	// Aba 1: Estoque e Vendas
	http.HandleFunc("/produtos", hProdutos.ListarProdutos)
	http.HandleFunc("/produtos/novo", hProdutos.Criar)
	http.HandleFunc("/produtos/vender", hProdutos.Vender)
	http.HandleFunc("/produtos/deletar", hProdutos.Deletar)

	// Aba 2: Minha Loja (Configurações e Tarefas)
	http.HandleFunc("/loja/config", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			hLoja.ObterConfig(w, r)
		} else {
			hLoja.AtualizarConfig(w, r)
		}
	})
	http.HandleFunc("/loja/tarefas", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			hLoja.ListarTarefas(w, r)
		} else {
			hLoja.CriarTarefa(w, r)
		}
	})

	// Aba 3: Caderno de Fiado
	http.HandleFunc("/clientes/novo", hFiado.CadastrarCliente)
	http.HandleFunc("/fiados", hFiado.ListarFiados)
	http.HandleFunc("/fiados/novo", hFiado.NovoFiado)
	http.HandleFunc("/fiados/pagar", hFiado.DarBaixa)

	// ROTA DO FRONT-END HTML
	http.Handle("/", http.FileServer(http.Dir("./static")))

	// ==========================================
	fmt.Println("🚀 Backend Completo Online em http://localhost:8080")
	fmt.Println("📦 Modos ativos: Estoque | Minha Loja | Caderno de Fiado")

	// Inicia o servidor e libera a porta caso falhe
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Servidor parou. Use 'fuser -k 8080/tcp' se a porta estiver travada. Erro:", err)
	}
}
