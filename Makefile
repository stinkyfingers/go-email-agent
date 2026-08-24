BINARY_NAME=gmail_token
BUILD_DIR=bin
MAIN_PACKAGE=./cmd/oauth_token

.PHONY: build
build:
	@echo "Building gmail_token"
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)

.PHONY: build-linux
build-linux:
	@echo "Building gmail_token"
	mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux $(MAIN_PACKAGE)

.PHONY: build-windows
build-windows:
	@echo "Building gmail_token"
	mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows $(MAIN_PACKAGE)

AWS_REGION ?= us-west-1
AWS_ACCOUNT_ID ?= $(shell aws sts get-caller-identity --query Account --output text)
ECR_REPO ?= go-email-agent
IMAGE_TAG ?= main
ECR_REGISTRY := $(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_REGION).amazonaws.com
ECR_IMAGE := $(ECR_REGISTRY)/$(ECR_REPO):$(IMAGE_TAG)

.PHONY: docker-build
docker-build:
	docker build -t $(ECR_IMAGE) .

.PHONY: docker-build-linux
docker-build-linux:
	docker build --platform=linux/amd64 -t $(ECR_IMAGE) .

.PHONY: ecr-login
ecr-login:
	aws ecr get-login-password --region $(AWS_REGION) | docker login --username AWS --password-stdin $(ECR_REGISTRY)

.PHONY: ecr-push
ecr-push: docker-build-linux
	@echo "Pushing $(ECR_IMAGE)"
	$(MAKE) ecr-login
	docker push $(ECR_IMAGE)