// package handler é a camada HTTP — a "porta de entrada" da aplicação.
// Ele só sabe lidar com requisições e respostas HTTP.
// Não contém regra de negócio — só recebe, valida o formato, chama o service e responde.
//
// No Node.js com Express seria equivalente ao seu controller:
//   router.post('/users', async (req, res) => { ... })
package handler

import (
	"net/http"
	"exercicio/internal/domain"
	"exercicio/internal/service"

	"github.com/gin-gonic/gin"
)

// UserHandler guarda as dependências necessárias para lidar com requests de usuário.
// Neste caso, só precisa do service.
//
// No Node.js seria:
//   class UserController {
//     constructor(private userService: UserService) {}
//   }
type UserHandler struct {
	service *service.UserService
	// ↑ ponteiro para o service — compartilha a mesma instância,
	// não cria uma nova para cada request.
}

// NewUserHandler é o construtor do handler.
// Recebe o service já configurado e retorna o handler pronto para usar.
//
// No Node.js seria:
//   export function createUserHandler(service: UserService) {
//     return new UserController(service)
//   }
func NewUserHandler(service *service.UserService) *UserHandler {
	return &UserHandler{service: service}
}

// RegisterRoutes registra todas as rotas relacionadas a usuários no roteador do Gin.
// É chamado uma única vez no main, quando o servidor está subindo.
//
// (h *UserHandler) = receiver, o "this" do Go. "h" de UserHandler.
//
// No Node.js com Express seria:
//   const router = express.Router()
//   router.post('/', createUser)
//   app.use('/users', router)
func (h *UserHandler) RegisterRoutes(r *gin.Engine) {
	users := r.Group("/users")
	// ↑ Group cria um prefixo de rota — todas as rotas dentro terão /users na frente.
	// Equivalente ao express.Router() com app.use('/users', router).

	{
		users.POST("/", h.CreateUser)
		// ↑ POST /users/ → chama o método CreateUser deste handler.
		// h.CreateUser é a referência ao método — não chama, só aponta.
		// O Gin vai chamar quando a request chegar.
	}
}

// CreateUser lida com a requisição POST /users/
// Valida o body, chama o service e devolve a resposta apropriada.
//
// *gin.Context tem tudo da requisição: body, headers, params, query string...
// E os métodos para responder: c.JSON(), c.String(), c.Status()...
//
// No Node.js com Express seria:
//   async createUser(req: Request, res: Response) {
//     const { name, email, password } = req.body
//     // validar...
//     const user = await this.userService.createUser({ name, email, password })
//     res.status(201).json(user)
//   }
func (h *UserHandler) CreateUser(c *gin.Context) {

	// PASSO 1: Ler e validar o corpo da requisição.
	var input domain.CreateUserInput
	// ↑ declara uma variável vazia do tipo CreateUserInput.
	// Vamos preencher ela com os dados do body JSON logo abaixo.

	if err := c.ShouldBindJSON(&input); err != nil {
		// ↑ ShouldBindJSON faz duas coisas ao mesmo tempo:
		// 1. Deserializa o JSON do body para a struct input.
		// 2. Valida as regras das struct tags: required, email, min=8...
		//
		// & passa o ENDEREÇO de input — ShouldBindJSON precisa escrever nela.
		//
		// No Node.js você faria:
		//   const { error } = schema.validate(req.body)
		//   if (error) return res.status(400).json({ error: error.message })

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		// ↑ http.StatusBadRequest = 400
		// gin.H é atalho para map[string]interface{} — monta o JSON de erro.
		// err.Error() converte o erro para string legível.
		// Exemplos de mensagem: "Key: 'email' Error: must be a valid email address"

		return
		// ↑ CRÍTICO: sem esse return, o código continuaria executando.
		// Em Go não existe o "early return automático" do Express — você precisa
		// sempre dar return explicitamente após responder com erro.
		// No Node.js o return no res.json() já fazia isso.
	}

	// PASSO 2: Chamar o service com o input já validado.
	user, err := h.service.CreateUser(input)
	// ↑ O handler delega a lógica pro service.
	// Recebe o usuário criado OU um erro de negócio/infraestrutura.

	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		// ↑ http.StatusConflict = 409 — usado quando algo já existe.
		// Neste caso: email duplicado.
		// Em produção você teria erros tipados para decidir o status code correto
		// (409 para conflito, 500 para erro de infra, etc.)
		return
	}

	// PASSO 3: Responder com sucesso.
	c.JSON(http.StatusCreated, user)
	// ↑ http.StatusCreated = 201 — padrão REST para criação bem-sucedida.
	// user é a struct *domain.User — o Gin serializa para JSON automaticamente.
	// Como PasswordHash tem json:"-" no domain, ele NUNCA aparece aqui.
	//
	// Resposta: {"id": "uuid", "name": "Cláudio", "email": "claudio@email.com"}
	//
	// No Node.js seria: res.status(201).json(user)
}
