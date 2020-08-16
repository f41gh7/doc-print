REPO=github.com/f41gh7/doc-print
GOCMD=GO111MODULE=on go
GOOS ?= linux
GOARCH ?= amd64
GOBUILD=CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH}  $(GOCMD) build -trimpath ${LDFLAGS}
GOCLEAN=$(GOCMD) clean
BINARY_NAME=doc-print


build:
	$(GOBUILD) $(REPO)