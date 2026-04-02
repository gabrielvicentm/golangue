// package main é o ponto de entrada da aplicação.
// Em Go, todo executável precisa ter um package main com uma função main().
// É equivalente ao index.js/server.js do Node.js.
//
// O main não tem lógica de negócio — ele só:
// 1. Carrega configurações
// 2. Conecta ao banco
// 3. Monta as dependências (injeção de dependência manual)
// 4. Registra as rotas
// 5. Sobe o servidor
package main

import (
	"exercicio/config"
	"exercicio/internal/domain"
	"exercicio/internal/handler"
	"exercicio/internal/repository"
	"exercicio/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {

	// PASSO 1: Carregar variáveis de ambiente do arquivo .env
	// Deve ser o primeiro passo — tudo que vem depois depende das env vars.
	//
	// No Node.js seria a primeira linha: import 'dotenv/config'
	config.LoadEnv()

	// PASSO 2: Conectar ao banco de dados
	// Cria o pool de conexões com as credenciais do .env.
	//
	// No Node.js seria:
	//   const pool = new Pool({ connectionString: process.env.DATABASE_URL })
	db := config.NewDBConnection()

	defer db.Close()
	// ↑ defer = "execute isso quando a função main() encerrar".
	// Garante que o pool de conexões é fechado corretamente quando o servidor para.
	// É o equivalente ao try/finally do Node — só que mais elegante.
	// Funciona mesmo se um panic (erro fatal) acontecer.

	// PASSO 3: Montar a cadeia de dependências (Injeção de Dependência manual)
	// Cada camada recebe a de baixo como parâmetro — ninguém se instancia sozinho.
	// Ordem obrigatória: repository → service → handler (de baixo pra cima)
	//
	// No Node.js seria:
	//   const userRepository = createUserRepository(pool)
	//   const userService = createUserService(userRepository)
	//   const userHandler = createUserHandler(userService)
	userRepo    := repository.NewUserRepository(db)
	// ↑ repository recebe o pool de conexões

	userService := service.NewUserService(userRepo)
	// ↑ service recebe o repository (via interface domain.UserRepository)

	userHandler := handler.NewUserHandler(userService)
	// ↑ handler recebe o service

	// PASSO 4: Criar o servidor HTTP com o Gin
	//
	// No Node.js seria: const app = express()
	r := gin.Default()
	// ↑ gin.Default() cria o roteador com dois middlewares já configurados:
	// - Logger: loga cada requisição (método, rota, status, tempo)
	// - Recovery: captura panics e evita que o servidor crave (equivalente ao crash do Node)

	// PASSO 5: Registrar as rotas
	// Cada handler registra suas próprias rotas — o main só chama o método.
	// Se amanhã tiver productHandler, authHandler, etc., cada um tem seu RegisterRoutes().
	//
	// No Node.js seria:
	//   app.use('/users', userRouter)
	userHandler.RegisterRoutes(r)

	// Apenas para garantir que domain é usado (AutoMigrate em projetos reais usaria isso)
	_ = domain.User{}

	// PASSO 6: Subir o servidor
	// Bloqueia aqui e fica escutando requisições pra sempre (até o processo ser encerrado).
	//
	// No Node.js seria: app.listen(8080, () => console.log('Server running'))
	r.Run(":8080")
	// ↑ :8080 = escuta em todas as interfaces de rede na porta 8080.
	// Para definir host específico: r.Run("127.0.0.1:8080")
}
