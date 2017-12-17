CWD     := $(shell pwd)
NAME    := ovfenv-installer
VERSION := 1.0.1
RELEASE := $(shell git rev-list HEAD --count)

LDFLAGS := -s -w \
           -X 'main.BuildVersion=$(VERSION)' \
           -X 'main.BuildGitRev=$(RELEASE)' \
           -X 'main.BuildGitCommit=$(shell git describe --abbrev=0 --always)' \
           -X 'main.BuildDate=$(shell date -u -R)'

PACKAGES := $(shell go list ./... | grep -v /vendor/)

default:
	@ echo "no default target for Makefile"

clean:
	rm -rf $(NAME) ./releases ./rpmbuild

glide-vc:
	glide-vc --only-code --no-tests --no-legal-files

fmt:
	go fmt $(PACKAGES)

lint: fmt
	go vet $(PACKAGES)

build: clean fmt
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o releases/$(NAME)-$(VERSION)

rpm: build
	mkdir -p rpmbuild
	cp -f releases/$(NAME)-$(VERSION) rpmbuild/$(NAME)

	rpmbuild -bb rpm.spec \
		--define="_topdir  $(CWD)/rpmbuild" \
		--define="_version $(VERSION)" \
		--define="_release $(RELEASE)"

	cp -f rpmbuild/RPMS/x86_64/*.rpm releases/

md5sum: build rpm
	for f in $(shell ls ./releases); do \
		cd $(CWD)/releases && md5sum "$$f" >> $$f.md5; \
	done

release: md5sum
