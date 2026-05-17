# Criação do Bucket S3
resource "aws_s3_bucket" "drive_storage" {
  bucket = var.bucket_name

  tags = {
    Name        = "DriveStorage"
    Environment = "Dev"
  }
}

resource "aws_s3_bucket_versioning" "drive_versioning" {
  bucket = aws_s3_bucket.drive_storage.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_lifecycle_configuration" "drive_lifecycle" {
  bucket = aws_s3_bucket.drive_storage.id

  rule {
    id     = "abort-incomplete-multipart-uploads"
    status = "Enabled"

    # Um filtro vazio indica para a AWS que essa regra se aplica a todos os arquivos do bucket
    filter {}

    # Se um multipart upload ficar travado por mais de 3 dias, a AWS apaga automaticamente e você não paga por lixo
    abort_incomplete_multipart_upload {
      days_after_initiation = 3
    }
  }
}

resource "aws_s3_bucket_cors_configuration" "drive_cors" {
  bucket = aws_s3_bucket.drive_storage.id

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["GET", "PUT", "POST", "DELETE"]
    allowed_origins = ["*"] # URL para o Front-end em produção
    expose_headers  = ["ETag"]
    max_age_seconds = 3000
  }
}
