# 💼 Boutique ERP SaaS & Financial Dashboard

Um sistema completo de Ponto de Venda (PDV), Gestão de Estoque, Inteligência de Negócios (BI) e **Catálogo Online** construído com **Go**, **PostgreSQL** e **Vanilla JS**.  
Agora preparado como um **SaaS Multi-Tenant**, permitindo que vários feirantes/lojas rodem de forma isolada na mesma infraestrutura.

## 🚀 Visão Geral

- Painel administrativo blindado (login + cookies HttpOnly + JWT) para operar caixa, estoque e relatórios.
- Cada lojista (tabela `usuarios`) possui seus próprios **produtos**, **movimentações de caixa** e **configuração de catálogo**, isolados por `usuario_id`.
- Geração automática de um **Catálogo Público** por lojista, consumido via `catalogo.html?usuario_id=123`, com pedidos chegando direto no WhatsApp.

## ✨ Principais Funcionalidades

- **🔒 Trava Operacional de Caixa:** exige registro de fundo de troco inicial antes de liberar o painel financeiro.
- **📊 BI em tempo real:** dashboards com receita, despesas, ticket médio e gráficos em 7 dias.
- **📦 Estoque Inteligente:** categorias, alertas de ruptura, patrimônio imobilizado e tela de reposição com custo médio automático.
- **🧾 Livro Caixa & Relatórios:** extrato dos últimos 30 dias, exportação `.csv` e PDF executivo para diretoria/contador.
- **🛍️ Catálogo Online (White-label):** página pública por lojista com tema, nome, Instagram e mensagem personalizados.
- **📱 Integração WhatsApp:** recibos pós-venda e pedidos do catálogo enviados direto para o WhatsApp.
- **🧑‍💼 Multi-Tenant Real:** todas as queries protegidas filtram por `usuario_id`; um lojista nunca acessa dados de outro.
- **🖼️ Múltiplas Fotos por Produto:** campo `url_imagem` como `JSONB` (array de base64/URLs) com controle `visivel_catalogo`.

## 🛠️ Tecnologias Utilizadas

**Backend**
- Go (Golang)
- PostgreSQL
- Gorilla Mux
- JWT (cookies HttpOnly)

**Frontend**
- HTML5 + Vanilla JS
- Tailwind CSS
- Chart.js
- Phosphor Icons

## ⚙️ Como Executar o Projeto

### 1. Pré-requisitos
- Go 1.20+
- PostgreSQL rodando local (ex.: porta 5432)

### 2. Clonar o repositório

```bash
git clone https://github.com/theo-guerra/simple-shop.git
cd simple-shop
```

### 3. Configurar o banco de dados

Crie o banco:

```bash
createdb simple_shop
```

Depois, rode a migração mínima (ajuste usuário/senha conforme seu ambiente):

```bash
sudo -u postgres psql -d simple_shop -c "
CREATE TABLE IF NOT EXISTS usuarios (
  id    SERIAL PRIMARY KEY,
  email VARCHAR(255) NOT NULL UNIQUE,
  senha VARCHAR(255) NOT NULL DEFAULT ''
);

INSERT INTO usuarios (id, email, senha)
VALUES (1, 'mestre@loja.com', '')
ON CONFLICT (id) DO NOTHING;

ALTER TABLE produtos ADD COLUMN IF NOT EXISTS usuario_id INTEGER NOT NULL DEFAULT 1 REFERENCES usuarios(id);
ALTER TABLE produtos ADD COLUMN IF NOT EXISTS visivel_catalogo BOOLEAN NOT NULL DEFAULT true;
ALTER TABLE produtos ADD COLUMN IF NOT EXISTS url_imagem JSONB DEFAULT '[]'::jsonb;

ALTER TABLE caixa_movimentos ADD COLUMN IF NOT EXISTS usuario_id INTEGER NOT NULL DEFAULT 1 REFERENCES usuarios(id);

ALTER TABLE loja_config ADD COLUMN IF NOT EXISTS usuario_id INTEGER NOT NULL DEFAULT 1 REFERENCES usuarios(id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_loja_config_usuario_id ON loja_config(usuario_id);
"
```

> Para uma migração mais completa (incluindo conversão de colunas antigas), veja `docs/MULTI-TENANT-MIGRATION.md`.

### 4. Configurar a conexão com o banco

No arquivo `cmd/api/main.go`, ajuste a string de conexão:

```go
db, err := database.Conectar("user=postgres password=SUA_SENHA dbname=simple_shop host=localhost sslmode=disable")
```

### 5. Subir o servidor

```bash
go run cmd/api/main.go
```

A API sobe em `http://localhost:8080`.

### 6. Acessar o painel e o catálogo

- Painel (login):  
  `http://localhost:7000/login.html`

- Após logar, você é redirecionado para o painel (`/`) e todas as operações passam a usar o `usuario_id` do token.

- Catálogo público para o lojista `usuario_id = 1`:  
  `http://localhost:7000/catalogo.html?usuario_id=1`

## 👨‍💻 Autor

Theo Guerra  
LinkedIn: TheoGuerra71  
GitHub: @theo-guerra
