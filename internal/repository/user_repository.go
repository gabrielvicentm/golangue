// package repository é responsável APENAS por falar com o banco de dados.
// Ele não sabe nada de regra de negócio — só executa SQL.
// No Node.js seria equivalente ao seu arquivo de queries/DAL (Data Access Layer).
package repository

import (
	"context"
	"exercicio/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

// userRepository é a struct que guarda a conexão com o banco.
// Letra minúscula = privada ao package — ninguém de fora instancia diretamente.
//
// No Node.js seria algo como:
//   class UserRepository {
//     constructor(private db: Pool) {}
//   }
type userRepository struct {
	db *pgxpool.Pool
	// ↑ *pgxpool.Pool é um ponteiro para o pool de conexões.
	// Pool = conjunto de conexões abertas e reutilizáveis com o PostgreSQL.
	// Em vez de abrir/fechar uma conexão por request (lento),
	// o pool empresta uma conexão existente e devolve ao terminar.
}

// NewUserRepository é o CONSTRUTOR — padrão Go para criar structs.
// Recebe o pool de conexões e retorna a interface (não a struct concreta).
//
// Retornar domain.UserRepository (interface) em vez de *userRepository (struct concreta)
// é essencial: o service vai depender da interface, não da implementação.
// Amanhã você pode criar um MockUserRepository para testes
// e o service não precisa mudar uma linha.
//
// No Node.js seria:
//   export function createUserRepository(db: Pool): UserRepository {
//     return new UserRepository(db)
//   }
func NewUserRepository(db *pgxpool.Pool) domain.UserRepository {
	return &userRepository{db: db}
	// ↑ & cria um ponteiro para a struct.
	// {db: db} inicializa o campo db com o pool recebido.
	// É como o "new" do JS — aloca a struct na memória e retorna o endereço.
}

// Save insere um novo usuário no banco de dados.
//
// (r *userRepository) é o RECEIVER — é o "this" do Go.
// "r" é o apelido escolhido para acessar a struct dentro da função.
// Convenção Go: usar a inicial minúscula do nome da struct.
//
// No Node.js seria:
//   async save(user: User): Promise<void> {
//     await db.query(
//       'INSERT INTO users (id, name, email, password_hash) VALUES ($1, $2, $3, $4)',
//       [user.id, user.name, user.email, user.passwordHash]
//     )
//   }
func (r *userRepository) Save(user domain.User) error {
	_, err := r.db.Exec(
		// ↑ Exec executa SQL que não retorna linhas (INSERT, UPDATE, DELETE).
		// _ descarta o primeiro retorno (resultado da execução — não precisamos).
		// err guarda o erro, se houver.

		context.Background(),
		// ↑ context.Background() é um contexto vazio, sem timeout.
		// Em produção você usaria context.WithTimeout() para cancelar queries lentas.
		// É padrão passar o context como primeiro argumento em Go.

		`INSERT INTO users (id, name, email, password_hash)
		 VALUES ($1, $2, $3, $4)`,
		// ↑ $1, $2, $3, $4 são parâmetros — o pgx substitui com segurança.
		// NUNCA concatene strings SQL — isso causa SQL Injection.
		// No Node com pg você usa $1/$2 igual. Com mysql seria ?.

		user.ID, user.Name, user.Email, user.PasswordHash,
		// ↑ os valores que substituem $1, $2, $3, $4 — na mesma ordem.
	)
	return err
	// ↑ Se Exec funcionou, err é nil — retornar nil = "deu tudo certo".
	// Se falhou (email duplicado, banco fora do ar...), err tem a descrição do problema.
}

// FindByEmail busca um usuário pelo email.
// Retorna (*User, error) — dois valores, padrão Go.
//
// No Node.js seria:
//   async findByEmail(email: string): Promise<User | null> {
//     const result = await db.query('SELECT...', [email])
//     return result.rows[0] || null
//   }
func (r *userRepository) FindByEmail(email string) (*domain.User, error) {
	var user domain.User
	// ↑ declara uma variável local do tipo User com valores zerados.
	// string zerada = "", int zerado = 0.
	// Vamos preencher os campos com Scan() logo abaixo.

	err := r.db.QueryRow(
		// ↑ QueryRow busca no máximo UMA linha.
		// Use Query() quando pode retornar múltiplas linhas.

		context.Background(),
		`SELECT id, name, email, password_hash
		 FROM users
		 WHERE email = $1`,
		// ↑ buscamos o password_hash também — vai ser necessário no login futuramente.

		email,
	).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash)
	// ↑ Scan lê as colunas do resultado e escreve nos campos da struct.
	// & passa o ENDEREÇO de cada campo — Scan precisa do endereço para escrever.
	// A ordem DEVE bater com o SELECT:
	//   id → &user.ID | name → &user.Name | email → &user.Email | password_hash → &user.PasswordHash

	if err != nil && err.Error() == "no rows in result set" {
		// ↑ Quando nenhuma linha é encontrada, pgx retorna esse erro específico.
		// Não é um erro real — é só "email não existe no banco".
		// Retornamos nil, nil: "sem usuário, sem erro".
		return nil, nil
	}

	if err != nil {
		// ↑ Qualquer outro erro é problema real: banco fora do ar, SQL inválido, etc.
		// Retornamos nil pro usuário e o erro para quem chamou tratar.
		return nil, err
	}

	return &user, nil
	// ↑ & retorna o ENDEREÇO da struct local — um ponteiro.
	// nil no segundo valor = sem erro.
	// O Go garante que a struct não some da memória mesmo após a função terminar.
}
