# Builder base image
FROM registry.erda.cloud/retag/golang:1.16-alpine3.14 as builder

WORKDIR /workspace
# Args
ARG GOPROXY
ARG ARCH

# Envoriment
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=${ARCH:-amd64} \
    GOPROXY=${GOPROXY:-https://goproxy.cn}

# Copy modules mainfest
COPY go.mod go.mod
COPY go.sum go.sum

# Cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the soruce
COPY cmd/ cmd/
COPY api/ api/
COPY pkg/ pkg/
COPY pkg/controllers/ controllers/

# Build the binary
RUN go build -a -o manager cmd/manager/main.go

ARG BASE_IMAGE
FROM ${BASE_IMAGE:-registry.erda.cloud/retag/distroless-static:nonroot}

WORKDIR /

COPY --from=builder /workspace/manager .
USER 65532:65532

ENTRYPOINT ["./manager"]
