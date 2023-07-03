ifeq ($(OS),Windows_NT)
    detected_os := windows
		ifeq ($(PROCESSOR_ARCHITEW6432),AMD64)
        detected_arch := amd64
    else
        ifeq ($(PROCESSOR_ARCHITECTURE),AMD64)
            detected_arch := amd64
        endif
        ifeq ($(PROCESSOR_ARCHITECTURE),x86)
            detected_arch := ia32
        endif
    endif
else
    UNAME_S := $(shell uname -s)
    ifeq ($(UNAME_S),Linux)
        detected_os := linux
    endif
    ifeq ($(UNAME_S),Darwin)
        detected_os := macOS
    endif
    UNAME_P := $(shell uname -p)
    ifeq ($(UNAME_P),x86_64)
        detected_arch := amd64
    endif
    ifneq ($(filter %86,$(UNAME_P)),)
        detected_arch := ia32
    endif
    ifneq ($(filter arm%,$(UNAME_P)),)
        detected_arch := arm
    endif
endif

ifeq ($(detected_os),windows)
		RIMRAF := rmdir /S /Q
else
		RIMRAF := rm -rf
endif

ifeq ($(detected_os),windows)
		PATH_DELIM := \\
else
		PATH_DELIM := /
endif

.PHONY: clean build lint test

clean:
	$(RIMRAF) .$(PATH_DELIM)bin

build:
  # go commands use unix-style paths, even on windows 
	go build -o ./bin/ ./...

lint:
  # docker commands use unix-style paths, even on windows
	docker run -t --rm -v $${PWD}:/app -w /app golangci/golangci-lint:v1.53.3 golangci-lint run

test:
  # go commands use unix-style paths, even on windows
	go test -v --race -run=Test_Unit ./...

