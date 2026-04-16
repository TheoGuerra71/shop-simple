package handlers

import (
	"fmt"
	"strings"
)

// validarCampoObrigatorio verifica se um campo string não está vazio
func validarCampoObrigatorio(campo, nome string) string {
	if strings.TrimSpace(campo) == "" {
		return "Campo '" + nome + "' é obrigatório"
	}
	return ""
}

// validarTamanhoMax verifica se um campo não excede o tamanho máximo
func validarTamanhoMax(campo, nome string, max int) string {
	if len(campo) > max {
		return fmt.Sprintf("Campo '%s' excede o tamanho máximo de %d caracteres", nome, max)
	}
	return ""
}

// validarEmail verifica formato mínimo de email
func validarEmail(email string) string {
	email = strings.TrimSpace(email)
	if email == "" {
		return "E-mail é obrigatório"
	}
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return "E-mail inválido"
	}
	if len(email) > 255 {
		return "E-mail muito longo"
	}
	return ""
}

// validarSenha verifica requisitos mínimos da senha
func validarSenha(senha string) string {
	if len(senha) < 6 {
		return "Senha deve ter no mínimo 6 caracteres"
	}
	if len(senha) > 72 { // limite do bcrypt
		return "Senha muito longa"
	}
	return ""
}
