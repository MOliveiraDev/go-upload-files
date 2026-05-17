variable "aws_region" {
  description = "A região da AWS onde os recursos serão provisionados"
  type        = string
  default     = "us-east-2"
}

variable "bucket_name" {
  description = "S3 Bucket para armazenar os arquivos"
  type        = string
  default     = "drive-storage-br-2026" 
}
