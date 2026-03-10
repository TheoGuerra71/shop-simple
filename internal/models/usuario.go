package models

// Usuario representa as credenciais de quem tem acesso ao painel (Você).
type Usuario struct {
	Email string `json:"email"`
	Senha string `json:"senha"`
}
