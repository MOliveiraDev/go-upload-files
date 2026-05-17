package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	AuthCookieName = "auth_token"
	InvalidAuthenticationTokenMessage = "invalid authentication token"
)

type contextKey string

const (
	userIDContextKey    contextKey = "auth.user_id"
	userEmailContextKey contextKey = "auth.user_email"
)

func GenerateToken(userID uuid.UUID, email string) (string, error) {
	claims := jwt.MapClaims{
		"sub":   userID.String(),
		"email": email,
		"exp":   time.Now().Add(time.Hour * 72).Unix(),
		"iat":   time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		return "", fmt.Errorf("chave secreta JWT não encontrada")
	}

	return token.SignedString([]byte(secretKey))
}

func SetAuthCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     AuthCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   shouldUseSecureCookie(),
		MaxAge:   int((72 * time.Hour).Seconds()),
	})
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString, err := tokenFromRequestCookie(r)
		if err != nil {
			WriteAppError(w, r, NewUnauthorizedError(InvalidAuthenticationTokenMessage, err))
			return
		}

		claims, err := parseToken(tokenString)
		if err != nil {
			WriteAppError(w, r, NewUnauthorizedError(InvalidAuthenticationTokenMessage, err))
			return
		}

		userIDValue, ok := claims["sub"].(string)
		if !ok || userIDValue == "" {
			WriteAppError(w, r, NewUnauthorizedError(InvalidAuthenticationTokenMessage, errors.New("missing sub claim")))
			return
		}

		userID, err := uuid.Parse(userIDValue)
		if err != nil {
			WriteAppError(w, r, NewUnauthorizedError(InvalidAuthenticationTokenMessage, err))
			return
		}

		ctx := context.WithValue(r.Context(), userIDContextKey, userID)
		if email, ok := claims["email"].(string); ok && email != "" {
			ctx = context.WithValue(ctx, userEmailContextKey, email)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	if ctx == nil {
		return uuid.Nil, false
	}

	userID, ok := ctx.Value(userIDContextKey).(uuid.UUID)
	return userID, ok
}

func UserEmailFromContext(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}

	email, ok := ctx.Value(userEmailContextKey).(string)
	return email, ok
}

func tokenFromRequestCookie(r *http.Request) (string, error) {
	if r == nil {
		return "", errors.New("request is nil")
	}

	cookie, err := r.Cookie(AuthCookieName)
	if err != nil {
		return "", err
	}

	token := strings.TrimSpace(cookie.Value)
	if token == "" {
		return "", errors.New("empty auth cookie")
	}

	return token, nil
}

func parseToken(tokenString string) (jwt.MapClaims, error) {
	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		return nil, errors.New("jwt secret key not configured")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %s", token.Method.Alg())
		}

		return []byte(secretKey), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

func shouldUseSecureCookie() bool {
	appEnv := strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV")))
	if appEnv == "production" || appEnv == "staging" {
		return true
	}

	if secureCookie := strings.ToLower(strings.TrimSpace(os.Getenv("COOKIE_SECURE"))); secureCookie != "" {
		return secureCookie == "true" || secureCookie == "1"
	}

	return false
}
