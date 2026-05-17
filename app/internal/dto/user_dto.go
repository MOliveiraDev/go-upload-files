package dto

type CreateUserRequest struct {
	Name     string `json:"name" validate:"required,min=3,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type CreateUserResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type LoginResponse struct {
	JWT string `json:"jwt"`
}

func (req *CreateUserRequest) ValidateRequest() error {
	return Validate.Struct(req)
}

func (req *LoginRequest) ValidateRequest() error {
	return Validate.Struct(req)
}
