CWD     := $(shell pwd)
NAME    := ovfenv-installer
VERSION := 1.0.0

GOPATH  := /tmp/go
GOPWD   := $(GOPATH)/src/github.com/subchen
GOCWD   := $(GOPWD)/$(NAME)

PATH    := $(GOPATH)/bin:$(PATH)
export PATH

LDFLAGS := -s -w \
           -X 'main.BuildVersion=$(VERSION)' \
           -X 'main.BuildGitRev=$(shell git rev-list HEAD --count)' \
           -X 'main.BuildGitCommit=$(shell git describe --abbrev=0 --always)' \
           -X 'main.BuildDate=$(shell date -u -R)'

PACKAGES := $(shell go list ./... | grep -v /vendor/)

default:
	@ echo "no default target for Makefile"

init:
	mkdir -p $(GOPWD)
	ln -sf $(CWD) $(GOCWD)

clean:
	@ rm -rf $(NAME) ./releases

pre-install:
	[ -n "$(shell type -P glide)" ]    || go get -u github.com/Masterminds/glide/...
	[ -n "$(shell type -P glide-vc)" ] || go get -u github.com/sgotti/glide-vc/...

glide-update:
	glide update

glide-vc:
	glide-vc --only-code --no-tests --no-legal-files

fmt:
	@ go fmt $(PACKAGE)

lint: fmt
	@ go vet $(PACKAGE)

build: clean init fmt
	cd $(GOCWD) && GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o releases/$(NAME)-$(VERSION)

md5sum: build
	cd $(CWD)/releases; md5sum "$(NAME)-$(VERSION)" > "$(NAME)-$(VERSION).md5"

release: md5sum
