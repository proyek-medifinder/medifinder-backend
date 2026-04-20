package dto

type VerifyAdminRequest struct {
	AdminID string `json:"admin_id" binding:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
	Action  string `json:"action" binding:"required" example:"approved"`
	Notes   string `json:"notes" example:"Dokumen lengkap dan valid"`
}

type ChangeAdminStatusRequest struct {
	Status string `json:"status" binding:"required" example:"suspended"`
}

type AdminRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// DTO untuk Update Admin
type UpdateAdminRequest struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
}
