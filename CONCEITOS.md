# Go para quem vem do Node.js — Conceitos Fundamentais

> Guia preparado para o grupo antes do exercício prático.
> Todos os exemplos usam comparações com Node.js/TypeScript.

---

## 1. go.mod — o package.json do Go

Antes de escrever qualquer código, você inicializa o módulo. É o equivalente ao `npm init`.

```bash
# Node.js
npm init

# Go
go mod init nomedoprojeto
```

Isso cria o arquivo `go.mod`:

```
module fruit-api      ← nome do módulo (usado nos imports internos)

go 1.22               ← versão do Go

require (             ← dependências (preenchido automaticamente)
  github.com/gin-gonic/gin v1.9.1
  github.com/jackc/pgx/v5  v5.5.0
)
```

| Node.js | Go |
|---|---|
| `package.json` | `go.mod` |
| `package-lock.json` | `go.sum` |
| `node_modules/` | cache global em `~/go/pkg/mod/` |
| `npm install axios` | `go get github.com/gin-gonic/gin` |
| `npm install` | `go mod tidy` |

> **Diferença importante:** em Go não existe pasta `node_modules`. As dependências ficam em um cache global na sua máquina e são compiladas junto com seu código em um **único binário**.

---

## 2. Tipos de import — stdlib vs libs externas

Go tem dois tipos de import: bibliotecas que **já vêm com o Go** (stdlib) e bibliotecas **externas** que você instala.

```go
import (
    // ── STDLIB ─────────────────────────────────────────────
    // Vêm instaladas com o Go — não precisa de go get
    "context"       // controle de timeout e cancelamento
    "fmt"           // formatação de strings (Sprintf, Println)
    "net/http"      // servidor e cliente HTTP
    "os"            // variáveis de ambiente, arquivos
    "errors"        // criar erros simples

    // ── LIBS EXTERNAS ──────────────────────────────────────
    // Precisam de: go get <caminho>
    "github.com/gin-gonic/gin"        // go get github.com/gin-gonic/gin
    "github.com/jackc/pgx/v5/pgxpool" // go get github.com/jackc/pgx/v5
    "github.com/joho/godotenv"         // go get github.com/joho/godotenv
    "github.com/google/uuid"           // go get github.com/google/uuid

    // ── IMPORTS INTERNOS ───────────────────────────────────
    // Seus próprios arquivos — usa o nome do módulo do go.mod
    "fruit-api/internal/domain"
    "fruit-api/internal/service"
)
```

```bash
# instalar dependência externa (equivalente ao npm install)
go get github.com/gin-gonic/gin

# remover imports não usados e organizar o go.mod
go mod tidy
```

> **Regra do Go:** se você importar algo e não usar, o código **não compila**. O compilador é rígido com isso — diferente do Node que apenas ignora imports não utilizados.

---

## 3. Structs e Interfaces — o centro de tudo

Em Go não existe `class`. No lugar, usamos **structs** (para dados) e **interfaces** (para comportamento).

### Struct — o molde dos dados

```go
// Go
type Fruit struct {
    ID          string  `json:"id"`
    Name        string  `json:"name"`
    Color       string  `json:"color"`
    Price       float64 `json:"price"`
    WeightGrams int     `json:"weight_grams"`
}
```

```typescript
// TypeScript equivalente
interface Fruit {
    id: string
    name: string
    color: string
    price: number
    weightGrams: number
}
```

As **struct tags** (entre crases) são instruções para bibliotecas externas:

```go
Name  string `json:"name"  binding:"required"`
//             ↑             ↑
//     chave no JSON     validação do Gin
//     "name" minúsculo  campo obrigatório
```

```go
Price float64 `json:"price" binding:"required,gt=0"`
//                                           ↑
//                               gt=0 = greater than zero
//                               valor precisa ser maior que zero
```

A tag especial `json:"-"` **esconde o campo** do JSON — nunca aparece na resposta:

```go
PasswordHash string `json:"-"`
// mesmo que você retorne a struct inteira, esse campo nunca vaza
```

### Interface — o contrato

