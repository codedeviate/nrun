all: local util go

local:
	go build -o nrun main.go

util:
	go build -o ~/Utils/nrun main.go

go:
	go build -o ~/go/bin/nrun main.go