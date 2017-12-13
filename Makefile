CWD     := $(shell pwd)
NAME    := ovfenv-installer
VERSION := 1.0.1

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
	[ -n "$(shell type -P fpm)" ]      || gem install fpm

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

rpm: build
	@ mkdir -p rpmbuild/usr/local/bin/
	@ cp -f releases/$(NAME)-$(VERSION) rpmbuild/usr/local/bin/$(NAME)
	@ fpm -s dir -t rpm \
		--rpm-os linux \
		--name $(NAME) --version $(VERSION) --iteration $(shell git rev-list HEAD --count) \
		--maintainer "subchen@gmail.com" --vendor "Guoqiang Chen" \
		--license "Apache 2" \
		--url "https://github.com/subchen/$(NAME)" \
		--description "Configure networking from vSphere ovfEnv properties" \
		-C rpmbuild/ \
		--package ./releases/

md5sum: build rpm
	@ for f in $(shell ls ./releases); do \
		cd $(CWD)/releases && md5sum "$$f" >> $$f.md5; \
	done

release: md5sum
