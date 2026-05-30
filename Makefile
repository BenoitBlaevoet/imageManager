GOFLAGS := -mod=mod
LDFLAGS := -ldflags="-s -w"

.PHONY: all windows mac linux run

all: windows mac linux

windows:
	GOOS=windows GOARCH=amd64 go build $(GOFLAGS) $(LDFLAGS) -o dist/imagemanager.exe .

mac:
	GOOS=darwin GOARCH=arm64 go build $(GOFLAGS) $(LDFLAGS) -o dist/imagemanager-mac .

mac-intel:
	GOOS=darwin GOARCH=amd64 go build $(GOFLAGS) $(LDFLAGS) -o dist/imagemanager-mac-intel .

linux:
	GOOS=linux GOARCH=amd64 go build $(GOFLAGS) $(LDFLAGS) -o dist/imagemanager-linux .

run:
	go run $(GOFLAGS) . --port 8080 --data-dir .
