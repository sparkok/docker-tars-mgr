
win:
	CGO_ENABLED=0 GOARCH=amd64 GOOS=windows go build -o build/bk-dock.exe -ldflags "-s -w" docker-tars-mgr/main

linux:
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -o build/bk-dock -ldflags "-s -w" docker-tars-mgr/main

mac:
	CGO_ENABLED=0 GOARCH=amd64 GOOS=darwin go build -o build/bk-dock-mac -ldflags "-s -w" docker-tars-mgr/main

.PHONY: everything win linux mac

all: win linux mac