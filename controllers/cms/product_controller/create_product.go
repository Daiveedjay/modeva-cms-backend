package product_controller

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var cloudinaryService *services.CloudinaryService

func InitCloudinary(cloudName, apiKey, apiSecret string) error {
	var err error
	cloudinaryService, err = services.NewCloudinaryService(cloudName, apiKey, apiSecret)
	return err
}

// CreateProduct godoc
// @Summary Create a new product
// @Description Create a new product with Cloudinary URLs (optimized flow)
// @Tags CMS - Products
// @Accept json
// @Produce json
// @Param product body models.ProductRequest true "Product details with Cloudinary URLs"
// @Success 201 {object} models.ApiResponse
// @Failure 400 {object} models.ApiResponse
// @Failure 500 {object} models.ApiResponse
// @Router /api/v1/admin/products [post]
func CreateProduct(c *gin.Context) {
	overallStart := time.Now()
	log.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Printf("[PERF] CREATE PRODUCT START (GORM + UUID v7)")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	// Step 1: Parse JSON request
	var req models.ProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[ERROR] Invalid request: %v", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Invalid request: "+err.Error()))
		return
	}

	// Step 2: Set default status if not provided
	if req.Status == "" {
		req.Status = "Draft"
	}

	// Step 3: Validate subcategory exists
	validationStart := time.Now()
	var subCategory models.Category
	if err := config.CmsGorm.WithContext(ctx).
		Select("id, name, parent_id, parent_name").
		First(&subCategory, "id = ?", req.SubCategoryID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("[ERROR] Invalid sub_category_id: %s", req.SubCategoryID)
			c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Invalid sub_category_id"))
		} else {
			log.Printf("[ERROR] Database error: %v", err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Database error"))
		}
		return
	}
	log.Printf("[PERF] ⏱️  Category validation: %v", time.Since(validationStart))

	// Step 4: Validate media URLs exist
	if req.Media.Primary.URL == "" {
		log.Printf("[ERROR] Primary image URL is missing")
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Primary image URL is required"))
		return
	}

	log.Printf("[PERF] 📸 Primary URL: %s", req.Media.Primary.URL)
	if len(req.Media.Other) > 0 {
		log.Printf("[PERF] 📸 Other images: %d URLs", len(req.Media.Other))
	}

	// Step 5: Create product model (UUID v7 auto-generated in BeforeCreate hook)
	product := models.Product{
		Name:          req.Name,
		Description:   req.Description,
		Composition:   models.CompositionList(req.Composition),
		Price:         req.Price,
		SubCategoryID: req.SubCategoryID,
		Status:        req.Status,
		Tags:          models.TagsList(req.Tags),
		Media:         req.Media,
		Variants:      models.VariantsList(req.Variants),
		Inventory:     models.InventoryList(req.Inventory),
		SEO:           req.SEO,
		Views:         0,
	}

	// Step 6: Save to database
	dbStart := time.Now()
	if err := config.CmsGorm.WithContext(ctx).Create(&product).Error; err != nil {
		log.Printf("[ERROR] Failed to create product: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to create product: "+err.Error()))
		return
	}
	dbDuration := time.Since(dbStart)
	log.Printf("[PERF] ⏱️  Database insert: %v", dbDuration)
	log.Printf("[PERF] 🆔 Product ID (UUID v7): %s", product.ID)

	// Step 7: Load subcategory relationship for response
	if err := config.CmsGorm.WithContext(ctx).
		Preload("SubCategory", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, name, parent_id, parent_name")
		}).
		First(&product, "id = ?", product.ID).Error; err != nil {
		log.Printf("[ERROR] Failed to reload product: %v", err)
		// Product is created, just missing relationship - still return success
	}

	totalDuration := time.Since(overallStart)
	log.Printf("[PERF] ⏱️  ⭐ TOTAL TIME: %v (Database only, images already in Cloudinary)", totalDuration)
	log.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	c.JSON(http.StatusCreated, models.SuccessResponse(c, "Product created successfully", product))
}

// ════════════════════════════════════════════════════════════
// CLEANUP ENDPOINT
// ════════════════════════════════════════════════════════════

// CleanupFolderRequest represents the request to delete a folder
type CleanupFolderRequest struct {
	FolderPath string `json:"folder_path" binding:"required"`
}

// CleanupOrphanedFolder godoc
// @Summary Delete orphaned product folder from Cloudinary
// @Description Deletes entire product folder when backend save fails after upload succeeds
// @Tags CMS - Products
// @Accept json
// @Produce json
// @Param request body CleanupFolderRequest true "Folder path to delete"
// @Success 200 {object} models.ApiResponse
// @Failure 400 {object} models.ApiResponse
// @Failure 403 {object} models.ApiResponse
// @Router /api/v1/admin/products/cleanup-folder [post]
func CleanupOrphanedFolder(c *gin.Context) {
	var req CleanupFolderRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Invalid request: "+err.Error()))
		return
	}

	if req.FolderPath == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Folder path is required"))
		return
	}

	// Security: Only allow cleanup of product folders
	if !strings.HasPrefix(req.FolderPath, "modeva/products/") {
		log.Printf("[Cleanup] ⚠️  Blocked attempt to delete non-product folder: %s", req.FolderPath)
		c.JSON(http.StatusForbidden, models.ErrorResponse(c, "Can only cleanup product folders"))
		return
	}

	// Validate folder path format (should be modeva/products/{uuid})
	parts := strings.Split(req.FolderPath, "/")
	if len(parts) != 3 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Invalid folder path format"))
		return
	}

	log.Printf("[Cleanup] Folder deletion requested: %s", req.FolderPath)

	// Delete folder in background (don't block response)
	go func(folderPath string) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		err := cloudinaryService.DeleteFolder(ctx, folderPath)
		if err != nil {
			log.Printf("[Cleanup] ❌ Failed to delete folder %s: %v", folderPath, err)
		} else {
			log.Printf("[Cleanup] ✓ Successfully deleted orphaned folder: %s", folderPath)
		}
	}(req.FolderPath)

	c.JSON(http.StatusOK, models.SuccessResponse(c, "Folder cleanup initiated", map[string]string{
		"folder": req.FolderPath,
		"status": "deleting",
	}))
}
