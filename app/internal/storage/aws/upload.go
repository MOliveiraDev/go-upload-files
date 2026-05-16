package aws

import (
	"bytes"
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// InitMultipartUpload inicia o processo no S3 e retorna o ID da sessão de upload
func (s *S3Storage) InitMultipartUpload(ctx context.Context, fileKey string) (string, error) {
	input := &s3.CreateMultipartUploadInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(fileKey),
	}

	result, err := s.client.CreateMultipartUpload(ctx, input)
	if err != nil {
		return "", fmt.Errorf("erro ao iniciar multipart upload: %w", err)
	}

	return *result.UploadId, nil
}

// UploadPart envia um pedaço do arquivo para o S3 e retorna a ETag (recibo de confirmação)
func (s *S3Storage) UploadPart(ctx context.Context, fileKey, uploadID string, partNumber int32, chunk []byte) (string, error) {
	input := &s3.UploadPartInput{
		Bucket:     aws.String(s.bucket),
		Key:        aws.String(fileKey),
		UploadId:   aws.String(uploadID),
		PartNumber: aws.Int32(partNumber),
		Body:       bytes.NewReader(chunk),
	}

	result, err := s.client.UploadPart(ctx, input)
	if err != nil {
		return "", fmt.Errorf("erro ao fazer upload da parte %d: %w", partNumber, err)
	}

	return *result.ETag, nil
}

// CompleteMultipartUpload monta as partes no S3 e finaliza o upload
func (s *S3Storage) CompleteMultipartUpload(ctx context.Context, fileKey, uploadID string, parts []Part) error {
	// A AWS exige que as partes sejam enviadas num formato específico (CompletedPart)
	var completedParts []types.CompletedPart
	for _, p := range parts {
		completedParts = append(completedParts, types.CompletedPart{
			PartNumber: aws.Int32(p.PartNumber),
			ETag:       aws.String(p.ETag),
		})
	}

	input := &s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(s.bucket),
		Key:      aws.String(fileKey),
		UploadId: aws.String(uploadID),
		MultipartUpload: &types.CompletedMultipartUpload{
			Parts: completedParts,
		},
	}

	_, err := s.client.CompleteMultipartUpload(ctx, input)
	if err != nil {
		return fmt.Errorf("erro ao completar o multipart upload: %w", err)
	}

	return nil
}

// AbortMultipartUpload cancela a sessão e apaga todas as partes enviadas previamente
func (s *S3Storage) AbortMultipartUpload(ctx context.Context, fileKey, uploadID string) error {
	input := &s3.AbortMultipartUploadInput{
		Bucket:   aws.String(s.bucket),
		Key:      aws.String(fileKey),
		UploadId: aws.String(uploadID),
	}

	_, err := s.client.AbortMultipartUpload(ctx, input)
	if err != nil {
		return fmt.Errorf("erro ao abortar multipart upload: %w", err)
	}

	return nil
}
