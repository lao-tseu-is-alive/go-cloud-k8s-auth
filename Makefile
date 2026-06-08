#!make
SHELL := /bin/bash
# on Linux with rancher-desktop no need for docker
DOCKER_BIN := nerdctl
VER_SOURCE_CODE := pkg/version/version.go
APP_NAME := $(shell grep -E 'APP\s+=' $(VER_SOURCE_CODE)| awk '{ print $$3 }'  | tr -d '"')
APP_VERSION := $(shell grep -E 'VERSION\s+=' $(VER_SOURCE_CODE)| awk '{ print $$3 }'  | tr -d '"')
APP_REPOSITORY := $(shell grep -E 'REPOSITORY\s+=' $(VER_SOURCE_CODE)| awk '{ print $$3 }'  | tr -d '"')
$(info  Found APP_NAME:'$(APP_NAME)', APP_VERSION:'$(APP_VERSION)', APP_REPOSITORY:'$(APP_REPOSITORY)',  in file: $(VER_SOURCE_CODE) )
ifneq ("$(wildcard .env)","")
	ENV_EXISTS := "TRUE"
	include .env
	# next line allows to export env variables to external process (like your Go app)
	export $(shell sed 's/=.*//' .env)
else
    $(info env file was not found using default values for undefined variables )
    ENV_EXISTS := "FALSE"
	DB_DRIVER ?= postgres
	DB_HOST ?= 127.0.0.1
	DB_PORT ?= 5432
	DB_NAME ?= go_cloud_auth
	DB_USER ?= go_cloud_auth
	# DB_PASSWORD should be defined in your env or in github secrets
	DB_SSL_MODE ?= disable
endif
# uncomment line above to debug the value of env variable
#$(info $$ENV_EXISTS = $(ENV_EXISTS) )
APP_EXECUTABLE := $(APP_NAME)Server
APP_REVISION := $(shell git describe --dirty --always)
BUILD := $(shell date -u '+%Y-%m-%d_%I:%M:%S%p')
PACKAGES := $(shell go list ./... | grep -v /vendor/)
LDFLAGS := -ldflags "-X ${APP_REPOSITORY}/pkg/version.REVISION=${APP_REVISION} -X ${APP_REPOSITORY}/pkg/version.BuildStamp=${BUILD}"
#$(info $$LDFLAGS = $(LDFLAGS) )
PID_FILE := "./$(APP).pid"
APP_DSN := $(DB_DRIVER)://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSL_MODE)
ifeq ($(ENV_EXISTS),"TRUE")
	# or download your release from here : https://github.com/golang-migrate/migrate/releases
	# for ubuntu & debian : wget https://github.com/golang-migrate/migrate/releases/download/v4.15.1/migrate.linux-amd64.deb
	MIGRATE := /usr/local/bin/migrate -database "$(APP_DSN)" -path=db/migrations/
else
	# using golang-migrate https://github.com/golang-migrate/migrate/tree/master/cmd/migrate
	MIGRATE := $(DOCKER_BIN) run -v $(shell pwd)/db/migrations:/migrations --network host migrate/migrate:v4.10.0 -path=/db/migrations/ -database "$(APP_DSN)"
endif

# Make is verbose in Linux. Make it silent.
MAKEFLAGS += --silent

# because it is the first target in this Makefile this is also the default rule
.PHONY: run
## run:	will run a dev version of your Go application [DEFAULT RULE]
run: check-env mod-download openapi-codegen
	go run $(LDFLAGS) cmd/$(APP_EXECUTABLE)/${APP_EXECUTABLE}.go

.PHONY: mod-download
mod-download:
	@echo "  >  Downloading go modules dependencies..."
	go mod download

## build:	will compile your server app binary and place it in the bin sub-folder
build: check-env clean mod-download test openapi-codegen
	@echo "  >  Building your app binary inside bin directory..."
	CGO_ENABLED=0 go build ${LDFLAGS} -a -o bin/$(APP_EXECUTABLE) cmd/$(APP_EXECUTABLE)/${APP_EXECUTABLE}.go

.PHONY: build-example-client
## build-example-client:	will build the example client binary
build-example-client: mod-download
	@echo "  >  Building example client binary inside bin directory..."
	CGO_ENABLED=0 go build -o bin/exampleClient cmd/exampleClient/main.go

.PHONY: run-example-client
## run-example-client:	will run the example client with arguments (e.g. ARGS="-cmd start-oauth")
run-example-client:
	go run cmd/exampleClient/main.go $(ARGS)

.PHONY: exec-bin
## exec-bin:	will execute app binary with .env variables in current directory
exec-bin: bin/$(APP_EXECUTABLE)
	@echo "  >  executing your app binary ..."
	eval $(egrep -v '^#' .env | xargs -0) bin/$(APP_EXECUTABLE)


## old test using
 # go test -c -coverpkg=$(go list https://github.com/lao-tseu-is-alive/go-cloud-k8s-auth/...|tr "\n" ",") -race -covermode=atomic -o ./___goCloudAuthServer_test_go.test ./...; \
 #	go tool test2json -t ./___goCloudAuthServer_test_go.test -test.v -test.paniconexit0 -test.run ^\QTestMainExec\E -test.coverprofile coverage.out
 # GOROOT=/usr/local/go GOPATH=/home/cgil/go /usr/local/go/bin/go test -c -coverpkg=../../../go-cloud-auth/... -covermode=atomic -o ./___goCloudAuthServer_test_go.test https://github.com/lao-tseu-is-alive/go-cloud-k8s-auth/cmd/goCloudAuthServer; \
    #	go tool test2json -t ./___goCloudAuthServer_test_go.test -test.v -test.paniconexit0 -test.run ^\QTestMainExec\E$$ -test.coverprofile ./coverage.out
 # go test -race -coverprofile coverage.out -coverpkg=$(go list https://github.com/lao-tseu-is-alive/go-cloud-k8s-auth/...|tr "\n" ",") ./...


