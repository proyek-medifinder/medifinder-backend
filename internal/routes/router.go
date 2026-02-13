package routes

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/sasaefulanwar/medifinder/internal/handler"
	"github.com/sasaefulanwar/medifinder/internal/middleware"
	"github.com/sasaefulanwar/medifinder/internal/repository"
	"github.com/sasaefulanwar/medifinder/internal/service"
)

func SetupRouter(db *sqlx.DB) *gin.Engine {

	r := gin.Default()

	// ================= REPOSITORY =================
	userRepo := &repository.UserRepository{DB: db}
	apotekRepo := &repository.ApotekRepository{DB: db}
	obatRepo := &repository.ObatRepository{DB: db}
	cartRepo := &repository.CartRepository{DB: db}
	transaksiRepo := &repository.TransaksiRepository{DB: db}
	resepRepo := &repository.ResepRepository{DB: db}

	// ================= SERVICE =================
	authService := &service.AuthService{UserRepo: userRepo}
	apotekService := &service.ApotekService{Repo: apotekRepo}
	obatService := &service.ObatService{
		ObatRepo:   obatRepo,
		ApotekRepo: apotekRepo,
	}
	cartService := &service.CartService{
		CartRepo: cartRepo,
		ObatRepo: obatRepo,
	}
	transaksiService := &service.TransaksiService{
		DB:         db,
		Repo:       transaksiRepo,
		ApotekRepo: apotekRepo,
	}
	paymentService := &service.PaymentService{DB: db}
	resepService := &service.ResepService{
		Repo: resepRepo,
	}
	superAdminService := &service.SuperAdminService{
		UserRepo: userRepo,
	}

	// ================= HANDLER =================
	authHandler := &handler.AuthHandler{Service: authService}
	apotekHandler := &handler.ApotekHandler{Service: apotekService}
	obatHandler := &handler.ObatHandler{Service: obatService}
	cartHandler := &handler.CartHandler{Service: cartService}
	transaksiHandler := &handler.TransaksiHandler{Service: transaksiService}
	paymentHandler := &handler.PaymentHandler{Service: paymentService}
	resepHandler := &handler.ResepHandler{
		Service: resepService,
	}
	superAdminHandler := &handler.SuperAdminHandler{
		Service: superAdminService,
	}

	// ================= BACKGROUND JOB =================
	go func() {
		for {
			err := transaksiService.CancelExpiredTransactions()
			if err != nil {
				log.Println("Auto cancel error:", err)
			}
			time.Sleep(1 * time.Minute)
		}
	}()

	// ================= PUBLIC =================
	r.POST("/register", authHandler.Register)
	r.POST("/login", authHandler.Login)
	r.GET("/apotek/:id/obat", obatHandler.GetByApotekPublic)
	r.POST("/payment/notify", paymentHandler.Notification)
	r.GET("/apotek", apotekHandler.SearchNearby)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// ================= USER =================
	userGroup := r.Group("/")
	userGroup.Use(middleware.AuthMiddleware())
	userGroup.Use(middleware.Authorize("user"))
	{
		userGroup.POST("/cart/items", cartHandler.AddToCart)
		userGroup.GET("/cart", cartHandler.GetCart)
		userGroup.PUT("/cart/items/:id", cartHandler.UpdateItem)
		userGroup.DELETE("/cart/items/:id", cartHandler.DeleteItem)
		userGroup.POST("/cart/checkout", cartHandler.Checkout)

		userGroup.GET("/transaksi", transaksiHandler.UserHistory)
		userGroup.GET("/transaksi/:id", transaksiHandler.Detail)

		userGroup.POST("/resep", resepHandler.Upload)
	}

	// ================= ADMIN =================
	adminGroup := r.Group("/admin")
	adminGroup.Use(middleware.AuthMiddleware())
	adminGroup.Use(middleware.Authorize("admin_apotek"))
	{
		adminGroup.POST("/obat", obatHandler.Create)
		adminGroup.POST("/apotek", apotekHandler.Create)
		adminGroup.GET("/apotek", apotekHandler.GetMyApotek)
		adminGroup.GET("/transaksi", transaksiHandler.AdminHistory)
		adminGroup.PUT("/resep/:id", resepHandler.UpdateStatus)
		adminGroup.GET("/resep", resepHandler.List)
	}

	// ================= SUPER ADMIN =================
	superAdminGroup := r.Group("/superadmin")
	superAdminGroup.Use(middleware.AuthMiddleware())
	superAdminGroup.Use(middleware.Authorize("super_admin"))
	{
		superAdminGroup.GET("/transaksi", transaksiHandler.SuperAdminHistory)

		superAdminGroup.GET("/admin", superAdminHandler.ListAdmin)
		superAdminGroup.POST("/admin", superAdminHandler.CreateAdmin)
		superAdminGroup.PUT("/admin/:id", superAdminHandler.UpdateAdmin)
		superAdminGroup.DELETE("/admin/:id", superAdminHandler.DeleteAdmin)
	}

	return r
}
