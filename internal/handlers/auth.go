package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/theo-guerra/simple-shop/internal/models"
)

// 1. A ESTRUTURA DO HANDLER (Resolveu o "undefined: AuthHandler")
type AuthHandler struct {
	DB *sql.DB
}

// 2. CHAVE DE SEGURANÇA E ESTRUTURA DO TOKEN (Resolveu o "undefined: auth")
// Em produção, isso ficaria num arquivo .env
var jwtKey = []byte("sua_chave_secreta_boutique_2026")

type Claims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

// --- FUNÇÕES INTERNAS DE JWT ---

func gerarTokenJWT(email string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

func validarTokenJWT(tokenString string) (bool, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		return false, err
	}
	return true, nil
}

// --- ROTAS DA API ---

// Login: Recebe as credenciais e devolve um Cookie Seguro
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var creds models.Usuario
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Requisição inválida", http.StatusBadRequest)
		return
	}

	// Aqui entraria a verificação da senha no Banco de Dados.
	// Para não travar seu teste agora, estamos autorizando se o email for preenchido.
	if creds.Email == "" {
		http.Error(w, "Credenciais inválidas", http.StatusUnauthorized)
		return
	}

	// Gera o Token JWT
	tokenString, err := gerarTokenJWT(creds.Email)
	if err != nil {
		http.Error(w, "Erro interno ao gerar credencial", http.StatusInternalServerError)
		return
	}

	// 🛡️ A MÁGICA: Injeta o Cookie HttpOnly no navegador do usuário
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    tokenString,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,  // Impede que ataques de JavaScript (XSS) roubem o token
		Secure:   false, // Coloque 'true' quando subir para a nuvem com HTTPS
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
	})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"mensagem": "Acesso autorizado"})
}

// Logout: Mata o Cookie do navegador
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Expires:  time.Unix(0, 0), // Data no passado mata o cookie instantaneamente
		HttpOnly: true,
		Path:     "/",
	})
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"mensagem": "Sessão encerrada"})
}

// --- MIDDLEWARE DE PROTEÇÃO ---

// AuthMiddleware: Tranca as rotas da API exigindo o Cookie
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Extrai o token do Cookie (que o navegador manda automaticamente)
		cookie, err := r.Cookie("auth_token")
		if err != nil {
			http.Error(w, "Acesso Negado: Cookie não encontrado", http.StatusUnauthorized)
			return
		}

		// Valida a assinatura do Token
		valido, err := validarTokenJWT(cookie.Value)
		if err != nil || !valido {
			http.Error(w, "Acesso Negado: Token inválido ou expirado", http.StatusUnauthorized)
			return
		}

		// Tudo certo, libera o tráfego
		next.ServeHTTP(w, r)
	})
}