# Check if .env_testing exists and include it if it does
ifneq ("$(wildcard .env_testing)","")
include .env_testing
env-test-export:
	@echo "Exporting environment variables from .env_testing..."
	sed -ne '/^export / {p;d}; /.*=/ s/^/export / p' .env_testing > .env-testing-export

.PHONY: test
test: clean mod-download env-test-export
	@echo "  >  Running all tests code..."
	. .env-testing-export && go test -race -coverprofile coverage.out -coverpkg=./... ./...

.PHONY: test-all
test-all: clean mod-download env-test-export
	@echo "  >  Running all tests code..."
	@echo "mode: count" > coverage-all.out
	@$(foreach pkg,$(PACKAGES), \
		. .env-testing-export && go test -race -p=1 -cover -covermode=atomic -coverprofile=coverage.out ${pkg}; \
		tail -n +2 coverage.out >> coverage-all.out;)


else
env-test-export:
	@echo ".env_testing file does not exist, skipping export..."

.PHONY: test
test: clean mod-download env-test-export
	@echo "  >  Running all tests code..."
	go test -race -coverprofile coverage.out -coverpkg=./... ./...

.PHONY: test-all
test-all: clean mod-download env-test-export
	@echo "  >  Running all tests code..."
	@echo "mode: count" > coverage-all.out
	@$(foreach pkg,$(PACKAGES), \
		go test -race -p=1 -cover -covermode=atomic -coverprofile=coverage.out ${pkg}; \
		tail -n +2 coverage.out >> coverage-all.out;)



endif


.PHONY: clean
## clean:	will delete you server app binary and remove temporary files like coverage output
clean:
	@echo "  >  Removing binaries and coverage from build directory..."
	rm -rf bin/$(APP_EXECUTABLE) bin/exampleClient coverage.out coverage-all.out ___goCloudAuthServer_test_go.test

.PHONY: release
## release:	will build & tag a clean repo with a version release and push the tag to the remote git
release: build
	@echo "  >  Preparing release $(APP_EXECUTABLE) v$(APP_VERSION) rev: $(APP_REVISION) ..."
ifeq ($(shell git status -s),)  # check if repo is clean
	echo "OK : your repo is clean"
	@git fetch  ||  (echo "ERROR : git fetch failed" && exit 1)
	@git tag -l  "v${APP_VERSION}"  ||  (echo "ERROR : this git tag v${APP_VERSION} already exist" && exit 1)
	git tag "v${APP_VERSION}" -m "v${APP_VERSION} bump"
else
	(echo "ERROR : your local git repo is dirty : it contains modified and/or untracked files :" && ( git status -s) && exit 1)
endif
	#git push origin $(APP_VERSION)

# check some dependencies
.PHONY: dependencies-openapi
dependencies-openapi:
	@command -v oapi-codegen >/dev/null 2>&1 || { printf >&2 "oapi-codegen is not installed, please run: go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest\n"; exit 1; }

.PHONY: dependencies-xo
dependencies-xo:
	@command -v oapi-codegen >/dev/null 2>&1 || { printf >&2 "xo is not installed, please run: go install github.com/xo/xo@latest\n"; exit 1; }


.PHONY: check-env
check-env:
	@echo "  >  checking if env APP_NAME exist..."
ifndef APP_NAME
	# if this variable is not defined via ./scripts/getAppInfo.sh
	$(error APP_NAME is undefined)
endif
	@echo "  >  checking if env DB_PASSWORD exist..."
ifndef DB_PASSWORD
	# if this variable is not defined you cannot initialise the postgres db correctly
	$(error DB_PASSWORD is undefined)
endif


# for reason to use .Phony see : https://www.gnu.org/software/make/manual/html_node/Phony-Targets.html
.PHONY: openapi-codegen
## openapi-codegen:	will generate helper Go code for types & server based on OpenApi spec in api/app.yml
openapi-codegen: dependencies-openapi
	oapi-codegen --old-config-style -generate types -o pkg/go_cloud_auth/go_cloud_auth_types.gen.go -package go_cloud_auth api/openapi/go_cloud_auth.swagger.yaml


.PHONY: build-container-image
## build-container-image:	will build a local container multi-stage image from your app
build-container-image:
	echo "  >  Building your container image using scripts/01_build_image_locally.sh"
	$(shell ./scripts/01_build_image_locally.sh)


.PHONY: xo-codegen
## xo-codegen:	will generate helper Go code for database queries in models directory
xo-codegen: dependencies-xo
	xo schema schema -s public --go-pkg=go_cloud_auth --src templates_xo -o models ${DB_DRIVER}://${DB_USER}:=@${DB_HOST}/${DB_NAME}
	#xo schema schema -s public --go-pkg=go_cloud_auth --src templates_xo -o pkg/go_cloud_auth ${DB_DRIVER}://${DB_USER}:=@${DB_HOST}/${DB_NAME}

.PHONY: help
help: Makefile
	@echo
	@echo " Choose a make target from one of  :"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo
