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