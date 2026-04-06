# Exercício Frutas — Pseudocódigo comentado
> Mistura de português + Go para facilitar a leitura do grupo.

---

## Conceitos que você precisa saber antes de ler o código

| Conceito | O que é |
|---|---|
| **variável** | caixinha que guarda um valor: `nome := "Manga"` |
| **struct** | agrupamento de variáveis relacionadas — o "molde" de um dado |
| **interface** | lista de métodos que uma struct DEVE ter — um contrato |
| **função** | bloco de código com nome que executa uma tarefa |
| **método** | função que pertence a uma struct — tem um "dono" |
| **ponteiro `*`** | no tipo: significa que pode ser `nil` (vazio/inexistente) |
| **endereço `&`** | pega o endereço de memória de uma variável (manda o original, não a cópia) |
| **`error`** | tipo nativo do Go para representar erros — retornado junto com o valor |
| **`nil`** | ausência de valor — equivalente ao `null` do JavaScript |
| **JSON tag** | instrução entre crases que diz como o campo aparece no JSON da API |
| **binding tag** | instrução do Gin para validar o campo automaticamente |
| **`return`** | encerra a função e devolve os valores indicados |

---

## fruta.go — domain
> Define os "moldes" dos dados e o contrato do banco. Todos os outros arquivos dependem deste.

```
PACOTE domain

// ─── MOLDE 1 ────────────────────────────────────────────────────────────────
// Representa uma fruta como ela existe no banco de dados.
// Quando você faz SELECT no banco, o resultado vira uma struct Fruta.

MOLDE Fruta {
    ID           texto     → aparece no JSON como "id"
    Name         texto     → aparece no JSON como "name"
    Color        texto     → aparece no JSON como "color"
    Price        decimal   → aparece no JSON como "price"
    Weight_grams inteiro   → aparece no JSON como "weight_grams"
}

// ─── MOLDE 2 ────────────────────────────────────────────────────────────────
// Representa o que o CLIENTE manda no body da requisição.
// Separado do Fruta porque o cliente não manda o ID (gerado pelo servidor).

MOLDE FrutaInput {
    Name         texto     → obrigatório
    Color        texto     → obrigatório
    Price        decimal   → obrigatório, mínimo 1
    Weight_grams inteiro   → obrigatório, mínimo 1
}

// ─── CONTRATO ────────────────────────────────────────────────────────────────
// Define O QUE o repository precisa saber fazer.
// Não diz COMO — isso é problema do repository.
// O service só conhece esse contrato, nunca o repository diretamente.

CONTRATO FrutaRepository {
    Save(fruta Fruta) → retorna erro ou nil
    FindByName(name texto) → retorna *Fruta (pode ser nil) + erro ou nil
}
```

---

## fruta_handler.go — handler
> Equivalente ao routes + controller do Node.js. Recebe a requisição HTTP, valida o body e responde.

```
PACOTE handler

IMPORTA:
    net/http              → códigos HTTP (200, 201, 400...)
    domain                → os moldes (FrutaInput, Fruta...)
    service               → as regras de negócio
    gin                   → o framework HTTP

// ─── STRUCT ──────────────────────────────────────────────────────────────────
// Guarda a dependência do handler — ele só precisa do service.

MOLDE FrutaHandler {
    service → ponteiro para FrutaService
}

// ─── CONSTRUTOR ──────────────────────────────────────────────────────────────
// Cria um FrutaHandler recebendo o service já pronto.
// Chamado uma vez no main.go.

FUNÇÃO NewFrutaHandler(service) → retorna *FrutaHandler
    retorna FrutaHandler{ service: service }

// ─── ROTAS ───────────────────────────────────────────────────────────────────
// Registra quais URLs existem e qual função cada uma chama.
// Equivalente ao router.post('/registraFrutinha', controller) do Express.

MÉTODO (h FrutaHandler) RegisterRoutes(r motor do gin)
    grupo "/frutas"
        POST "/registraFrutinha" → chama h.CreateFruta

// ─── HANDLER DA ROTA ─────────────────────────────────────────────────────────
// Executado toda vez que alguém faz POST /frutas/registraFrutinha.

MÉTODO (h FrutaHandler) CreateFruta(c contexto da requisição)

    // PASSO 1 — ler e validar o body JSON
    variável input do tipo FrutaInput (vazia)

    SE body inválido ou campos faltando
        responde 400 Bad Request com { "error": "..." }
        para aqui  ← sem esse return o código continuaria

    // PASSO 2 — chamar o service com o input validado
    fruta, erro ← service.CreateFruta(input)

    SE erro não for nil
        responde 409 Conflict com { "error": "..." }
        para aqui

    // PASSO 3 — responder com sucesso
    responde 201 Created com o JSON da fruta criada
```

