# Conceitos fundamentais de Go — para quem vem do Node.js

> Cada conceito explicado com analogias, comparações com Node.js e exemplos do nosso projeto.
> Leia na ordem — cada conceito prepara o próximo.

---

## Índice

1. [Variáveis e tipos](#1-variáveis-e-tipos)
2. [Structs — o molde dos dados](#2-structs--o-molde-dos-dados)
3. [Funções e métodos](#3-funções-e-métodos)
4. [Ponteiros — `*` e `&`](#4-ponteiros----e-)
5. [Interfaces — o contrato](#5-interfaces--o-contrato)
6. [Tratamento de erros](#6-tratamento-de-erros)
7. [JSON tags e binding tags](#7-json-tags-e-binding-tags)
8. [Imports — stdlib vs externos](#8-imports--stdlib-vs-externos)
9. [Context](#9-context)
10. [Como tudo se conecta](#10-como-tudo-se-conecta)

---

## 1. Variáveis e tipos

No Node.js você declara variáveis com `let`, `const` ou `var` — o tipo é inferido automaticamente e pode mudar.

```javascript
// Node.js — tipo dinâmico
let nome = "Manga"   // JS decide que é string
nome = 42            // JS deixa mudar o tipo — sem erro
```

Em Go, toda variável tem um tipo fixo. Você pode deixar o Go inferir, mas depois não muda.

```go
// Go — tipo fixo
nome := "Manga"   // Go infere: é string
nome = 42         // ERRO — nome é string, não aceita int
```

Os tipos mais comuns que você vai usar:

```go
nome   := "Manga"     // string  — texto
preco  := 4.99        // float64 — número decimal
peso   := 400         // int     — número inteiro
ativo  := true        // bool    — verdadeiro/falso
```

Quando vai zero, Go tem um valor padrão para cada tipo — diferente do JS que usa `undefined`:

```go
var nome  string   // ""    (string vazia)
var preco float64  // 0.0
var peso  int      // 0
var ativo bool     // false
```

---

## 2. Structs — o molde dos dados

Em Go não existe `class`. No lugar, usamos **struct** — um agrupamento de variáveis relacionadas.

```javascript
// Node.js — você usava assim:
const fruta = {
    id: "abc-123",
    name: "Manga",
    price: 4.99
}

// Ou com TypeScript:
interface Fruta {
    id: string
    name: string
    price: number
}
```

```go
// Go — struct é o equivalente:
type Fruta struct {
    ID    string
    Name  string
    Price float64
}

// Criando uma fruta:
fruta := Fruta{
    ID:    "abc-123",
    Name:  "Manga",
    Price: 4.99,
}

// Acessando campos — igual ao JS:
fmt.Println(fruta.Name)  // "Manga"
```

A diferença principal: em Go o "molde" (struct) e o "dado" (instância) são conceitos bem separados. Você define o molde uma vez no `domain/` e usa em todo o projeto.

No nosso projeto temos dois moldes para fruta:

```go
// Fruta — como existe no banco (tem ID)
type Fruta struct {
    ID           string
    Name         string
    Color        string
    Price        float64
    Weight_grams int
}

// FrutaInput — o que o cliente manda (sem ID — gerado pelo servidor)
type FrutaInput struct {
    Name         string
    Color        string
    Price        float64
    Weight_grams int
}
```

---

## 3. Funções e métodos

### Função simples — sem dono

```javascript
// Node.js
function somar(a, b) {
    return a + b
}
```

```go
// Go
func Somar(a int, b int) int {
    return a + b
}
```

### Múltiplos retornos — exclusivo do Go

Go permite retornar mais de um valor. Isso é muito usado para retornar o resultado **e** um possível erro juntos.

```javascript
// Node.js — você teria que retornar um objeto
async function buscarFruta(nome) {
    try {
        const fruta = await db.query(...)
        return { fruta, erro: null }
    } catch (erro) {
        return { fruta: null, erro }
    }
}
```

```go
// Go — retorno múltiplo nativo
func BuscarFruta(nome string) (*Fruta, error) {
    // retorna a fruta E o erro separados — sem objeto wrapper
    return &fruta, nil   // deu certo
    return nil, err      // deu errado
}

// Quem chama recebe os dois:
fruta, err := BuscarFruta("Manga")
```

### Método — função com dono (a struct)

Quando uma função pertence a uma struct, ela vira um **método**. O "dono" é declarado entre `func` e o nome:

```javascript
// Node.js — o dono é implícito (this)
class FrutaService {
    createFruta(input) {
        this.repo.save(...)   // "this" é o dono
    }
}
```

```go
// Go — o dono é explícito (o receiver)
func (s *FrutaService) CreateFruta(input FrutaInput) (*Fruta, error) {
//    ↑
//    "s" é o apelido para acessar a struct — equivalente ao "this"
//    convenção: usar a inicial minúscula do nome da struct
    s.repo.Save(...)
}
```

### Como ler qualquer função em Go

```
func  (s *FrutaService)  CreateFruta  (input FrutaInput)  (*Fruta, error)
  ↑          ↑               ↑               ↑                  ↑
"é uma    "quem é       "nome da        "o que entra"       "o que sai"
função"   o dono"        função"
         (opcional)                                          (opcional)
```

Tem `()` antes do nome → é um **método** (tem dono).
Não tem → é uma **função pura** (sem dono).

---

## 4. Ponteiros — `*` e `&`

Essa é a parte que mais confunde quem vem do Node. Mas a lógica é simples quando você entende o motivo.

### O que acontece quando você passa uma variável pra uma função

Imagina que sua variável é um documento. Quando você chama uma função em Go, Go **tira uma xerox** desse documento e manda a xerox pra função. A função trabalha na xerox — o original fica intacto.

```go
type Fruta struct {
    Name string
}

func renomear(f Fruta) {
    f.Name = "Banana"
    // mudou a XEROX — o original não sabe disso
}

fruta := Fruta{Name: "Manga"}
renomear(fruta)

fmt.Println(fruta.Name) // "Manga" — original intacto
// a função mudou a xerox, não adiantou nada
```

```javascript
// No Node.js objetos NÃO são xerocados — JS manda o original automaticamente
function renomear(f) {
    f.name = "Banana"  // muda o original — JS fez isso por você
}

const fruta = { name: "Manga" }
renomear(fruta)
console.log(fruta.name)  // "Banana"
```

**Esse é o ponto.** JS te protegeu disso automaticamente. Go não — em Go você decide.

### Quando a xerox vira problema

Pensa no nosso projeto. O `main` cria o pool de conexões com o banco:

```go
db := config.NewDBConnection()
// pool de conexões — coisa grande, representa conexões abertas com o PostgreSQL
```

Agora passa pro repository:

```go
repo := NewFrutaRepository(db)
```

**Sem ponteiro**, Go tiraria uma xerox do pool inteiro e mandaria pro repository. Uma xerox de conexões abertas com o banco — que não funciona, porque as conexões reais estão no original. É como xerocar uma chave: a cópia não abre a porta.

**Com ponteiro**, você manda o endereço do original. Repository e main acessam o mesmo pool — as mesmas conexões reais.

### O que é esse "endereço de memória"

Esquece "memória" por um segundo. Pensa assim:

Você tem um apartamento. O apartamento é a sua variável. O endereço (`Rua X, nº 42`) é o ponteiro.

```
Sem ponteiro:
  main tem o apartamento original
  manda uma PLANTA BAIXA (xerox) pro repository
  repository olha a planta — mas não consegue entrar no apartamento real

Com ponteiro:
  main tem o apartamento original
  manda o ENDEREÇO pro repository
  repository vai até o endereço e entra no apartamento real
```

### O que `*` e `&` fazem

```go
nome := "Manga"

&nome   // & = "me dá o ENDEREÇO dessa variável"
        // resultado: algo como 0xc000 (onde "Manga" está na memória)

*(&nome) // * = "me dá o VALOR que está nesse endereço"
         // resultado: "Manga"
```

```
Resumo:
  &  →  de valor para endereço    (vai pro original)
  *  →  de endereço para valor    (lê o original)
```

### Os três momentos onde você usa no projeto

**1 — Passar algo pesado sem copiar**

```go
// sem ponteiro — xerox do pool (não funciona)
func NewFrutaRepository(db pgxpool.Pool) { }

// com ponteiro — endereço do pool original (funciona)
func NewFrutaRepository(db *pgxpool.Pool) { }
```

**2 — Representar ausência de valor (nil)**

```go
func (r *frutaRepository) FindByName(name string) (*Fruta, error) {
//                                                  ↑
// *Fruta pode ser nil = "não encontrei nada"
// Fruta sem * não pode ser nil — Go exigiria uma struct mesmo que vazia
// Como você diria "não existe" sem ponteiro? Não tem como.

    return nil, nil    // não encontrou — não é erro
    return &fruta, nil // encontrou
}
```

**3 — Deixar uma função escrever na sua variável**

```go
var input FrutaInput  // variável vazia

c.ShouldBindJSON(&input)
//               ↑
// ShouldBindJSON precisa ESCREVER dentro de input
// sem & ela receberia a xerox — preencheria a xerox, jogaria fora
// com & ela vai no original e escreve lá
```

### Por que o JS nunca te forçou a pensar nisso

```javascript
// JS tem duas regras que aplica sozinho:
// Primitivos (number, string, boolean) → sempre copia
// Objetos e arrays → sempre manda o original

let a = 42
let b = a
b = 100
console.log(a) // 42 — primitivo, copiou

let fruta = { name: "Manga" }
let outra = fruta
outra.name = "Banana"
console.log(fruta.name) // "Banana" — objeto, mandou o original
```

Go não tem essa distinção automática. Tudo é copiado por padrão. Se você quer o original, diz explicitamente com `&`. Mais trabalhoso, mas você sempre sabe exatamente o que está acontecendo.

> **Resumo em uma frase:** ponteiro existe porque Go copia tudo por padrão, e às vezes você precisa do original — seja para não duplicar algo pesado, para representar ausência com `nil`, ou para deixar uma função escrever na sua variável.

---

## 5. Interfaces — o contrato

Interface define **o que** um tipo deve saber fazer, sem dizer **como**.

```typescript
// TypeScript — você conhece assim:
interface FrutaRepository {
    save(fruta: Fruta): Promise<void>
    findByName(name: string): Promise<Fruta | null>
}
```

```go
// Go — mesma ideia:
type FrutaRepository interface {
    Save(fruta Fruta) error
    FindByName(name string) (*Fruta, error)
}
```

Qualquer struct que tiver esses dois métodos **automaticamente** satisfaz a interface — sem precisar declarar `implements`.

```go
// frutaRepository tem os dois métodos → satisfaz FrutaRepository automaticamente
type frutaRepository struct { db *pgxpool.Pool }

func (r *frutaRepository) Save(fruta Fruta) error { ... }
func (r *frutaRepository) FindByName(name string) (*Fruta, error) { ... }
```

### Por que isso importa na prática

O service guarda a **interface**, não a struct concreta:

```go
type FrutaService struct {
    repo domain.FrutaRepository  // interface — não sabe se é postgres, mysql, mock...
}
```

Isso significa que você pode trocar o banco de dados sem tocar no service. Ou criar um repositório falso para testes. O service não liga — ele só conhece o contrato.

```
frutaRepository (PostgreSQL)  ┐
frutaRepositoryMySQL          ├── todos satisfazem FrutaRepository
MockFrutaRepository (testes)  ┘
```

---

## 6. Tratamento de erros

Go não tem `try/catch`. Erros são **valores** — retornados como segundo valor e verificados com `if`.

```javascript
// Node.js
try {
    const fruta = await service.createFruta(input)
    res.status(201).json(fruta)
} catch (error) {
    res.status(409).json({ error: error.message })
}
```

```go
// Go — erro é um retorno normal
fruta, err := service.CreateFruta(input)

if err != nil {
    c.JSON(409, gin.H{"error": err.Error()})
    return  // ← obrigatório — sem isso o código continua
}

c.JSON(201, fruta)
```

### Os três retornos possíveis

```go
return &fruta, nil  // ✅ deu certo — tem valor, sem erro
return nil, nil     // 🔍 não encontrou — sem valor, sem erro
return nil, err     // ❌ deu errado — sem valor, com erro
```

### Criando erros

```go
import "errors"

return nil, errors.New("Fruta já cadastrada paizão")

// com formatação (como template string):
return nil, fmt.Errorf("fruta com id %s não encontrada", id)
```

### Quando a função não retorna error

Se a operação não depende de nada externo (banco, rede, arquivo), não precisa retornar error:

```go
// pode falhar (banco) → retorna error
func (r *frutaRepository) Save(fruta Fruta) error { }

// não pode falhar (só registra rota) → sem error
func (h *FrutaHandler) RegisterRoutes(r *gin.Engine) { }
```

---

## 7. JSON tags e binding tags

As instruções entre crases nas structs dizem como os campos se comportam — no JSON e na validação.

```go
type FrutaInput struct {
    Name         string  `json:"name"  binding:"required"`
    Price        float64 `json:"price" binding:"required,gt=0"`
    Weight_grams int     `json:"weight_grams" binding:"required,min=1"`
}
```

**JSON tags** — controlam como o campo aparece no JSON:

```go
Name string `json:"name"`
// struct Go: Name (maiúsculo)
// JSON:      "name" (minúsculo) — convencional em APIs

PasswordHash string `json:"-"`
// json:"-" = NUNCA aparece no JSON
// mesmo retornando a struct inteira, esse campo fica escondido
```

**Binding tags** — validação automática do Gin:

```go
`binding:"required"`        // campo obrigatório
`binding:"required,email"`  // obrigatório + deve ser email válido
`binding:"required,min=8"`  // obrigatório + mínimo 8 caracteres
`binding:"required,gt=0"`   // obrigatório + maior que zero
```

Se a validação falhar, `ShouldBindJSON` retorna erro automaticamente — sem precisar validar na mão.

---

## 8. Imports — stdlib vs externos

```go
import (
    // ── STDLIB ─────────────────────────────────────────────────
    // Vêm com o Go — não precisa instalar nada
    "context"      // controle de timeout/cancelamento em operações
    "errors"       // criar erros simples
    "fmt"          // formatação de strings
    "net/http"     // códigos HTTP (StatusOK, StatusCreated...)
    "os"           // variáveis de ambiente, arquivos

    // ── EXTERNOS ───────────────────────────────────────────────
    // Precisam de: go get <caminho>
    "github.com/gin-gonic/gin"         // framework HTTP
    "github.com/jackc/pgx/v5/pgxpool"  // driver PostgreSQL
    "github.com/joho/godotenv"          // ler arquivo .env
    "github.com/google/uuid"            // gerar UUIDs

    // ── INTERNOS ───────────────────────────────────────────────
    // Seus próprios arquivos — usa o nome do go.mod
    "exercicio/internal/domain"
    "exercicio/internal/service"
)
```

```bash
# instalar dependência externa — equivalente ao npm install
go get github.com/gin-gonic/gin

# limpar imports não usados e organizar go.mod — equivalente ao npm install após editar package.json
go mod tidy
```

> **Regra importante:** se você importar algo e não usar, o código **não compila**. O Go é rígido com isso.

---

## 9. Context

`context` é o mecanismo do Go para controlar o tempo de vida de uma operação. Toda query no banco deve receber um context.

```go
// Context mais simples — sem timeout, sem cancelamento
// Use quando não há contexto pai — é o ponto de partida
context.Background()

// Context com timeout — cancela automaticamente após 3 segundos
ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
defer cancel()
// se a query demorar mais que 3s → cancelada automaticamente
// defer cancel() → libera os recursos quando a função terminar
```

No nosso projeto aparece em todas as queries:

```go
r.db.Exec(
    context.Background(),  // sempre primeiro argumento
    "INSERT INTO frutas...",
    fruta.ID, fruta.Name,
)
```

```javascript
// No Node.js você não passa context — a lib cuida disso internamente
// Go é explícito — você sempre sabe o comportamento de cada operação
```

---

## 10. Como tudo se conecta

Agora que você conhece todos os conceitos, a imagem completa:

```
go.mod
└── define o nome do módulo (usado nos imports internos)
    sem ele nenhum import funciona — primeiro passo sempre

.env
└── credenciais do banco — nunca no git
    carregado pelo config.go antes de qualquer coisa

config/config.go
└── lê o .env → monta a string de conexão → cria o pool
    equivalente ao db.js do Node.js

internal/domain/fruta.go  ← O CENTRO DE TUDO
└── define as structs (Fruta, FrutaInput)
    define a interface (FrutaRepository)
    todos os outros importam este — nenhuma lógica aqui

internal/repository/fruta_repository.go
└── implementa FrutaRepository com SQL puro
    único que fala com o banco
    não sabe que o service existe

internal/service/fruta_service.go
└── regras de negócio ("fruta já existe? retorna erro")
    não sabe de HTTP, não sabe de SQL
    fala só com a interface — não com o repository diretamente

internal/handler/fruta_handler.go
└── equivalente ao routes + controller do Node.js
    recebe req → valida body → chama service → responde

cmd/api/main.go  ← PONTO DE ENTRADA
└── roda uma vez quando o servidor sobe
    monta a cadeia de dependências → sobe o Gin
    depois disso, cada requisição vai direto pro handler
```

### A cadeia de dependências montada no main

```go
// Cada camada recebe a de baixo — ninguém se cria sozinho
db      := config.NewDBConnection()        // pool de conexões
repo    := NewFrutaRepository(db)          // recebe o pool
service := NewFrutaService(repo)           // recebe o repository
handler := NewFrutaHandler(service)        // recebe o service

// No Node.js você fazia igual — Go só te força a ser explícito:
// const repo    = createFrutaRepository(pool)
// const service = createFrutaService(repo)
// const handler = createFrutaHandler(service)
```

### O fluxo de uma requisição

```
POST /frutas/registraFrutinha + JSON no body
        │
        ▼
    handler → valida o body (ShouldBindJSON)
        │
        ▼
    service → verifica se já existe, gera UUID
        │
        ▼
    repository → executa INSERT no PostgreSQL
        │
        ▼
    PostgreSQL salva
        │
        │  (sobe de volta)
        ▼
    handler → responde 201 + JSON da fruta
        │
        ▼
    Cliente recebe {"id":"...","name":"Manga",...}
```
