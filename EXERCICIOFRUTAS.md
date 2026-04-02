# 🏋️ Exercício — POST em Go com Gin

> Exercício proposto por Cláudio — baseado no projeto de registro de usuário que estudamos juntos.

---

## 🎯 Objetivo

Criar um endpoint `POST /frutas/` que cadastra uma fruta no banco de dados.  
Você vai replicar a mesma estrutura de camadas que vimos: **domain → repository → service → handler**.

---

## 📦 O que cadastrar

Uma **fruta** com os seguintes campos:

| Campo | Tipo | Regra |
|---|---|---|
| `id` | string (UUID) | gerado pelo servidor |
| `name` | string | obrigatório |
| `color` | string | obrigatório |
| `price` | float64 | obrigatório, maior que zero |
| `weight_grams` | int | obrigatório, maior que zero |

---

## 🗄️ Tabela no banco

```sql
CREATE TABLE IF NOT EXISTS fruits (
    id           TEXT    PRIMARY KEY,
    name         TEXT    NOT NULL,
    color        TEXT    NOT NULL,
    price        NUMERIC NOT NULL,
    weight_grams INT     NOT NULL
);
```

---

## 📡 Comportamento esperado

**Requisição:**
```bash
curl -X POST http://localhost:8080/frutas/ \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Manga",
    "color": "Amarela",
    "price": 4.99,
    "weight_grams": 400
  }'
```

**Resposta de sucesso — 201 Created:**
```json
{
  "id": "uuid-gerado-aqui",
  "name": "Manga",
  "color": "Amarela",
  "price": 4.99,
  "weight_grams": 400
}
```

**Se o nome já existir — 409 Conflict:**
```json
{
  "error": "fruta já cadastrada"
}
```

**Se algum campo obrigatório faltar — 400 Bad Request:**
```json
{
  "error": "..."
}
```

---

## 📁 Estrutura de arquivos que você vai criar

```
fruit-api/
├── cmd/api/
│   └── main.go
├── config/
│   └── config.go
├── internal/
│   ├── domain/
│   │   └── fruit.go        ← você cria
│   ├── handler/
│   │   └── fruit_handler.go ← você cria
│   ├── service/
│   │   └── fruit_service.go ← você cria
│   └── repository/
│       └── fruit_repository.go ← você cria
├── .env
└── go.mod
```

---

## 💡 Dicas

**1. Comece pelo 1. go.mod ***
   └── sem ele nenhum import funciona — é o primeiro passo sempre

2. .env 
   └── sem as credenciais o config.go não tem o que ler

3. migration.sql → rodar no banco
   └── sem a tabela o INSERT vai falhar

4. internal/domain/fruit.go
   └── define a struct Fruit, CreateFruitInput e a interface FruitRepository
   └── todos os outros arquivos importam esse — tem que existir primeiro

5. internal/repository/fruit_repository.go
   └── implementa a interface definida no domain
   └── depende só do domain e do pgx

6. internal/service/fruit_service.go
   └── depende do domain (interface) e do uuid
   └── não sabe que o repository existe — só conhece a interface

7. config/config.go
   └── lê o .env e cria o pool de conexões
   └── depende só de libs externas (pgx, godotenv)

8. internal/handler/fruit_handler.go
   └── depende do domain e do service
   └── última camada antes do main

9. cmd/api/main.go
   └── depende de tudo — só escreva ele quando tudo acima estiver pronto
   └── é o único que conhece todas as camadas e monta a cadeia

**2. O `price` é `float64` — no JSON e no SQL:**
```go
// na struct:
Price float64 `json:"price" binding:"required,gt=0"`
//                                            ↑ gt=0 = greater than zero
```

**3. O `weight_grams` é `int`:**
```go
WeightGrams int `json:"weight_grams" binding:"required,gt=0"`
```

**4. No SQL, use `NUMERIC` para preço e `INT` para peso:**
```sql
INSERT INTO fruits (id, name, color, price, weight_grams)
VALUES ($1, $2, $3, $4, $5)
```

**5. A regra de negócio do service é simples:**
- Verificar se já existe uma fruta com esse nome
- Gerar o UUID
- Chamar o repository para salvar

**6. Não esqueça do `go.mod`:**
```bash
go mod init fruit-api
go get github.com/gin-gonic/gin
go get github.com/jackc/pgx/v5
go get github.com/joho/godotenv
go get github.com/google/uuid
```

---

## ✅ Checklist — seu exercício está completo quando:

- [ ] `POST /frutas/` retorna **201** com o JSON da fruta criada
- [ ] Campos vazios retornam **400**
- [ ] `price` ou `weight_grams` negativos/zero retornam **400**
- [ ] Nome duplicado retorna **409**
- [ ] `id` é gerado pelo servidor (não vem do cliente)
- [ ] Os dados aparecem na tabela do PostgreSQL após a requisição

---

## 🚀 Bônus (opcional)

Se terminar antes do grupo:

- Adicione um endpoint `GET /frutas/` que lista todas as frutas cadastradas
- No repository, implemente um método `FindAll() ([]Fruit, error)`
- No service, crie `GetAllFruits() ([]Fruit, error)`
- No handler, crie `ListFruits(c *gin.Context)` que responde com array JSON