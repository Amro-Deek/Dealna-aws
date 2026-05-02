$ErrorActionPreference = "Continue"

Write-Host "Logging into ECR..."
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin 015615541352.dkr.ecr.us-east-1.amazonaws.com

Write-Host "Ensuring ECR repository exists..."
aws ecr create-repository --repository-name dealna-search-worker --region us-east-1

Write-Host "Building Docker image..."
docker build --provenance=false -t dealna-search-worker .

Write-Host "Tagging image..."
docker tag dealna-search-worker:latest 015615541352.dkr.ecr.us-east-1.amazonaws.com/dealna-search-worker:latest

Write-Host "Pushing image to ECR..."
docker push 015615541352.dkr.ecr.us-east-1.amazonaws.com/dealna-search-worker:latest

Write-Host "Deployment push complete."
