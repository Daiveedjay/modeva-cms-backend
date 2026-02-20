package product_controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UpdateProduct godoc
// @Summary Update an existing product
// @Description Update product details by ID with support for both text and image updates
// @Tags CMS - Products
// @Accept json
// @Produce json
// @Param id path string true "Product ID (UUID)"
// @Param product body models.UpdateProductRequest true "Product update fields"
// @Success 200 {object} models.ApiResponse
// @Failure 400 {object} models.ApiResponse
// @Failure 404 {object} models.ApiResponse
// @Router /api/v1/admin/products/{id} [patch]
func UpdateProduct(c *gin.Context) {
	// Check if this is multipart (images) or JSON (text only)
	contentType := c.GetHeader("Content-Type")
	isMultipart := strings.Contains(contentType, "multipart/form-data")

	idParam := c.Param("id")
	productID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Invalid product ID"))
		return
	}

	if isMultipart {
		updateProductWithImages(c, productID)
	} else {
		updateProductTextOnly(c, productID)
	}
}

// updateProductTextOnly handles JSON updates without image changes
func updateProductTextOnly(c *gin.Context, productID uuid.UUID) {
	var input models.UpdateProductRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Invalid request: "+err.Error()))
		return
	}

	// Step 1: Find existing product
	var product models.Product
	if err := config.CmsGorm.
		First(&product, "id = ?", productID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse(c, "Product not found"))
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Database error"))
		}
		return
	}

	// Step 2: Validate subcategory if provided
	if input.SubCategoryID != nil {
		var subCategory models.Category
		if err := config.CmsGorm.
			First(&subCategory, "id = ?", *input.SubCategoryID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Invalid sub_category_id"))
			} else {
				c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Database error"))
			}
			return
		}
	}

	// Step 3: Build update map (only non-nil fields)
	updates := make(map[string]interface{})

	if input.Name != nil {
		updates["name"] = *input.Name
	}
	if input.Description != nil {
		updates["description"] = *input.Description
	}
	if input.Composition != nil {
		updates["composition"] = models.CompositionList(*input.Composition)
	}
	if input.Price != nil {
		updates["price"] = *input.Price
	}
	if input.SubCategoryID != nil {
		updates["sub_category_id"] = *input.SubCategoryID
	}
	if input.Status != nil {
		updates["status"] = *input.Status
	}
	if input.Tags != nil {
		updates["tags"] = models.TagsList(*input.Tags)
	}
	// Only update media if it's explicitly provided AND has valid data
	if input.Media != nil && input.Media.Primary.URL != "" {
		updates["media"] = *input.Media
	}
	if input.Variants != nil {
		updates["variants"] = models.VariantsList(*input.Variants)
	}
	if input.Inventory != nil {
		updates["inventory"] = models.InventoryList(*input.Inventory)
	}
	if input.SEO != nil {
		updates["seo"] = *input.SEO
	}

	// Step 4: Update product
	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "No fields to update"))
		return
	}

	if err := config.CmsGorm.
		Model(&product).
		Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to update product"))
		return
	}

	// Step 5: Reload with subcategory
	if err := config.CmsGorm.
		Preload("SubCategory", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, name, parent_id, parent_name")
		}).
		First(&product, "id = ?", productID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to reload product"))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(c, "Product updated successfully", product))
}

