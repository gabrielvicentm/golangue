# Registro de Usuário em Go + Gin — Resumo do Fluxo Completo

## Estrutura de arquivos

```
exercicio/
├── cmd/api/main.go                   → ponto de entrada, monta tudo
├── config/config.go                  → .env e conexão com banco
├── internal/
│   ├── domain/user.go                → structs + interface (o "contrato")
│   ├── handler/user_handler.go       → camada HTTP (recebe req, responde)
│   ├── service/user_service.go       → regras de negócio + hash da senha
│   └── repository/user_repository.go → SQL com o PostgreSQL
├── .env                              → credenciais (nunca no git)
└── migration.sql                     → cria a tabela no banco
```

---

## O fluxo de uma requisição POST /users/

```
                    CLIENTE (curl, Postman, frontend...)
                              │
                              │  POST /users/
                              │  Body: { "name": "Cláudio",
                              │          "email": "claudio@email.com",
                              │          "password": "minhasenha123" }
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│  GIN (servidor HTTP)                                        │
│  r.Run(":8080") — fica escutando requisições               │
│  Bate na rota: POST /users/ → chama CreateUser()           │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│  HANDLER  (user_handler.go)                                 │
│                                                             │
│  1. c.ShouldBindJSON(&input)                                │
│     → deserializa o JSON para a struct CreateUserInput      │
│     → valida: name obrigatório? ✅                          │
│     → valida: email é email válido? ✅                      │
│     → valida: password tem 8+ caracteres? ✅                │
│     → se falhar: responde 400 e para aqui                   │
│                                                             │
│  2. h.service.CreateUser(input) → chama o service          │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│  SERVICE  (user_service.go)                                 │
│                                                             │
│  1. repo.FindByEmail(input.Email)                           │
│     → pergunta ao repository se o email já existe           │
│     → se existir: retorna erro "email já cadastrado"        │
│                                                             │
│  2. hashPassword(input.Password)                            │
│     → bcrypt.GenerateFromPassword([]byte(senha), custo 10)  │
│     → "minhasenha123" → "$2a$10$N9qo8uLO..."               │
│     → leva ~100ms de propósito (dificulta força bruta)      │
│                                                             │
│  3. Monta a struct User completa:                           │
│     ID:           uuid.NewString()  → gera UUID             │
│     Name:         input.Name                                │
│     Email:        input.Email                               │
│     PasswordHash: hash gerado acima                         │
│                                                             │
│  4. repo.Save(user) → manda pro repository salvar           │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│  REPOSITORY  (user_repository.go)                           │
│                                                             │
│  FindByEmail:                                               │
│    SELECT id, name, email, password_hash                    │
│    FROM users WHERE email = $1                              │
│    → retorna *User ou nil                                   │
│                                                             │
│  Save:                                                      │
│    INSERT INTO users (id, name, email, password_hash)       │
│    VALUES ($1, $2, $3, $4)                                  │
│    → retorna error (nil = sucesso)                          │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
                    POSTGRESQL
                    (salva a linha na tabela users)
                              │
                              │  (sobe de volta por todas as camadas)
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│  HANDLER  (responde ao cliente)                             │
│                                                             │
│  c.JSON(201, user)                                          │
│  → serializa a struct User para JSON                        │
│  → json:"-" no PasswordHash garante que ele NÃO aparece    │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
                    CLIENTE RECEBE:
                    HTTP 201 Created
                    {
                      "id": "a1b2c3d4-e5f6-...",
                      "name": "Cláudio",
                      "email": "claudio@email.com"
                      // password_hash: nunca aparece — json:"-"
                    }
```

---

## Go vs Node.js — comparação direta

| Conceito | Node.js | Go |
|---|---|---|
| Ponto de entrada | `index.js` / `server.js` | `cmd/api/main.go` |
| Framework HTTP | Express, Fastify | Gin |
| Variáveis de ambiente | `dotenv` + `process.env.X` | `godotenv` + `os.Getenv("X")` |
| Validação do body | Zod, Joi, class-validator | struct tags `binding:"required,email"` |
| Hash de senha | `bcryptjs` | `golang.org/x/crypto/bcrypt` |
| UUID | `crypto.randomUUID()` | `uuid.NewString()` |
| Retorno de erro | `throw new Error(...)` | `return nil, errors.New(...)` |
| Tratar erro | `try/catch` | `if err != nil { }` |
| "Não encontrado" | `return null` | `return nil, nil` |
| Esconder campo no JSON | `delete user.passwordHash` | `json:"-"` na struct tag |
| Injeção de dependência | manual ou NestJS DI | manual via construtores |
| Rodar servidor | `app.listen(8080)` | `r.Run(":8080")` |

---

## Por que bcrypt e não SHA256?

```
SHA256 (errado para senhas):
  → projetado para ser RÁPIDO
  → GPU moderna: 10.000.000.000 hashes/segundo
  → atacante testa 10 bilhões de senhas por segundo

bcrypt (correto para senhas):
  → projetado para ser LENTO de propósito
  → custo 10: ~100ms por hash
  → atacante testa ~10 senhas por segundo
  → diferença: 1.000.000.000x mais seguro
```

---

## Como testar

```bash
# Subir a API
go run ./cmd/api/

# Criar usuário (sucesso → 201)
curl -X POST http://localhost:8080/users/ \
  -H "Content-Type: application/json" \
  -d '{"name":"Cláudio","email":"claudio@email.com","password":"minhasenha123"}'

# Resposta:
# {"id":"uuid-aqui","name":"Cláudio","email":"claudio@email.com"}

# Mesmo email de novo (conflito → 409)
# {"error":"email já cadastrado"}

# Email inválido (validação → 400)
curl -d '{"name":"Cláudio","email":"nao-e-email","password":"12345678"}'
# {"error":"Key: 'Email' Error:Field validation for 'Email' failed on the 'email' tag"}

# Senha curta (validação → 400)
curl -d '{"name":"Cláudio","email":"a@b.com","password":"123"}'
# {"error":"Key: 'Password' Error:Field validation for 'Password' failed on the 'min' tag"}
```
