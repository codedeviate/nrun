all: local util go

local:
	go build -o nrun main.go
	GOOS=linux GOARCH=amd64 go build -o bin/linux/nrun main.go
	GOOS=linux GOARCH=arm64 go build -o bin/linux.arm/nrun main.go
	GOOS=windows GOARCH=amd64 go build -o bin/windows/nrun.exe main.go
	GOOS=windows GOARCH=arm64 go build -o bin/windows.arm/nrun.exe main.go
	GOOS=darwin GOARCH=amd64 go build -o bin/darwin/nrun main.go
	GOOS=darwin GOARCH=arm64 go build -o bin/darwin.arm/nrun main.go

util:
	go build -o ~/Utils/nrun main.go

go:
	go build -o ~/go/bin/nrun main.go