// package config centraliza tudo relacionado à configuração da aplicação:
// leitura do .env e conexão com o banco de dados.
// Mantido separado para não poluir o main.go.
package config

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

// LoadEnv lê o arquivo .env e carrega as variáveis no ambiente do processo.
// Após chamar essa função, os.Getenv("DB_HOST") funciona em qualquer lugar do código.
//
// No Node.js seria:
//   import 'dotenv/config'  // ou require('dotenv').config()
func LoadEnv() {
	if err := godotenv.Load("../../.env"); err != nil {
		// ↑ godotenv.Load() procura o arquivo .env na pasta atual
		// e chama os.Setenv() para cada linha: DB_HOST=localhost → os.Setenv("DB_HOST", "localhost")

		log.Fatal("Erro ao carregar .env: ", err)
		// ↑ log.Fatal = imprime a mensagem + chama os.Exit(1).
		// Mata o processo imediatamente — sem .env não tem como continuar.
		// Equivalente ao process.exit(1) do Node.
	}
}

// NewDBConnection cria e retorna um pool de conexões com o PostgreSQL.
// Usa as variáveis de ambiente carregadas pelo LoadEnv().
//
// Retorna *pgxpool.Pool (ponteiro) — todas as camadas compartilham
// o MESMO pool, não criam conexões novas a cada request.
func NewDBConnection() *pgxpool.Pool {

	// Monta a string de conexão (DSN = Data Source Name) com os dados do .env.
	// Formato: postgres://usuario:senha@host:porta/banco
	//
	// No Node.js com pg seria:
	//   const pool = new Pool({
	//     host: process.env.DB_HOST,
	//     port: process.env.DB_PORT,
	//     ...
	//   })
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		// ↑ fmt.Sprintf funciona como template string do JS:
		// `postgres://${user}:${pass}@${host}:${port}/${db}`
		// Cada %s é substituído pelo argumento correspondente abaixo.

		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
		// ↑ os.Getenv lê variável de ambiente — equivalente ao process.env.DB_USER do Node.
		// Resultado: "postgres://postgres:senha123@localhost:5432/userapi"
	)

	pool, err := pgxpool.New(context.Background(), dsn)
	// ↑ pgxpool.New cria o pool de conexões com o PostgreSQL.
	// Não abre uma conexão agora — abre sob demanda quando a primeira query rodar.
	//
	// Pool de conexões = conjunto de conexões abertas e reutilizáveis.
	// Imagine 10 conexões abertas esperando. Cada request pega uma, usa e devolve.
	// Muito mais eficiente do que abrir/fechar uma conexão por request.
	//
	// context.Background() = contexto raiz, sem timeout nem cancelamento.

	if err != nil {
		log.Fatal("Erro ao conectar no banco: ", err)
		// ↑ Se não conseguiu criar o pool, algo está errado (DSN inválido, banco inacessível).
		// Mata o processo — não adianta subir a API sem banco.
	}

	return pool
	// ↑ Retorna o ponteiro pro pool.
	// main.go vai passar esse mesmo pool para o repository.
}
