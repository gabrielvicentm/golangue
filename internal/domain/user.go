// package domain é o coração da aplicação.
// Ele não depende de ninguém — nem do banco, nem do gin, nem de nada externo.
// É aqui que vivem as "regras" e "formatos" dos dados.
//
// No Node.js, isso seria equivalente aos seus models/types do TypeScript.
package domain

// User representa um usuário como ele existe no banco de dados.
// Essa struct é o "molde" — define quais campos um usuário tem e seus tipos.
//
// No Node.js seria algo como:
//   interface User {
//     id: string
//     name: string
//     email: string
//     passwordHash: string  // nunca exposto na API
//   }
type User struct {
	ID           string `json:"id"`
	// ↑ json:"id" é uma struct tag — instrução pro encoder de JSON.
	// Sem ela, o JSON sairia com "ID" maiúsculo.
	// Com ela: {"id": "abc-123"} — segue o padrão camelCase/lowercase do JSON.

	Name         string `json:"name"`
	Email        string `json:"email"`

	PasswordHash string `json:"-"`
	// ↑ json:"-" é especial: diz pro encoder "NUNCA inclua esse campo no JSON".
	// Não importa o que aconteça, PasswordHash jamais vai aparecer na resposta da API.
	// No Node.js você faria isso manualmente com delete user.passwordHash antes de res.json()
	// Aqui o Go faz isso automaticamente pela tag.
}

// CreateUserInput representa os dados que o CLIENTE envia no corpo da requisição.
// É separado do User porque o cliente não manda o ID (gerado pelo servidor)
// e não manda o PasswordHash (gerado pelo servidor a partir da senha).
//
// No Node.js seria:
//   interface CreateUserInput {
//     name: string      // obrigatório
//     email: string     // obrigatório, formato email
//     password: string  // obrigatório, mínimo 8 caracteres
//   }
type CreateUserInput struct {
	Name     string `json:"name"     binding:"required"`
	// ↑ binding:"required" é uma tag do Gin.
	// Se o campo vier vazio ou ausente, o Gin rejeita a requisição
	// e devolve 400 automaticamente — sem precisar validar na mão.

	Email    string `json:"email"    binding:"required,email"`
	// ↑ binding:"required,email" — duas regras separadas por vírgula:
	// 1. required: campo obrigatório
	// 2. email: valida o formato (precisa ter @, domínio, etc.)
	// "claudio" → rejeitado | "claudio@email.com" → aceito

	Password string `json:"password" binding:"required,min=8"`
	// ↑ min=8 exige mínimo 8 caracteres.
	// "123" → rejeitado | "minhasenha" → aceito
}

// UserRepository é uma INTERFACE — define um contrato:
// "quem quiser ser um UserRepository PRECISA ter esses dois métodos".
//
// A interface fica aqui no domain (e não no repository) porque
// o service precisa depender de uma abstração, não de uma implementação concreta.
// Isso permite trocar o banco de dados sem tocar no service.
//
// No Node.js com TypeScript seria:
//   interface UserRepository {
//     save(user: User): Promise<void>
//     findByEmail(email: string): Promise<User | null>
//   }
type UserRepository interface {
	Save(user User) error
	// ↑ recebe um User completo (com hash já gerado) e salva no banco.
	// Retorna error — se der certo, retorna nil (equivalente ao null/undefined do JS).

	FindByEmail(email string) (*User, error)
	// ↑ *User é um PONTEIRO para User — pode ser nil.
	// nil aqui significa "não encontrei nenhum usuário com esse email".
	// Retorna dois valores: o usuário (ou nil) E um possível erro.
	// Combinações possíveis:
	//   (*User, nil)  → encontrou, sem erro
	//   (nil, nil)    → não encontrou, sem erro — email não existe no banco
	//   (nil, error)  → problema no banco (conexão, query errada, etc.)
}
