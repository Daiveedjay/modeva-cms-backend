package services

import (
	"context"
	"fmt"
	"mime/multipart"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api"
	"github.com/cloudinary/cloudinary-go/v2/api/admin"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type CloudinaryService struct {
	cld *cloudinary.Cloudinary
}

func NewCloudinaryService(cloudName, apiKey, apiSecret string) (*CloudinaryService, error) {
	cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		return nil, err
	}
	return &CloudinaryService{cld: cld}, nil
}

// UploadImage uploads a single image to Cloudinary and returns the secure URL
func (s *CloudinaryService) UploadImage(ctx context.Context, file multipart.File, filename string, folder string) (string, error) {
	// Use pointer booleans as required by the cloudinary SDK
	unique := true
	overwrite := false
	uploadParams := uploader.UploadParams{
		Folder:       folder,
		ResourceType: "image",
		// Removed Transformation - apply on delivery instead for faster uploads!
		UniqueFilename: &unique,
		Overwrite:      &overwrite,
	}

	// Only set PublicID if filename is provided
	if filename != "" {
		uploadParams.PublicID = filename
	}

	result, err := s.cld.Upload.Upload(ctx, file, uploadParams)
	if err != nil {
		return "", fmt.Errorf("failed to upload image: %w", err)
	}

	// Ensure we return the secure URL
	if result.SecureURL == "" {
		return "", fmt.Errorf("upload successful but no URL returned")
	}

	return result.SecureURL, nil
}

// UploadMultipleImages uploads multiple images and returns their URLs
func (s *CloudinaryService) UploadMultipleImages(ctx context.Context, files []*multipart.FileHeader, folder string) ([]string, error) {
	var urls []string

	for i, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open file %s: %w", fileHeader.Filename, err)
		}
		defer file.Close()

		filename := fmt.Sprintf("%s_%d", fileHeader.Filename, i)
		url, err := s.UploadImage(ctx, file, filename, folder)
		if err != nil {
			return nil, err
		}

		urls = append(urls, url)
	}

	return urls, nil
}

// DeleteImage deletes an image from Cloudinary using its public ID
func (s *CloudinaryService) DeleteImage(ctx context.Context, publicID string) error {
	_, err := s.cld.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID: publicID,
	})
	return err
}

// DeleteFolder deletes an entire folder and all its contents from Cloudinary
// This uses the Admin API to delete all assets with a given prefix
func (s *CloudinaryService) DeleteFolder(ctx context.Context, folderPath string) error {
	// Method 1: Delete all assets by prefix (this deletes the actual files)
	// The Prefix parameter expects an api.CldAPIArray
	result, err := s.cld.Admin.DeleteAssetsByPrefix(ctx, admin.DeleteAssetsByPrefixParams{
		Prefix: api.CldAPIArray{folderPath},
	})
	if err != nil {
		return fmt.Errorf("failed to delete assets in folder %s: %w", folderPath, err)
	}

	// Log how many assets were deleted
	if result != nil {
		fmt.Printf("Deleted assets from folder %s\n", folderPath)
	}

	// Method 2: Try to delete the folder structure itself
	// Note: Cloudinary usually auto-removes empty folders, so this might not be needed
	// We'll try anyway and ignore errors

	// Delete subfolders first, then parent
	foldersToDelete := []string{
		folderPath + "/primary",
		folderPath + "/other",
		folderPath,
	}

	for _, folder := range foldersToDelete {
		// Try to delete each folder, ignore errors
		s.cld.Admin.DeleteFolder(ctx, admin.DeleteFolderParams{
			Folder: folder,
		})
	}

	return nil
}