Interface define **o que** um tipo deve fazer, sem dizer **como**:

```go
// Go — define o contrato
type FruitRepository interface {
    Save(fruit Fruit) error
    FindByName(name string) (*Fruit, error)
}
```

```typescript
// TypeScript equivalente
interface FruitRepository {
    save(fruit: Fruit): Promise<void>
    findByName(name: string): Promise<Fruit | null>
}
```

Qualquer struct que tiver esses métodos **automaticamente** satisfaz a interface — sem precisar declarar `implements` como no Java ou TypeScript.

> **Por que isso importa?** O service depende da interface, não da implementação concreta. Você pode trocar PostgreSQL por MySQL, ou criar um mock para testes, sem tocar no service.

---

## 4. Ponteiros — `*` e `&`

Ponteiro é o **endereço de memória** de uma variável. Em vez de copiar o valor, você passa o endereço — quem recebe acessa o original.

```go
nome := "Cláudio"
// nome está guardado em algum endereço, ex: 0xc000

// & = "me dá o ENDEREÇO dessa variável"
endereco := &nome
// endereco = 0xc000

// * = "me dá o VALOR que está nesse endereço"
valor := *endereco
// valor = "Cláudio"
```

```
Memória:
  endereço 0xc000  →  "Cláudio"

  &nome    →  0xc000        (o endereço)
  *(&nome) →  "Cláudio"     (o valor no endereço)
```

### Onde aparece no projeto:

**1. No tipo — significa "pode ser nil":**
```go
func FindByName(name string) (*Fruit, error)
//                            ↑
// *Fruit = ponteiro pra Fruit
// pode retornar nil = "não encontrei nenhuma fruta"
// Fruit sem * não pode ser nil — seria uma struct vazia
```

**2. No retorno — criando ponteiro de uma struct:**
```go
fruta := Fruit{Name: "Manga", ...}
return &fruta, nil
//      ↑
// & retorna o endereço da struct
// quem recebe acessa a struct original, não uma cópia
```

**3. No ShouldBindJSON — para escrever na variável:**
```go
var input CreateFruitInput
c.ShouldBindJSON(&input)
//               ↑
// & passa o endereço de input
// ShouldBindJSON precisa do endereço para ESCREVER os dados nela
// sem &, receberia uma cópia — as mudanças se perderiam
```

```typescript
// No Node.js você não pensa nisso — JS passa objetos por referência automaticamente
// Go é explícito: você decide quando quer referência (&) e quando quer cópia
```

---

## 5. Tipos de função

Go tem quatro formas de declarar funções. Vindo do Node, a estrutura parece diferente no início.

```javascript
// Node.js — você conhece assim:
function somar(a, b) { return a + b }
const somar = (a, b) => a + b
```

```go
// Go — a estrutura completa:
func NomeFuncao(parametro tipo) tipoRetorno {
    return valor
}
```

### Tipo 1 — Função simples

```go
func Somar(a int, b int) int {
    return a + b
}

// Chamada:
resultado := Somar(2, 3)  // 5
```

### Tipo 2 — Múltiplos retornos (muito comum em Go)

```go
// Go permite retornar mais de um valor — nativo da linguagem
func Dividir(a int, b int) (int, error) {
//                          ↑    ↑
//                    valor1  valor2

    if b == 0 {
        return 0, errors.New("divisão por zero")
    }
    return a / b, nil
}

// Chamada — você recebe os dois valores:
resultado, err := Dividir(10, 2)

// No Node.js você teria que retornar um objeto: { result, error }
// Em Go múltiplos retornos são nativos
```

### Tipo 3 — Método (função com dono — a struct)

```go
//       ↓ isso é o RECEIVER — define quem "dono" da função
func (s *FruitService) CreateFruit(input CreateFruitInput) (*Fruit, error) {
//    ↑      ↑
//  apelido  struct dona
    s.repo.FindByName(...)  // s = acesso à struct, igual ao "this" do JS
}
```

```javascript
// Equivalente em JS:
class FruitService {
    createFruit(input) {
        this.repo.findByName(...)  // "this" é implícito
    }
}
```

