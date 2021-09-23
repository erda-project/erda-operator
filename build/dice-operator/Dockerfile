FROM registry.erda.cloud/erda/golang:1.16.6-stretch AS builder

ARG GO_PROJECT_ROOT
ARG GO_PROXY

WORKDIR /go/src/${GO_PROJECT_ROOT}

ENV GO111MODULE=on
ENV GOPATH=/go

RUN go env -w GOPROXY=${GO_PROXY}

COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/dice-operator cmd/dice-operator/main.go

FROM registry.erda.cloud/erda/alpine:3.14

ARG GO_PROJECT_ROOT

COPY --from=builder /go/src/${GO_PROJECT_ROOT}/bin/dice-operator /app/dice-operator

RUN chmod +x /app/dice-operator
ENV TZ=Asia/Shanghai

ENTRYPOINT ["/app/dice-operator"]
