package handlers

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/big"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/theo-guerra/simple-shop/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// Chave para guardar o ID do usuário na memória do Go (Contexto)
type contextKey string

const CtxKeyUsuarioID contextKey = "usuario_id"

// Memória temporária para guardar os códigos de recuperação (Simulando uma tabela de tokens)
var codigosRecuperacao = make(map[string]string)

// jwtKey é configurada em tempo de inicialização via SetJWTSecret
var jwtKey []byte

// isProduction controla flags de segurança (cookie Secure, etc.)
var isProduction bool

// SetJWTSecret configura a chave JWT a partir do config (chamado no main)
func SetJWTSecret(secret string) {
	jwtKey = []byte(secret)
}

// SetProduction configura o modo de produção
func SetProduction(prod bool) {
	isProduction = prod
}

type AuthHandler struct {
	DB *sql.DB
}

// Claims: O conteúdo do nosso Token JWT
type Claims struct {
	Email     string `json:"email"`
	UsuarioID int    `json:"usuario_id"`
	jwt.RegisteredClaims
}

// --- UTILITÁRIOS DE SEGURANÇA ---

func gerarTokenJWT(email string, usuarioID int) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Email:     email,
		UsuarioID: usuarioID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

func validarTokenJWT(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}
	return claims, nil
}

func gerarCodigoPIN() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(900000))
	return fmt.Sprintf("%06d", n.Int64()+100000)
}

// --- HANDLERS PRINCIPAIS ---

// 🔒 LOGIN (REVISADO: Agora compara senhas criptografadas)
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var creds models.Usuario
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Requisição inválida", http.StatusBadRequest)
		return
	}

	if msg := validarEmail(creds.Email); msg != "" {
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	if creds.Senha == "" {
		http.Error(w, "Senha é obrigatória", http.StatusBadRequest)
		return
	}

	var usuarioID int
	var senhaHash string

	// Busca o usuário no banco
	err := h.DB.QueryRow("SELECT id, senha FROM usuarios WHERE email = $1", creds.Email).Scan(&usuarioID, &senhaHash)
	if err != nil {
		// Por segurança, não dizemos se o e-mail existe ou não, apenas "erro nas credenciais"
		http.Error(w, "E-mail ou senha incorretos", http.StatusUnauthorized)
		return
	}

	// 🛡️ COMPARAÇÃO SEGURA: Verifica se a senha digitada bate com o Hash do banco
	err = bcrypt.CompareHashAndPassword([]byte(senhaHash), []byte(creds.Senha))
	if err != nil {
		http.Error(w, "E-mail ou senha incorretos", http.StatusUnauthorized)
		return
	}

	tokenString, err := gerarTokenJWT(creds.Email, usuarioID)
	if err != nil {
		http.Error(w, "Erro interno", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    tokenString,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   isProduction,
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
	})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"mensagem": "Acesso autorizado"})
}

// 🆕 CADASTRO DE LOJISTA
func (h *AuthHandler) Cadastro(w http.ResponseWriter, r *http.Request) {
	var creds models.Usuario
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Dados inválidos", http.StatusBadRequest)
		return
	}

	if msg := validarEmail(creds.Email); msg != "" {
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	if msg := validarSenha(creds.Senha); msg != "" {
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	// Verifica duplicidade
	var id int
	err := h.DB.QueryRow("SELECT id FROM usuarios WHERE email = $1", creds.Email).Scan(&id)
	if err == nil {
		http.Error(w, "Este e-mail já está cadastrado.", http.StatusConflict)
		return
	}

	// Criptografa a senha antes de salvar
	hashedSenha, err := bcrypt.GenerateFromPassword([]byte(creds.Senha), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Erro ao processar segurança", http.StatusInternalServerError)
		return
	}

	_, err = h.DB.Exec("INSERT INTO usuarios (email, senha) VALUES ($1, $2)", creds.Email, string(hashedSenha))
	if err != nil {
		http.Error(w, "Erro ao salvar no banco", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// 🚪 LOGOUT
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:    "auth_token",
		Value:   "",
		Expires: time.Unix(0, 0),
		Path:    "/",
	})
	w.WriteHeader(http.StatusOK)
}

// --- FLUXO DE RECUPERAÇÃO DE SENHA (OTP) ---

func (h *AuthHandler) SolicitarRecuperacao(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Dados inválidos", http.StatusBadRequest)
		return
	}

	var id int
	err := h.DB.QueryRow("SELECT id FROM usuarios WHERE email = $1", req.Email).Scan(&id)
	if err != nil {
		http.Error(w, "E-mail não encontrado.", http.StatusNotFound)
		return
	}

	codigo := gerarCodigoPIN()
	codigosRecuperacao[req.Email] = codigo

	// Simulação de envio por e-mail (trocar por serviço real em produção)
	slog.Info("Código de recuperação gerado", "email", req.Email, "codigo", codigo)

	w.WriteHeader(http.StatusOK)
}

func (h *AuthHandler) ValidarRecuperacao(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email     string `json:"email"`
		Codigo    string `json:"codigo"`
		NovaSenha string `json:"nova_senha"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Dados inválidos", http.StatusBadRequest)
		return
	}

	codigoSalvo, existe := codigosRecuperacao[req.Email]
	if !existe || codigoSalvo != req.Codigo {
		http.Error(w, "Código inválido ou expirado.", http.StatusUnauthorized)
		return
	}

	hashedSenha, _ := bcrypt.GenerateFromPassword([]byte(req.NovaSenha), bcrypt.DefaultCost)
	_, err := h.DB.Exec("UPDATE usuarios SET senha = $1 WHERE email = $2", string(hashedSenha), req.Email)
	if err != nil {
		http.Error(w, "Erro ao atualizar senha", http.StatusInternalServerError)
		return
	}

	delete(codigosRecuperacao, req.Email)
	w.WriteHeader(http.StatusOK)
}

// --- MIDDLEWARE E HELPERS ---

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("auth_token")
		if err != nil {
			http.Error(w, "Faça login para continuar", http.StatusUnauthorized)
			return
		}

		claims, err := validarTokenJWT(cookie.Value)
		if err != nil || claims == nil {
			http.Error(w, "Sessão expirada", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), CtxKeyUsuarioID, claims.UsuarioID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UsuarioIDFromContext(ctx context.Context) (int, bool) {
	id, ok := ctx.Value(CtxKeyUsuarioID).(int)
	return id, ok
}