// // @title Modeva CMS API
// // @version 1.0
// // @description Modeva CMS Backend API Documentation
// // @host localhost:8081
// // @BasePath /api/v1
// // @schemes http
// package main

// import (
// 	"fmt"
// 	"log"
// 	"os"
// 	"time"

// 	"github.com/Modeva-Ecommerce/modeva-cms-backend/controllers/cms/product_controller"
// 	_ "github.com/Modeva-Ecommerce/modeva-cms-backend/controllers/cms/product_controller"
// 	"github.com/Modeva-Ecommerce/modeva-cms-backend/routes/cms_routes"
// 	"github.com/Modeva-Ecommerce/modeva-cms-backend/routes/ecommerce_routes"

// 	"github.com/gin-gonic/gin"
// 	"github.com/joho/godotenv"
// 	swaggerFiles "github.com/swaggo/files"
// 	ginSwagger "github.com/swaggo/gin-swagger"

// 	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
// 	"github.com/Modeva-Ecommerce/modeva-cms-backend/services"

// 	_ "github.com/Modeva-Ecommerce/modeva-cms-backend/docs"
// 	"github.com/Modeva-Ecommerce/modeva-cms-backend/middleware"
// 	"github.com/gin-contrib/cors"
// )

// func init() {
// 	_ = godotenv.Load()
// }

// func CORSMiddleware() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		origin := c.Request.Header.Get("Origin")
// 		allowedOrigins := []string{
// 			"http://localhost:3000",
// 			"http://localhost:3001",
// 		}

// 		// Check if the origin is in the allowed list
// 		for _, allowedOrigin := range allowedOrigins {
// 			if origin == allowedOrigin {
// 				c.Header("Access-Control-Allow-Origin", origin)
// 				break
// 			}
// 		}

// 		c.Header("Access-Control-Allow-Credentials", "true")
// 		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
// 		c.Header("Access-Control-Allow-Methods", "POST,HEAD,PATCH, OPTIONS, GET, PUT")

// 		if c.Request.Method == "OPTIONS" {
// 			c.AbortWithStatus(204)
// 			return
// 		}

// 		c.Next()
// 	}
// }

// func main() {
// 	// Connect to DB
// 	config.InitDB()
// 	// Redis connection
// 	config.ConnectRedis()

// 	// Initialize Cloudinary service
// 	cloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
// 	apiKey := os.Getenv("CLOUDINARY_API_KEY")
// 	apiSecret := os.Getenv("CLOUDINARY_API_SECRET")

// 	if err := product_controller.InitCloudinary(cloudName, apiKey, apiSecret); err != nil {
// 		log.Fatalf("Failed to initialize Cloudinary: %v", err)
// 	}

// 	// ✅ Initialize JWT Service for Admin Auth
// 	jwtSecret := os.Getenv("JWT_SECRET")
// 	if jwtSecret == "" {
// 		log.Fatal("❌ JWT_SECRET environment variable not set")
// 	}
// 	if err := services.InitJWTService(jwtSecret); err != nil {
// 		log.Fatalf("Failed to initialize JWT service: %v", err)
// 	}
// 	log.Println("✅ JWT Service initialized")

// 	corsCfg := cors.Config{
// 		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:3001"},
// 		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
// 		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-CSRF-Token", "X-Requested-With"},
// 		AllowCredentials: true,
// 		MaxAge:           12 * time.Hour,
// 	}

// 	// ✅ Initialize Google OAuth
// 	config.InitGoogleOAuth()

// 	router := gin.Default()
// 	router.Use(cors.New(corsCfg))

// 	// Register API routes
// 	api := router.Group("/api/v1")

// 	// ✅ Setup Admin Routes (at /api/v1/admin prefix)
// 	cms_routes.SetupAdminRoutes(api)
// 	log.Println("✅ Admin routes registered")

// 	// Register CMS routes (at /api/v1/admin prefix)
// 	adminGroup := api.Group("/admin")
// 	adminGroup.Use(middleware.RateLimiter(100, time.Minute))

// 	cms_routes.SetupCategoryRoutes(adminGroup)
// 	cms_routes.SetupProductRoutes(adminGroup)
// 	cms_routes.SetupOrderRoutes(adminGroup)
// 	cms_routes.SetupCustomerRoutes(adminGroup)
// 	cms_routes.SetupAnalyticsRoutes(adminGroup)

// 	// Public storefront (no rate limiter)
// 	ecommerce_routes.SetupUserRoutes(api)
// 	ecommerce_routes.SetupAuthRoutes(api)
// 	ecommerce_routes.SetupStorefrontRoutes(api)

// 	// Swagger docs
// 	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

// 	fmt.Println("🚀 Server is running on http://localhost:8081")
// 	router.Run(":8081")
// }

// @title Modeva CMS API
// @version 1.0
// @description Modeva CMS Backend API Documentation
// @host localhost:8081
// @BasePath /api/v1
// @schemes http
package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/controllers/cms/product_controller"
	_ "github.com/Modeva-Ecommerce/modeva-cms-backend/controllers/cms/product_controller"
	_ "github.com/Modeva-Ecommerce/modeva-cms-backend/docs"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/middleware"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/routes/cms_routes"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/routes/ecommerce_routes"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/services"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func init() {
	_ = godotenv.Load()
}

func main() {
	// Connect to DB
	config.InitDB()
	// Redis connection
	config.ConnectRedis()
	// Initialize Cloudinary service
	cloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
	apiKey := os.Getenv("CLOUDINARY_API_KEY")
	apiSecret := os.Getenv("CLOUDINARY_API_SECRET")
	if err := product_controller.InitCloudinary(cloudName, apiKey, apiSecret); err != nil {
		log.Fatalf("Failed to initialize Cloudinary: %v", err)
	}

	// ✅ Initialize JWT Service for Admin Auth
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("❌ JWT_SECRET environment variable not set")
	}
	if err := services.InitJWTService(jwtSecret); err != nil {
		log.Fatalf("Failed to initialize JWT service: %v", err)
	}
	log.Println("✅ JWT Service initialized")

	// ✅ Configure CORS properly for all content types including PDFs
	corsCfg := cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:3001"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-CSRF-Token", "X-Requested-With"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
		ExposeHeaders:    []string{"Content-Disposition", "Content-Length"}, // Expose these headers for downloads
	}

	// ✅ Initialize Google OAuth
	config.InitGoogleOAuth()

	router := gin.Default()

	// ✅ Use ONLY the cors.New() middleware - single CORS config
	router.Use(cors.New(corsCfg))

	// Register API routes
	api := router.Group("/api/v1")

	// ✅ Setup Admin Routes (at /api/v1/admin prefix)
	cms_routes.SetupAdminRoutes(api)
	log.Println("✅ Admin routes registered")

	// Register CMS routes (at /api/v1/admin prefix)
	adminGroup := api.Group("/admin")
	adminGroup.Use(middleware.RateLimiter(100, time.Minute))
	cms_routes.SetupCategoryRoutes(adminGroup)
	cms_routes.SetupProductRoutes(adminGroup)
	cms_routes.SetupOrderRoutes(adminGroup)
	cms_routes.SetupCustomerRoutes(adminGroup)
	cms_routes.SetupAnalyticsRoutes(adminGroup)

	// Public storefront (no rate limiter)
	ecommerce_routes.SetupUserRoutes(api)
	ecommerce_routes.SetupAuthRoutes(api)
	ecommerce_routes.SetupStorefrontRoutes(api)

	// Swagger docs
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	fmt.Println("🚀 Server is running on http://localhost:8081")
	router.Run(":8081")
}
