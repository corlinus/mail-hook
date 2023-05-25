FROM golang:1.20
WORKDIR /usr/src/app
COPY go.mod go.sum .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build .

FROM alpine:latest
WORKDIR /app
COPY --from=0 /usr/src/app/smtp-hook /usr/src/app/example.yml .
CMD /app/smtp-hook -c example.yml
