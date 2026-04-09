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
	Email string `json:"email"`
	Role  string `json:"role" example:"user"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type GoogleLoginRequest struct {
	Token string `json:"token" binding:"required"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"` // min=6 artinya minimal 6 karakter
}

type RegisterAdminRequest struct {
	Name        string  `json:"name" binding:"required"`
	Email       string  `json:"email" binding:"required,email"`
	Password    string  `json:"password" binding:"required,min=6"`
	NamaApotek  string  `json:"nama_apotek" binding:"required"`
	Alamat      string  `json:"alamat" binding:"required"`
	Latitude    float64 `json:"latitude" binding:"required"`
	Longitude   float64 `json:"longitude" binding:"required"`
	PhoneNumber string  `json:"phone_number" binding:"required"`
	Deskripsi   string  `json:"deskripsi"`
}

type VerifyAdminRequest struct {
	AdminID string `json:"admin_id" binding:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
	Action  string `json:"action" binding:"required" example:"approved"` // approved atau rejected
	Notes   string `json:"notes" example:"Dokumen apotek valid dan lengkap"`
}