// updateProductWithImages handles multipart form updates with image changes
func updateProductWithImages(c *gin.Context, productID uuid.UUID) {
	// Parse multipart form
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Failed to parse form data"))
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	// Step 1: Fetch existing product
	var product models.Product
	if err := config.CmsGorm.WithContext(ctx).
		First(&product, "id = ?", productID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse(c, "Product not found"))
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Database error"))
		}
		return
	}

	existingMedia := product.Media

	// Get form fields
	updates := make(map[string]interface{})

	if name := c.PostForm("name"); name != "" {
		updates["name"] = name
	}
	if description := c.PostForm("description"); description != "" {
		updates["description"] = description
	}
	if priceStr := c.PostForm("price"); priceStr != "" {
		var price float64
		if _, err := fmt.Sscanf(priceStr, "%f", &price); err == nil {
			updates["price"] = price
		}
	}
	if subCategoryIDStr := c.PostForm("sub_category_id"); subCategoryIDStr != "" {
		if subCatID, err := uuid.Parse(subCategoryIDStr); err == nil {
			// Validate subcategory exists
			var subCategory models.Category
			if err := config.CmsGorm.WithContext(ctx).
				First(&subCategory, "id = ?", subCatID).Error; err == nil {
				updates["sub_category_id"] = subCatID
			}
		}
	}
	if status := c.PostForm("status"); status != "" {
		updates["status"] = status
	}

	// Parse JSON fields from form
	if compositionStr := c.PostForm("composition"); compositionStr != "" {
		var composition []models.Composition
		if err := json.Unmarshal([]byte(compositionStr), &composition); err == nil {
			updates["composition"] = models.CompositionList(composition)
		}
	}
	if tagsStr := c.PostForm("tags"); tagsStr != "" {
		var tags []string
		if err := json.Unmarshal([]byte(tagsStr), &tags); err == nil {
			updates["tags"] = models.TagsList(tags)
		}
	}
	if variantsStr := c.PostForm("variants"); variantsStr != "" {
		var variants []models.ProductVariant
		if err := json.Unmarshal([]byte(variantsStr), &variants); err == nil {
			updates["variants"] = models.VariantsList(variants)
		}
	}
	if inventoryStr := c.PostForm("inventory"); inventoryStr != "" {
		var inventory []models.InventoryField
		if err := json.Unmarshal([]byte(inventoryStr), &inventory); err == nil {
			updates["inventory"] = models.InventoryList(inventory)
		}
	}
	if seoStr := c.PostForm("seo"); seoStr != "" {
		var seo models.Seo
		if err := json.Unmarshal([]byte(seoStr), &seo); err == nil {
			updates["seo"] = seo
		}
	}

	// Product folder for Cloudinary
	productFolder := fmt.Sprintf("modeva/products/%s", productID.String())

	// Handle media updates
	var newMedia models.ProductMedia

	// === PRIMARY IMAGE HANDLING ===
	primaryImageFile, _, err := c.Request.FormFile("primaryImage")
	if err == nil {
		defer primaryImageFile.Close()

		// Delete old primary image if exists
		if existingMedia.Primary.URL != "" {
			publicID := extractPublicIDFromURL(existingMedia.Primary.URL)
			if publicID != "" {
				// Delete in background
				go func(pid string) {
					deleteCtx, deleteCancel := context.WithTimeout(context.Background(), 10*time.Second)
					defer deleteCancel()
					_ = cloudinaryService.DeleteImage(deleteCtx, pid)
				}(publicID)
			}
		}

		// Upload new primary image
		primaryURL, err := cloudinaryService.UploadImage(ctx, primaryImageFile, "primary", productFolder+"/primary")
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to upload primary image: "+err.Error()))
			return
		}
		newMedia.Primary = models.MediaURL{URL: primaryURL}
	} else {
		// No new primary image - check if keeping existing
		primaryImageURL := c.PostForm("primaryImageUrl")
		if primaryImageURL != "" {
			// Keep existing primary image
			newMedia.Primary = models.MediaURL{URL: primaryImageURL}
		} else {
			// Fallback to existing from database
			newMedia.Primary = existingMedia.Primary
		}
	}

	// === OTHER IMAGES HANDLING ===
	// First, get list of existing images to keep
	existingOtherImagesStr := c.PostForm("existingOtherImages")
	var existingImagesToKeep []models.MediaURL
	if existingOtherImagesStr != "" {
		_ = json.Unmarshal([]byte(existingOtherImagesStr), &existingImagesToKeep)
	}

	// Find images to delete (images in DB but not in "keep" list)
	imagesToDelete := findImagesToDelete(existingMedia.Other, existingImagesToKeep)
	for _, img := range imagesToDelete {
		publicID := extractPublicIDFromURL(img.URL)
		if publicID != "" {
			// Delete in background
			go func(pid string) {
				deleteCtx, deleteCancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer deleteCancel()
				_ = cloudinaryService.DeleteImage(deleteCtx, pid)
			}(publicID)
		}
	}

	// Start with kept images
	newMedia.Other = existingImagesToKeep

	// Upload new "other" images
	form, _ := c.MultipartForm()
	imageFiles := form.File["otherImages"]
	if len(imageFiles) > 0 {
		newImageURLs, err := cloudinaryService.UploadMultipleImages(ctx, imageFiles, productFolder+"/other")
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to upload other images: "+err.Error()))
			return
		}

		// Add new images to the list
		startOrder := len(newMedia.Other)
		for i, url := range newImageURLs {
			order := startOrder + i
			newMedia.Other = append(newMedia.Other, models.MediaURL{
				URL:   url,
				Order: &order,
			})
		}
	}

	// Update media in updates map
	updates["media"] = newMedia

	// Step 2: Update database
	if len(updates) > 0 {
		if err := config.CmsGorm.WithContext(ctx).
			Model(&product).
			Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to update product: "+err.Error()))
			return
		}
	}

	// Step 3: Reload with subcategory
	if err := config.CmsGorm.WithContext(ctx).
		Preload("SubCategory", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, name, parent_id, parent_name")
		}).
		First(&product, "id = ?", productID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to reload product"))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(c, "Product updated successfully", product))
}

// ═══════════════════════════════════════════════════════════
// Helper Functions
// ═══════════════════════════════════════════════════════════

// extractPublicIDFromURL extracts the Cloudinary public ID from a full URL
// Example: https://res.cloudinary.com/demo/image/upload/v1234/modeva/products/test/primary.jpg
// Returns: modeva/products/test/primary
func extractPublicIDFromURL(url string) string {
	if url == "" {
		return ""
	}

	// Find the position after "/upload/"
	uploadIndex := strings.Index(url, "/upload/")
	if uploadIndex == -1 {
		return ""
	}

	// Get everything after "/upload/"
	afterUpload := url[uploadIndex+8:] // +8 to skip "/upload/"

	// Skip version if present (e.g., "v1234567890/")
	if strings.HasPrefix(afterUpload, "v") {
		versionEndIndex := strings.Index(afterUpload, "/")
		if versionEndIndex != -1 {
			afterUpload = afterUpload[versionEndIndex+1:]
		}
	}

	// Remove file extension
	lastDotIndex := strings.LastIndex(afterUpload, ".")
	if lastDotIndex != -1 {
		afterUpload = afterUpload[:lastDotIndex]
	}

	return afterUpload
}

// findImagesToDelete finds images that exist in the database but are not in the keep list
func findImagesToDelete(existingImages, keepImages []models.MediaURL) []models.MediaURL {
	var toDelete []models.MediaURL

	// Create a map of URLs to keep for fast lookup
	keepMap := make(map[string]bool)
	for _, img := range keepImages {
		keepMap[img.URL] = true
	}

	// Find images to delete
	for _, existing := range existingImages {
		if !keepMap[existing.URL] {
			toDelete = append(toDelete, existing)
		}
	}

	return toDelete
}
