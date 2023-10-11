# Copyright (c) 2021 Terminus, Inc.
#
# This program is free software: you can use, redistribute, and/or modify
# it under the terms of the GNU Affero General Public License, version 3
# or later ("AGPL"), as published by the Free Software Foundation.
#
# This program is distributed in the hope that it will be useful, but WITHOUT
# ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
# FITNESS FOR A PARTICULAR PURPOSE.
#
# You should have received a copy of the GNU Affero General Public License
# along with this program. If not, see <http://www.gnu.org/licenses/>.

GO_PROJECT_ROOT := github.com/erda-project/erda
ARCH ?= amd64

REGISTRY ?= registry.erda.cloud/erda
VERSION ?= $(shell cat ./VERSION)
BUILD_TIME := $(shell date -u '+%Y-%m-%d %H:%M:%S')
GIT_SHORT_COMMIT := $(shell git rev-parse --short HEAD)
GIT_COMMIT := $(shell git rev-parse HEAD)
DEFAULT_IMAGE_TAG := ${VERSION}-$(shell date -u +"%Y%m%d%H%M%S")-${GIT_SHORT_COMMIT}
IMAGE_TAG ?= $(DEFAULT_IMAGE_TAG)
IMAGE := ${REGISTRY}/dice-operator:$(IMAGE_TAG)
IMAGE_VERSION_LATEST := ${REGISTRY}/dice-operator:${VERSION}

ifeq ($(GO_PROXY_ENV),)
	GO_PROXY := "https://goproxy.cn,direct"
else
	GO_PROXY := $(GO_PROXY_ENV)
endif

build-version:
	@echo Arch: ${ARCH}
	@echo Version: ${VERSION}
	@echo Build Time: ${BUILD_TIME}
	@echo Git Commit: ${GIT_COMMIT}
	@echo Image: ${IMAGE}

default: build

build: build-version
	@echo "build dice-operator"
	@CGO_ENABLED=0 GOARCH=${ARCH} go build -o bin/dice-operator-${ARCH} ./cmd/dice-operator

docker-build: build-version
	@docker build -t $(IMAGE)	\
	  --build-arg ARCH=$(ARCH) \
	  --build-arg GO_PROJECT_ROOT=$(GO_PROJECT_ROOT) \
	  --build-arg GO_PROXY=$(GO_PROXY) .

docker-build-push: docker-build
	@docker push $(IMAGE)
	@echo action meta: image=$(IMAGE)

docker-version-latest: docker-build-push
	@echo "retag image $(IMAGE) to $(IMAGE_VERSION_LATEST)"
	@docker tag $(IMAGE) $(IMAGE_VERSION_LATEST)
	@docker push $(IMAGE_VERSION_LATEST)
	@echo action meta: image_version_latest=$(IMAGE_VERSION_LATEST)
