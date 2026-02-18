package service

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"
	"user-service/internal/config"
	"user-service/internal/models"
	"user-service/internal/repository"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type ImageService struct {
	minioClient *minio.Client
	repo        *repository.UserRepository
	cfg         *config.Config
}

func NewImageService(repo *repository.UserRepository, cfg *config.Config) (*ImageService, error) {
	// Инициализируем MinIO клиент
	minioClient, err := minio.New(cfg.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioAccessKey, cfg.MinioSecretKey, ""),
		Secure: cfg.MinioUseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize MinIO client: %v", err)
	}

	return &ImageService{
		minioClient: minioClient,
		repo:        repo,
		cfg:         cfg,
	}, nil
}

// UploadImage загружает изображение в MinIO и сохраняет метаданные в БД
// imageType: avatar, venue-logo, venue-cover, venue-photo, event-cover
func (s *ImageService) UploadImage(file *multipart.FileHeader, imageType string) (*models.Image, error) {
	// Проверяем тип файла
	if !isValidImageType(file.Filename) {
		return nil, fmt.Errorf("invalid file type, allowed: jpg, jpeg, png, gif, webp")
	}

	// Проверяем размер (максимум 10MB)
	if file.Size > 10*1024*1024 {
		return nil, fmt.Errorf("file too large, maximum size is 10MB")
	}

	// Определяем бакет по типу изображения
	bucketName := getBucketByImageType(imageType)
	if bucketName == "" {
		return nil, fmt.Errorf("invalid image type: %s", imageType)
	}

	// Открываем файл
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer src.Close()

	// Генерируем уникальное имя файла
	ext := filepath.Ext(file.Filename)
	uniqueFileName := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	objectName := uniqueFileName // Без префикса, просто UUID + расширение

	// Определяем content-type
	contentType := getContentType(file.Filename)

	// Загружаем в MinIO
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err = s.minioClient.PutObject(ctx, bucketName, objectName, src, file.Size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload to MinIO: %v", err)
	}

	// Сохраняем метаданные в БД
	image := &models.Image{
		FileName:   file.Filename,
		FilePath:   objectName,
		FileType:   contentType,
		ImageType:  imageType,
		BucketName: bucketName,
	}

	if err := s.repo.CreateImage(image); err != nil {
		// Если не удалось сохранить в БД, пытаемся удалить файл из MinIO
		s.minioClient.RemoveObject(context.Background(), bucketName, objectName, minio.RemoveObjectOptions{})
		return nil, fmt.Errorf("failed to save image metadata: %v", err)
	}

	return image, nil
}

// DeleteImage удаляет изображение из MinIO и БД
func (s *ImageService) DeleteImage(imageID int) error {
	// Получаем информацию об изображении
	image, err := s.repo.GetImageByID(imageID)
	if err != nil {
		return fmt.Errorf("image not found: %v", err)
	}

	// Удаляем из MinIO (используем bucket_name из БД)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = s.minioClient.RemoveObject(ctx, image.BucketName, image.FilePath, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete from MinIO: %v", err)
	}

	// Удаляем из БД
	if err := s.repo.DeleteImage(imageID); err != nil {
		return fmt.Errorf("failed to delete image metadata: %v", err)
	}

	return nil
}

// GetImage возвращает изображение из MinIO в виде байтов
func (s *ImageService) GetImage(imageID int) (*models.Image, []byte, error) {
	image, err := s.repo.GetImageByID(imageID)
	if err != nil {
		return nil, nil, fmt.Errorf("image not found in DB: %v", err)
	}

	ctx := context.Background()

	// Получаем объект из MinIO (используем bucket_name из БД)
	object, err := s.minioClient.GetObject(ctx, image.BucketName, image.FilePath, minio.GetObjectOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get object from MinIO: %v", err)
	}
	defer object.Close()

	// Читаем весь объект в память
	data, err := io.ReadAll(object)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read image data: %v", err)
	}

	if len(data) == 0 {
		return nil, nil, fmt.Errorf("image data is empty (0 bytes)")
	}

	return image, data, nil
}

// Вспомогательные функции

func getBucketByImageType(imageType string) string {
	bucketMap := map[string]string{
		"avatar":       "creator-avatars",
		"venue-logo":   "venue-logos",
		"venue-cover":  "venue-cover-photos",
		"venue-photo":  "venue-photos",
		"event-cover":  "event-covers",
	}
	return bucketMap[imageType]
}

func isValidImageType(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	validExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
	}
	return validExtensions[ext]
}

func getContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	contentTypes := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
	}
	if ct, ok := contentTypes[ext]; ok {
		return ct
	}
	return "application/octet-stream"
}
