package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq" // Driver oficial para PostgreSQL
)

// Conectar abre e valida a conexão com o banco de dados.
func Conectar(connStr string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Ping valida se as credenciais e o banco estão realmente ativos.
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("conexão falhou: %v", err)
	}

	return db, nil
}
