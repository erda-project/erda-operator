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

ifeq ($(GO_PROXY_ENV),)
	GO_PROXY := "https://proxy.golang.org,direct"
else
	GO_PROXY := $(GO_PROXY_ENV)
endif

ifeq ($(REGISTRY_HOST),)
    REGISTRY := registry.erda.cloud/erda
else
    REGISTRY := $(REGISTRY_HOST)
endif

BUILD_DIR := ./build
TARGETS_DIR := dice-operator
IMAGE_PREFIX ?= $(strip )
IMAGE_SUFFIX ?= $(strip )

IMAGE_TAG ?= "$(shell cat VERSION)-$(shell date -u +%Y%m%d)-$(shell git rev-parse --short HEAD --dirty)"
DOCKER_LABELS ?= git-describe="$(IMAGE_TAG)"

GO_OPTIONS ?= -mod=vendor -count=1
SHELLOPTS := errexit

container:
	@for target in $(TARGETS_DIR); do                                                  \
	  image=$(IMAGE_PREFIX)$${target}$(IMAGE_SUFFIX);                                  \
	  docker build -t $(REGISTRY)/$${image}:$(IMAGE_TAG)                               \
	    --build-arg GO_PROJECT_ROOT=$(GO_PROJECT_ROOT)                                 \
	    --build-arg GO_PROXY=$(GO_PROXY)                                 \
	    --label $(DOCKER_LABELS)                                                       \
	    -f $(BUILD_DIR)/$${target}/Dockerfile .;                                       \
	  echo "image=$(REGISTRY)/$${image}:$(IMAGE_TAG)" >> $${METAFILE};		   \
	done

push: container
	@for target in $(TARGETS_DIR); do                                                  \
	  image=$(IMAGE_PREFIX)$${target}$(IMAGE_SUFFIX);                                  \
	  docker push $(REGISTRY)/$${image}:$(IMAGE_TAG);                                  \
	done
