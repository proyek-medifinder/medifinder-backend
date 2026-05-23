package helper

import (
	"context"
	"os"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

func UploadToCloudinary(file interface{}) (string, error) {
	// 1. Cek dulu semua env var ada isinya
	cloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
	apiKey := os.Getenv("CLOUDINARY_API_KEY")
	apiSecret := os.Getenv("CLOUDINARY_API_SECRET")

	// 2. Inisialisasi Cloudinary dengan error handling
	cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		return "", err // Kalau gagal inisialisasi, balikin error-nya
	}

	ctx := context.Background()

	// 3. Upload dengan parameter folder
	resp, err := cld.Upload.Upload(ctx, file, uploader.UploadParams{
		Folder: "medifinder_assets", // Nama folder di Cloudinary
	})
	if err != nil {
		return "", err
	}

	return resp.SecureURL, nil
}
