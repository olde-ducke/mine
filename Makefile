NAME          := mine
OUTPUT        := ./bin
VERSION       := main.version=$(shell git describe --tags --abbrev=0 HEAD 2>/dev/null)
DEBUG_VERSION := $(VERSION)-debug
BUILD_INFO    := -X "main.buildDate=$(shell date --iso-8601=s)"
BUILD_INFO    += -X "main.commit=$(shell git rev-parse --short=8 HEAD 2>/dev/null)"
BUILD_INFO    += -X "main.builtBy=$(shell echo `whoami`)"

GOBUILD       := CGO_ENABLED=0 go build $(BUILD_FLAGS)

pre:
	go mod tidy

build: build_darwin build_linux build_windows
build_windows: EXT := .exe 
build_darwin build_linux build_windows: build_%: clean pre
	GOOS=$* GOARCH=amd64 $(GOBUILD) --ldflags '-s -w $(BUILD_INFO) -X "$(VERSION)"' --trimpath -v -o $(OUTPUT)/$(NAME)-$*-amd64$(EXT) ./main.go

debug: clean pre
ifeq ($(OS),Windows_NT)
	$(eval EXT := .exe)
endif
	$(GOBUILD) --ldflags '$(BUILD_INFO) -X "$(DEBUG_VERSION)"' --gcflags "all=-N -l" -v -o $(OUTPUT)/$(NAME)$(EXT) ./main.go

clean:
	rm -f $(OUTPUT)/$(NAME)-windows-amd64.exe \
      $(OUTPUT)/$(NAME)-darwin-amd64 \
      $(OUTPUT)/$(NAME)-linux-amd64 \
      $(OUTPUT)/$(NAME).exe \
      $(OUTPUT)/$(NAME)

