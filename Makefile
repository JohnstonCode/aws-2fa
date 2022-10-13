.PHONY: osx
osx:
	GOOS=darwin GOARCH=amd64 go build -o bin/osx/aws-2fa main.go

.PHONY: linux
linux:
	GOOS=linux GOARCH=amd64 go build -o bin/linux/aws-2fa main.go
