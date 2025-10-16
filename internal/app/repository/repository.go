package repository

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Repository struct {
	db         *gorm.DB
	minio      *minio.Client
	bucketName string
}

func New(dsn string) (*Repository, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Получаем настройки Minio из переменных окружения
	minioEndpoint := getEnv("MINIO_ENDPOINT", "minio:9000")
	minioAccessKey := getEnv("MINIO_ACCESS_KEY", "minioadmin")
	minioSecretKey := getEnv("MINIO_SECRET_KEY", "minioadmin")
	bucketName := getEnv("MINIO_BUCKET", "blood-loss-images")
	useSSL := getEnv("MINIO_USE_SSL", "false") == "true"

	// Инициализация Minio клиента
	minioClient, err := minio.New(minioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(minioAccessKey, minioSecretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Minio client: %w", err)
	}

	// Проверяем доступность бакета
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	exists, err := minioClient.BucketExists(ctx, bucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		return nil, fmt.Errorf("bucket %s does not exist", bucketName)
	}

	return &Repository{
		db:         db,
		minio:      minioClient,
		bucketName: bucketName,
	}, nil
}

// UploadFile загружает файл в Minio
func (r *Repository) UploadFile(ctx context.Context, fileName string, file io.Reader, fileSize int64) (string, error) {
	contentType := "application/octet-stream"

	// Определяем Content-Type по расширению
	if strings.HasSuffix(strings.ToLower(fileName), ".jpg") || strings.HasSuffix(strings.ToLower(fileName), ".jpeg") {
		contentType = "image/jpeg"
	} else if strings.HasSuffix(strings.ToLower(fileName), ".png") {
		contentType = "image/png"
	} else if strings.HasSuffix(strings.ToLower(fileName), ".gif") {
		contentType = "image/gif"
	}

	_, err := r.minio.PutObject(ctx, r.bucketName, fileName, file, fileSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	// Возвращаем URL для доступа к файлу
	return fmt.Sprintf("http://localhost:9000/%s/%s", r.bucketName, fileName), nil
}

// DeleteFile удаляет файл из Minio
func (r *Repository) DeleteFile(ctx context.Context, fileName string) error {
	err := r.minio.RemoveObject(ctx, r.bucketName, fileName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// GeneratePresignedURL создает временную ссылку для доступа
func (r *Repository) GeneratePresignedURL(ctx context.Context, fileName string, expiry time.Duration) (string, error) {
	url, err := r.minio.PresignedGetObject(ctx, r.bucketName, fileName, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}
	return url.String(), nil
}

// Вспомогательная функция для получения переменных окружения
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
