# Migração Multi-Tenant e Novas Funcionalidades

## Comandos ALTER TABLE (e CREATE) — copiar e colar no PostgreSQL

Execute no DBeaver (ou `psql`) na ordem abaixo. Se já tiver dados, ajuste o `DEFAULT 1` e faça um `UPDATE` para associar às contas corretas antes de remover o default.

```sql
-- 1) Tabela de usuários (feirantes)
CREATE TABLE IF NOT EXISTS usuarios (
  id    SERIAL PRIMARY KEY,
  email VARCHAR(255) NOT NULL UNIQUE,
  senha VARCHAR(255) NOT NULL DEFAULT ''
);

-- 2) Produtos: usuario_id, url_imagem como JSONB, visivel_catalogo
ALTER TABLE produtos ADD COLUMN IF NOT EXISTS usuario_id INTEGER NOT NULL DEFAULT 1 REFERENCES usuarios(id);
ALTER TABLE produtos ADD COLUMN IF NOT EXISTS visivel_catalogo BOOLEAN NOT NULL DEFAULT true;

-- url_imagem como JSONB:
-- Se você JÁ TEM a coluna url_imagem como TEXT: descomente e execute o bloco abaixo (e comente a linha seguinte).
-- ALTER TABLE produtos ADD COLUMN IF NOT EXISTS url_imagem_jsonb JSONB DEFAULT '[]'::jsonb;
-- UPDATE produtos SET url_imagem_jsonb = CASE WHEN url_imagem IS NULL OR url_imagem = '' THEN '[]'::jsonb ELSE to_jsonb(ARRAY[url_imagem]) END;
-- ALTER TABLE produtos DROP COLUMN IF EXISTS url_imagem;
-- ALTER TABLE produtos RENAME COLUMN url_imagem_jsonb TO url_imagem;
-- Se url_imagem ainda não existe ou você já renomeou:
ALTER TABLE produtos ADD COLUMN IF NOT EXISTS url_imagem JSONB DEFAULT '[]'::jsonb;

-- 3) Caixa: usuario_id
ALTER TABLE caixa_movimentos ADD COLUMN IF NOT EXISTS usuario_id INTEGER NOT NULL DEFAULT 1 REFERENCES usuarios(id);

-- 4) Loja config: usuario_id e índice único para UPSERT
ALTER TABLE loja_config ADD COLUMN IF NOT EXISTS usuario_id INTEGER NOT NULL DEFAULT 1 REFERENCES usuarios(id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_loja_config_usuario_id ON loja_config(usuario_id);
```

Antes de remover os `DEFAULT 1`, crie ao menos um usuário e use o `id` dele:

```sql
INSERT INTO usuarios (email, senha) VALUES ('admin@feira.com', 'sua_senha') ON CONFLICT (email) DO NOTHING;
-- Depois: UPDATE produtos SET usuario_id = (SELECT id FROM usuarios WHERE email = 'admin@feira.com') WHERE usuario_id = 1;
-- Idem para caixa_movimentos e loja_config.
```

---

## Detalhamento das alterações (referência)

Execute na ordem abaixo. Ajuste o nome do banco se necessário.

### 1. Tabela de usuários (feirantes)

```sql
-- Tabela de usuários para login e isolamento multi-tenant
CREATE TABLE IF NOT EXISTS usuarios (
  id    SERIAL PRIMARY KEY,
  email VARCHAR(255) NOT NULL UNIQUE,
  senha VARCHAR(255) NOT NULL DEFAULT ''
);
```

### 2. Coluna `usuario_id` e isolamento

