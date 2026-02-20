package dto

type RegisterRequest struct {
	Name     string `json:"name" example:"Budi"`
	Email    string `json:"email" example:"budi@email.com"`
	Password string `json:"password" example:"password123"`
}

type LoginRequest struct {
	Email    string `json:"email" example:"budi@email.com"`
	Password string `json:"password" example:"password123"`
}

type AuthResponse struct {
	Token string `json:"token" example:"jwt_token_here"`
	Name  string `json:"name" example:"Budi"`
	Role  string `json:"role" example:"user"`
}
