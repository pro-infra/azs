all: azs.darwin_arm64 azs.darwin_amd64 azs.linux_amd64

clean:
	find . -type f -a \( -name azs.darwin_arm64 -o -name azs.darwin_amd64 -o -name azs.linux_amd64 -o -name azs \) -delete	

azs.darwin_arm64: $(wildcard *.go)
	GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.version=`git describe --tags HEAD`" -o azs.darwin_arm64

azs.darwin_amd64: $(wildcard *.go)
	GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=`git describe --tags HEAD`" -o azs.darwin_amd64

azs.linux_amd64: $(wildcard *.go)
	GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=`git describe --tags HEAD`" -o azs.linux_amd64

.PHONY: all clean
