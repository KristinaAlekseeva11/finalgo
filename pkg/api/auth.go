package api

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func pass() string { return os.Getenv("TODO_PASSWORD") }

func hashPass(p string) string {
	sum := sha256.Sum256([]byte(p))
	return hex.EncodeToString(sum[:])
}

// генерим токен: подпись = пароль, claim "ph" = хэш пароля, exp = 8 часов
func makeTokenFromPassword(p string) (string, error) {
	h := hashPass(p)
	claims := jwt.MapClaims{
		"ph": h,
		"exp": time.Now().Add(8 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(p))
}

func validateTokenWithPassword(tok, p string) bool {
	if tok == "" || p == "" {
		return false
	}
	parsed, err := jwt.Parse(tok, func(t *jwt.Token) (any, error) {
		// HS256
		return []byte(p), nil
	})
	if err != nil || !parsed.Valid {
		return false
	}
	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return false
	}
	// сверяем хэш текущего пароля и хэш из токена
	cur := hashPass(p)
	tokHash, _ := claims["ph"].(string)
	return tokHash == cur
}

// middleware: если пароль задан — требуем валидную куку token
func auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p := pass()
		if p == "" {
			next(w, r)
			return
		}
		c, err := r.Cookie("token")
		if err != nil || !validateTokenWithPassword(c.Value, p) {
			http.Error(w, "Authentification required", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}
