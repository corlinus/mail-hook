FROM golang:1.20-alpine3.18 AS builder
ARG GOOS
ARG GOARCH
WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN apk add --no-cache file
RUN GOOS=${GOOS} GOARCH=${GOARCH} go build -ldflags "-s -w" -o /smtp-hook

FROM alpine:3.18
WORKDIR /app
COPY --from=builder /smtp-hook /app
COPY envcfg.yml /app/config.yml
ENTRYPOINT [ "/app/smtp-hook", "-c", "config.yml" ]
