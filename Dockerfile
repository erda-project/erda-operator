FROM registry.erda.cloud/erda-x/golang:1.24 AS builder

ARG GO_PROJECT_ROOT
ARG GO_PROXY

WORKDIR /go/src/${GO_PROJECT_ROOT}

ENV GO111MODULE=on \
    GOPATH=/go \
    CGO_ENABLED=0

RUN go env -w GOPROXY=${GO_PROXY}

COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download

COPY pkg pkg
COPY cmd cmd

RUN go build -o bin/dice-operator cmd/dice-operator/main.go

FROM registry.erda.cloud/erda-x/debian-bookworm:12

ARG GO_PROJECT_ROOT

COPY --from=builder /go/src/${GO_PROJECT_ROOT}/bin/dice-operator /app/dice-operator

RUN chmod +x /app/dice-operator
ENV TZ=Asia/Shanghai

ENTRYPOINT ["/app/dice-operator"]
