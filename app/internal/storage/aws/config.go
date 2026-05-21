package aws

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/MOliveiraDev/go-upload-files/internal/models"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type FileStorage interface {
	InitMultipartUpload(ctx context.Context, fileKey string) (string, error)
	UploadPart(ctx context.Context, fileKey, uploadID string, partNumber int32, chunk []byte) (string, error)
	CompleteMultipartUpload(ctx context.Context, fileKey, uploadID string, parts []Part) error
	AbortMultipartUpload(ctx context.Context, fileKey, uploadID string) error
	GetDownloadURL(ctx context.Context, file *models.File, expires time.Duration) (string, error)
}

// Struct genérica exigida pelo método "Complete" para juntar as partes
type Part struct {
	PartNumber int32
	ETag       string
}

// S3Storage implementa a interface FileStorage usando a AWS SDK
type S3Storage struct {
	client   *s3.Client
	bucket   string
	region   string
	endpoint string
}

// NewS3Storage incializa a conexão lendo nossas variáveis de ambiente (.env)
func NewS3Storage() (*S3Storage, error) {
	region := os.Getenv("AWS_REGION")
	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	bucket := os.Getenv("AWS_BUCKET_NAME")
	endpoint := os.Getenv("AWS_ENDPOINT_URL") // Extremamente útil para usar MinIO localmente em vez da AWS paga

	// Cria as configurações da nuvem
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("Falha de contato com AWS S3: %w", err)
	}

	// Se existir um Endpoint específico (ex: MinIO rodando em localhost:9000), redirecionamos o tráfego da AWS para lá.
	if endpoint != "" {
		cfg.BaseEndpoint = aws.String(endpoint)
	}

	// Inicializa de fato o Cliente S3
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	log.Println("S3 Storage Driver: Conectado e configurado com sucesso!")

	return &S3Storage{
		client:   client,
		bucket:   bucket,
		region:   region,
		endpoint: endpoint,
	}, nil
}

func (s *S3Storage) GetDownloadURL(ctx context.Context, file *models.File, expires time.Duration) (string, error) {
	if file == nil || file.Path == "" {
		return "", fmt.Errorf("arquivo inválido para geração de URL")
	}

	presignClient := s3.NewPresignClient(s.client)

	req, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(file.Path),
	}, s3.WithPresignExpires(expires))
	if err != nil {
		return "", fmt.Errorf("gerar URL assinada: %w", err)
	}

	return req.URL, nil
}
