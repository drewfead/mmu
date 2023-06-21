BUILD_DIR := build

clean:
	@rm -rf ${BUILD_DIR}

build-lambda-on-windows:
	set GOARCH=amd64
	set GOOS=linux
	set CGO_ENABLED=0
	go build -o ../bin/mmu main.go
	${HOMEPATH}\go\bin\build-lambda-zip.exe -o ../bin/lambda-handler.zip ../bin/mmu