```sql
-- Produtos: um feirante por produto
ALTER TABLE produtos
  ADD COLUMN IF NOT EXISTS usuario_id INTEGER NOT NULL DEFAULT 1 REFERENCES usuarios(id);
-- Remova o DEFAULT após popular os dados existentes com o usuario_id correto:
-- UPDATE produtos SET usuario_id = 1 WHERE usuario_id IS NULL;
-- ALTER TABLE produtos ALTER COLUMN usuario_id DROP DEFAULT;

-- Caixa/movimentos: um feirante por movimento
ALTER TABLE caixa_movimentos
  ADD COLUMN IF NOT EXISTS usuario_id INTEGER NOT NULL DEFAULT 1 REFERENCES usuarios(id);

-- Loja config: uma config por feirante (usuario_id como chave)
ALTER TABLE loja_config
  ADD COLUMN IF NOT EXISTS usuario_id INTEGER NOT NULL DEFAULT 1 REFERENCES usuarios(id);
-- Garantir uma linha por feirante (para ON CONFLICT no handler)
CREATE UNIQUE INDEX IF NOT EXISTS idx_loja_config_usuario_id ON loja_config(usuario_id);
```

### 3. Múltiplas fotos (JSONB) e vitrine

```sql
-- Trocar url_imagem de TEXT/VARCHAR para JSONB (array de strings)
-- Se a coluna já existir como texto, migrar para JSONB:
ALTER TABLE produtos
  ADD COLUMN IF NOT EXISTS url_imagem_new JSONB DEFAULT '[]'::jsonb;

-- Migrar dados antigos: uma única URL vira um array de um elemento
UPDATE produtos
SET url_imagem_new = CASE
  WHEN url_imagem IS NULL OR url_imagem = '' THEN '[]'::jsonb
  ELSE jsonb_build_array(url_imagem)
END
WHERE url_imagem_new IS NULL OR url_imagem_new = '[]'::jsonb;

-- Remover coluna antiga e renomear (se sua coluna atual se chama url_imagem)
ALTER TABLE produtos DROP COLUMN IF EXISTS url_imagem;
ALTER TABLE produtos RENAME COLUMN url_imagem_new TO url_imagem;

-- Se você ainda não tinha a coluna url_imagem:
-- ALTER TABLE produtos ADD COLUMN IF NOT EXISTS url_imagem JSONB DEFAULT '[]'::jsonb;

-- Controle de vitrine: só produtos com visivel_catalogo = true e quantidade > 0 aparecem no catálogo público
ALTER TABLE produtos
  ADD COLUMN IF NOT EXISTS visivel_catalogo BOOLEAN NOT NULL DEFAULT true;
```

### 4. Resumo em um único bloco (para banco novo ou já com usuario_id)

Se você **já tem** a coluna `usuario_id` em todas as tabelas e só precisa de JSONB + vitrine:

```sql
-- 1) Usuários (se ainda não existir)
CREATE TABLE IF NOT EXISTS usuarios (
  id    SERIAL PRIMARY KEY,
  email VARCHAR(255) NOT NULL UNIQUE,
  senha VARCHAR(255) NOT NULL DEFAULT ''
);

-- 2) Produtos: usuario_id (se não existir)
ALTER TABLE produtos ADD COLUMN IF NOT EXISTS usuario_id INTEGER REFERENCES usuarios(id);
ALTER TABLE produtos ADD COLUMN IF NOT EXISTS visivel_catalogo BOOLEAN NOT NULL DEFAULT true;

-- 3) Converter url_imagem para JSONB (se ainda for texto)
-- Opção A: coluna nova e migração
ALTER TABLE produtos ADD COLUMN IF NOT EXISTS url_imagem_jsonb JSONB DEFAULT '[]'::jsonb;
UPDATE produtos SET url_imagem_jsonb = to_jsonb(ARRAY[url_imagem]) WHERE url_imagem IS NOT NULL AND url_imagem != '' AND (url_imagem_jsonb IS NULL OR url_imagem_jsonb = '[]');
ALTER TABLE produtos DROP COLUMN IF EXISTS url_imagem;
ALTER TABLE produtos RENAME COLUMN url_imagem_jsonb TO url_imagem;

-- Opção B: se não existir coluna url_imagem
-- ALTER TABLE produtos ADD COLUMN IF NOT EXISTS url_imagem JSONB DEFAULT '[]'::jsonb;

-- 4) Caixa
ALTER TABLE caixa_movimentos ADD COLUMN IF NOT EXISTS usuario_id INTEGER REFERENCES usuarios(id);

-- 5) Loja config
ALTER TABLE loja_config ADD COLUMN IF NOT EXISTS usuario_id INTEGER REFERENCES usuarios(id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_loja_config_usuario_id ON loja_config(usuario_id);
```

