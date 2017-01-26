VERSION?="0.1.0"
DEST?=./bin

default: install

test:
	echo "==> Running tests..."
	go test -cover -v `go list ./... | grep -v /vendor/`

build:
	echo "==> Build binaries..."
	go build -v -ldflags "-s -w -X main.version=${VERSION}" -o ${DEST}/broforce broforce.go

install: test build
	echo "==> Copy binaries to \$GOPATH/bin/..."
	cp ${DEST}/* ${GOPATH}/bin/