---

## fruta_repository.go — repository
> Única camada que fala com o banco. Escreve e lê SQL. Não tem regra de negócio.

```
PACOTE repository

IMPORTA:
    context               → necessário para queries no pgx
    domain                → os moldes
    pgxpool               → pool de conexões com o PostgreSQL

// ─── STRUCT ──────────────────────────────────────────────────────────────────
// Guarda o pool de conexões com o banco.
// Letra minúscula = privado, só usado dentro deste pacote.

MOLDE frutaRepository {
    db → ponteiro para o pool de conexões (pgxpool.Pool)
}

// ─── CONSTRUTOR ──────────────────────────────────────────────────────────────
// Recebe o pool vindo do main e retorna um frutaRepository pronto.
// Retorna a INTERFACE (FrutaRepository), não a struct concreta.
// Isso permite trocar o banco sem mudar o service.

FUNÇÃO NewFrutaRepository(db pool) → retorna FrutaRepository
    retorna frutaRepository{ db: db }

// ─── SAVE ────────────────────────────────────────────────────────────────────
// Executa o INSERT no banco. Não retorna dados, só erro ou nil.

MÉTODO (r frutaRepository) Save(fruta Fruta) → retorna erro

    executa no banco:
        INSERT INTO frutas(id, name, color, price, weight_grams)
        VALUES (fruta.ID, fruta.Name, fruta.Color, fruta.Price, fruta.Weight_grams)

    retorna o erro (nil se deu certo)

// ─── FINDBYNAME ──────────────────────────────────────────────────────────────
// Busca uma fruta pelo nome. Retorna nil se não existir — não é um erro.

MÉTODO (r frutaRepository) FindByName(name texto) → retorna *Fruta, erro

    variável fruta do tipo Fruta (vazia)

    executa no banco:
        SELECT id, name, color, price, weight_grams
        FROM frutas
        WHERE name = $1
    preenche os campos de fruta com o resultado (Scan)

    SE erro for "no rows in result set"  ← não encontrou, não é erro real
        retorna nil, nil   ← "não existe" + "sem erro"

    SE qualquer outro erro
        retorna nil, erro  ← problema no banco

    retorna &fruta, nil    ← encontrou + sem erro
                    ↑
                    & = manda o endereço da fruta, não uma cópia
```

---

## fruta_service.go — service
> Contém as regras de negócio. Não sabe de HTTP, não sabe de SQL — só sabe O QUE fazer.

```
PACOTE service

IMPORTA:
    errors                → criar erros simples
    domain                → os moldes
    uuid                  → gerar IDs únicos

// ─── STRUCT ──────────────────────────────────────────────────────────────────
// Guarda o repository — acessa o banco através do contrato (interface).
// Não sabe se é PostgreSQL, MySQL, ou mock de teste por baixo.

MOLDE FrutaService {
    repo → contrato FrutaRepository (interface)
}

// ─── CONSTRUTOR ──────────────────────────────────────────────────────────────

FUNÇÃO NewFrutaService(repo FrutaRepository) → retorna *FrutaService
    retorna FrutaService{ repo: repo }

// ─── CREATEFRUTA ─────────────────────────────────────────────────────────────
// Orquestra a criação de uma fruta aplicando as regras de negócio.

MÉTODO (s FrutaService) CreateFruta(input FrutaInput) → retorna *Fruta, erro

    // REGRA 1 — nome não pode ser duplicado
    existing, erro ← repo.FindByName(input.Name)

    SE erro não for nil
        retorna nil, erro  ← erro do banco, repassa pra cima

    SE existing não for nil  ← já existe uma fruta com esse nome
        retorna nil, erro("Fruta já cadastrada paizão")

    // monta a fruta completa com os dados do input
    fruta = Fruta {
        ID:           gera um UUID novo
        Name:         input.Name
        Color:        input.Color
        Price:        input.Price
        Weight_grams: input.Weight_grams
    }

    // REGRA 2 — salva no banco
    SE repo.Save(fruta) retornar erro
        retorna nil, erro

    retorna &fruta, nil  ← fruta criada com sucesso
```

---

## Como os arquivos se conectam

```
main.go
  │
  ├── cria pool de conexões (config.go)
  │
  ├── NewFrutaRepository(pool)   → repository conhece o banco
  │         │
  ├── NewFrutaService(repo)      → service conhece o repository
  │         │
  └── NewFrutaHandler(service)   → handler conhece o service
            │
            └── RegisterRoutes(gin)  → gin conhece as rotas

Requisição POST /frutas/registraFrutinha
    → handler valida o body
        → service aplica as regras
            → repository executa o SQL
                → banco salva
            ← repository retorna resultado
        ← service retorna fruta ou erro
    ← handler responde 201 ou 4xx
```
