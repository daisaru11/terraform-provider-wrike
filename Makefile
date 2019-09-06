
PROVIDER_VERSION ?= "v0.0.0-beta.1"
PROVIDER_NAME ?= "terraform-provider-wrike"

export GO111MODULE=on

build:
	OS="`go env GOOS`" ARCH="`go env GOARCH`" go build -o $(PROVIDER_NAME)

test:
	go test -v .

lint:
	golangci-lint run

build-release:
	mkdir -p pkg
	go build -o pkg/$(PROVIDER_NAME)_$(PROVIDER_VERSION)_$(GOOS)-$(GOARCH)

build-release-cross: build-release-linux build-release-darwin build-release-windows 

build-release-linux: 
	GOOS=linux GOARCH=386   make build-release
	GOOS=linux GOARCH=amd64 make build-release

build-release-darwin:
	GOOS=darwin GOARCH=386   make build-release
	GOOS=darwin GOARCH=amd64 make build-release

build-release-windows:
	GOOS=windows GOARCH=386   make build-release
	GOOS=windows GOARCH=amd64 make build-release

clean:
	rm -f $(PROVIDER_NAME)
	rm -Rf pkg/*