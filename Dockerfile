FROM golang:1.20-alpine3.18
ARG GOOS
ARG GOARCH
WORKDIR /usr/src/app
COPY go.mod go.sum .
RUN go mod download
COPY . .
RUN apk add file
RUN go build -ldflags "-s -w" .

FROM alpine:3.18
WORKDIR /app
COPY --from=0 /usr/src/app/smtp-hook .
COPY --from=0 /usr/src/app/envcfg.yml config.yml

RUN echo "file :"
RUN file /app/smtp-hook
ENTRYPOINT [ "/app/smtp-hook", "-c", "config.yml" ]
