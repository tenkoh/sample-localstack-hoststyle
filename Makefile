.PHONY: release

DIST_DIR=dist
ZIP_NAME=lambda.zip
BINARY_NAME=bootstrap

build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -tags lambda.norpc --trimpath --ldflags '-w -s' -o $(BINARY_NAME) ./lambda/main.go

zip: build
	@mkdir -p $(DIST_DIR)
	zip $(DIST_DIR)/$(ZIP_NAME) $(BINARY_NAME)
	@rm -f $(BINARY_NAME)

clean:
	@rm -f $(BINARY_NAME) $(DIST_DIR)/$(ZIP_NAME)

release: clean zip