---

## Frontend (index.html) — mudanças para múltiplas fotos e vitrine

### 1. Formato de `url_imagem`: de string para array

- **Antes:** o front enviava e recebia `url_imagem` como uma única string (ex.: base64).
- **Agora:** a API espera e devolve `url_imagem` como **array de strings** `string[]`.

Exemplos de payload:

- Criar produto:  
  `{ "nome": "...", "preco": 10, "url_imagem": ["data:image/jpeg;base64,...", "data:image/png;base64,..."], "visivel_catalogo": true }`
- Editar produto:  
  idem; pode enviar `"url_imagem": []` para remover fotos.

No JavaScript, ao montar o objeto para enviar:

- Use um **array** para as fotos, por exemplo `url_imagem: arrayDeBase64`.
- Se hoje você tem um único `<input type="file">`, pode continuar enviando um array com um elemento:  
  `url_imagem: base64 ? [base64] : []`.
- Para múltiplas fotos: vários `<input type="file">` ou um único com `multiple`; leia cada arquivo com `FileReader.readAsDataURL`, acumule em um array e envie em `url_imagem`.

### 2. Exibir múltiplas fotos no painel

- Ao receber o produto da API, `p.url_imagem` é um **array**.
- Exemplo: primeira foto como miniatura  
  `const src = Array.isArray(p.url_imagem) && p.url_imagem.length > 0 ? p.url_imagem[0] : null;`
- Para galeria: fazer um `p.url_imagem.forEach(...)` e renderizar um `<img>` por item.

### 3. Campo `visivel_catalogo`

- Inclua no formulário de criar/editar produto um checkbox (ou toggle) ligado a `visivel_catalogo`.
- Ao **criar**: envie `visivel_catalogo: true` ou `false` (default no backend é `true` se omitir).
- Ao **editar**: envie o valor atual do checkbox para atualizar no backend.
- O catálogo público (**GET /api/produtos?usuario_id=X**) só mostra itens com `visivel_catalogo === true` e `quantidade > 0`.

### 4. Catálogo público e `usuario_id`

- A rota pública de produtos exige o parâmetro na query:  
  `GET /api/produtos?usuario_id=1`
- A rota pública da loja também:  
  `GET /api/loja?usuario_id=1`
- No front do **catálogo** (ex.: `catalogo.html`), você precisa saber o `usuario_id` do feirante (por URL, config ou login). Exemplo:  
  `fetch('/api/produtos?usuario_id=' + feiranteId)` e `fetch('/api/loja?usuario_id=' + feiranteId)`.

### 5. Resumo das mudanças no JS

| Onde | Antes | Depois |
|------|--------|--------|
| Enviar produto (criar/editar) | `url_imagem: string` (uma base64) | `url_imagem: string[]` (array de base64 ou URLs) |
| Receber produto (listar/editar) | `p.url_imagem` string | `p.url_imagem` array; usar `p.url_imagem[0]` ou iterar |
| Novo campo | — | Enviar e exibir `visivel_catalogo` (boolean) |
| Chamada catálogo público | `/api/produtos`, `/api/loja` | `/api/produtos?usuario_id=X`, `/api/loja?usuario_id=X` |

Com isso, o front fica alinhado ao multi-tenant, múltiplas fotos (JSONB) e controle de vitrine (`visivel_catalogo`).
