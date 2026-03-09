# 🛒 Shop-Simple (Gestão de Estoque Full-Stack)

Um sistema completo de gerenciamento de estoque e PDV (Ponto de Venda) desenvolvido em **Go (Golang)** e **PostgreSQL**, com uma interface web interativa. 

Este projeto foi construído para demonstrar conceitos avançados de backend, incluindo **Transações ACID** no banco de dados, criação de uma API RESTful do zero (sem frameworks) e integração com um front-end em Vanilla JavaScript.

## 🚀 Tecnologias Utilizadas

* **Backend:** Go (Golang) - Roteamento nativo (`net/http`)
* **Banco de Dados:** PostgreSQL (Driver: `lib/pq`)
* **Frontend:** HTML5, CSS3, Vanilla JavaScript (Fetch API)
* **Infraestrutura:** Linux (Ubuntu) / Git / GitHub

## ✨ Funcionalidades (CRUD Completo)

- **Listagem (GET):** Visualização em tempo real de todos os produtos cadastrados e seus respectivos estoques.
- **Cadastro (POST):** Inserção de novos itens na base de dados.
- **Venda com Proteção de Estoque (POST):** Utilização de **Transações SQL** para garantir que o estoque seja reduzido com segurança. O sistema bloqueia vendas se o estoque for insuficiente.
- **Remoção (DELETE):** Exclusão permanente de itens fora de linha.

## ⚙️ Arquitetura do Projeto

O código foi estruturado visando manutenibilidade e escalabilidade, separando as responsabilidades de forma clara:

    /shop-simple
    ├── cmd/
    │   └── api/
    │       └── main.go         # Ponto de entrada, configuração do servidor e rotas
    ├── internal/
    │   ├── database/
    │   │   └── db.go           # Gerenciamento da conexão com o PostgreSQL
    │   ├── handlers/
    │   │   └── produtos.go     # Regras de negócio, manipulação de JSON e Transações
    │   └── models/
    │       └── produto.go      # Estruturas de dados (Structs) e tags JSON
    ├── static/
    │   └── index.html          # Interface gráfica do usuário servida pelo Go
    ├── go.mod                  # Gerenciador de dependências do Go
    └── README.md               # Documentação do projeto

## 🛠️ Como rodar o projeto localmente

### Pré-requisitos
* Go 1.20+ instalado
* PostgreSQL rodando localmente

### 1. Configuração do Banco de Dados
No seu terminal PostgreSQL (`psql`), crie o banco e a tabela:
```sql
CREATE DATABASE simple_shop;

\c simple_shop;

CREATE TABLE produtos (
    id SERIAL PRIMARY KEY,
    nome VARCHAR(100) NOT NULL,
    preco NUMERIC(10, 2) NOT NULL,
    quantidade INT NOT NULL
);
```

### 2. Configurando as Credenciais
No arquivo `cmd/api/main.go`, altere a variável `connStr` para incluir a sua senha real do PostgreSQL:
```go
connStr := "user=postgres password=SUA_SENHA dbname=simple_shop host=localhost sslmode=disable"
```

### 3. Executando o Servidor
No terminal, na raiz do projeto, instale as dependências e inicie o motor:
```bash
go mod tidy
go run cmd/api/main.go
```

Acesse o painel visual no navegador: **`http://localhost:8080`**

---
Desenvolvido por [Theo Guerra](https://github.com/TheoGuerra71).
