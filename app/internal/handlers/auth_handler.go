package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/MOliveiraDev/go-upload-files/internal/dto"
	"github.com/MOliveiraDev/go-upload-files/internal/middleware"
	"github.com/MOliveiraDev/go-upload-files/internal/services"
)

type AuthHandler struct {
	userService *services.UserService
}

func NewAuthHandler(userService *services.UserService) *AuthHandler {
	return &AuthHandler{userService: userService}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) error {
	if h.userService == nil {
		return middleware.NewInternalError("O serviço de usuário não está configurado", errors.New("service de usuário nulo"))
	}

	var req dto.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return middleware.NewBadRequestError("corpo da requisição inválido", err)
	}

	user, err := h.userService.RegisterUser(r.Context(), &req)
	if err != nil {
		if errors.Is(err, services.ErrUserEmailAlreadyExists) {
			return middleware.NewConflictError("email já cadastrado", err)
		}

		return middleware.NewBadRequestError("falha ao registrar usuário", err)
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"data": dto.CreateUserResponse{
			ID:    user.ID.String(),
			Name:  user.Name,
			Email: user.Email,
		},
	})

	return nil
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) error {
	if h.userService == nil {
		return middleware.NewInternalError("O serviço de usuário não está configurado", errors.New("nil user service"))
	}

	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return middleware.NewBadRequestError("corpo da requisição inválido", err)
	}

	response, token, err := h.userService.LoginUser(r.Context(), &req)
	if err != nil {
		return middleware.NewUnauthorizedError("email ou senha inválidos", err)
	}

	middleware.SetAuthCookie(w, token)
	writeJSON(w, http.StatusOK, map[string]any{
		"data": response,
	})

	return nil
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) error {
	middleware.ClearAuthCookie(w)
	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]string{
			"message": "logout feito com sucesso",
		},
	})

	return nil
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
