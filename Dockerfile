FROM golang:1.20-alpine3.18
ARG GOOS
ARG GOARCH
WORKDIR /usr/src/app
COPY go.mod go.sum .
RUN go mod download
COPY . .
RUN go build .

FROM alpine:3.18
WORKDIR /app
COPY --from=0 /usr/src/app/smtp-hook .
COPY --from=0 /usr/src/app/envcfg.yml config.yml
CMD /app/smtp-hook -c config.yml
