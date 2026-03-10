# 💼 Boutique ERP & Financial Dashboard

Um sistema completo de Ponto de Venda (PDV), Gestão de Estoque e Inteligência de Negócios (BI) construído com **Go** e **PostgreSQL**. Desenhado para unir a potência de um terminal financeiro (wall-street style) com a elegância visual de uma marca de grife.

## 🚀 Visão Geral

Este projeto foi desenvolvido com a mentalidade de um Diretor Financeiro (CFO). Ele não apenas registra vendas, mas exige disciplina operacional (Trava de Abertura/Fechamento de Caixa), calcula patrimônio imobilizado em tempo real e projeta o fluxo de caixa através de gráficos interativos.

Além do painel administrativo, o sistema gera automaticamente um **Catálogo Online** (Estilo E-commerce) personalizável para que o lojista receba pedidos direto no WhatsApp.

## ✨ Principais Funcionalidades

- **🔒 Trava Operacional de Caixa:** O sistema só opera após o registro do Fundo de Troco inicial, garantindo balanços perfeitos no fim do dia.
- **📊 Business Intelligence (BI):** Dashboards interativos com Chart.js (Gráficos de linha de fluxo e barras comparativas de Receita vs Despesa).
- **📦 Estoque Inteligente:** Filtro dinâmico por categorias, alertas automáticos de ruptura de estoque e cálculo do patrimônio imobilizado (Preço de Custo x Quantidade).
- **📋 Logística e Checklist:** Sistema integrado de To-Do list para operações diárias (compras, contatos com fornecedores) salvo no *Local Storage*.
- **📈 Exportação Contábil:** Geração de relatórios financeiros e extratos detalhados com exportação em um clique para planilhas `.csv`.
- **🛍️ Catálogo Online (White-label):** Vitrine digital para clientes, onde o lojista pode alterar pelo painel a cor principal do tema, o nome e as redes sociais.
- **📱 Integração WhatsApp:** Geração automática de recibos pós-venda enviados diretamente para o WhatsApp do cliente.

## 🛠️ Tecnologias Utilizadas

**Backend:**
- [Go (Golang)](https://golang.org/) - Alta performance, tipagem forte e concorrência.
- [PostgreSQL](https://www.postgresql.org/) - Banco de dados relacional robusto e seguro.
- Arquitetura RESTful com autenticação JWT.

**Frontend:**
- HTML5 / Vanilla JS - Zero dependências pesadas (sem frameworks JS), focando em velocidade bruta.
- [Tailwind CSS](https://tailwindcss.com/) - Estilização utility-first para um design elegante e responsivo (Desktop & Mobile).
- [Chart.js](https://www.chartjs.org/) - Renderização de gráficos financeiros em tempo real.
- [Phosphor Icons](https://phosphoricons.com/) - Iconografia limpa e moderna.

## ⚙️ Como Executar o Projeto

### Pré-requisitos
- Go 1.20+ instalado
- PostgreSQL rodando localmente ou em nuvem

### Passo a Passo

1. **Clone o repositório**
   ```bash
   git clone [https://github.com/theo-guerra/simple-shop.git](https://github.com/theo-guerra/simple-shop.git)
   cd simple-shop
Configure o Banco de Dados
Crie um banco de dados no PostgreSQL (ex: simple_shop) e execute os scripts de criação de tabelas (produtos, caixa_movimentos, sessoes_caixa, loja_config, etc).

Configure as Variáveis de Conexão
No arquivo cmd/api/main.go, atualize a string de conexão com suas credenciais do banco:

Go
db, _ := database.Conectar("user=postgres password=SUA_SENHA dbname=simple_shop host=localhost sslmode=disable")
Inicie o Servidor

Bash
go run cmd/api/main.go
Acesse a Aplicação
Abra o navegador em http://localhost:8080/login.html

👨‍💻 Autor
Theo Guerra

LinkedIn: TheoGuerra71

GitHub: @theo-guerra
