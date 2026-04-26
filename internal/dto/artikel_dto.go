package dto

type CreateArtikelRequest struct {
	Judul    string `form:"judul" binding:"required"`
	Konten   string `form:"konten" binding:"required"`
	Kategori string `form:"kategori" binding:"required"`
	Status   string `form:"status" binding:"required,oneof=DRAFT PUBLISHED"`
}

type UpdateArtikelRequest struct {
	Judul    string `form:"judul" binding:"required"`
	Konten   string `form:"konten" binding:"required"`
	Kategori string `form:"kategori" binding:"required"`
	Status   string `form:"status" binding:"required,oneof=DRAFT PUBLISHED"`
}
