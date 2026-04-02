// package service contém as REGRAS DE NEGÓCIO da aplicação.
// Ele não sabe como os dados chegam (HTTP, gRPC, CLI...) — isso é problema do handler.
// Ele não sabe como os dados são salvos (PostgreSQL, MySQL...) — isso é problema do repository.
// O service só sabe O QUE fazer, não COMO fazer nas pontas.
//
// No Node.js seria equivalente ao seu arquivo service/userService.js.
package service

import (
	"errors"
	"exercicio/internal/domain"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	// ↑ golang.org/x/crypto é mantido pela equipe oficial do Go.
	// Não é uma lib externa de terceiro — é a extensão oficial da stdlib.
	// bcrypt é o algoritmo correto para hash de senhas:
	// foi projetado para ser LENTO de propósito (~100ms por hash),
	// dificultando ataques de força bruta.
	// SHA256, MD5 são rápidos demais para senhas (bilhões de tentativas/segundo com GPU).
)

// UserService é a struct que guarda as dependências do service.
// Maiúsculo = exportado — outros packages (o handler) podem usar.
//
// No Node.js seria:
//   class UserService {
//     constructor(private repo: UserRepository) {}
//   }
type UserService struct {
	repo domain.UserRepository
	// ↑ Guarda a INTERFACE, não a implementação concreta.
	// O service não sabe se é PostgreSQL, MySQL, ou um mock de testes por baixo.
	// Isso é injeção de dependência — o service recebe o que precisa, não vai buscar.
}

// NewUserService é o construtor do service.
// Recebe o repository (que implementa a interface) e retorna o service pronto.
//
// No Node.js seria:
//   export function createUserService(repo: UserRepository): UserService {
//     return new UserService(repo)
//   }
func NewUserService(repo domain.UserRepository) *UserService {
	return &UserService{repo: repo}
	// ↑ & cria um ponteiro para a struct UserService.
	// Retornamos ponteiro (*UserService) porque a struct pode ser grande
	// e não queremos copiar ela toda vez que for passada pra uma função.
}

// CreateUser orquestra a criação de um usuário.
// Recebe o input validado, aplica as regras de negócio e salva no banco.
//
// (s *UserService) = receiver, o "this" do Go. "s" de UserService.
//
// No Node.js seria:
//   async createUser(input: CreateUserInput): Promise<User> {
//     const existing = await this.repo.findByEmail(input.email)
//     if (existing) throw new Error('email já cadastrado')
//     const passwordHash = await bcrypt.hash(input.password, 10)
//     const user = { id: uuid(), ...input, passwordHash }
//     await this.repo.save(user)
//     return user
//   }
func (s *UserService) CreateUser(input domain.CreateUserInput) (*domain.User, error) {

	// REGRA 1: email não pode ser duplicado.
	existing, err := s.repo.FindByEmail(input.Email)
	// ↑ chama o repository — que vai ao banco verificar se o email já existe.
	// existing será *User ou nil.

	if err != nil {
		// Erro de infraestrutura — banco fora do ar, timeout, etc.
		// O service não sabe tratar isso — repassa pro handler decidir o que fazer.
		return nil, err
	}

	if existing != nil {
		// existing != nil significa que FindByEmail retornou um usuário — email já existe.
		// Essa é uma REGRA DE NEGÓCIO — o service É responsável por ela.
		// Retornamos nil pro usuário e um erro descritivo.
		return nil, errors.New("email já cadastrado")
	}

	// REGRA 2: nunca salvar senha em texto puro — sempre fazer o hash.
	passwordHash, err := hashPassword(input.Password)
	// ↑ chama a função privada definida abaixo neste mesmo arquivo.
	// Separamos em função própria para deixar CreateUser legível.
	if err != nil {
		return nil, err
	}

	// Monta a struct User completa, pronta para ser salva.
	// O ID é gerado aqui pelo servidor — o cliente não manda isso.
	user := domain.User{
		ID:           uuid.NewString(),
		// ↑ gera um UUID v4 aleatório: "a1b2c3d4-e5f6-..."
		// No Node.js você usaria: crypto.randomUUID() ou a lib uuid

		Name:         input.Name,
		Email:        input.Email,
		PasswordHash: passwordHash,
		// ↑ hash da senha — nunca a senha original.
		// json:"-" no domain garante que isso nunca vaza na resposta JSON.
	}

	// Salva no banco via repository.
	if err := s.repo.Save(user); err != nil {
		// ↑ "if err := x(); err != nil" é o padrão Go mais compacto:
		// declara err dentro do if e já verifica na mesma linha.
		return nil, err
	}

	return &user, nil
	// ↑ Retorna ponteiro para o user criado + nil (sem erro).
	// & pega o endereço da variável local user.
	// O Go gerencia a memória — user não some quando a função termina.
}

// hashPassword gera o hash bcrypt de uma senha.
// É privada (minúsculo) — só o service usa. Não faz sentido expor isso.
//
// No Node.js seria:
//   async function hashPassword(password: string): Promise<string> {
//     return bcrypt.hash(password, 10)  // 10 = custo/rounds
//   }
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		// ↑ bcrypt recebe []byte (slice de bytes), não string.
		// []byte("minhasenha") converte: cada caractere vira seu valor em byte.
		// No Node.js o bcrypt aceita string diretamente — Go é mais explícito.

		bcrypt.DefaultCost,
		// ↑ custo do algoritmo = 10 (valor padrão recomendado).
		// Cada +1 de custo dobra o tempo de processamento:
		//   custo 10 → ~100ms por hash
		//   custo 11 → ~200ms
		//   custo 12 → ~400ms
		// Lento o suficiente para inviabilizar força bruta em massa.
	)

	if err != nil {
		return "", err
		// ↑ string vazia + erro — padrão Go quando algo dá errado e
		// não temos um valor útil para retornar.
	}

	return string(bytes), nil
	// ↑ converte []byte de volta para string para guardar no banco.
	// O hash final tem este formato:
	// "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"
	//   ↑    ↑  ↑
	// versão custo  hash de 53 caracteres
}
