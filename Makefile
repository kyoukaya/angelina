PKGNAME := angelina
VERSION := 0.1-alpha
TARGET := cmd/main.go
TARGETOS := windows
TARGETARCH := amd64
GOFLAGS := -v
LDFLAGS := -s -w
BIN := main.exe
TEMP_DIR := ${PKGNAME}
ADDITIONAL_RELEASE_FILES := data readme.md
ARCHIVE_NAME := ${PKGNAME}-${VERSION}-${TARGETOS}-${TARGETARCH}.zip

default: release

clean:
	rm -rf $(ARCHIVE_NAME) $(TEMP_DIR) $(BIN)

release:
	rm -rf $(TEMP_DIR) $(ARCHIVE_NAME) && mkdir $(TEMP_DIR)
	GOOS=$(TARGETOS) GOARCH=$(TARGETARCH) go build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o $(TEMP_DIR)/$(BIN) cmd/main.go
	cd $(TEMP_DIR)
	cp -r $(ADDITIONAL_RELEASE_FILES) $(TEMP_DIR)
	zip -9 -r $(ARCHIVE_NAME) $(TEMP_DIR)
	rm -rf $(TEMP_DIR) $(BIN)
