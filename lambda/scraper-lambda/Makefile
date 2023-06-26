BUILD_DIR := bin

all: clean build-lambda

clean:
	@rm -rf ${BUILD_DIR}

build-lambda:
	env GOOS=linux GOARCH=amd64 go build -o ../bin/mmu main.go
	cd ../bin && zip -r lambda-handler.zip .