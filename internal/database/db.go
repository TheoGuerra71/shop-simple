package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq" // Driver oficial para PostgreSQL
)

// Conectar abre, configura o pool e valida a conexão com o banco de dados.
func Conectar(connStr string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Pool de conexões — evita estourar o PostgreSQL sob carga
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("conexão falhou: %v", err)
	}

	return db, nil
}
