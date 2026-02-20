package product_controller

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DeleteProduct godoc
// @Summary Delete a product
// @Description Delete a product by ID and its associated Cloudinary folder
// @Tags CMS - Products
// @Produce json
// @Param id path string true "Product ID (UUID)"
// @Success 200 {object} models.ApiResponse
// @Failure 400 {object} models.ApiResponse
// @Failure 404 {object} models.ApiResponse
// @Router /api/v1/admin/products/{id} [delete]
func DeleteProduct(c *gin.Context) {
	// Step 1: Parse and validate product ID
	idParam := c.Param("id")
	productID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(c, "Invalid product ID"))
		return
	}

	ctx, cancel := config.WithTimeout()
	defer cancel()

	// Step 2: Find product and check if it has Cloudinary images
	var product models.Product
	if err := config.CmsGorm.WithContext(ctx).
		Select("id, media").
		First(&product, "id = ?", productID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse(c, "Product not found"))
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Database error"))
		}
		return
	}

	// Step 3: Check if product has Cloudinary images
	hasCloudinaryImages := false
	if product.Media.Primary.URL != "" || len(product.Media.Other) > 0 {
		hasCloudinaryImages = true
	}

	// Step 4: Delete from database
	if err := config.CmsGorm.WithContext(ctx).Delete(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to delete product: "+err.Error()))
		return
	}

	// Step 5: Delete Cloudinary folder in background (don't block response)
	if hasCloudinaryImages && cloudinaryService != nil {
		go func(prodID uuid.UUID) {
			// Create folder path: modeva/products/{productId}
			folderPath := fmt.Sprintf("modeva/products/%s", prodID.String())

			// Create context with timeout for deletion
			deleteCtx, deleteCancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer deleteCancel()

			// Delete the entire folder
			if err := cloudinaryService.DeleteFolder(deleteCtx, folderPath); err != nil {
				// Log error but don't fail the delete operation
				fmt.Printf("⚠️  Warning: Failed to delete Cloudinary folder %s: %v\n", folderPath, err)
			} else {
				fmt.Printf("✓ Successfully deleted Cloudinary folder: %s\n", folderPath)
			}
		}(productID)
	}

	c.JSON(http.StatusOK, models.SuccessResponse(c, "Product deleted successfully", map[string]string{
		"id": productID.String(),
	}))
}
