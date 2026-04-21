package routes

import (
	"log"
	"time"

	"github.com/gin-contrib/cors"
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

	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:3000",
			"http://127.0.0.1:3000",
			"http://localhost:5173",
			"https://nonregressive-kyoko-supercelestially.ngrok-free.dev",
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "ngrok-skip-browser-warning"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

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
	r.POST("/register-admin", authHandler.RegisterAdmin)
	r.POST("/login", authHandler.Login)
	r.POST("/google-login", authHandler.GoogleLogin)
	// ++++++++++++++++++++++++++++++++++++++++++
	r.POST("/forgot-password", authHandler.ForgotPassword)
	r.POST("/reset-password", authHandler.ResetPassword)
	// ++++++++++++++++++++++++++++++++++++++++++
	r.GET("/apotek/:id/obat", obatHandler.GetByApotekPublic)
	r.GET("/apotek/nearby", apotekHandler.SearchNearby)
	// ++++++++++++++++++++++++++++++++++++++++++
	r.POST("/payment/notify", paymentHandler.Notification)
	// +++++++++++++++++++++++++++++++++++++++++
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// ================= USER =================
	userGroup := r.Group("/")
	userGroup.Use(middleware.AuthMiddleware())
	userGroup.Use(middleware.Authorize("user"))
	{
		// ++++++++++ CART +++++++++++
		userGroup.POST("/cart/items", cartHandler.AddToCart)
		userGroup.GET("/cart", cartHandler.GetCart)
		userGroup.PUT("/cart/items/:id", cartHandler.UpdateItem)
		userGroup.DELETE("/cart/items/:id", cartHandler.DeleteItem)
		userGroup.POST("/cart/checkout", cartHandler.Checkout)
		// ++++++++++ TRANSAKSI +++++++++++
		userGroup.GET("/transaksi", transaksiHandler.UserHistory)
		userGroup.GET("/transaksi/:id", transaksiHandler.Detail)
		// ++++++++++ RESEP +++++++++++
		userGroup.POST("/resep", resepHandler.Upload)
	}

	// ================= ADMIN =================
	adminGroup := r.Group("/admin")
	adminGroup.Use(middleware.AuthMiddleware())
	adminGroup.Use(middleware.Authorize("admin_apotek"))
	{
		// +++++++++++ OBAT +++++++++++
		adminGroup.POST("/obat", obatHandler.Create)
		adminGroup.GET("/obat", obatHandler.GetMyObat)
		adminGroup.PUT("/obat/:id", obatHandler.Update)
		adminGroup.DELETE("/obat/:id", obatHandler.Delete)
		// ++++++++++ APOTEK +++++++++++
		adminGroup.GET("/apotek", apotekHandler.GetMyApotek)
		adminGroup.PUT("/apotek", apotekHandler.UpdateMyApotek)
		// ++++++++++ TRANSAKSI +++++++++++
		adminGroup.GET("/transaksi", transaksiHandler.AdminHistory)
		// ++++++++++ RESEP +++++++++++
		adminGroup.PUT("/resep/:id", resepHandler.UpdateStatus)
		adminGroup.GET("/resep", resepHandler.List)
	}

	// ================= SUPER ADMIN =================
	superAdminGroup := r.Group("/superadmin")
	superAdminGroup.Use(middleware.AuthMiddleware())
	superAdminGroup.Use(middleware.Authorize("super_admin"))
	{
		// ++++++++++ TRANSAKSI +++++++++++
		superAdminGroup.GET("/transaksi", transaksiHandler.SuperAdminHistory)
		// ++++++++++ ADMIN MANAGEMENT +++++++++++
		superAdminGroup.GET("/admin", superAdminHandler.ListAdmin)
		superAdminGroup.POST("/admin", superAdminHandler.CreateAdmin)
		superAdminGroup.PUT("/admin/:id", superAdminHandler.UpdateAdmin)
		superAdminGroup.DELETE("/admin/:id", superAdminHandler.DeleteAdmin)
		superAdminGroup.PATCH("/admin/:id/status", superAdminHandler.ChangeAdminStatus)
		// +++++++++ VERIFIKASI ADMIN +++++++++++
		superAdminGroup.GET("/pengajuan", superAdminHandler.GetPendingAdmins)
		superAdminGroup.POST("/verifikasi", superAdminHandler.VerifyAdmin)
	}

	// ================= AUTH PROTECTED (SEMUA ROLE) =================
	protectedAuth := r.Group("/")
	protectedAuth.Use(middleware.AuthMiddleware()) // Wajib masukin Token JWT!
	{
		protectedAuth.POST("/change-password", authHandler.ChangePassword)
		protectedAuth.GET("/me", authHandler.GetMe)
	}

	return r
}
