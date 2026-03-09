package models

// LojaConfig armazena os dados principais do seu negócio.
// As tags json são essenciais para conversar com o front-end depois.
type LojaConfig struct {
	NomeLoja  string `json:"nome_loja"`
	Whatsapp  string `json:"whatsapp"`
	Instagram string `json:"instagram"`
}

// Tarefa representa um item na sua To-Do list de lojista.
type Tarefa struct {
	ID        int    `json:"id"`
	Descricao string `json:"descricao"`
	Concluida bool   `json:"concluida"`
}
