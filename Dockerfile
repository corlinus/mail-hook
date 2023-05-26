FROM golang:1.20-alpine
WORKDIR /usr/src/app
COPY go.mod go.sum .
RUN go mod download
COPY . .
RUN go build .

FROM alpine:latest
WORKDIR /app
COPY --from=0 /usr/src/app/smtp-hook .
COPY --from=0 /usr/src/app/envcfg.yml config.yml
CMD /app/smtp-hook -c config.yml
