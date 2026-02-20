package cms_routes

import (
	admin_controller "github.com/Modeva-Ecommerce/modeva-cms-backend/controllers/cms/admin_controller"
	admin_auth "github.com/Modeva-Ecommerce/modeva-cms-backend/controllers/cms/admin_controller/auth"
	admin_auth_controller "github.com/Modeva-Ecommerce/modeva-cms-backend/controllers/cms/admin_controller/auth"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/controllers/cms/order_controller"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/middleware"
	"github.com/gin-gonic/gin"
)

// SetupAdminRoutes sets up all admin routes with appropriate middleware
func SetupAdminRoutes(rg *gin.RouterGroup) {
	// ════════════════════════════════════════════════════════════
	// Base Admin Group
	// ════════════════════════════════════════════════════════════

	admin := rg.Group("/admin")

	// ════════════════════════════════════════════════════════════
	// Public Routes (No Auth Required)
	// ════════════════════════════════════════════════════════════

	// Auth
	admin.POST("/login", admin_auth.AdminLogin)
	admin.POST("/accept-invite", admin_auth.AcceptAdminInvite)

	// ════════════════════════════════════════════════════════════
	// Protected Routes (Auth Required)
	// ════════════════════════════════════════════════════════════

	protected := admin.Group("")
	protected.Use(middleware.AdminAuthMiddleware())
	{
		// Auth
		protected.POST("/logout", admin_auth.AdminLogout)
		protected.GET("/me", admin_auth_controller.GetAdminMe)

		// Profile
		protected.PATCH("/profile", admin_controller.UpdateAdminProfile)

		// Admins
		protected.GET("/admins", admin_controller.GetAdmins)
		protected.GET("/admins/:id", admin_controller.GetAdmin)

		// Activity logs
		protected.GET("/admins/activity-logs", admin_controller.GetAllAdminActivityLogs)
		protected.GET("/admins/activity-logs/search", admin_controller.SearchAdminActivityLogs)
		protected.GET("/admins/:id/activity-logs", admin_controller.GetSingleAdminActivityLogs)

		// Stats
		protected.GET("/stats", admin_controller.GetAdminStats)

		// Orders
		protected.POST("/orders/:id/send-invoice", order_controller.SendOrderInvoicePDF)
		protected.GET("/orders/:id/download-invoice", order_controller.DownloadOrderInvoicePDF)

	}

	// ════════════════════════════════════════════════════════════
	// Super Admin Only Routes
	// ════════════════════════════════════════════════════════════

	superAdmin := admin.Group("")
	superAdmin.Use(
		middleware.AdminAuthMiddleware(),
		middleware.RequireSuperAdminMiddleware(),
	)
	{
		// Invitations
		superAdmin.POST("/invite", admin_auth.CreateAdminInvite)

		// Admin management
		superAdmin.POST("/admins/:id/suspend", admin_controller.SuspendAdmin)
		superAdmin.POST("/admins/:id/unsuspend", admin_controller.UnsuspendAdmin)
	}
}
