BUILD_DIR := bin

all: clean build-lambda-on-mac-linux

clean:
	@rm -rf ${BUILD_DIR}

build-lambda-on-windows:
	set GOARCH=amd64
	set GOOS=linux
	set CGO_ENABLED=0
	go build -o ../bin/mmu main.go
	${HOMEPATH}\go\bin\build-lambda-zip.exe -o ../bin/lambda-handler.zip ../bin/mmu

build-lambda-on-mac-linux:
	env GOOS=linux GOARCH=amd64 go build -o ../bin/mmu main.go
	cd ../bin && zip -r lambda-handler.zip .