### Tipo 4 — Sem retorno

```go
func (h *FruitHandler) RegisterRoutes(r *gin.Engine) {
// sem tipo de retorno = não retorna nada (equivalente ao void do TypeScript)
    r.POST("/frutas/", h.CreateFruit)
}
```

---

## 6. O receiver — as letras soltas `r`, `s`, `h`, `c`

Essa é a parte que mais confunde quem vem do Node. Aquelas letras soltas são o **receiver** — o "this" do Go, só que você escolhe o nome.

```go
func (s *FruitService) CreateFruit(...) { }
//    ↑
//    "s" é o apelido para acessar a FruitService dentro da função
//    É o "this" — você acessa os campos com s.repo, s.qualquerCampo

func (r *fruitRepository) Save(...) { }
//    ↑
//    "r" de repository

func (h *FruitHandler) RegisterRoutes(...) { }
//    ↑
//    "h" de handler

func (c *gin.Context) ... { }
//    ↑
//    "c" de context — vem do Gin, representa a requisição HTTP
```

**A convenção é usar a inicial minúscula da struct.** Não é obrigatório, mas é tão universal que você vai ver em todo código Go.

```go
// Por que com * no receiver?
func (s *FruitService) CreateFruit(...) {
//    ↑
// Com *: acessa a struct ORIGINAL — mudanças persistem
// Sem *: receberia uma CÓPIA — mudanças se perderiam
// Quase sempre você quer *, exceto para structs pequenas e imutáveis
```

---

## 7. Tratamento de erros

Go não tem `try/catch`. Erros são **valores** — retornados como segundo valor de uma função e verificados com `if`.

```javascript
// Node.js
try {
    const user = await createUser(input)
    res.json(user)
} catch (error) {
    res.status(500).json({ error: error.message })
}
```

```go
// Go — erros são retornos normais
user, err := service.CreateUser(input)
//     ↑
//     err é nil se deu certo, ou contém o erro se deu errado

if err != nil {
    // deu errado — trata aqui
    c.JSON(500, gin.H{"error": err.Error()})
    return
}
// chegou aqui = deu certo, user tem o valor
c.JSON(201, user)
```

### O padrão mais compacto:

```go
// Declara e verifica na mesma linha:
if err := repo.Save(fruit); err != nil {
//  ↑ err só existe dentro deste if
    return nil, err
}
```

### Criando erros:

```go
import "errors"

// erro simples:
return nil, errors.New("fruta já cadastrada")

// erro com formatação (como template string):
return nil, fmt.Errorf("fruta com id %s não encontrada", id)
```

### Os três retornos possíveis com ponteiro + error:

```go
return &fruit, nil   // ✅ encontrou, sem erro
return nil, nil      // 🔍 não encontrou, sem erro (não existe)
return nil, err      // ❌ erro real (banco fora do ar, etc.)
```

---

## 8. Context

`context` é o sistema do Go para **controlar o tempo de vida de uma operação**. Toda query no banco, toda chamada HTTP externa deve receber um context.

```go
import "context"

// Context mais simples — sem timeout, sem cancelamento:
context.Background()
// Use quando não há nenhum contexto pai
// É o ponto de partida — o "contexto raiz"

// Context com timeout — cancela automaticamente após X tempo:
ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
defer cancel()
// Se a query demorar mais que 3 segundos, é cancelada automaticamente
// defer cancel() garante que os recursos são liberados quando a função terminar
```

No projeto de vocês, aparece em todas as queries:

```go
r.db.Exec(
    context.Background(),  // ← sempre o primeiro argumento
    "INSERT INTO fruits...",
    fruit.ID, fruit.Name,
)
```

```javascript
// No Node.js você não passa context explicitamente
// O Go é explícito — você decide o comportamento de cada operação
```

> Em produção, o context vem da requisição HTTP (`c.Request.Context()`) e é passado até o banco — se o cliente cancelar a requisição, as queries pendentes são canceladas automaticamente também.

---

## 9. Como tudo se conecta — visão geral

Agora que você conhece os conceitos, veja como as peças se encaixam:

