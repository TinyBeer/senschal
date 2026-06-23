APP_NAME = seneschal
BUILD_DIR = build

.PHONY: gen build build-all clean

gen:
	@go generate ./...

build: gen
	@go build -o $(BUILD_DIR)/$(APP_NAME) .

build-linux: gen
	@GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 .

build-darwin: gen
	@GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64 .

build-windows: gen
	@GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe .

build-all: build-linux build-darwin build-windows

clean:
	@rm -rf $(BUILD_DIR)