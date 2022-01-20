REPO=github.com/f41gh7/doc-print
GOCMD=GO111MODULE=on go
GOBUILD=CGO_ENABLED=0 $(GOCMD) build -trimpath ${LDFLAGS}
GOCLEAN=$(GOCMD) clean
BINARY_NAME=doc-print


build:
	$(GOBUILD) $(REPO)
test:
	$(GOCMD) test .