```
go.mod
└── define o nome do módulo e lista as dependências
    sem ele, nenhum import funciona

.env
└── guarda as credenciais do banco (nunca no git)
    carregado pelo config.go antes de qualquer coisa

config/config.go
└── lê o .env com godotenv
    cria o pool de conexões com o PostgreSQL
    equivalente ao db.js do Node.js

internal/domain/fruit.go  ← O CENTRO DE TUDO
└── define as structs (Fruit, CreateFruitInput)
    define a interface (FruitRepository)
    todos os outros arquivos importam esse
    nenhuma lógica aqui — só "formatos" e "contratos"

internal/repository/fruit_repository.go
└── implementa a interface FruitRepository
    é o único que fala com o banco
    escreve o SQL na mão (sem ORM)
    não sabe que o service existe

internal/service/fruit_service.go
└── contém as regras de negócio
    "fruta já existe? retorna erro"
    "gera o UUID, chama o repository"
    não sabe que o handler existe
    não sabe como os dados chegam (HTTP? CLI? tanto faz)

internal/handler/fruit_handler.go
└── equivalente ao routes + controller do Node.js
    recebe a requisição HTTP
    valida o body com ShouldBindJSON
    chama o service
    devolve a resposta (201, 400, 409...)

cmd/api/main.go  ← O PONTO DE ENTRADA
└── roda uma vez quando o servidor sobe
    carrega .env → conecta banco → monta as dependências → sobe o Gin
    é o único que conhece TODAS as camadas
    depois que o Gin sobe, o main entrega o controle pra ele
    cada requisição vai direto pro handler — não passa pelo main
```

### O fluxo de uma requisição POST /frutas/

```
Cliente
  │  POST /frutas/ + JSON
  ▼
Gin (servidor HTTP)
  │  identifica a rota, chama o handler
  ▼
Handler
  │  valida o JSON (ShouldBindJSON)
  │  chama service.CreateFruit(input)
  ▼
Service
  │  verifica se nome já existe (repo.FindByName)
  │  gera UUID
  │  chama repo.Save(fruit)
  ▼
Repository
  │  executa INSERT no PostgreSQL
  ▼
PostgreSQL
  │  salva a linha, retorna sucesso
  │
  │  (sobe de volta)
  ▼
Handler
  │  c.JSON(201, fruit)
  ▼
Cliente recebe
  {"id":"...","name":"Manga","color":"Amarela","price":4.99,"weight_grams":400}
```

### A injeção de dependência no main.go

```go
// Cada camada recebe a de baixo — ninguém se instancia sozinho:
db          := config.NewDBConnection()
fruitRepo   := repository.NewFruitRepository(db)      // recebe o banco
fruitSvc    := service.NewFruitService(fruitRepo)     // recebe o repository
fruitHandler := handler.NewFruitHandler(fruitSvc)     // recebe o service

// No Node.js sem framework você faria a mesma coisa manualmente:
// const repo    = createFruitRepository(pool)
// const service = createFruitService(repo)
// const handler = createFruitHandler(service)
```

---

## Resumo rápido para ter sempre à mão

| Conceito | O que é | Quando usar |
|---|---|---|
| `go mod init` | inicializa o módulo | primeiro passo em todo projeto |
| `go get` | instala dependência externa | quando precisar de lib externa |
| `struct` | agrupa dados (como interface do TS) | modelar dados |
| `interface` | define contrato de métodos | desacoplar camadas |
| `*Tipo` | ponteiro — pode ser nil | retorno que pode não existir |
| `&variavel` | endereço de memória | passar referência, não cópia |
| `(s *Struct)` | receiver — o "this" do Go | método pertence a uma struct |
| `valor, err :=` | múltiplos retornos | padrão Go em toda função |
| `if err != nil` | verificar erro | substitui o try/catch |
| `context.Background()` | contexto raiz | primeiro argumento de queries |
| `json:"nome"` | struct tag — chave no JSON | serialização |
| `binding:"required"` | struct tag — validação do Gin | validar campos do body |
| `json:"-"` | struct tag — esconde do JSON | campos sensíveis (hash de senha) |