FROM golang:1.12
WORKDIR /alice
COPY . /alice/
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /alice/
COPY --from=0 /alice/alice .
CMD ["./alice"]