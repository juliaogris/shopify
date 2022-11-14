FROM golang:1.19-alpine AS builder

WORKDIR /src
RUN apk add git make

ARG VERSION
ENV VERSION=${VERSION}

COPY go.mod go.sum Makefile main.go .
RUN make install

FROM alpine:3.16
WORKDIR /app
COPY --from=builder /go/bin/shopify .
ENTRYPOINT ["/app/shopify"]
