ARG ARCH

FROM --platform=${ARCH} registry.erda.cloud/retag/golang:1.19-bullseye  as builder

ARG GO_PROJECT_ROOT
ARG GO_PROXY

WORKDIR /go/src/${GO_PROJECT_ROOT}

ENV GO111MODULE=on
ENV GOPATH=/go

RUN go env -w GOPROXY=${GO_PROXY}

COPY go.mod go.sum ./

RUN go mod download

COPY pkg pkg
COPY cmd cmd

RUN CGO_ENABLED=0 GOOS=linux GOARCH=${ARCH} \
    go build -o bin/dice-operator cmd/dice-operator/main.go

FROM --platform=${ARCH} registry.erda.cloud/retag/debian:bullseye-slim

ARG GO_PROJECT_ROOT

COPY --from=builder /go/src/${GO_PROJECT_ROOT}/bin/dice-operator /app/dice-operator

RUN chmod +x /app/dice-operator
ENV TZ=Asia/Shanghai

ENTRYPOINT ["/app/dice-operator"]
