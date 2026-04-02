-- Migration: criação da tabela de usuários
-- Executar manualmente no PostgreSQL antes de subir a API
-- No Node.js com ORM (Prisma, Sequelize) isso seria gerado automaticamente.
-- Em Go sem ORM, você escreve o SQL na mão — mais controle, mais clareza.

CREATE TABLE IF NOT EXISTS users (
    id            TEXT        PRIMARY KEY,
    name          TEXT        NOT NULL,
    email         TEXT        UNIQUE NOT NULL,
    password_hash TEXT        NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
