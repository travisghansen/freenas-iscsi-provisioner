BIN=bin

all: build

fmt:
	go fmt ./...

tmp:
	mkdir -p tmp/

vendor:
	glide install -v --strip-vcs

$(BIN)/freenas-provisioner build: vendor $(BIN) $(shell find . -name "*.go")
	env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags '-extldflags "-static"' -o $(BIN)/freenas-iscsi-provisioner .

darwin: vendor $(BIN) $(shell find . -name "*.go")
	env CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -a -ldflags '-extldflags "-static"' -o $(BIN)/freenas-iscsi-provisioner-darwin .

freebsd: vendor $(BIN) $(shell find . -name "*.go")
	env CGO_ENABLED=0 GOOS=freebsd GOARCH=amd64 go build -a -ldflags '-extldflags "-static"' -o $(BIN)/freenas-iscsi-provisioner-freebsd .

clean:
	go clean -i
	rm -rf $(BIN)
	rm -rf tmp/
	rm -rf vendor

$(BIN):
	mkdir -p $(BIN)

.PHONY: all fmt clean
